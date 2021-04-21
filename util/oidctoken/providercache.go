package oidctoken

import (
	"encoding/json"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/payfazz/go-errors"

	"github.com/payfazz/fazz-ecr/config"
	"github.com/payfazz/fazz-ecr/util/jsonfile"
)

type providerCache map[string]providerCacheItem

type providerCacheItem struct {
	AuthURL  string
	TokenURL string
	Exp      int64
}

func loadProviderCache() providerCache {
	var ret providerCache
	if err := jsonfile.Read(config.CacheFileOIDCProvider, &ret); err != nil {
		return make(providerCache)
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
	jsonfile.Write(config.CacheFileOIDCProvider, c)
}

func (c providerCache) ensure(issuer string) error {
	if v, ok := c[issuer]; ok {
		if v.Exp < time.Now().Unix() {
			return nil
		}
	}

	u, err := url.Parse(issuer)
	if err != nil {
		return errors.Wrap(err)
	}
	if u.Scheme != "https" {
		return errors.Errorf("oidc provider must use https")
	}
	u, _ = u.Parse(path.Join("/", u.EscapedPath(), ".well-known/openid-configuration"))

	resp, err := http.Get(u.String())
	if err != nil {
		return errors.Wrap(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.Errorf("oidc config url not returning 200")
	}

	var config struct {
		Issuer   string   `json:"issuer"`
		AuthURL  string   `json:"authorization_endpoint"`
		TokenURL string   `json:"token_endpoint"`
		RespType []string `json:"response_types_supported"`
		Scopes   []string `json:"scopes_supported"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return errors.Wrap(err)
	}

	if config.Issuer != issuer {
		return errors.Errorf("oidc config issuer is not match")
	}

	authURL, err := url.Parse(config.AuthURL)
	if err != nil {
		return errors.NewWithCause("Cannot parse oidc auth url", err)
	}
	if authURL.Scheme != "https" {
		return errors.Errorf("auth url is not using https")
	}

	tokenURL, err := url.Parse(config.TokenURL)
	if err != nil {
		return errors.NewWithCause("Cannot parse oidc token url", err)
	}
	if tokenURL.Scheme != "https" {
		return errors.Errorf("auth url is not using https")
	}

	if !inArray("code", config.RespType) {
		return errors.Errorf("oidc config doesn't support \"code\" response_type")
	}

	for _, v := range []string{"email", "openid", "groups"} {
		if !inArray(v, config.Scopes) {
			return errors.Errorf("oidc scopes doesn't support \"%s\" scope", v)
		}
	}

	c[issuer] = providerCacheItem{
		AuthURL:  config.AuthURL,
		TokenURL: config.TokenURL,
		Exp:      time.Now().Add(24 * time.Hour).Unix(),
	}

	c.save()

	return nil
}

func (c providerCache) getAuthUri(issuer string, clientID string, redirectURI string, state string) (string, error) {
	if err := c.ensure(issuer); err != nil {
		return "", errors.Wrap(err)
	}

	u, _ := url.Parse(c[issuer].AuthURL)

	q := u.Query()
	q.Set("client_id", clientID)
	q.Set("response_type", "code")
	q.Set("redirect_uri", redirectURI)
	q.Set("scope", "openid email groups")
	q.Set("state", state)

	u.RawQuery = q.Encode()
	return u.String(), nil
}

func (c providerCache) getIDToken(issuer string, clientID string, redirectURI string, code string) (string, error) {
	if err := c.ensure(issuer); err != nil {
		return "", errors.Wrap(err)
	}

	q := make(url.Values)
	q.Set("client_id", clientID)
	q.Set("grant_type", "authorization_code")
	q.Set("redirect_uri", redirectURI)
	q.Set("code", code)

	resp, err := http.Post(
		c[issuer].TokenURL,
		"application/x-www-form-urlencoded",
		strings.NewReader(q.Encode()),
	)
	if err != nil {
		return "", errors.Wrap(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", errors.Errorf("token url not returning 200")
	}

	var result struct {
		IDToken string `json:"id_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", errors.Wrap(err)
	}

	return result.IDToken, nil
}
