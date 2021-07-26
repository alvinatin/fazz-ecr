package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/payfazz/go-errors/v2"

	"github.com/payfazz/fazz-ecr/util/logerr"
	"github.com/payfazz/fazz-ecr/util/oidctoken"
)

func main() {
	if err := errors.Catch(run); err != nil {
		logerr.Log(err)
		fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "USAGE: %s <repo-name>\n", os.Args[0])
		os.Exit(1)
	}

	repo := os.Args[1]
	repo = strings.TrimPrefix(repo, "https://")
	repo = strings.Split(repo, ":")[0]

	processToken := func(IDToken string) (string, error) {
		if err := createRepo(IDToken, repo); err != nil {
			return "", errors.Errorf("cannot create repo: %w\n", err)
		}

		var resp strings.Builder
		fmt.Fprintf(&resp, "fazz-ecr-create-repo\n")
		fmt.Fprintf(&resp, "====================\n")

		msg := fmt.Sprintf("Repo %s created", repo)
		fmt.Fprintf(&resp, "\n%s\n", msg)
		fmt.Printf("%s\n", msg)

		fmt.Fprintf(&resp, "\nYou can now close this window\n")

		return resp.String(), nil
	}

	return oidctoken.GetToken(processToken)
}
