package platform

import (
	"runtime"
	"testing"
)

func TestDetect(t *testing.T) {
	p := Detect()
	if p.OS == "" {
		t.Fatal("OS should not be empty")
	}
	if p.Arch == "" {
		t.Fatal("Arch should not be empty")
	}
}

func TestDetectMatchesRuntime(t *testing.T) {
	p := Detect()
	if p.OS != runtime.GOOS {
		t.Errorf("OS = %q, want %q", p.OS, runtime.GOOS)
	}
	if p.Arch != runtime.GOARCH {
		t.Errorf("Arch = %q, want %q", p.Arch, runtime.GOARCH)
	}
}
