package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

type TokenManagerInt interface {
	getToken(cnf *Config, url string) (*DIDResponse, error)
	getRefresh(current **DIDResponse, url string) error
}

type TokenManager struct{}

func (tm *TokenManager) getToken(cnf *Config, url string) (*DIDResponse, error) {
	tc := tokenCreate{cnf.Identifier, cnf.GetSecret()}

	requestBody, err := json.Marshal(tc)
	if err != nil {
		slog.Error("Failed to marshal request body")
		return nil, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		slog.Error("Failed to send request")
		slog.Error(err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	slog.Debug("Session server", "HTTP", resp.StatusCode)
	for k, v := range resp.Header {
		slog.Debug("sessionServer: Header", k, v)
	}

	if resp.StatusCode != http.StatusOK {
		slog.Error("Unexpected status code returned", "code", resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			slog.Warn("Unable to read response body")
		}
		slog.Warn("Response body", "content", string(body))

		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var tokenResponse DIDResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		slog.Error("Failed to marshall respond to DIDResponse type")
		return nil, err
	}

	return &tokenResponse, nil
}

func (tm *TokenManager) getRefresh(current **DIDResponse, url string) error {
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
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
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
