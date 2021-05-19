package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/payfazz/fazz-ecr/pkg/types"
	"github.com/payfazz/go-errors/v2"
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

func resp400(msg string) ([]byte, error) {
	return resp(400, msg)
}

func resp401() ([]byte, error) {
	return resp(401, "401 Unauthorized")
}

func resp404() ([]byte, error) {
	return resp(404, "404 Not Found")
}

func resp500(err error) ([]byte, error) {
	logBody, _ := json.Marshal(struct {
		Time   time.Time
		Error  string
		Detail string
	}{
		Time:   time.Now(),
		Error:  err.Error(),
		Detail: errors.FormatWithFilterPkgs(err, "github.com/payfazz/fazz-ecr"),
	})
	fmt.Fprintln(os.Stderr, string(logBody))
	return resp(500, "500 Internal Server error")
}
