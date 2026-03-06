package httpclient

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDoBasicGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if !ok || u != "user" || p != "pass" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	resp, err := DoBasicGet(srv.URL, "user", "pass")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestDoBearerGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer mytoken" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	resp, err := DoBearerGet(srv.URL, "mytoken")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestDoBasicPost(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		u, p, ok := r.BasicAuth()
		if !ok || u != "user" || p != "pass" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if r.Header.Get("Content-Type") != "application/json" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	resp, err := DoBasicPost(srv.URL, "user", "pass", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusCreated)
	}
}

func TestDoBearerPost(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if r.Header.Get("Authorization") != "Bearer mytoken" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if r.Header.Get("Content-Type") != "application/json" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	resp, err := DoBearerPost(srv.URL, "mytoken", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusCreated)
	}
}

func TestCheckStatus(t *testing.T) {
	tests := []struct {
		status  int
		wantErr bool
	}{
		{http.StatusOK, false},
		{http.StatusCreated, false},
		{http.StatusUnauthorized, true},
		{http.StatusForbidden, true},
		{http.StatusNotFound, true},
		{http.StatusInternalServerError, true},
	}

	for _, tt := range tests {
		resp := &http.Response{
			StatusCode: tt.status,
			Body:       io.NopCloser(strings.NewReader("")),
		}
		err := CheckStatus(resp)
		if tt.wantErr && err == nil {
			t.Errorf("status %d: expected error, got nil", tt.status)
		}
		if !tt.wantErr && err != nil {
			t.Errorf("status %d: unexpected error: %v", tt.status, err)
		}
	}
}
