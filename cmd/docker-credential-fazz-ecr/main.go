package main

import (
	"fmt"
	"strings"

	"github.com/docker/docker-credential-helpers/credentials"
	"github.com/payfazz/go-errors"

	"github.com/payfazz/fazz-ecr/util/oidctoken"
)

var errNotImplemented = fmt.Errorf("not implemented")

func main() {
	credentials.Serve(h{})
}

type h struct{}

func (h) Add(*credentials.Credentials) error {
	return errors.Wrap(errNotImplemented)
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
			return "", errors.Wrap(err)
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
		return "", "", errors.Wrap(err)
	}

	item := cache[serverURL]
	return item.User, item.Pass, nil
}
