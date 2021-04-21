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

	"github.com/payfazz/fazz-ecr/config"
	"github.com/payfazz/fazz-ecr/util/randstring"
)

func GetToken(callback func(string) (string, error)) error {
	tokenCache := loadTokenCache()

	if v, ok := tokenCache[config.OIDCIssuer]; ok {
		if _, err := callback(v.IDToken); err != nil {
			return errors.Wrap(err)
		}
		return nil
	}

	providerCache := loadProviderCache()

	errCh := make(chan error, 1)

	redirectURI := fmt.Sprintf("http://localhost:%d", config.OIDCCallbackPort)
	state := randstring.Get(16)

	callbackServer := http.Server{Addr: fmt.Sprintf("localhost:%d", config.OIDCCallbackPort)}

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

		token, err := providerCache.getIDToken(config.OIDCIssuer, config.OIDCClientID, redirectURI, r.URL.Query().Get("code"))
		if err != nil {
			return errResponse(errCh, errors.Wrap(err))
		}

		jwt, err := jose.ParseSigned(token)
		if err != nil {
			return errResponse(errCh, errors.Wrap(err))
		}

		var jwtBody struct {
			Exp int64 `json:"exp"`
		}
		if err := json.Unmarshal(jwt.UnsafePayloadWithoutVerification(), &jwtBody); err != nil {
			return errResponse(errCh, errors.Wrap(err))
		}

		tokenCache[config.OIDCIssuer] = tokenCacheItem{
			IDToken: token,
			Exp:     jwtBody.Exp,
		}
		tokenCache.save()

		res, err := callback(token)
		if err != nil {
			return errResponse(errCh, errors.Wrap(err))
		}

		errCh <- nil
		return defresponse.Text(200, res)
	})

	authURI, err := providerCache.getAuthUri(config.OIDCIssuer, config.OIDCClientID, redirectURI, state)
	if err != nil {
		return errors.Wrap(err)
	}

	go func() {
		if err := callbackServer.ListenAndServe(); err != http.ErrServerClosed {
			errCh <- errors.Wrap(err)
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
