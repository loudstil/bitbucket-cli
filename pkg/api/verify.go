package api

import (
	"encoding/json"
	"fmt"

	"github.com/loudstil/bb/pkg/api/httpclient"
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
	resp, err := httpclient.DoBasicGet(endpoint, email, token)
	if err != nil {
		return nil, fmt.Errorf("verify cloud: %w", err)
	}
	defer resp.Body.Close()

	if err := httpclient.CheckStatus(resp); err != nil {
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
	resp, err := httpclient.DoBearerGet(endpoint, token)
	if err != nil {
		return nil, fmt.Errorf("verify datacenter: %w", err)
	}
	defer resp.Body.Close()

	if err := httpclient.CheckStatus(resp); err != nil {
		return nil, fmt.Errorf("verify datacenter: %w", err)
	}

	// DC does not return user info from this endpoint.
	// Return a placeholder; the username entered by the user is kept as-is.
	return &VerifiedUser{}, nil
}
