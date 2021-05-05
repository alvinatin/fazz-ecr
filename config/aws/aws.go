package aws

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var (
	invalidChar = regexp.MustCompile(`[^a-z0-9-]`)
	charLimit   = 50

	region = "ap-southeast-1"

	roleNameTemplate = `fazz-ecr-%s`
	roleArnTemplate  = `arn:aws:iam::322727087874:role/%s`
	repoTemplate     = `arn:aws:ecr:ap-southeast-1:322727087874:repository/%s/*`
	policyTemplate   = `` +
		`{"Version":"2012-10-17","Statement":[` +
		`{"Sid":"Login","Effect":"Allow","Action":"ecr:GetAuthorizationToken","Resource":"*"},` +
		`{"Sid":"CreatePushPull","Effect":"Allow","Action":` +
		`[` +
		`"ecr:CreateRepository",` +
		`"ecr:BatchGetImage",` +
		`"ecr:GetDownloadUrlForLayer",` +
		`"ecr:BatchCheckLayerAvailability",` +
		`"ecr:InitiateLayerUpload",` +
		`"ecr:UploadLayerPart",` +
		`"ecr:CompleteLayerUpload",` +
		`"ecr:PutImage"` +
		`]` +
		`,"Resource":%s}]}`
)

func normalize(value string) string {
	value = invalidChar.ReplaceAllString(strings.ToLower(value), "-")
	if len(value) > charLimit {
		value = value[:charLimit]
	}
	return value
}

func repoFor(namespace string) string {
	return fmt.Sprintf(repoTemplate, normalize(namespace))
}

func Region() string { return region }

func RoleNameFor(email string) string {
	return fmt.Sprintf(roleNameTemplate, normalize(email))
}

func RoleArnFor(email string) string {
	return fmt.Sprintf(roleArnTemplate, RoleNameFor(email))
}

func PolicyDocumentFor(email string, groups []string) string {
	resources := []string{repoFor(email)}
	for _, g := range groups {
		resources = append(resources, repoFor(g))
	}
	resourcesEncoded, _ := json.Marshal(resources)
	return fmt.Sprintf(policyTemplate, resourcesEncoded)
}

func RepoListFromPolicyDoc(document string) []string {
	var doc struct {
		Statement []struct {
			Sid      string
			Resource json.RawMessage
		}
	}
	json.Unmarshal([]byte(document), &doc)
	for _, v := range doc.Statement {
		if v.Sid == "CreatePushPull" {
			var ret []string
			json.Unmarshal(v.Resource, &ret)
			return ret
		}
	}
	return nil
}
