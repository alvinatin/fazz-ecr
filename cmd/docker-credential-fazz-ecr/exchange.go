package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/payfazz/go-errors/v2"

	"github.com/payfazz/fazz-ecr/config/endpoint"
	"github.com/payfazz/fazz-ecr/pkg/types"
)

func exchageToken(IDToken string) (types.Cred, error) {
	var cred types.Cred

	retryCount := 0
	for {
		retry, err := func() (bool, error) {
			req, err := http.NewRequest("GET", endpoint.DockerLoginExchange, nil)
			if err != nil {
				return false, errors.Trace(err)
			}
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", IDToken))

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return false, errors.Trace(err)
			}
			defer resp.Body.Close()

			if 500 <= resp.StatusCode && resp.StatusCode < 600 {
				retryCount++
				if retryCount < 5 {
					time.Sleep(10 * time.Second)
					return true, nil
				}
			}

			if resp.StatusCode != 200 {
				return false, errors.Errorf("exchange endpoint returning http code: %d", resp.StatusCode)
			}

			if err := json.NewDecoder(resp.Body).Decode(&cred); err != nil {
				return false, errors.Trace(err)
			}

			if cred.User == "" || cred.Pass == "" {
				return false, errors.Errorf("username or password is empty from exchange endpoint")
			}

			return false, nil
		}()

		if !retry {
			return cred, err
		}
	}
}
