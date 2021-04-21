package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/payfazz/go-errors"

	"github.com/payfazz/fazz-ecr/config"
	"github.com/payfazz/fazz-ecr/pkg/types"
)

func exchageToken(IDToken string) (types.Cred, error) {
	var cred types.Cred

	req, err := http.NewRequest("GET", config.IDTokenExchangeEndpoint, nil)
	if err != nil {
		return types.Cred{}, errors.Wrap(err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", IDToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return types.Cred{}, errors.Wrap(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return types.Cred{}, errors.Errorf("exchange endpoint not returning 200")
	}

	if err := json.NewDecoder(resp.Body).Decode(&cred); err != nil {
		return types.Cred{}, errors.Wrap(err)
	}

	if cred.User == "" || cred.Pass == "" {
		return types.Cred{}, errors.Errorf("username or password is empty from exchange endpoint")
	}

	return cred, nil
}
