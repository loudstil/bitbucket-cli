package git

import (
	"testing"
)

func TestParseRemoteURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		workspace string
		slug      string
		wantErr   bool
	}{
		{
			name:      "SSH SCP-style Cloud",
			url:       "git@bitbucket.org:myworkspace/myrepo.git",
			workspace: "myworkspace",
			slug:      "myrepo",
		},
		{
			name:      "SSH SCP-style no .git suffix",
			url:       "git@bitbucket.org:myworkspace/myrepo",
			workspace: "myworkspace",
			slug:      "myrepo",
		},
		{
			name:      "HTTPS Cloud",
			url:       "https://bitbucket.org/myworkspace/myrepo.git",
			workspace: "myworkspace",
			slug:      "myrepo",
		},
		{
			name:      "HTTPS Cloud with user",
			url:       "https://user@bitbucket.org/myworkspace/myrepo.git",
			workspace: "myworkspace",
			slug:      "myrepo",
		},
		{
			name:      "HTTPS DC /scm/ path",
			url:       "https://bitbucket.example.com/scm/PROJ/myrepo.git",
			workspace: "PROJ",
			slug:      "myrepo",
		},
		{
			name:      "SSH DC ssh:// scheme",
			url:       "ssh://git@bitbucket.example.com/PROJ/myrepo.git",
			workspace: "PROJ",
			slug:      "myrepo",
		},
		{
			name:    "invalid URL no segments",
			url:     "https://bitbucket.org/",
			wantErr: true,
		},
		{
			name:    "invalid SCP-style no path",
			url:     "git@bitbucket.org:",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseRemoteURL(tt.url)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Workspace != tt.workspace {
				t.Errorf("workspace: got %q, want %q", got.Workspace, tt.workspace)
			}
			if got.Slug != tt.slug {
				t.Errorf("slug: got %q, want %q", got.Slug, tt.slug)
			}
		})
	}
}
