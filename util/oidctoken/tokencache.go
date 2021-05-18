package oidctoken

import (
	"time"

	"github.com/payfazz/fazz-ecr/config/cachefile"
	"github.com/payfazz/fazz-ecr/util/jsonfile"
)

type tokenCache map[string]tokenCacheItem

type tokenCacheItem struct {
	IDToken string
	Exp     int64
}

func loadTokenCache() tokenCache {
	var ret tokenCache
	if err := jsonfile.Read(cachefile.OIDCToken, &ret); err != nil {
		return make(tokenCache)
	}

	now := time.Now().Unix()
	for k, v := range ret {
		if v.Exp <= now {
			delete(ret, k)
		}
	}

	return ret
}

func (c tokenCache) save() {
	jsonfile.Write(cachefile.OIDCToken, c)
}
