package postoffice

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/mr-stringer/punkbot/config"
	"github.com/mr-stringer/punkbot/global"
)

func getToken(cnf *config.Config) (*global.DIDResponse, error) {
	requestBody, err := json.Marshal(map[string]string{
		"identifier": cnf.Identifier,
		"password":   cnf.GetSecret(),
	})
	if err != nil {
		slog.Error("Failed to marshal request body")
		return nil, err
	}

	url := fmt.Sprintf("%s/%s", global.ApiUrl, global.CreateSessionEndpoint)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		slog.Error("Failed to send request")
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("Unexpected status code returned", "code", resp.StatusCode)
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var tokenResponse global.DIDResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		slog.Error("Failed to marshall respond to DIDResponse type")
		return nil, err
	}

	return &tokenResponse, nil
}

func getRefresh(current *global.DIDResponse) (*global.DIDResponse, error) {
	url := fmt.Sprintf("%s/%s", global.ApiUrl, global.RefreshEndpoint)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		slog.Error("Error creating request", "error", err)
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", current.RefreshJwt))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("Error sending request", "error", err)
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		slog.Error("Unexpected status code", "status", resp)
		return nil, err
	}

	var tokenResponse global.DIDResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		slog.Error("Failed to marshall respond to DIDResponse type")
		return nil, err
	}

	slog.Debug("Refreshed Access Token", "newToken", tokenResponse.AccessJwt)
	return &tokenResponse, nil
}
