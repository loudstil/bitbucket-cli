// Package git provides utilities for detecting the current git repository's
// remote configuration.
package git

import (
	"fmt"
	"net/url"
	"os/exec"
	"strings"
)

// RepoInfo holds the workspace (project key) and slug parsed from a remote URL.
type RepoInfo struct {
	Workspace string
	Slug      string
}

// Detect runs `git remote get-url origin` in the current directory and
// parses the URL to extract workspace and slug.
func Detect() (*RepoInfo, error) {
	out, err := exec.Command("git", "remote", "get-url", "origin").Output()
	if err != nil {
		return nil, fmt.Errorf("git: could not get remote URL: %w", err)
	}
	rawURL := strings.TrimSpace(string(out))
	return parseRemoteURL(rawURL)
}

// parseRemoteURL handles:
//   - SSH:         git@bitbucket.org:workspace/repo.git
//   - HTTPS Cloud: https://[user@]bitbucket.org/workspace/repo.git
//   - HTTPS DC:    https://host/scm/PROJECT/repo.git
//   - SSH DC alt:  ssh://git@host/PROJECT/repo.git
func parseRemoteURL(rawURL string) (*RepoInfo, error) {
	// SCP-style SSH: git@host:path/to/repo.git (no "://" scheme)
	if !strings.Contains(rawURL, "://") && strings.Contains(rawURL, ":") {
		parts := strings.SplitN(rawURL, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("git: cannot parse remote URL: %q", rawURL)
		}
		segments := strings.Split(strings.TrimPrefix(parts[1], "/"), "/")
		if len(segments) < 2 {
			return nil, fmt.Errorf("git: cannot parse remote URL: %q", rawURL)
		}
		return &RepoInfo{
			Workspace: segments[0],
			Slug:      stripGit(segments[1]),
		}, nil
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("git: cannot parse remote URL %q: %w", rawURL, err)
	}

	// Split path into non-empty segments.
	var parts []string
	for _, s := range strings.Split(strings.TrimPrefix(u.Path, "/"), "/") {
		if s != "" {
			parts = append(parts, s)
		}
	}

	// HTTPS DC: /scm/PROJECT/repo.git
	for i, p := range parts {
		if strings.EqualFold(p, "scm") && i+2 < len(parts) {
			return &RepoInfo{
				Workspace: parts[i+1],
				Slug:      stripGit(parts[i+2]),
			}, nil
		}
	}

	// Standard two-segment path: /workspace/repo.git (Cloud HTTPS or ssh://git@host/PROJECT/repo.git)
	if len(parts) >= 2 {
		return &RepoInfo{
			Workspace: parts[0],
			Slug:      stripGit(parts[1]),
		}, nil
	}

	return nil, fmt.Errorf("git: cannot determine workspace/slug from URL %q", rawURL)
}

func stripGit(s string) string {
	return strings.TrimSuffix(s, ".git")
}
