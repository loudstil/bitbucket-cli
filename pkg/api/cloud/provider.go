// Package cloud implements api.BitbucketClient for Bitbucket Cloud.
package cloud

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/loudstil/bb/pkg/api"
	"github.com/loudstil/bb/pkg/api/httpclient"
)

const baseAPI = "https://api.bitbucket.org"

// CloudProvider implements api.BitbucketClient for Bitbucket Cloud.
type CloudProvider struct {
	email string
	token string
}

// New creates a new CloudProvider with the given credentials.
func New(email, token string) *CloudProvider {
	return &CloudProvider{email: email, token: token}
}

func (c *CloudProvider) ProviderType() api.ProviderType { return api.ProviderCloud }
func (c *CloudProvider) BaseURL() string                { return baseAPI }

// ListRepositories fetches all repositories in the given workspace using
// cursor-based ("next" URL) pagination.
func (c *CloudProvider) ListRepositories(workspace string) ([]api.Repository, error) {
	var all []api.Repository
	url := fmt.Sprintf("%s/2.0/repositories/%s?pagelen=100", baseAPI, workspace)
	for url != "" {
		page, next, err := c.fetchPage(url)
		if err != nil {
			return nil, err
		}
		all = append(all, page...)
		url = next
	}
	return all, nil
}

type cloudPage struct {
	Values []struct {
		Slug        string `json:"slug"`
		FullName    string `json:"full_name"`
		Description string `json:"description"`
		IsPrivate   bool   `json:"is_private"`
		Links       struct {
			HTML  struct{ Href string } `json:"html"`
			Clone []struct {
				Name string `json:"name"`
				Href string `json:"href"`
			} `json:"clone"`
		} `json:"links"`
	} `json:"values"`
	Next string `json:"next"`
}

// fetchPage retrieves one page and returns repos plus the next-page URL (empty = done).
func (c *CloudProvider) fetchPage(url string) ([]api.Repository, string, error) {
	resp, err := httpclient.DoBasicGet(url, c.email, c.token)
	if err != nil {
		return nil, "", fmt.Errorf("cloud: list repos: %w", err)
	}
	defer resp.Body.Close()

	if err := httpclient.CheckStatus(resp); err != nil {
		return nil, "", fmt.Errorf("cloud: list repos: %w", err)
	}

	var page cloudPage
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		return nil, "", fmt.Errorf("cloud: list repos: decode: %w", err)
	}

	repos := make([]api.Repository, 0, len(page.Values))
	for _, v := range page.Values {
		r := api.Repository{
			Slug:        v.Slug,
			FullName:    v.FullName,
			Description: v.Description,
			IsPrivate:   v.IsPrivate,
			WebURL:      v.Links.HTML.Href,
		}
		for _, cl := range v.Links.Clone {
			if cl.Name == "https" {
				r.CloneURL = cl.Href
				break
			}
		}
		repos = append(repos, r)
	}
	return repos, page.Next, nil
}

// ListPullRequests fetches all pull requests for the given workspace/slug.
// state is one of OPEN, MERGED, DECLINED, or ALL (ALL omits the state filter).
func (c *CloudProvider) ListPullRequests(workspace, slug, state string) ([]api.PullRequest, error) {
	var all []api.PullRequest
	url := fmt.Sprintf("%s/2.0/repositories/%s/%s/pullrequests?pagelen=100", baseAPI, workspace, slug)
	if state != "ALL" {
		url += "&state=" + state
	}
	for url != "" {
		page, next, err := c.fetchPRPage(url)
		if err != nil {
			return nil, err
		}
		all = append(all, page...)
		url = next
	}
	return all, nil
}

type cloudPRPage struct {
	Values []struct {
		ID    int    `json:"id"`
		Title string `json:"title"`
		State string `json:"state"`
		Author struct {
			DisplayName string `json:"display_name"`
		} `json:"author"`
		Source struct {
			Branch struct {
				Name string `json:"name"`
			} `json:"branch"`
		} `json:"source"`
		Destination struct {
			Branch struct {
				Name string `json:"name"`
			} `json:"branch"`
		} `json:"destination"`
		Links struct {
			HTML struct {
				Href string `json:"href"`
			} `json:"html"`
		} `json:"links"`
	} `json:"values"`
	Next string `json:"next"`
}

