package main

import (
	"time"

	"github.com/payfazz/fazz-ecr/config/cachefile"
	"github.com/payfazz/fazz-ecr/pkg/types"
	"github.com/payfazz/fazz-ecr/util/jsonfile"
)

type cache map[string]types.Cred

func loadCache() cache {
	var ret cache
	if err := jsonfile.Read(cachefile.DockerCreds, &ret); err != nil {
		return make(cache)
	}

	now := time.Now().Unix()
	for k, v := range ret {
		if v.Exp <= now {
			delete(ret, k)
		}
	}

	return ret
}

func (c cache) save() {
	jsonfile.Write(cachefile.DockerCreds, c)
}
