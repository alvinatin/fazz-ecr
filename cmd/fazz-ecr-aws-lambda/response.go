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

func resp(status int, msg string) []byte {
	body, _ := json.Marshal(struct{ Message string }{Message: msg})

	ret, _ := json.Marshal(events.APIGatewayV2HTTPResponse{
		StatusCode: status,
		Headers:    map[string]string{"Content-type": "application/json"},
		Body:       string(body),
	})

	return ret
}

func respCred(cred types.Cred) []byte {
	body, _ := json.Marshal(cred)

	ret, _ := json.Marshal(events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Headers:    map[string]string{"Content-type": "application/json"},
		Body:       string(body),
	})

	return ret
}

func respErr(err error) []byte {
	logBody, _ := json.Marshal(struct {
		Time   time.Time
		Error  string
		Detail string
	}{
		Time:   time.Now(),
		Error:  err.Error(),
		Detail: errors.FormatWithFilterPkgs(err, "main", "github.com/payfazz/fazz-ecr"),
	})

	fmt.Fprintln(os.Stderr, string(logBody))

	return resp(500, "500 Internal Server error")
}
