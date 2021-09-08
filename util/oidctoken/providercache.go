package oidctoken

import (
	"encoding/json"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/payfazz/go-errors/v2"

	"github.com/payfazz/fazz-ecr/config/cachefile"
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
	if err := jsonfile.Read(cachefile.OIDCProvider, &ret); err != nil {
		return make(providerCache)
	}

	now := time.Now().Unix()
	for k, v := range ret {
		if v.Exp <= now {
			delete(ret, k)
		}
	}

	return ret
}

func (c providerCache) save() {
	jsonfile.Write(cachefile.OIDCProvider, c)
}

func (c providerCache) ensure(issuer string) error {
	if v, ok := c[issuer]; ok {
		if v.Exp < time.Now().Unix() {
			return nil
		}
	}

	u, err := url.Parse(issuer)
	if err != nil {
		return errors.Trace(err)
	}
	if u.Scheme != "https" {
		return errors.Errorf("oidc provider must use https")
	}
	if !strings.HasPrefix(u.EscapedPath(), "/") && u.EscapedPath() != "" {
		return errors.Errorf("oidc provider must use absolute path")
	}
	u, _ = u.Parse(path.Join(u.EscapedPath(), ".well-known/openid-configuration"))

	resp, err := http.Get(u.String())
	if err != nil {
		return errors.Trace(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.Errorf("oidc config url endpoint returning http code: %d", resp.StatusCode)
	}

	var config struct {
		Issuer   string   `json:"issuer"`
		AuthURL  string   `json:"authorization_endpoint"`
		TokenURL string   `json:"token_endpoint"`
		RespType []string `json:"response_types_supported"`
		Scopes   []string `json:"scopes_supported"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return errors.Trace(err)
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
		return errors.Errorf(`oidc config doesn't support "code" response_type`)
	}

	for _, v := range []string{"openid", "email", "groups"} {
		if !inArray(v, config.Scopes) {
			return errors.Errorf(`oidc scopes doesn't support "%s" scope`, v)
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

func (c providerCache) getAuthUri(issuer string, clientID string, redirect string, state string) (string, error) {
	if err := c.ensure(issuer); err != nil {
		return "", err
	}

	u, _ := url.Parse(c[issuer].AuthURL)

	q := u.Query()
	q.Set("client_id", clientID)
	q.Set("response_type", "code")
	q.Set("redirect_uri", redirect)
	q.Set("scope", "openid email groups offline_access")
	q.Set("state", state)

	u.RawQuery = q.Encode()
	return u.String(), nil
}

func (c providerCache) refreshIDToken(issuer string, clientID string, refreshToken string) (string, string, error) {
	if err := c.ensure(issuer); err != nil {
		return "", "", err
	}

	q := make(url.Values)
	q.Set("client_id", clientID)
	q.Set("grant_type", "refresh_token")
	q.Set("refresh_token", refreshToken)

	return requestToken(c[issuer].TokenURL, q)
}

func (c providerCache) getIDToken(issuer string, clientID string, redirect string, code string) (string, string, error) {
	if err := c.ensure(issuer); err != nil {
		return "", "", err
	}

	q := make(url.Values)
	q.Set("client_id", clientID)
	q.Set("grant_type", "authorization_code")
	q.Set("redirect_uri", redirect)
	q.Set("code", code)

	return requestToken(c[issuer].TokenURL, q)
}

func requestToken(url string, body url.Values) (string, string, error) {
	resp, err := http.Post(
		url,
		"application/x-www-form-urlencoded",
		strings.NewReader(body.Encode()),
	)
	if err != nil {
		return "", "", errors.Trace(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", "", errors.Errorf("token url endpoint returning http code: %d", resp.StatusCode)
	}

	var result struct {
		IDToken      string `json:"id_token"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", errors.Trace(err)
	}

	return result.IDToken, result.RefreshToken, nil
}
