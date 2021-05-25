package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/payfazz/go-errors/v2"

	"github.com/payfazz/fazz-ecr/config/endpoint"
	"github.com/payfazz/fazz-ecr/pkg/types"
)

func exchageToken(IDToken string) (types.Cred, error) {
	var cred types.Cred

	req, err := http.NewRequest("GET", endpoint.DockerLoginExchange, nil)
	if err != nil {
		return cred, errors.Trace(err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", IDToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return cred, errors.Trace(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return cred, errors.Errorf("exchange endpoint returning http code: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&cred); err != nil {
		return cred, errors.Trace(err)
	}

	if cred.User == "" || cred.Pass == "" {
		return cred, errors.Errorf("username or password is empty from exchange endpoint")
	}

	return cred, nil
}
