package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/payfazz/go-errors/v2"
	"gopkg.in/square/go-jose.v2"

	oidcconfig "github.com/payfazz/fazz-ecr/config/oidc"
)

var cache struct {
	mu         sync.RWMutex
	checkAgain time.Time
	keys       jose.JSONWebKeySet
}

func getJwtKeyByID(kid string) (*jose.JSONWebKey, error) {
	cache.mu.RLock()
	keys := cache.keys.Key(kid)
	checkAgain := cache.checkAgain
	cache.mu.RUnlock()

	if len(keys) != 0 {
		return &keys[0], nil
	}

	if !time.Now().After(checkAgain) {
		return nil, nil
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()

	if len(keys) != 0 {
		return &keys[0], nil
	}

	if !time.Now().After(checkAgain) {
		return nil, nil
	}

	u, err := url.Parse(oidcconfig.Issuer)
	if err != nil {
		return nil, errors.Trace(err)
	}
	if u.Scheme != "https" {
		return nil, errors.Errorf("oidc provider must use https")
	}
	if !strings.HasPrefix(u.EscapedPath(), "/") && u.EscapedPath() != "" {
		return nil, errors.Errorf("oidc provider must use absolute path")
	}
	u, _ = u.Parse(path.Join(u.EscapedPath(), ".well-known/openid-configuration"))

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, errors.Trace(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.Errorf("oidc config url endpoint returning http code: %d", resp.StatusCode)
	}

	var config struct {
		Issuer  string `json:"issuer"`
		JwksURI string `json:"jwks_uri"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, errors.Trace(err)
	}

	if config.Issuer != oidcconfig.Issuer {
		return nil, errors.Errorf("oidc config issuer is not match")
	}

	jwkURL, err := url.Parse(config.JwksURI)
	if err != nil {
		return nil, errors.Trace(err)
	}

	if jwkURL.Scheme != "https" {
		return nil, errors.Errorf("jwks_uri endpoint must use https")
	}

	req, err := http.NewRequest("GET", config.JwksURI, nil)
	if err != nil {
		return nil, errors.Trace(err)
	}
	ctx, cancelFn := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancelFn()
	res, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, errors.Trace(err)
	}
	if res.StatusCode != 200 {
		return nil, errors.Errorf("jwks_uri endpoint returning http code: %d", resp.StatusCode)
	}
	defer res.Body.Close()

	if err := json.NewDecoder(res.Body).Decode(&cache.keys); err != nil {
		return nil, errors.Trace(err)
	}

	checkAgain = time.Now().Add(5 * time.Minute)
	keys = cache.keys.Key(kid)
	if len(keys) != 0 {
		return &keys[0], nil
	}
	return nil, nil
}
