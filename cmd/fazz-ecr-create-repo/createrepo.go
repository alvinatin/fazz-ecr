package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/payfazz/fazz-ecr/config/endpoint"
	"github.com/payfazz/go-errors/v2"
)

func createRepo(IDToken string, repo string) error {
	body, _ := json.Marshal(repo)
	req, err := http.NewRequest("POST", endpoint.CreateRepo, bytes.NewReader(body))
	if err != nil {
		return errors.Trace(err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", IDToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Trace(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		return errors.Errorf("create repo endpoint returning http code: %d", resp.StatusCode)
	}

	return nil
}
