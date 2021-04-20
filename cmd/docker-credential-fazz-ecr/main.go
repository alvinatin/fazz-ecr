package main

import (
	stderrors "errors"

	"github.com/docker/docker-credential-helpers/credentials"
	"github.com/payfazz/go-errors"
)

var errNotImplemented = stderrors.New("not implemented")

func main() { credentials.Serve(h{}) }

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

	// TODO
	return "", "", errors.Wrap(errNotImplemented)
}
