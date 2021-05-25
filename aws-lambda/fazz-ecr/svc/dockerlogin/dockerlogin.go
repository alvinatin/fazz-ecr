package dockerlogin

import (
	"encoding/base64"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/payfazz/go-errors/v2"

	iamutil "github.com/payfazz/fazz-ecr/aws-lambda/fazz-ecr/util/iam"
	awsconfig "github.com/payfazz/fazz-ecr/config/aws"
	"github.com/payfazz/fazz-ecr/pkg/types"
)

func GetCredFor(email string, groups []string) (cred types.Cred, err error) {
	envSession, err := iamutil.EnvSession()
	if err != nil {
		return cred, errors.Trace(err)
	}

	roleName := awsconfig.RoleNameFor(email)
	roleArn := awsconfig.RoleArnFor(email)

	iamsvc := iam.New(envSession)
	stssvc := sts.New(envSession)

	doAction := func(
		action func() (interface{}, error),
		onOk func(interface{}) error,
		expectedFailCode string,
		onFail func() error,
	) error {
		alreadyHandled := false
		for {
			result, err := action()
			if err == nil {
				return onOk(result)
			}

			if alreadyHandled {
				return err
			}

			var awserr awserr.Error
			if !errors.As(err, &awserr) {
				return err
			}

			if awserr.Code() != expectedFailCode {
				return err
			}

			if err := onFail(); err != nil {
				return err
			}

			alreadyHandled = true
		}
	}

	createRole := func() error {
		result, err := iamsvc.CreateRole(&iam.CreateRoleInput{
			RoleName:                 aws.String(roleName),
			AssumeRolePolicyDocument: aws.String(awsconfig.AssumePolicyDoc()),
		})
		if err != nil {
			return errors.Trace(err)
		}
		if *result.Role.Arn != roleArn {
			return errors.New("created role arn didn't match")
		}
		return nil
	}

	createRolePolicy := func() error {
		return doAction(
			func() (interface{}, error) {
				result, err := iamsvc.GetRole(&iam.GetRoleInput{
					RoleName: aws.String(roleName),
				})
				if err != nil {
					return nil, errors.Trace(err)
				}
				return result, nil
			},
			func(r interface{}) error {
				roleResult := r.(*iam.GetRoleOutput)
				if *roleResult.Role.Arn != roleArn {
					return errors.New("existing role arn didn't match")
				}
				_, err := iamsvc.PutRolePolicy(&iam.PutRolePolicyInput{
					RoleName:       aws.String(roleName),
					PolicyName:     aws.String(awsconfig.InlinePolicyName()),
					PolicyDocument: aws.String(awsconfig.PolicyDocumentFor(email, groups)),
				})
				if err != nil {
					return errors.Trace(err)
				}
				return nil
			},
			"NoSuchEntity",
			createRole,
		)
	}

	doGetCred := func() error {
		return doAction(
			func() (interface{}, error) {
				result, err := iamsvc.GetRolePolicy(&iam.GetRolePolicyInput{
					RoleName:   aws.String(roleName),
					PolicyName: aws.String(awsconfig.InlinePolicyName()),
				})
				if err != nil {
					return nil, errors.Trace(err)
				}
				return result, nil
			},
			func(r interface{}) error {
				rolePolicyResult := r.(*iam.GetRolePolicyOutput)
				doc, err := url.QueryUnescape(*rolePolicyResult.PolicyDocument)
				if err != nil {
					return errors.Trace(err)
				}

				assumeRoleResult, err := stssvc.AssumeRole(&sts.AssumeRoleInput{
					RoleArn:         aws.String(roleArn),
					RoleSessionName: aws.String(email),
				})
				if err != nil {
					return errors.Trace(err)
				}

				assumedSession := envSession.Copy(aws.NewConfig().WithCredentials(credentials.NewStaticCredentials(
					*assumeRoleResult.Credentials.AccessKeyId,
					*assumeRoleResult.Credentials.SecretAccessKey,
					*assumeRoleResult.Credentials.SessionToken,
				)))

				ecrsvc := ecr.New(assumedSession)
				authTokenResult, err := ecrsvc.GetAuthorizationToken(&ecr.GetAuthorizationTokenInput{})
				if err != nil {
					return errors.Trace(err)
				}

				token, err := base64.StdEncoding.DecodeString(*authTokenResult.AuthorizationData[0].AuthorizationToken)
				if err != nil {
					return errors.Trace(err)
				}
				tokenParts := strings.SplitN(string(token), ":", 2)

				cred.User = tokenParts[0]
				cred.Pass = tokenParts[1]
				cred.Access = awsconfig.RepoListPatternFromPolicyDoc(doc)
				cred.Exp = authTokenResult.AuthorizationData[0].ExpiresAt.Unix()

				return nil
			},
			"NoSuchEntity",
			createRolePolicy,
		)
	}

	if err := doGetCred(); err != nil {
		return cred, err
	}

	return cred, nil
}
