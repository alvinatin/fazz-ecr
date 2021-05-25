package iam

import (
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"

	awsconfig "github.com/payfazz/fazz-ecr/config/aws"
)

var envSession_ struct {
	once sync.Once
	err  error
	ses  *session.Session
}

func EnvSession() (*session.Session, error) {
	envSession_.once.Do(func() {
		envSession_.ses, envSession_.err = session.NewSession(
			&aws.Config{Region: aws.String(awsconfig.Region())},
		)
	})
	return envSession_.ses, envSession_.err
}
