package metadata

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchLatestRelease_Success(t *testing.T) {
	fixture := githubRelease{
		TagName: "v4.3.0",
		Assets: []githubAsset{
			{Name: "Godot_v4.3.0-stable_linux.x86_64.zip", URL: "https://example.com/godot.zip"},
		},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") != "application/vnd.github+json" {
			t.Errorf("missing Accept header")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(fixture)
	}))
	defer srv.Close()

	rel, err := FetchLatestRelease(srv.URL, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rel.TagName != fixture.TagName {
		t.Errorf("got tag %q, want %q", rel.TagName, fixture.TagName)
	}
	wantName := fixture.Assets[0].Name
	if _, ok := rel.Assets[wantName]; !ok || len(rel.Assets) != 1 {
		t.Errorf("unexpected assets: %+v", rel.Assets)
	}
}

func TestFetchLatestRelease_WithToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer testtoken" {
			t.Errorf("got Authorization %q, want Bearer testtoken", auth)
		}
		json.NewEncoder(w).Encode(githubRelease{TagName: "v4.3.0"})
	}))
	defer srv.Close()

	if _, err := FetchLatestRelease(srv.URL, "testtoken"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFetchLatestRelease_NonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	_, err := FetchLatestRelease(srv.URL, "")
	if err == nil {
		t.Fatal("expected error for 403 response")
	}
}

func TestFetchLatestRelease_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("not-json"))
	}))
	defer srv.Close()

	_, err := FetchLatestRelease(srv.URL, "")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestFetchLatestRelease_BadURL(t *testing.T) {
	_, err := FetchLatestRelease("://bad-url", "")
	if err == nil {
		t.Fatal("expected error for bad URL")
	}
}
