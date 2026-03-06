package api

// ProviderType identifies which Bitbucket variant is being used.
type ProviderType string

const (
	ProviderCloud      ProviderType = "cloud"
	ProviderDataCenter ProviderType = "datacenter"
)

// Repository is a provider-agnostic representation of a Bitbucket repository.
type Repository struct {
	// Slug is the URL-friendly name (both Cloud and DC use this).
	Slug string
	// FullName is "workspace/slug" for Cloud, "PROJECT/slug" for DC.
	FullName    string
	Description string
	CloneURL    string
	WebURL      string
	IsPrivate   bool
}

// PullRequest is a provider-agnostic representation of a Bitbucket PR.
type PullRequest struct {
	ID          int
	Title       string
	Description string
	State       string // OPEN, MERGED, DECLINED
	AuthorName  string
	SourceBranch string
	TargetBranch string
	WebURL      string
}

// BitbucketClient is the primary interface for all Bitbucket operations.
// Both the CloudProvider and DataCenterProvider must implement this interface.
//
// All commands in cmd/ interact exclusively with this interface, never with
// provider-specific implementations directly.
type BitbucketClient interface {
	// Metadata
	ProviderType() ProviderType
	BaseURL() string

	// Repository operations
	ListRepositories(workspace string) ([]Repository, error)
	GetRepository(workspace, slug string) (*Repository, error)
	CreateRepository(workspace, slug string, opts CreateRepoRequest) (*Repository, error)

	// Pull Request operations
	// state is one of OPEN, MERGED, DECLINED, or ALL (ALL omits the filter).
	ListPullRequests(workspace, slug, state string) ([]PullRequest, error)
	GetPullRequest(workspace, slug string, id int) (*PullRequest, error)
	CreatePullRequest(workspace, slug string, pr CreatePRRequest) (*PullRequest, error)
	MergePullRequest(workspace, slug string, id int) error
}

// CreateRepoRequest holds the data needed to create a repository.
type CreateRepoRequest struct {
	Description string
	IsPrivate   bool
}

// CreatePRRequest holds the data needed to open a pull request.
type CreatePRRequest struct {
	Title        string
	Description  string
	SourceBranch string
	TargetBranch string
	ReviewerIDs  []string // Account IDs (Cloud) or user slugs (DC)
}
