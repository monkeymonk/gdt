package plugins

import "testing"

func TestExtractRepoSlug(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"monkeymonk/gdt-plugin-assets", "monkeymonk/gdt-plugin-assets"},
		{"https://github.com/monkeymonk/gdt-plugin-assets", "monkeymonk/gdt-plugin-assets"},
		{"https://github.com/monkeymonk/gdt-plugin-assets.git", "monkeymonk/gdt-plugin-assets"},
		{"https://github.com/monkeymonk/gdt-plugin-assets/", "monkeymonk/gdt-plugin-assets"},
		{"http://github.com/foo/bar", "foo/bar"},
		{"not-a-slug", ""},
	}

	for _, tt := range tests {
		got := extractRepoSlug(tt.input)
		if got != tt.want {
			t.Errorf("extractRepoSlug(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestBuildAssetCandidates(t *testing.T) {
	candidates := buildAssetCandidates("assets", "linux", "amd64")
	found := false
	for _, c := range candidates {
		if c == "assets-linux-amd64" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected assets-linux-amd64 in candidates, got %v", candidates)
	}

	candidates = buildAssetCandidates("assets", "darwin", "arm64")
	foundMacos := false
	foundDarwin := false
	for _, c := range candidates {
		if c == "assets-macos-arm64" {
			foundMacos = true
		}
		if c == "assets-darwin-aarch64" {
			foundDarwin = true
		}
	}
	if !foundMacos {
		t.Errorf("expected assets-macos-arm64 in candidates, got %v", candidates)
	}
	if !foundDarwin {
		t.Errorf("expected assets-darwin-aarch64 in candidates, got %v", candidates)
	}
}

func TestDetectRepoSlug(t *testing.T) {
	// detectRepoSlug on a non-git directory should return empty
	slug := detectRepoSlug(t.TempDir())
	if slug != "" {
		t.Errorf("expected empty slug for non-git dir, got %q", slug)
	}
}
