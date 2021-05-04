package oidctoken

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/payfazz/go-errors/v2"
	"github.com/payfazz/go-handler"
	"github.com/payfazz/go-handler/defresponse"

	"github.com/payfazz/fazz-ecr/config"
	"github.com/payfazz/fazz-ecr/util/randstring"
)

func GetToken(callback func(string) (string, error)) error {
	cache := loadTokenCache()

	if v, ok := cache[config.OIDCIssuer]; ok {
		if _, err := callback(v.IDToken); err != nil {
			return errors.Wrap(err)
		}
		return nil
	}

	provider := loadProviderCache()

	redirect := fmt.Sprintf("http://localhost:%d", config.OIDCCallbackPort)
	state := randstring.Get(16)

	server := http.Server{Addr: fmt.Sprintf("localhost:%d", config.OIDCCallbackPort)}

	done := uint32(0)

	var shutdownOnce sync.Once
	shutdown := func() { shutdownOnce.Do(func() { server.Shutdown(context.Background()) }) }

	errCh := make(chan error, 1)

	server.Handler = handler.Of(func(r *http.Request) handler.Response {
		if r.URL.EscapedPath() != "/" {
			return defresponse.Status(404)
		}

		if r.URL.Query().Get("state") != state {
			return defresponse.Text(400, `invalid "state"`)
		}

		if !atomic.CompareAndSwapUint32(&done, 0, 1) {
			return defresponse.Text(400, "cannot call callback multiple time")
		}

		go shutdown()

		token, err := provider.getIDToken(config.OIDCIssuer, config.OIDCClientID, redirect, r.URL.Query().Get("code"))
		if err != nil {
			return errResponse(errCh, errors.Wrap(err))
		}

		tokenParts := strings.Split(token, ".")
		if len(tokenParts) < 2 {
			return errResponse(errCh, errors.Errorf("invalid token from oidc"))
		}

		tokenBodyRaw, err := base64.RawURLEncoding.DecodeString(tokenParts[1])
		if err != nil {
			return errResponse(errCh, errors.Wrap(err))
		}

		var tokenBody struct {
			Exp int64 `json:"exp"`
		}
		if err := json.Unmarshal(tokenBodyRaw, &tokenBody); err != nil {
			return errResponse(errCh, errors.Wrap(err))
		}

		cache[config.OIDCIssuer] = tokenCacheItem{
			IDToken: token,
			Exp:     tokenBody.Exp,
		}
		cache.save()

		res, err := callback(token)
		if err != nil {
			return errResponse(errCh, errors.Wrap(err))
		}

		errCh <- nil
		return defresponse.Text(200, res)
	})

	auth, err := provider.getAuthUri(config.OIDCIssuer, config.OIDCClientID, redirect, state)
	if err != nil {
		return errors.Wrap(err)
	}

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			errCh <- errors.Wrap(err)
		}
	}()

	// if no error after 100ms, open browser
	// this is to make sure that server is running first
	select {
	case err = <-errCh:
	case <-time.After(100 * time.Millisecond):
		openBrowser(auth)
	}

	if err == nil {
		select {
		case err = <-errCh:
		case <-time.After(5 * time.Minute):
			err = errors.Errorf("timed out after waiting for 5 minutes")
		}
	}

	shutdown()
	return err
}
