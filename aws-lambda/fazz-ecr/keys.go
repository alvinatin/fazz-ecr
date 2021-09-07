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

	// multiple key with same kid is not supported
	if len(keys) != 0 {
		return &keys[0], nil
	}

	if !time.Now().After(checkAgain) {
		return nil, nil
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()

	keys = cache.keys.Key(kid)
	checkAgain = cache.checkAgain

	if len(keys) != 0 {
		return &keys[0], nil
	}

	if !time.Now().After(checkAgain) {
		return nil, nil
	}

	cache.checkAgain = time.Now().Add(5 * time.Minute)

	ctx, canceCtx := context.WithTimeout(context.Background(), 2*time.Minute)
	defer canceCtx()

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

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, errors.Trace(err)
	}
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
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

	u, err = url.Parse(config.JwksURI)
	if err != nil {
		return nil, errors.Trace(err)
	}

	if u.Scheme != "https" {
		return nil, errors.Errorf("jwks_uri endpoint must use https")
	}

	req, err = http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, errors.Trace(err)
	}
	resp, err = http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, errors.Trace(err)
	}
	if resp.StatusCode != 200 {
		return nil, errors.Errorf("jwks_uri endpoint returning http code: %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&cache.keys); err != nil {
		return nil, errors.Trace(err)
	}

	keys = cache.keys.Key(kid)
	if len(keys) != 0 {
		return &keys[0], nil
	}
	return nil, nil
}
