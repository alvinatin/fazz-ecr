package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker-credential-helpers/credentials"
	"github.com/payfazz/go-errors/v2"

	"github.com/payfazz/fazz-ecr/util/jsonfile"
	"github.com/payfazz/fazz-ecr/util/logerr"
	"github.com/payfazz/fazz-ecr/util/oidctoken"
)

func main() {
	if err := errors.Catch(main2); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
		logerr.Log(err)
		os.Exit(1)
	}
}

func main2() error {
	if len(os.Args) > 1 && os.Args[1] == "update-config" {
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
		}
		credHelpers, ok := credHelpersMap.(map[string]interface{})
		if !ok {
			return errors.New("invalid .credHelpers")
		}

		credHelpers["322727087874.dkr.ecr.ap-southeast-1.amazonaws.com"] = "fazz-ecr"

		config["credHelpers"] = credHelpers
		if err := jsonfile.Write(configPath, config); err != nil {
			return errors.Trace(err)
		}

		return nil
	}

	credentials.Serve(h{})
	return nil
}

type h struct{}

func (h) Add(*credentials.Credentials) error {
	return logerr.Log(errors.New("Not Implemented"))
}

func (h) Delete(serverURL string) error {
	cache := loadCache()
	delete(cache, serverURL)
	cache.save()
	return nil
}

func (h) List() (map[string]string, error) {
	cache := loadCache()

	ret := make(map[string]string)
	for k, v := range cache {
		ret[k] = v.User
	}
	return ret, nil
}

func (h) Get(serverURL string) (string, string, error) {
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
		fmt.Fprintf(&resp, "\nYou are now logged in into registry %s\n", strings.TrimPrefix(serverURL, "https://"))
		fmt.Fprintf(&resp, "\nYou should have push/pull permission for following repositories:\n")
		for _, v := range result.Access {
			fmt.Fprintf(&resp, "- %s\n", v)
		}
		fmt.Fprintf(&resp, "\nYou can now close this window\n")

		return resp.String(), nil
	}

	if err := oidctoken.GetToken(processToken); err != nil {
		return "", "", logerr.Log(err)
	}

	item := cache[serverURL]
	return item.User, item.Pass, nil
}
