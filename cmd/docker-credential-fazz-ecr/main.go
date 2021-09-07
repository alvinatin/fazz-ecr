package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker-credential-helpers/credentials"
	"github.com/payfazz/go-errors/v2"

	"github.com/payfazz/fazz-ecr/config/cachefile"
	"github.com/payfazz/fazz-ecr/util/jsonfile"
	"github.com/payfazz/fazz-ecr/util/logerr"
	"github.com/payfazz/fazz-ecr/util/oidctoken"
)

var repoList = []string{
	"322727087874.dkr.ecr.ap-southeast-1.amazonaws.com",
}

func main() {
	if err := errors.Catch(run); err != nil {
		logerr.Log(err)
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) != 2 {
		return usage()
	}

	switch os.Args[1] {
	case "update-config":
		return updateConfig()
	case "login":
		return login()
	case "list-access":
		return listAccess()
	case "get":
		return credentials.HandleCommand(h{}, os.Args[1], os.Stdin, os.Stdout)
	default:
		return usage()
	}
}

func usage() error {
	fmt.Fprintf(os.Stderr, "Usage: %s <update-config|login|list-access|get>\n", os.Args[0])
	os.Exit(1)
	return nil
}

func updateConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return errors.Trace(err)
	}

	configPath := filepath.Join(home, ".docker", "config.json")

	var config map[string]interface{}
	if err := jsonfile.Read(configPath, &config); err != nil {
		return errors.Trace(err)
	}

	credHelpersMap := config["credHelpers"]
	if credHelpersMap == nil {
		credHelpersMap = make(map[string]interface{})
		config["credHelpers"] = credHelpersMap
	}
	credHelpers, ok := credHelpersMap.(map[string]interface{})
	if !ok {
		return errors.New("invalid .credHelpers")
	}

	helperName := strings.TrimPrefix(filepath.Base(os.Args[0]), "docker-credential-")
	for _, repo := range repoList {
		credHelpers[repo] = helperName
	}

	if err := jsonfile.Write(configPath, config); err != nil {
		return errors.Trace(err)
	}

	return nil
}

func login() error {
	os.RemoveAll(cachefile.DockerCreds)
	os.RemoveAll(cachefile.OIDCToken)
	for _, repo := range repoList {
		_, _, err := h{}.Get(repo)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h) Get(serverURL string) (string, string, error) {
	serverURL = strings.TrimPrefix(serverURL, "https://")
	serverURL = strings.Split(serverURL, "/")[0]

	cache := loadCache()

	if item, ok := cache[serverURL]; ok {
		return item.User, item.Pass, nil
	}

	processToken := func(IDToken string) (string, error) {
		result, err := exchageToken(IDToken)
		if err != nil {
			return "", err
		}

		cache[serverURL] = result
		cache.save()

		var resp strings.Builder
		fmt.Fprintf(&resp, "docker-credential-fazz-ecr\n")
		fmt.Fprintf(&resp, "==========================\n")
		fmt.Fprintf(&resp, "\nYou are now logged in into registry %s\n", serverURL)
		fmt.Fprintf(&resp, "\nYou should have push/pull permission for following repositories:\n")
		for _, v := range result.Access {
			fmt.Fprintf(&resp, "- %s\n", v)
		}
		fmt.Fprintf(&resp, "\nYou can now close this window\n")

		return resp.String(), nil
	}

	if err := oidctoken.GetToken(processToken); err != nil {
		return "", "", err
	}

	item := cache[serverURL]
	return item.User, item.Pass, nil
}

func listAccess() error {
	c := loadCache()
	for _, v := range c {
		for _, r := range v.Access {
			fmt.Println(r)
		}
	}
	return nil
}

type h struct{}

var errNotImplemented = fmt.Errorf("not implemented")

func (h) Add(*credentials.Credentials) error { return errNotImplemented }
func (h) Delete(serverURL string) error      { return errNotImplemented }
func (h) List() (map[string]string, error)   { return nil, errNotImplemented }
