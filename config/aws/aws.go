package aws

import (
	"crypto/sha256"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var (
	invalidChar   = regexp.MustCompile(`[^a-z0-9-]`)
	charLimit     = 50
	charLimitHash = 10

	region = "ap-southeast-1"

	roleNameTemplate = `fazz-ecr-%s`
	roleArnTemplate  = `arn:aws:iam::322727087874:role/%s`
	repoTemplate     = `arn:aws:ecr:ap-southeast-1:322727087874:repository/%s/*`

	inlinePolicyName = "ecr"

	policySidForManageRepo = "PushPull"

	policyTemplate = `` +
		`{"Version":"2012-10-17","Statement":[` +
		`{"Sid":"Login","Effect":"Allow","Action":"ecr:GetAuthorizationToken","Resource":"*"},` +
		`{"Sid":"` + policySidForManageRepo + `","Effect":"Allow","Action":` +
		`[` +
		`"ecr:BatchGetImage",` +
		`"ecr:GetDownloadUrlForLayer",` +
		`"ecr:BatchCheckLayerAvailability",` +
		`"ecr:InitiateLayerUpload",` +
		`"ecr:UploadLayerPart",` +
		`"ecr:CompleteLayerUpload",` +
		`"ecr:PutImage"` +
		`]` +
		`,"Resource":%s}]}`

	assumePolicyDoc = `` +
		`{"Version": "2012-10-17","Statement": [` +
		`{"Effect": "Allow",` +
		`"Principal": {"AWS": "arn:aws:iam::322727087874:role/lambda-fazz-ecr"},` +
		`"Action": "sts:AssumeRole"}]}`
)

func normalize(value string) string {
	if len(value) > charLimit {
		sumBytes := sha256.Sum256([]byte(value))
		sum := base32.StdEncoding.EncodeToString(sumBytes[:])
		value = value[:charLimit-charLimitHash] + sum[:charLimitHash]
	}
	value = invalidChar.ReplaceAllString(strings.ToLower(value), "-")
	return value
}

func repoFor(namespace string) string {
	return fmt.Sprintf(repoTemplate, normalize(namespace))
}

func Region() string {
	return region
}

func AssumePolicyDoc() string {
	return assumePolicyDoc
}

func InlinePolicyName() string {
	return inlinePolicyName
}

func RoleNameFor(email string) string {
	return fmt.Sprintf(roleNameTemplate, normalize(email))
}

func RoleArnFor(email string) string {
	return fmt.Sprintf(roleArnTemplate, RoleNameFor(email))
}

func PolicyDocumentFor(email string, groups []string) string {
	resourceSet := make(map[string]struct{})
	resourceSet[repoFor(email)] = struct{}{}
	for _, g := range groups {
		resourceSet[repoFor(g)] = struct{}{}
	}
	resources := make([]string, 0, len(resourceSet))
	for k := range resourceSet {
		resources = append(resources, k)
	}
	resourcesEncoded, _ := json.Marshal(resources)
	return fmt.Sprintf(policyTemplate, resourcesEncoded)
}

func RepoListPatternFromPolicyDoc(document string) []string {
	var doc struct {
		Statement []struct {
			Sid      string
			Resource []string
		}
	}
	json.Unmarshal([]byte(document), &doc)
	for _, v := range doc.Statement {
		if v.Sid == policySidForManageRepo {
			var ret []string
			for _, x := range v.Resource {
				ret = append(ret, resourceToRepo(x))
			}
			return ret
		}
	}
	return nil
}

var resourceRegex = regexp.MustCompile(`^arn:aws:ecr:([^:]*):([^:]*):repository/(.*)$`)

func resourceToRepo(input string) string {
	matches := resourceRegex.FindStringSubmatch(input)
	return fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com/%s", matches[2], matches[1], matches[3])
}

var repoRegex = regexp.MustCompile(`^([^.]*).dkr.ecr.([^.]*).amazonaws.com/(.*)$`)

func RepoNameOnlyOf(input string) string {
	matches := repoRegex.FindStringSubmatch(input)
	if len(matches) != 4 {
		return ""
	}
	return matches[3]
}
