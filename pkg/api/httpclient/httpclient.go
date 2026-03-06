// Package httpclient provides shared HTTP helpers for Bitbucket API providers.
package httpclient

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is the shared HTTP client used by all providers.
var Client = &http.Client{Timeout: 10 * time.Second}

// DoBasicGet performs a GET request using HTTP Basic Auth (Cloud).
func DoBasicGet(url, username, password string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.SetBasicAuth(username, password)
	req.Header.Set("Accept", "application/json")

	resp, err := Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	return resp, nil
}

// DoBearerGet performs a GET request using a Bearer token (Data Center PAT).
func DoBearerGet(url, token string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	return resp, nil
}

// DoBasicPost sends a JSON POST with HTTP Basic Auth (Cloud).
func DoBasicPost(url, username, password string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.SetBasicAuth(username, password)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	return resp, nil
}

// DoBearerPost sends a JSON POST with Bearer token (Data Center).
func DoBearerPost(url, token string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	return resp, nil
}

// CheckStatus returns an error for non-2xx HTTP status codes.
func CheckStatus(resp *http.Response) error {
	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated:
		return nil
	case http.StatusUnauthorized:
		return fmt.Errorf("invalid token (HTTP 401 Unauthorized) – check that the token is correct and has not expired")
	case http.StatusForbidden:
		return fmt.Errorf("access denied (HTTP 403 Forbidden) – the token may lack the required scopes")
	case http.StatusNotFound:
		return fmt.Errorf("endpoint not found (HTTP 404) – verify the base URL is correct")
	default:
		return fmt.Errorf("unexpected status %d from server", resp.StatusCode)
	}
}
