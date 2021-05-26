package main

import (
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/payfazz/go-errors/v2"
	"gopkg.in/square/go-jose.v2"

	"github.com/payfazz/fazz-ecr/aws-lambda/fazz-ecr/util/iam"
	awsconfig "github.com/payfazz/fazz-ecr/config/aws"
	oidcconfig "github.com/payfazz/fazz-ecr/config/oidc"
)

func getAuth(token string) (string, []string, error) {
	jws, err := jose.ParseSigned(token)
	if err != nil {
		return "", nil, nil
	}

	if len(jws.Signatures) == 0 {
		return "", nil, nil
	}

	if jws.Signatures[0].Header.ExtraHeaders["typ"] == "statickey" {
		return authFromStaticKey(jws)
	}

	return authFromJwt(jws)
}

func authFromStaticKey(jwt *jose.JSONWebSignature) (string, []string, error) {
	envSession, err := iam.EnvSession()
	if err != nil {
		return "", nil, errors.Trace(err)
	}

	ddbsvc := dynamodb.New(envSession)
	result, err := ddbsvc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(awsconfig.StaticKeyTableName()),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {S: aws.String(string(jwt.UnsafePayloadWithoutVerification()))},
		},
	})
	if err != nil {
		return "", nil, errors.Trace(err)
	}

	if len(result.Item) == 0 {
		return "", nil, nil
	}

	return aws.StringValue(result.Item["email"].S), aws.StringValueSlice(result.Item["groups"].SS), nil
}

func authFromJwt(jwt *jose.JSONWebSignature) (string, []string, error) {
	sig := jwt.Signatures[0]

	if sig.Header.Algorithm == "none" || sig.Header.Algorithm == "" {
		return "", nil, nil
	}

	key, err := getJwtKeyByID(sig.Header.KeyID)
	if err != nil {
		return "", nil, err
	}
	if key == nil {
		return "", nil, nil
	}

	data, err := jwt.Verify(key)
	if err != nil {
		return "", nil, nil
	}

	var jwtBody struct {
		Iss    string   `json:"iss"`
		Aud    string   `json:"aud"`
		Exp    int64    `json:"exp"`
		Iat    int64    `json:"iat"`
		Email  string   `json:"email"`
		Groups []string `json:"groups"`
	}
	if err := json.Unmarshal(data, &jwtBody); err != nil {
		return "", nil, nil
	}

	if jwtBody.Iss != oidcconfig.Issuer || jwtBody.Aud != oidcconfig.ClientID {
		return "", nil, nil
	}

	now := time.Now().Unix()
	if !(jwtBody.Iat <= now && now <= jwtBody.Exp) {
		return "", nil, nil
	}

	return jwtBody.Email, jwtBody.Groups, nil
}
