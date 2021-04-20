package main

import (
	"time"

	"github.com/payfazz/fazz-ecr/util/cachedir"
)

var cacheFileName = "docker-creds.json"

type cache map[string]cacheItem

type cacheItem struct {
	User string
	Pass string
	Exp  int64
}

func loadCache() cache {
	var ret cache
	if err := cachedir.LoadJSONFile(cacheFileName, &ret); err != nil {
		return make(cache)
	}

	var expList []string
	now := time.Now().Unix()
	for k, v := range ret {
		if v.Exp <= now {
			expList = append(expList, k)
		}
	}
	for _, k := range expList {
		delete(ret, k)
	}

	return ret
}

func (c cache) save() {
	cachedir.SaveJSONFile(cacheFileName, c)
}
