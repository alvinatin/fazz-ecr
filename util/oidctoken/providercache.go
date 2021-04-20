package oidctoken

import (
	"net/url"
	"path"
	"time"

	"github.com/payfazz/fazz-ecr/util/cachedir"
)

var providerCacheFileName = "oidc.json"

type providerCache map[string]providerCacheItem

type providerCacheItem struct {
	AuthURL  string
	TokenURL string
	Exp      int64
}

func loadProviderCache() providerCache {
	var ret providerCache
	if err := cachedir.LoadJSONFile(providerCacheFileName, &ret); err != nil {
		return nil
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

func (c providerCache) save() {
	if c == nil {
		c = make(providerCache)
	}

	cachedir.SaveJSONFile(providerCacheFileName, c)
}

func (c providerCache) ensure(p string) error {
	if v, ok := c[p]; ok {
		if v.Exp < time.Now().Unix() {
			return nil
		}
	}

	u, err := url.Parse(p)
	if err != nil {
		return err
	}
	if u.Scheme != "https" {
		return
	}
	u, _ = u.Parse(path.Join("/", u.EscapedPath(), ".well-known/openid-configuration"))

}
