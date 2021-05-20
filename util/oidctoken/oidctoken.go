package oidctoken

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/payfazz/go-errors/v2"
	"github.com/payfazz/go-handler/v2"
	"github.com/payfazz/go-handler/v2/defresponse"

	oidcconfig "github.com/payfazz/fazz-ecr/config/oidc"
	"github.com/payfazz/fazz-ecr/util/randstring"
)

func GetToken(callback func(string) (string, error)) error {
	if token := os.Getenv("FAZZ_ECR_TOKEN"); token != "" {
		_, err := callback(token)
		return err
	}

	cache := loadTokenCache()

	if v, ok := cache[oidcconfig.Issuer]; ok {
		_, err := callback(v.IDToken)
		return err
	}

	redirect := fmt.Sprintf("http://localhost:%d", oidcconfig.CallbackPort)
	state := randstring.Get(16)

	handled := uint32(0)

	type resp struct {
		text   string
		status int
	}

	codeCh := make(chan string, 1)
	respCh := make(chan resp, 1)

	server := http.Server{
		Addr: fmt.Sprintf("localhost:%d", oidcconfig.CallbackPort),
		Handler: handler.Of(func(r *http.Request) http.HandlerFunc {
			if r.URL.EscapedPath() != "/" {
				return defresponse.Status(404)
			}

			if r.URL.Query().Get("state") != state {
				return defresponse.Text(400, `invalid "state"`)
			}

			code := r.URL.Query().Get("code")
			if code == "" {
				return defresponse.Text(400, `"code" is empty`)
			}

			if !atomic.CompareAndSwapUint32(&handled, 0, 1) {
				return defresponse.Text(400, "cannot call callback multiple time")
			}

			codeCh <- code
			resp := <-respCh

			return defresponse.Text(resp.status, resp.text)
		}),
	}

	serverErrCh := make(chan error, 1)
	go func() { serverErrCh <- errors.Trace(server.ListenAndServe()) }()
	defer server.Shutdown(context.Background())

	// this is to make sure that server is running first
	select {
	case err := <-serverErrCh:
		return err
	case <-time.After(500 * time.Millisecond):
	}

	provider := loadProviderCache()

	auth, err := provider.getAuthUri(oidcconfig.Issuer, oidcconfig.ClientID, redirect, state)
	if err != nil {
		return err
	}

	openBrowser(auth)

	processCode := func(code string) (string, error) {
		token, err := provider.getIDToken(oidcconfig.Issuer, oidcconfig.ClientID, redirect, code)
		if err != nil {
			return "", err
		}

		tokenParts := strings.Split(token, ".")
		if len(tokenParts) < 2 {
			return "", errors.Errorf("invalid token from oidc")
		}

		tokenBodyRaw, err := base64.RawURLEncoding.DecodeString(tokenParts[1])
		if err != nil {
			return "", errors.Trace(err)
		}

		var tokenBody struct {
			Exp int64 `json:"exp"`
		}
		if err := json.Unmarshal(tokenBodyRaw, &tokenBody); err != nil {
			return "", errors.Trace(err)
		}

		cache[oidcconfig.Issuer] = tokenCacheItem{
			IDToken: token,
			Exp:     tokenBody.Exp,
		}
		cache.save()

		res, err := callback(token)
		if err != nil {
			return "", err
		}

		return res, nil
	}

	select {
	case err := <-serverErrCh:
		return err
	case <-time.After(5 * time.Minute):
		return errors.Errorf("timed out after waiting for 5 minutes")
	case code := <-codeCh:
		text, err := processCode(code)
		if err != nil {
			respCh <- resp{text: err.Error(), status: 500}
		} else {
			respCh <- resp{text: text, status: 200}
		}
		return err
	}
}
