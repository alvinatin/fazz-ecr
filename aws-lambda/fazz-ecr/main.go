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
)

func main() {
	lambda.StartHandler(h{})
}

type h struct{}

func (h h) Invoke(ctx context.Context, payload []byte) ([]byte, error) {
	doInvoke := func() ([]byte, error) {
		var input events.APIGatewayV2HTTPRequest
		if err := json.Unmarshal(payload, &input); err != nil {
			return nil, errors.Trace(err)
		}

		email, groups, err := getAuth(strings.TrimPrefix(input.Headers["authorization"], "Bearer "))
		if err != nil {
			return nil, err
		}

		if email == "" {
			return resp(401, "unauthorized"), nil
		}

		switch input.RouteKey {
		case "GET /docker-login":
			cred, err := dockerlogin.GetCredFor(email, groups)
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

			if err := createrepo.CreateRepoFor(email, groups, repo); err != nil {
				if createrepo.IsAccessDenied(err) {
					return resp(403, "Access Denied"), nil
				}
				return nil, err
			}

			return resp(200, "OK"), nil

		default:
			return resp(404, "invalid path"), nil
		}
	}

	var ret []byte
	var err error
	err = errors.Catch(func() error { ret, err = doInvoke(); return err })
	if err != nil {
		return respErr(err), nil
	}

	return ret, nil
}
