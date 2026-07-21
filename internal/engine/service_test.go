package engine

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/monkeymonk/gdt/internal/config"
	"github.com/monkeymonk/gdt/internal/platform"
)

func TestActionableError(t *testing.T) {
	base := errors.New("boom")
	ae := &ActionableError{Err: base, Suggestion: "do X"}

	if ae.Error() != "boom" {
		t.Errorf("expected Error() to be %q, got %q", "boom", ae.Error())
	}
	if ae.Unwrap() != base {
		t.Errorf("expected Unwrap() to return base error, got %v", ae.Unwrap())
	}
	if !errors.Is(ae, base) {
		t.Error("expected errors.Is(ae, base) to be true")
	}
}

func TestCacheDir(t *testing.T) {
	svc := NewService("/home/x", platform.Info{OS: "linux", Arch: "amd64"}, &config.Config{})
	want := filepath.Join("/home/x", "cache")
	if got := svc.CacheDir(); got != want {
		t.Errorf("CacheDir() = %q, want %q", got, want)
	}
}

func TestCachePath(t *testing.T) {
	svc := NewService("/home/x", platform.Info{OS: "linux", Arch: "amd64"}, &config.Config{})
	want := filepath.Join("/home/x", "cache", "releases.json")
	if got := svc.CachePath(); got != want {
		t.Errorf("CachePath() = %q, want %q", got, want)
	}
}
