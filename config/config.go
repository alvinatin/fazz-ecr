package config

import (
	"os"
	"path/filepath"
)

var (
	CacheFileDockerCreds  = "docker-creds.json"
	CacheFileOIDCProvider = "oidc-provider.json"
	CacheFileOIDCToken    = "oidc-token.json"

	OIDCIssuer       = "https://dex.fazzfinancial.com"
	OIDCClientID     = "ecrhelper"
	OIDCCallbackPort = 3000
)

func init() {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		panic(err)
	}

	cacheDir = filepath.Join(cacheDir, "fazz-ecr")

	if err := os.MkdirAll(cacheDir, 0o700); err != nil {
		panic(err)
	}

	CacheFileDockerCreds = filepath.Join(cacheDir, CacheFileDockerCreds)
	CacheFileOIDCProvider = filepath.Join(cacheDir, CacheFileOIDCProvider)
	CacheFileOIDCToken = filepath.Join(cacheDir, CacheFileOIDCToken)
}
