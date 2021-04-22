package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/payfazz/go-errors"

	"github.com/payfazz/fazz-ecr/pkg/types"
)

var (
	roleArn        = "arn:aws:iam::322727087874:role/lambda-fazz-ecr"
	repoTemplate   = `arn:aws:ecr:ap-southeast-1:322727087874:repository/%s/*`
	policyTemplate = `{"Version":"2012-10-17","Statement":[` +
		`{"Effect":"Allow","Action":"ecr:GetAuthorizationToken","Resource":"*"},` +
		`{"Effect":"Allow","Action":"*","Resource":%s}]}`

	invalidCharRegex = regexp.MustCompile(`[^a-z0-9._-]`)
)

func normalizeAccessString(what string) string {
	return invalidCharRegex.ReplaceAllString(what, "-")
}

func assumeRole(name string, resourcePrefix []string) (*session.Session, error) {
	if resourcePrefix == nil {
		resourcePrefix = []string{}
	}

	envSession, err := session.NewSession()
	if err != nil {
		return nil, errors.Wrap(err)
	}

	stsSvc := sts.New(envSession)

	resourcePrefixEncoded, _ := json.Marshal(resourcePrefix)

	result, err := stsSvc.AssumeRole(&sts.AssumeRoleInput{
		RoleArn:         aws.String(roleArn),
		RoleSessionName: aws.String(name),
		Policy:          aws.String(fmt.Sprintf(policyTemplate, string(resourcePrefixEncoded))),
	})
	if err != nil {
		return nil, errors.Wrap(err)
	}

	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(
			*result.Credentials.AccessKeyId,
			*result.Credentials.SecretAccessKey,
			*result.Credentials.SessionToken,
		),
	})
	if err != nil {
		return nil, errors.Wrap(err)
	}

	return sess, nil
}

func getCredFor(email string, groups []string) (types.Cred, error) {
	resourcePrefix := []string{email}
	resourcePrefix = append(resourcePrefix, groups...)
	for i := range resourcePrefix {
		resourcePrefix[i] = fmt.Sprintf(repoTemplate, normalizeAccessString(resourcePrefix[i]))
	}

	tempSess, err := assumeRole(email, resourcePrefix)
	if err != nil {
		return types.Cred{}, errors.Wrap(err)
	}

	ecrSvc := ecr.New(tempSess)

	ecrLogin, err := ecrSvc.GetAuthorizationToken(&ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return types.Cred{}, errors.Wrap(err)
	}

	if len(ecrLogin.AuthorizationData) == 0 {
		return types.Cred{}, errors.Errorf("GetAuthorizationToken returing 0 authorization data")
	}

	authData := ecrLogin.AuthorizationData[0]
	cred, err := base64.StdEncoding.DecodeString(*authData.AuthorizationToken)
	if err != nil {
		return types.Cred{}, errors.Wrap(err)
	}
	credParts := strings.SplitN(string(cred), ":", 2)

	return types.Cred{
		User:   credParts[0],
		Pass:   credParts[1],
		Access: resourcePrefix,
		Exp:    authData.ExpiresAt.Unix(),
	}, nil
}
