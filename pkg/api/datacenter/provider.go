// Package datacenter implements api.BitbucketClient for Bitbucket Data Center.
package datacenter

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/loudstil/bb/pkg/api"
	"github.com/loudstil/bb/pkg/api/httpclient"
)

// DCProvider implements api.BitbucketClient for Bitbucket Data Center.
type DCProvider struct {
	baseURL string
	token   string
}

// New creates a new DCProvider with the given credentials.
func New(baseURL, token string) *DCProvider {
	return &DCProvider{baseURL: baseURL, token: token}
}

func (d *DCProvider) ProviderType() api.ProviderType { return api.ProviderDataCenter }
func (d *DCProvider) BaseURL() string                { return d.baseURL }

// ListRepositories fetches all accessible repositories using offset pagination.
// The workspace parameter is ignored – DC /rest/api/1.0/repos returns all repos
// the authenticated user can access.
func (d *DCProvider) ListRepositories(_ string) ([]api.Repository, error) {
	var all []api.Repository
	start := 0
	for {
		page, isLast, next, err := d.fetchPage(start)
		if err != nil {
			return nil, err
		}
		all = append(all, page...)
		if isLast {
			break
		}
		start = next
	}
	return all, nil
}

type dcPage struct {
	Values []struct {
		Slug        string `json:"slug"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Public      bool   `json:"public"`
		Project     struct {
			Key string `json:"key"`
		} `json:"project"`
		Links struct {
			Self  []struct{ Href string } `json:"self"`
			Clone []struct {
				Name string `json:"name"`
				Href string `json:"href"`
			} `json:"clone"`
		} `json:"links"`
	} `json:"values"`
	IsLastPage    bool `json:"isLastPage"`
	NextPageStart int  `json:"nextPageStart"`
}

// fetchPage retrieves one page starting at start and returns repos, isLastPage, nextPageStart.
func (d *DCProvider) fetchPage(start int) ([]api.Repository, bool, int, error) {
	url := fmt.Sprintf("%s/rest/api/1.0/repos?limit=100&start=%d", d.baseURL, start)
	resp, err := httpclient.DoBearerGet(url, d.token)
	if err != nil {
		return nil, false, 0, fmt.Errorf("datacenter: list repos: %w", err)
	}
	defer resp.Body.Close()

	if err := httpclient.CheckStatus(resp); err != nil {
		return nil, false, 0, fmt.Errorf("datacenter: list repos: %w", err)
	}

	var page dcPage
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		return nil, false, 0, fmt.Errorf("datacenter: list repos: decode: %w", err)
	}

	repos := make([]api.Repository, 0, len(page.Values))
	for _, v := range page.Values {
		r := api.Repository{
			Slug:        v.Slug,
			FullName:    v.Project.Key + "/" + v.Slug,
			Description: v.Description,
			IsPrivate:   !v.Public,
		}
		if len(v.Links.Self) > 0 {
			r.WebURL = v.Links.Self[0].Href
		}
		for _, cl := range v.Links.Clone {
			if cl.Name == "http" {
				r.CloneURL = cl.Href
				break
			}
		}
		repos = append(repos, r)
	}
	return repos, page.IsLastPage, page.NextPageStart, nil
}

// ListPullRequests fetches all pull requests for the given project/repo.
// state is one of OPEN, MERGED, DECLINED, or ALL (ALL omits the state filter).
func (d *DCProvider) ListPullRequests(workspace, slug, state string) ([]api.PullRequest, error) {
	var all []api.PullRequest
	start := 0
	for {
		page, isLast, next, err := d.fetchPRPage(workspace, slug, state, start)
		if err != nil {
			return nil, err
		}
		all = append(all, page...)
		if isLast {
			break
		}
		start = next
	}
	return all, nil
}

type dcPRPage struct {
	Values []struct {
		ID    int    `json:"id"`
		Title string `json:"title"`
		State string `json:"state"`
		Author struct {
			User struct {
				DisplayName string `json:"displayName"`
			} `json:"user"`
		} `json:"author"`
		FromRef struct {
			DisplayID string `json:"displayId"`
		} `json:"fromRef"`
		ToRef struct {
			DisplayID string `json:"displayId"`
		} `json:"toRef"`
		Links struct {
			Self []struct{ Href string } `json:"self"`
		} `json:"links"`
	} `json:"values"`
	IsLastPage    bool `json:"isLastPage"`
	NextPageStart int  `json:"nextPageStart"`
}

func (d *DCProvider) fetchPRPage(workspace, slug, state string, start int) ([]api.PullRequest, bool, int, error) {
	url := fmt.Sprintf("%s/rest/api/1.0/projects/%s/repos/%s/pull-requests?limit=100&start=%d",
		d.baseURL, workspace, slug, start)
	if state != "ALL" {
		url += "&state=" + state
	}
	resp, err := httpclient.DoBearerGet(url, d.token)
	if err != nil {
		return nil, false, 0, fmt.Errorf("datacenter: list prs: %w", err)
	}
	defer resp.Body.Close()

	if err := httpclient.CheckStatus(resp); err != nil {
		return nil, false, 0, fmt.Errorf("datacenter: list prs: %w", err)
	}

	var page dcPRPage
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		return nil, false, 0, fmt.Errorf("datacenter: list prs: decode: %w", err)
	}

	prs := make([]api.PullRequest, 0, len(page.Values))
	for _, v := range page.Values {
		pr := api.PullRequest{
			ID:           v.ID,
			Title:        v.Title,
			State:        v.State,
			AuthorName:   v.Author.User.DisplayName,
			SourceBranch: v.FromRef.DisplayID,
			TargetBranch: v.ToRef.DisplayID,
		}
		if len(v.Links.Self) > 0 {
			pr.WebURL = v.Links.Self[0].Href
		}
		prs = append(prs, pr)
	}
	return prs, page.IsLastPage, page.NextPageStart, nil
}

// dcRepo mirrors the JSON shape of a single repository response from DC.
type dcRepo struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Public      bool   `json:"public"`
	Project     struct {
		Key string `json:"key"`
	} `json:"project"`
	Links struct {
		Self  []struct{ Href string } `json:"self"`
		Clone []struct {
			Name string `json:"name"`
			Href string `json:"href"`
		} `json:"clone"`
	} `json:"links"`
}

func mapDCRepo(v dcRepo) api.Repository {
	r := api.Repository{
		Slug:        v.Slug,
		FullName:    v.Project.Key + "/" + v.Slug,
		Description: v.Description,
		IsPrivate:   !v.Public,
	}
	if len(v.Links.Self) > 0 {
		r.WebURL = v.Links.Self[0].Href
	}
	for _, cl := range v.Links.Clone {
		if cl.Name == "http" {
			r.CloneURL = cl.Href
			break
		}
	}
	return r
}

// GetRepository fetches a single repository by project key and slug.
func (d *DCProvider) GetRepository(workspace, slug string) (*api.Repository, error) {
	url := fmt.Sprintf("%s/rest/api/1.0/projects/%s/repos/%s", d.baseURL, workspace, slug)
	resp, err := httpclient.DoBearerGet(url, d.token)
	if err != nil {
		return nil, fmt.Errorf("datacenter: get repo: %w", err)
	}
	defer resp.Body.Close()

	if err := httpclient.CheckStatus(resp); err != nil {
		return nil, fmt.Errorf("datacenter: get repo: %w", err)
	}

	var v dcRepo
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, fmt.Errorf("datacenter: get repo: decode: %w", err)
	}
	r := mapDCRepo(v)
	return &r, nil
}

// CreateRepository creates a new repository under the given project key.
// IsPrivate is silently ignored for Data Center (access is project-level).
func (d *DCProvider) CreateRepository(workspace, slug string, opts api.CreateRepoRequest) (*api.Repository, error) {
	body, err := json.Marshal(map[string]interface{}{
		"name":        slug,
		"scmId":       "git",
		"description": opts.Description,
	})
	if err != nil {
		return nil, fmt.Errorf("datacenter: create repo: encode body: %w", err)
	}

	url := fmt.Sprintf("%s/rest/api/1.0/projects/%s/repos", d.baseURL, workspace)
	resp, err := httpclient.DoBearerPost(url, d.token, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("datacenter: create repo: %w", err)
	}
	defer resp.Body.Close()

	if err := httpclient.CheckStatus(resp); err != nil {
		return nil, fmt.Errorf("datacenter: create repo: %w", err)
	}

	var v dcRepo
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, fmt.Errorf("datacenter: create repo: decode: %w", err)
	}
	r := mapDCRepo(v)
	return &r, nil
}
func (d *DCProvider) GetPullRequest(_, _ string, _ int) (*api.PullRequest, error) {
	return nil, fmt.Errorf("not implemented")
}
func (d *DCProvider) CreatePullRequest(_, _ string, _ api.CreatePRRequest) (*api.PullRequest, error) {
	return nil, fmt.Errorf("not implemented")
}
func (d *DCProvider) MergePullRequest(_, _ string, _ int) error {
	return fmt.Errorf("not implemented")
}
