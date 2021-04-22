package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/payfazz/go-errors"

	"github.com/payfazz/fazz-ecr/pkg/types"
)

func resp(status int, msg string) ([]byte, error) {
	body, _ := json.Marshal(struct {
		Message string `json:"message"`
	}{
		Message: msg,
	})
	return json.Marshal(events.APIGatewayV2HTTPResponse{
		StatusCode: status,
		Headers: map[string]string{
			"content-type": "application/json",
		},
		Body: string(body),
	})
}

func resp400(msg string) ([]byte, error) {
	return resp(400, msg)
}

func resp401() ([]byte, error) {
	return resp(401, "401 Unauthorized")
}

func resp404() ([]byte, error) {
	return resp(401, "404 Not Found")
}

func resp500(err error) ([]byte, error) {
	fmt.Fprintln(os.Stderr, errors.Format(errors.Wrap(err)))
	return resp(500, err.Error())
}

func respOk(cred types.Cred) ([]byte, error) {
	body, _ := json.Marshal(cred)
	return json.Marshal(events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"content-type": "application/json",
		},
		Body: string(body),
	})
}
