package dockerlogin

import (
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/payfazz/go-errors/v2"

	awsconfig "github.com/payfazz/fazz-ecr/config/aws"
)

var envSession_ struct {
	once sync.Once
	err  error
	ses  *session.Session
}

func envSession() (*session.Session, error) {
	envSession_.once.Do(func() {
		envSession_.ses, envSession_.err = session.NewSession(
			&aws.Config{Region: aws.String(awsconfig.Region())},
		)
		if envSession_.err != nil {
			envSession_.err = errors.Trace(envSession_.err)
		}
	})
	return envSession_.ses, envSession_.err
}
