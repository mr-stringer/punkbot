package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

func getToken(cnf *Config) (*DIDResponse, error) {
	requestBody, err := json.Marshal(map[string]string{
		"identifier": cnf.Identifier,
		"password":   cnf.GetSecret(),
	})
	if err != nil {
		slog.Error("Failed to marshal request body")
		return nil, err
	}

	url := fmt.Sprintf("%s/%s", ApiUrl, CreateSessionEndpoint)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		slog.Error("Failed to send request")
		slog.Error(err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("Unexpected status code returned", "code", resp.StatusCode)
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var tokenResponse DIDResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		slog.Error("Failed to marshall respond to DIDResponse type")
		return nil, err
	}

	return &tokenResponse, nil
}

func getRefresh(current **DIDResponse) error {
	url := fmt.Sprintf("%s/%s", ApiUrl, RefreshEndpoint)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		slog.Error("Error creating request", "error", err)
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", (*current).RefreshJwt))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("Error sending request", "error", err)
		return err
	}
	if resp.StatusCode != http.StatusOK {
		slog.Error("Unexpected status code", "status", resp)
		return err
	}

	var tokenResponse DIDResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		slog.Error("Failed to marshall respond to DIDResponse type")
		return err
	}

	slog.Debug("Refreshed Access Token", "val", StrHash(tokenResponse.AccessJwt))
	*current = &tokenResponse
	return nil
}
