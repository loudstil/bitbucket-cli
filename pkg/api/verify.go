package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// VerifiedUser holds the identity returned by a successful credential check.
type VerifiedUser struct {
	Username    string
	DisplayName string
}

// VerifyCredentials makes a live API call to confirm the credentials are valid.
// It returns the authenticated user's identity so the caller can display it.
//
// Cloud:       GET https://api.bitbucket.org/2.0/user          (Basic auth: email:token)
// Data Center: GET <baseURL>/rest/api/1.0/profile/recent/repos  (Bearer token / PAT)
func VerifyCredentials(providerType ProviderType, baseURL, email, token string) (*VerifiedUser, error) {
	switch providerType {
	case ProviderCloud:
		return verifyCloud(email, token)
	case ProviderDataCenter:
		return verifyDC(baseURL, token)
	default:
		return nil, fmt.Errorf("verify: unknown provider type %q", providerType)
	}
}

// verifyCloud calls the /2.0/user endpoint using HTTP Basic Auth (email:token).
// Bitbucket Cloud API tokens are not Bearer tokens – they must be paired with
// the account email as Basic Auth credentials.
func verifyCloud(email, token string) (*VerifiedUser, error) {
	endpoint := "https://api.bitbucket.org/2.0/user"
	resp, err := doBasicGet(endpoint, email, token)
	if err != nil {
		return nil, fmt.Errorf("verify cloud: %w", err)
	}
	defer resp.Body.Close()

	if err := checkStatus(resp); err != nil {
		return nil, fmt.Errorf("verify cloud: %w", err)
	}

	var payload struct {
		Nickname    string `json:"nickname"`
		DisplayName string `json:"display_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("verify cloud: decode response: %w", err)
	}

	return &VerifiedUser{
		Username:    payload.Nickname,
		DisplayName: payload.DisplayName,
	}, nil
}

// verifyDC calls a lightweight authenticated endpoint on a Data Center instance.
// /rest/api/1.0/profile/recent/repos returns 401 for invalid tokens and a
// (possibly empty) list for valid ones, making it a reliable auth probe.
func verifyDC(baseURL, token string) (*VerifiedUser, error) {
	endpoint := baseURL + "/rest/api/1.0/profile/recent/repos?limit=1"
	resp, err := doBearerGet(endpoint, token)
	if err != nil {
		return nil, fmt.Errorf("verify datacenter: %w", err)
	}
	defer resp.Body.Close()

	if err := checkStatus(resp); err != nil {
		return nil, fmt.Errorf("verify datacenter: %w", err)
	}

	// DC does not return user info from this endpoint.
	// Return a placeholder; the username entered by the user is kept as-is.
	return &VerifiedUser{}, nil
}

var httpClient = &http.Client{Timeout: 10 * time.Second}

// doBasicGet performs a GET request using HTTP Basic Auth (Cloud).
func doBasicGet(url, username, password string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.SetBasicAuth(username, password)
	req.Header.Set("Accept", "application/json")
	
	

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	return resp, nil
}

// doBearerGet performs a GET request using a Bearer token (Data Center PAT).
func doBearerGet(url, token string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	return resp, nil
}

func checkStatus(resp *http.Response) error {
	switch resp.StatusCode {
	case http.StatusOK:
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
