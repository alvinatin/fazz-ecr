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

	IDTokenExchangeEndpoint = "https://9djfz7zb34.execute-api.ap-southeast-1.amazonaws.com/docker-login"
)

func init() {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		cacheDir = os.TempDir()
	}

	cacheDir = filepath.Join(cacheDir, "fazz-ecr")
	CacheFileDockerCreds = filepath.Join(cacheDir, CacheFileDockerCreds)
	CacheFileOIDCProvider = filepath.Join(cacheDir, CacheFileOIDCProvider)
	CacheFileOIDCToken = filepath.Join(cacheDir, CacheFileOIDCToken)
}
