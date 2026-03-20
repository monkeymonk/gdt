package config

import (
	"testing"
)

func TestDefaultHome(t *testing.T) {
	home := DefaultHome()
	if home == "" {
		t.Fatal("DefaultHome() returned empty string")
	}
}

func TestResolveHomeEnvOverride(t *testing.T) {
	t.Setenv("GDT_HOME", "/custom/path")
	got := ResolveHome()
	if got != "/custom/path" {
		t.Errorf("ResolveHome() = %q, want /custom/path", got)
	}
}

func TestResolveHomeFallback(t *testing.T) {
	t.Setenv("GDT_HOME", "")
	got := ResolveHome()
	if got == "" {
		t.Fatal("ResolveHome() returned empty")
	}
	if got != DefaultHome() {
		t.Errorf("ResolveHome() = %q, want %q", got, DefaultHome())
	}
}
