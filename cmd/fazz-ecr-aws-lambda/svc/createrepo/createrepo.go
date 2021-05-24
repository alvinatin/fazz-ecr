package createrepo

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/payfazz/go-errors/v2"

	iamutil "github.com/payfazz/fazz-ecr/cmd/fazz-ecr-aws-lambda/util/iam"
	awsconfig "github.com/payfazz/fazz-ecr/config/aws"
)

func CreateRepoFor(email string, groups []string, repo string) error {
	repo = strings.TrimPrefix(repo, "https://")
	repo = strings.Split(repo, ":")[0]

	haveAccess := false
	for _, l := range awsconfig.RepoListPatternFromPolicyDoc(awsconfig.PolicyDocumentFor(email, groups)) {
		if strings.HasPrefix(repo, strings.TrimSuffix(l, "*")) {
			haveAccess = true
			break
		}
	}

	if !haveAccess {
		return errors.Trace(errAccessDenied)
	}

	repo = awsconfig.RepoNameOnlyOf(repo)
	sess, err := iamutil.EnvSession()
	if err != nil {
		return errors.Trace(err)
	}

	ecrsvc := ecr.New(sess)

	_, err = ecrsvc.CreateRepository(&ecr.CreateRepositoryInput{
		RepositoryName:     aws.String(repo),
		ImageTagMutability: aws.String(ecr.ImageTagMutabilityImmutable),
	})
	if err != nil {
		var err2 awserr.Error
		if errors.As(err, &err2) {
			if err2.Code() == "RepositoryAlreadyExistsException" {
				err = nil
			}
		}
	}
	if err != nil {
		return errors.Trace(err)
	}

	return nil
}
