package main

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/payfazz/go-errors/v2"

	"github.com/payfazz/fazz-ecr/aws-lambda/fazz-ecr/util/iam"
	awsconfig "github.com/payfazz/fazz-ecr/config/aws"
	"github.com/payfazz/fazz-ecr/util/logerr"
	"github.com/payfazz/fazz-ecr/util/randstring"
)

func main() {
	if err := errors.Catch(run); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		logerr.Log(err)
		os.Exit(1)
	}
}

func run() error {
	envSession, err := iam.EnvSession()
	if err != nil {
		return errors.Trace(err)
	}

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s namespace1 [namespace2...]\n", os.Args[0])
		os.Exit(1)
	}

	statickey := randstring.Get(16)
	email := statickey + "@statickey"
	groups := os.Args[1:]

	token := base64.RawStdEncoding.EncodeToString([]byte(`{"typ":"statickey"}`)) +
		"." +
		base64.RawStdEncoding.EncodeToString([]byte(statickey)) +
		"."

	ddbsvc := dynamodb.New(envSession)
	_, err = ddbsvc.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(awsconfig.StaticKeyTableName()),
		Item: map[string]*dynamodb.AttributeValue{
			"id":       {S: aws.String(statickey)},
			"email":    {S: aws.String(email)},
			"groups":   {SS: aws.StringSlice(groups)},
			"jwttoken": {S: aws.String(token)},
		},
	})
	if err != nil {
		return errors.Trace(err)
	}

	fmt.Printf("statickey: %s\n", statickey)
	fmt.Printf("jwttoken: %s\n", token)

	return nil
}
