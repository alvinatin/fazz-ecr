package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/payfazz/go-errors/v2"

	oidcconfig "github.com/payfazz/fazz-ecr/config/oidc"
)

func main() {
	lambda.StartHandler(h{})
}

type h struct{}

func (h) Invoke(ctx context.Context, payload []byte) ([]byte, error) {
	var input events.APIGatewayV2HTTPRequest
	if err := json.Unmarshal(payload, &input); err != nil {
		return resp500(errors.Wrap(err))
	}

	if input.RequestContext.Authorizer == nil ||
		input.RequestContext.Authorizer.JWT == nil ||
		input.RequestContext.Authorizer.JWT.Claims == nil ||
		input.RequestContext.Authorizer.JWT.Claims["iss"] != oidcconfig.Issuer ||
		input.RequestContext.Authorizer.JWT.Claims["aud"] != oidcconfig.ClientID {
		return resp401()
	}

	// it is safe to trust content of jwt token, because api gateway already verify it for us

	authHeader := strings.SplitN(
		strings.TrimPrefix(input.Headers["authorization"], "Bearer "), ".", 3,
	)
	jwtBodyRaw, err := base64.RawURLEncoding.DecodeString(authHeader[1])
	if err != nil {
		return resp500(errors.Wrap(err))
	}

	var jwtBody struct {
		Email  string   `json:"email"`
		Groups []string `json:"groups"`
	}
	if err := json.Unmarshal(jwtBodyRaw, &jwtBody); err != nil {
		return resp500(errors.Wrap(err))
	}
	if jwtBody.Email == "" {
		return resp400("cannot accept jwt token without email")
	}

	if input.RouteKey == "GET /docker-login" {
		cred, err := getCredFor(jwtBody.Email, jwtBody.Groups)
		if err != nil {
			return resp500(errors.Wrap(err))
		}

		return respOk(cred)
	}

	return resp404()
}
