package datacenter

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/loudstil/bb/pkg/api"
)

func newTestProvider(srv *httptest.Server) *DCProvider {
	return New(srv.URL, "user", "tok")
}

func TestDCListRepositories(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"values": []map[string]interface{}{
				{
					"slug":        "my-repo",
					"name":        "My Repo",
					"description": "A DC repo",
					"public":      false,
					"project":     map[string]string{"key": "PROJ"},
					"links": map[string]interface{}{
						"self":  []map[string]string{{"href": "https://bitbucket.example.com/projects/PROJ/repos/my-repo"}},
						"clone": []map[string]string{{"name": "http", "href": "https://bitbucket.example.com/scm/PROJ/my-repo.git"}},
					},
				},
			},
			"isLastPage":    true,
			"nextPageStart": 0,
		})
	}))
	defer srv.Close()

	repos, err := newTestProvider(srv).ListRepositories("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repos) != 1 {
		t.Fatalf("got %d repos, want 1", len(repos))
	}
	r := repos[0]
	if r.Slug != "my-repo" {
		t.Errorf("slug: got %q, want %q", r.Slug, "my-repo")
	}
	if r.FullName != "PROJ/my-repo" {
		t.Errorf("full_name: got %q, want %q", r.FullName, "PROJ/my-repo")
	}
	if !r.IsPrivate {
		t.Error("expected IsPrivate=true (public=false)")
	}
	if r.CloneURL != "https://bitbucket.example.com/scm/PROJ/my-repo.git" {
		t.Errorf("clone URL: got %q", r.CloneURL)
	}
}

func TestDCGetRepository(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/api/1.0/projects/PROJ/repos/my-repo" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"slug":        "my-repo",
			"name":        "My Repo",
			"description": "desc",
			"public":      true,
			"project":     map[string]string{"key": "PROJ"},
			"links": map[string]interface{}{
				"self":  []map[string]string{{"href": "https://bitbucket.example.com/projects/PROJ/repos/my-repo"}},
				"clone": []map[string]string{{"name": "http", "href": "https://bitbucket.example.com/scm/PROJ/my-repo.git"}},
			},
		})
	}))
	defer srv.Close()

	repo, err := newTestProvider(srv).GetRepository("PROJ", "my-repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.Slug != "my-repo" {
		t.Errorf("slug: got %q, want %q", repo.Slug, "my-repo")
	}
	if repo.IsPrivate {
		t.Error("expected IsPrivate=false (public=true)")
	}
}

func TestDCGetRepository_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	_, err := newTestProvider(srv).GetRepository("PROJ", "missing")
	if err == nil {
		t.Fatal("expected error for 404, got nil")
	}
}

func TestDCCreateRepository(t *testing.T) {
	var gotBody map[string]interface{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/rest/api/1.0/projects/PROJ/repos" {
			http.NotFound(w, r)
			return
		}
		json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"slug":        "new-repo",
			"name":        "new-repo",
			"description": "dc repo",
			"public":      true,
			"project":     map[string]string{"key": "PROJ"},
			"links": map[string]interface{}{
				"self":  []map[string]string{{"href": "https://bitbucket.example.com/projects/PROJ/repos/new-repo"}},
				"clone": []map[string]string{},
			},
		})
	}))
	defer srv.Close()

	repo, err := newTestProvider(srv).CreateRepository("PROJ", "new-repo", api.CreateRepoRequest{
		Description: "dc repo",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.Slug != "new-repo" {
		t.Errorf("slug: got %q, want %q", repo.Slug, "new-repo")
	}
	if gotBody["name"] != "new-repo" {
		t.Errorf("body name: got %v, want new-repo", gotBody["name"])
	}
	if gotBody["scmId"] != "git" {
		t.Errorf("body scmId: got %v, want git", gotBody["scmId"])
	}
}