func (c *CloudProvider) fetchPRPage(url string) ([]api.PullRequest, string, error) {
	resp, err := httpclient.DoBasicGet(url, c.email, c.token)
	if err != nil {
		return nil, "", fmt.Errorf("cloud: list prs: %w", err)
	}
	defer resp.Body.Close()

	if err := httpclient.CheckStatus(resp); err != nil {
		return nil, "", fmt.Errorf("cloud: list prs: %w", err)
	}

	var page cloudPRPage
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		return nil, "", fmt.Errorf("cloud: list prs: decode: %w", err)
	}

	prs := make([]api.PullRequest, 0, len(page.Values))
	for _, v := range page.Values {
		prs = append(prs, api.PullRequest{
			ID:           v.ID,
			Title:        v.Title,
			State:        v.State,
			AuthorName:   v.Author.DisplayName,
			SourceBranch: v.Source.Branch.Name,
			TargetBranch: v.Destination.Branch.Name,
			WebURL:       v.Links.HTML.Href,
		})
	}
	return prs, page.Next, nil
}

// cloudRepo mirrors the JSON shape of a single repository response from Cloud.
type cloudRepo struct {
	Slug        string `json:"slug"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	IsPrivate   bool   `json:"is_private"`
	Links       struct {
		HTML  struct{ Href string } `json:"html"`
		Clone []struct {
			Name string `json:"name"`
			Href string `json:"href"`
		} `json:"clone"`
	} `json:"links"`
}

func mapCloudRepo(v cloudRepo) api.Repository {
	r := api.Repository{
		Slug:        v.Slug,
		FullName:    v.FullName,
		Description: v.Description,
		IsPrivate:   v.IsPrivate,
		WebURL:      v.Links.HTML.Href,
	}
	for _, cl := range v.Links.Clone {
		if cl.Name == "https" {
			r.CloneURL = cl.Href
			break
		}
	}
	return r
}

// GetRepository fetches a single repository by workspace and slug.
func (c *CloudProvider) GetRepository(workspace, slug string) (*api.Repository, error) {
	url := fmt.Sprintf("%s/2.0/repositories/%s/%s", baseAPI, workspace, slug)
	resp, err := httpclient.DoBasicGet(url, c.email, c.token)
	if err != nil {
		return nil, fmt.Errorf("cloud: get repo: %w", err)
	}
	defer resp.Body.Close()

	if err := httpclient.CheckStatus(resp); err != nil {
		return nil, fmt.Errorf("cloud: get repo: %w", err)
	}

	var v cloudRepo
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, fmt.Errorf("cloud: get repo: decode: %w", err)
	}
	r := mapCloudRepo(v)
	return &r, nil
}

// CreateRepository creates a new repository in the given workspace.
func (c *CloudProvider) CreateRepository(workspace, slug string, opts api.CreateRepoRequest) (*api.Repository, error) {
	body, err := json.Marshal(map[string]interface{}{
		"scm":         "git",
		"is_private":  opts.IsPrivate,
		"description": opts.Description,
	})
	if err != nil {
		return nil, fmt.Errorf("cloud: create repo: encode body: %w", err)
	}

	url := fmt.Sprintf("%s/2.0/repositories/%s/%s", baseAPI, workspace, slug)
	resp, err := httpclient.DoBasicPost(url, c.email, c.token, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("cloud: create repo: %w", err)
	}
	defer resp.Body.Close()

	if err := httpclient.CheckStatus(resp); err != nil {
		return nil, fmt.Errorf("cloud: create repo: %w", err)
	}

	var v cloudRepo
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, fmt.Errorf("cloud: create repo: decode: %w", err)
	}
	r := mapCloudRepo(v)
	return &r, nil
}
func (c *CloudProvider) GetPullRequest(_, _ string, _ int) (*api.PullRequest, error) {
	return nil, fmt.Errorf("not implemented")
}
func (c *CloudProvider) CreatePullRequest(_, _ string, _ api.CreatePRRequest) (*api.PullRequest, error) {
	return nil, fmt.Errorf("not implemented")
}
func (c *CloudProvider) MergePullRequest(_, _ string, _ int) error {
	return fmt.Errorf("not implemented")
}
