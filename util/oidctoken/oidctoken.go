package oidctoken

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/payfazz/go-errors"
	"github.com/payfazz/go-handler"
	"github.com/payfazz/go-handler/defresponse"
	"gopkg.in/square/go-jose.v2"

	"github.com/payfazz/fazz-ecr/util/randstring"
)

var (
	oidcIssuer       = "https://dex.fazzfinancial.com"
	oidcClientID     = "ecrhelper"
	oidcCallbackPort = 3000
)

func GetToken(callback func(string) (string, error)) error {
	tokenCache := loadTokenCache()

	if v, ok := tokenCache[oidcIssuer]; ok {
		callback(v.IDToken)
		return nil
	}

	providerCache := loadProviderCache()

	errCh := make(chan error)

	redirectURI := fmt.Sprintf("http://localhost:%d", oidcCallbackPort)
	state := randstring.Get(16)

	callbackServer := http.Server{Addr: fmt.Sprintf("localhost:%d", oidcCallbackPort)}

	handled := uint32(0)

	var shutdownOnce sync.Once
	shutdown := func() {
		shutdownOnce.Do(func() { callbackServer.Shutdown(context.Background()) })
	}

	callbackServer.Handler = handler.Of(func(r *http.Request) handler.Response {
		if r.URL.EscapedPath() != "/" {
			return defresponse.Status(404)
		}

		if r.URL.Query().Get("state") != state {
			return defresponse.Text(400, "invalid \"state\"")
		}

		if !atomic.CompareAndSwapUint32(&handled, 0, 1) {
			return defresponse.Text(400, "cannot call callback multiple time")
		}

		go shutdown()

		token, err := providerCache.getIDToken(oidcIssuer, oidcClientID, redirectURI, r.URL.Query().Get("code"))
		if err != nil {
			errCh <- errors.Wrap(err)
			return defresponse.Status(500)
		}

		jwt, err := jose.ParseSigned(token)
		if err != nil {
			errCh <- errors.Wrap(err)
			return defresponse.Status(500)
		}

		var jwtBody struct {
			Exp int64 `json:"exp"`
		}
		if err := json.Unmarshal(jwt.UnsafePayloadWithoutVerification(), &jwtBody); err != nil {
			errCh <- errors.Wrap(err)
			return defresponse.Status(500)
		}

		tokenCache[oidcIssuer] = tokenCacheItem{
			IDToken: token,
			Exp:     jwtBody.Exp,
		}
		tokenCache.save()

		res, err := callback(token)
		if err != nil {
			errCh <- errors.Wrap(err)
			return defresponse.Status(500)
		}

		return defresponse.Text(200, res)
	})

	authURI, err := providerCache.getAuthUri(oidcIssuer, oidcClientID, redirectURI, state)
	if err != nil {
		return errors.Wrap(err)
	}

	go func() {
		if err := callbackServer.ListenAndServe(); err != http.ErrServerClosed {
			errCh <- errors.Wrap(err)
		} else {
			errCh <- nil
		}
	}()

	select {
	case err = <-errCh:
	case <-time.After(100 * time.Millisecond):
		openBrowser(authURI)
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
