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
	if err := errors.Catch(main2); err != nil {
		logerr.Log(err)
		os.Exit(1)
	}
}

func main2() error {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "USAGE: %s <repo-name>\n", os.Args[0])
		return errors.New("invalid os.Args")
	}

	repo := os.Args[1]

	processToken := func(IDToken string) (string, error) {
		if err := createRepo(IDToken, repo); err != nil {
			fmt.Fprintf(os.Stderr, "Error on creating repo %s\n%s\n", repo, err.Error())
			return "", err
		}

		var resp strings.Builder
		fmt.Fprintf(&resp, "fazz-ecr-create-repo\n")
		fmt.Fprintf(&resp, "====================\n")

		msg := fmt.Sprintf("Repo %s created", repo)
		fmt.Fprintf(&resp, "\n%s\n", msg)
		fmt.Fprintf(os.Stderr, "%s\n", msg)

		fmt.Fprintf(&resp, "\nYou can now close this window\n")

		return resp.String(), nil
	}

	return oidctoken.GetToken(processToken)
}
