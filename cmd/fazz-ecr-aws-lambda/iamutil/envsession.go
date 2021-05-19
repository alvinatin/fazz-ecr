package iamutil

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"

	awsconfig "github.com/payfazz/fazz-ecr/config/aws"
	"github.com/payfazz/go-errors/v2"
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
	})
	return envSession_.ses, envSession_.err
}

func getPolicyDoc(svc *iam.IAM, email string, groups []string) (string, error) {
	rolePolicyOutput, err := svc.GetRolePolicy(&iam.GetRolePolicyInput{
		RoleName:   aws.String(awsconfig.RoleNameFor(email)),
		PolicyName: aws.String(awsconfig.InlinePolicyName()),
	})
	if err == nil && rolePolicyOutput.PolicyDocument != nil {
		return *rolePolicyOutput.PolicyDocument, nil
	}

	return "", nil
}

func getPolicyDocFor(svc *iam.IAM, email string) (string, error) {
	output, err := svc.GetRolePolicy(&iam.GetRolePolicyInput{
		RoleName:   aws.String(awsconfig.RoleNameFor(email)),
		PolicyName: aws.String(awsconfig.InlinePolicyName()),
	})
	if err != nil {
		return "", errors.Trace(err)
	}
	if output.PolicyDocument == nil {
		return "", errors.Errorf("output.PolicyDocument is nil")
	}
	return *output.PolicyDocument, nil
}

func createRoleFor(svc *iam.IAM, email string) error {
	// output, err := svc.CreateRole(&iam.CreateRoleInput{
	// 	RoleName:                 aws.String(awsconfig.RoleNameFor(email)),
	// 	AssumeRolePolicyDocument: aws.String(awsconfig.AssumePolicyDoc()),
	// })
	return nil
}

func Asdf() {
	envSession, err := envSession()
	check(err)
	iamSvc := iam.New(envSession)

	createRoleOutput, err := iamSvc.CreateRole(&iam.CreateRoleInput{
		RoleName:                 aws.String(awsconfig.RoleNameFor(email)),
		AssumeRolePolicyDocument: aws.String(awsconfig.AssumePolicyDoc()),
	})

	if err != nil {
		fmt.Println(err.Error())
		fmt.Printf("%#v\n", err)
		return
	}

	if createRoleOutput.Role == nil || createRoleOutput.Role.Arn == nil {
		fmt.Println("invalid createRoleOutput")
		return
	}

	if *createRoleOutput.Role.Arn != awsconfig.RoleArnFor(email) {
		fmt.Println("created role arn is not match")
		return
	}

	// policyDoc, err := getPolicyDocFor(iam, email)
	// if err != nil {

	// }

	return

	// output, err := iamSvc.GetRolePolicy(&iam.GetRolePolicyInput{
	// 	RoleName:   aws.String(awsconfig.RoleNameFor("gegel")),
	// 	PolicyName: aws.String("ecr"),
	// })
	// fmt.Printf("%v %#v\n", output, err)

	// stsSvc := sts.New(envSession)
	// _ = stsSvc
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
