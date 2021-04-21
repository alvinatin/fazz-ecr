package oidctoken

import (
	"time"

	"github.com/payfazz/fazz-ecr/util/cachedir"
)

var tokenCacheFileName = "oidc-token.json"

type tokenCache map[string]tokenCacheItem

type tokenCacheItem struct {
	IDToken string
	Exp     int64
}

func loadTokenCache() tokenCache {
	var ret tokenCache
	if err := cachedir.LoadJSONFile(tokenCacheFileName, &ret); err != nil {
		return make(tokenCache)
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

func (c tokenCache) save() {
	cachedir.SaveJSONFile(tokenCacheFileName, c)
}
