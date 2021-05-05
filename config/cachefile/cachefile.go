package cachefile

import (
	"fmt"
	"os"
	"path/filepath"
)

var (
	cacheDirName = "fazz-ecr"

	DockerCreds  = "docker-creds.json"
	OIDCProvider = "oidc-provider.json"
	OIDCToken    = "oidc-token.json"
)

func init() {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: cannot get cache base dir: %s", err.Error())
		os.Exit(1)
	}

	cacheDir = filepath.Join(cacheDir, cacheDirName)
	DockerCreds = filepath.Join(cacheDir, DockerCreds)
	OIDCProvider = filepath.Join(cacheDir, OIDCProvider)
	OIDCToken = filepath.Join(cacheDir, OIDCToken)
}
