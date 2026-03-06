package cloud

import (
	"encoding/json"
	"fmt"

	"github.com/loudstil/bb/pkg/api/httpclient"
)

// Workspace is a minimal representation of a Bitbucket Cloud workspace.
type Workspace struct {
	Slug string
	Name string
}

// ListWorkspaces returns all workspaces the authenticated user belongs to.
func ListWorkspaces(email, token string) ([]Workspace, error) {
	var all []Workspace
	url := "https://api.bitbucket.org/2.0/workspaces?pagelen=100"
	for url != "" {
		page, next, err := fetchWorkspacePage(email, token, url)
		if err != nil {
			return nil, err
		}
		all = append(all, page...)
		url = next
	}
	return all, nil
}

type workspacePage struct {
	Values []struct {
		Slug string `json:"slug"`
		Name string `json:"name"`
	} `json:"values"`
	Next string `json:"next"`
}

func fetchWorkspacePage(email, token, url string) ([]Workspace, string, error) {
	resp, err := httpclient.DoBasicGet(url, email, token)
	if err != nil {
		return nil, "", fmt.Errorf("list workspaces: %w", err)
	}
	defer resp.Body.Close()

	if err := httpclient.CheckStatus(resp); err != nil {
		return nil, "", fmt.Errorf("list workspaces: %w", err)
	}

	var page workspacePage
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		return nil, "", fmt.Errorf("list workspaces: decode: %w", err)
	}

	ws := make([]Workspace, 0, len(page.Values))
	for _, v := range page.Values {
		ws = append(ws, Workspace{Slug: v.Slug, Name: v.Name})
	}
	return ws, page.Next, nil
}
