package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/payfazz/go-errors/v2"

	"github.com/payfazz/fazz-ecr/cmd/fazz-ecr-aws-lambda/exchangesvc"
	oidcconfig "github.com/payfazz/fazz-ecr/config/oidc"
)

func main() {
	lambda.StartHandler(h{})
}

type h struct{}

func (h) Invoke(ctx context.Context, payload []byte) ([]byte, error) {
	f := func() ([]byte, error) {
		var input events.APIGatewayV2HTTPRequest
		if err := json.Unmarshal(payload, &input); err != nil {
			return nil, errors.Trace(err)
		}

		claims := input.RequestContext.Authorizer.JWT.Claims
		if claims["iss"] != oidcconfig.Issuer || claims["aud"] != oidcconfig.ClientID {
			return resp401(), nil
		}

		// it is safe to trust content of jwt token, because api gateway already verify it for us

		authParts := strings.Split(strings.TrimPrefix(input.Headers["authorization"], "Bearer "), ".")
		jwtBodyRaw, err := base64.RawURLEncoding.DecodeString(authParts[1])
		if err != nil {
			return nil, errors.Trace(err)
		}

		var jwtBody struct {
			Email  string   `json:"email"`
			Groups []string `json:"groups"`
		}
		if err := json.Unmarshal(jwtBodyRaw, &jwtBody); err != nil {
			return nil, errors.Trace(err)
		}
		if jwtBody.Email == "" {
			return resp400("cannot accept jwt token without email"), nil
		}

		if input.RouteKey == "GET /docker-login" {
			cred, err := exchangesvc.GetCredFor(jwtBody.Email, jwtBody.Groups)
			if err != nil {
				return nil, errors.Trace(err)
			}

			return respCred(cred), nil
		}

		return resp404(), nil
	}

	var ret []byte
	var err error
	err = errors.Catch(func() error { ret, err = f(); return err })
	if err != nil {
		return resp500(err), nil
	}
	return ret, nil
}
