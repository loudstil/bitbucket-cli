package cloud

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/loudstil/bb/pkg/api"
)

func newTestProvider(srv *httptest.Server) *CloudProvider {
	return &CloudProvider{base: srv.URL, email: "test@example.com", token: "tok"}
}

func TestCloudListRepositories(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"values": []map[string]interface{}{
				{
					"slug":        "my-repo",
					"full_name":   "myws/my-repo",
					"description": "A test repo",
					"is_private":  true,
					"links": map[string]interface{}{
						"html":  map[string]string{"href": "https://bitbucket.org/myws/my-repo"},
						"clone": []map[string]string{{"name": "https", "href": "https://bitbucket.org/myws/my-repo.git"}},
					},
				},
			},
			"next": "",
		})
	}))
	defer srv.Close()

	repos, err := newTestProvider(srv).ListRepositories("myws")
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
	if r.FullName != "myws/my-repo" {
		t.Errorf("full_name: got %q, want %q", r.FullName, "myws/my-repo")
	}
	if !r.IsPrivate {
		t.Error("expected IsPrivate=true")
	}
	if r.CloneURL != "https://bitbucket.org/myws/my-repo.git" {
		t.Errorf("clone URL: got %q", r.CloneURL)
	}
}

func TestCloudGetRepository(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2.0/repositories/myws/my-repo" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"slug":        "my-repo",
			"full_name":   "myws/my-repo",
			"description": "desc",
			"is_private":  false,
			"links": map[string]interface{}{
				"html":  map[string]string{"href": "https://bitbucket.org/myws/my-repo"},
				"clone": []map[string]string{{"name": "https", "href": "https://bitbucket.org/myws/my-repo.git"}},
			},
		})
	}))
	defer srv.Close()

	repo, err := newTestProvider(srv).GetRepository("myws", "my-repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.Slug != "my-repo" {
		t.Errorf("slug: got %q, want %q", repo.Slug, "my-repo")
	}
}

func TestCloudGetRepository_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	_, err := newTestProvider(srv).GetRepository("myws", "missing")
	if err == nil {
		t.Fatal("expected error for 404, got nil")
	}
}

func TestCloudCreateRepository(t *testing.T) {
	var gotBody map[string]interface{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"slug":        "new-repo",
			"full_name":   "myws/new-repo",
			"description": "my desc",
			"is_private":  true,
			"links": map[string]interface{}{
				"html":  map[string]string{"href": "https://bitbucket.org/myws/new-repo"},
				"clone": []map[string]string{},
			},
		})
	}))
	defer srv.Close()

	repo, err := newTestProvider(srv).CreateRepository("myws", "new-repo", api.CreateRepoRequest{
		Description: "my desc",
		IsPrivate:   true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.Slug != "new-repo" {
		t.Errorf("slug: got %q, want %q", repo.Slug, "new-repo")
	}
	if gotBody["is_private"] != true {
		t.Errorf("body is_private: got %v, want true", gotBody["is_private"])
	}
	if gotBody["scm"] != "git" {
		t.Errorf("body scm: got %v, want git", gotBody["scm"])
	}
}
