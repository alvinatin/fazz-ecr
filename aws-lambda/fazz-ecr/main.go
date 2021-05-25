package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/payfazz/go-errors/v2"

	"github.com/payfazz/fazz-ecr/aws-lambda/fazz-ecr/svc/createrepo"
	"github.com/payfazz/fazz-ecr/aws-lambda/fazz-ecr/svc/dockerlogin"
	oidcconfig "github.com/payfazz/fazz-ecr/config/oidc"
)

func main() {
	lambda.StartHandler(h{})
}

type h struct{}

func (h) Invoke(ctx context.Context, payload []byte) ([]byte, error) {
	var ret []byte
	if err := errors.Catch(func() error {
		var err error
		ret, err = func() ([]byte, error) {
			var input events.APIGatewayV2HTTPRequest
			if err := json.Unmarshal(payload, &input); err != nil {
				return nil, errors.Trace(err)
			}

			claims := input.RequestContext.Authorizer.JWT.Claims
			if claims["iss"] != oidcconfig.Issuer || claims["aud"] != oidcconfig.ClientID {
				return resp(401, "invalid iss or aud in jwt token"), nil
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
				return resp(400, "invalid token"), nil
			}
			if jwtBody.Email == "" {
				return resp(400, "cannot accept jwt token without email"), nil
			}

			switch input.RouteKey {
			case "GET /docker-login":
				cred, err := dockerlogin.GetCredFor(jwtBody.Email, jwtBody.Groups)
				if err != nil {
					return nil, err
				}

				return respCred(cred), nil

			case "POST /create-repo":
				var repo string
				inputBody := []byte(input.Body)
				if input.IsBase64Encoded {
					inputBody, err = base64.StdEncoding.DecodeString(string(inputBody))
					if err != nil {
						return nil, errors.Trace(err)
					}
				}
				if err := json.Unmarshal(inputBody, &repo); err != nil {
					return resp(400, "invalid json body"), nil
				}

				if err := createrepo.CreateRepoFor(jwtBody.Email, jwtBody.Groups, repo); err != nil {
					if createrepo.IsAccessDenied(err) {
						return resp(403, "Access Denied"), nil
					}
					return nil, err
				}

				return resp(200, "OK"), nil

			default:
				return resp(404, "invalid path"), nil
			}
		}()
		return err
	}); err != nil {
		return respErr(err), nil
	}

	return ret, nil
}
