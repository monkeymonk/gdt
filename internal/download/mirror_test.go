package download

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResolveURL_PrimaryOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	result := ResolveURL(srv.URL, []string{"http://should-not-be-used.invalid"})
	if result != srv.URL {
		t.Errorf("expected primary %s, got %s", srv.URL, result)
	}
}

func TestResolveURL_PrimaryFailMirrorOK(t *testing.T) {
	mirror := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer mirror.Close()

	result := ResolveURL("http://primary.invalid", []string{mirror.URL})
	if result != mirror.URL {
		t.Errorf("expected mirror %s, got %s", mirror.URL, result)
	}
}

func TestResolveURL_BothFail(t *testing.T) {
	primary := "http://primary.invalid"
	result := ResolveURL(primary, []string{"http://mirror.invalid"})
	if result != primary {
		t.Errorf("expected primary %s, got %s", primary, result)
	}
}
