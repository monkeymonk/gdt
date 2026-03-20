package editor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSupportedEditors(t *testing.T) {
	editors := SupportedEditors()
	if len(editors) != 3 {
		t.Fatalf("expected 3 editors, got %d", len(editors))
	}
	want := map[string]bool{"nvim": true, "helix": true, "vscode": true}
	for _, e := range editors {
		if !want[e] {
			t.Errorf("unexpected editor: %s", e)
		}
	}
}

func TestSnippet(t *testing.T) {
	for _, editor := range SupportedEditors() {
		s := Snippet(editor)
		if s == "" {
			t.Errorf("Snippet(%q) returned empty string", editor)
		}
	}
	if s := Snippet("unknown"); s != "" {
		t.Errorf("Snippet(unknown) should be empty, got %q", s)
	}
}

func TestSnippetContainsGdtLsp(t *testing.T) {
	for _, editor := range SupportedEditors() {
		s := Snippet(editor)
		if s == "" {
			continue
		}
		if !containsGdtLsp(s) {
			t.Errorf("Snippet(%q) does not reference 'gdt' and 'lsp'", editor)
		}
	}
}

func containsGdtLsp(s string) bool {
	return contains(s, "gdt") && contains(s, "lsp")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestSetupUnsupported(t *testing.T) {
	err := Setup("emacs", t.TempDir())
	if err == nil {
		t.Fatal("expected error for unsupported editor")
	}
}

func TestSetupVscode(t *testing.T) {
	dir := t.TempDir()
	if err := Setup("vscode", dir); err != nil {
		t.Fatalf("Setup(vscode) failed: %v", err)
	}

	path := filepath.Join(dir, ".vscode", "settings.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("settings.json not created: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("settings.json is empty")
	}

	// Second call should fail (file exists)
	if err := Setup("vscode", dir); err == nil {
		t.Fatal("expected error on second Setup(vscode) call")
	}
}

func TestSetupNonVscode(t *testing.T) {
	// Non-vscode editors just print to stdout, should not error
	dir := t.TempDir()
	if err := Setup("nvim", dir); err != nil {
		t.Fatalf("Setup(nvim) failed: %v", err)
	}
	if err := Setup("helix", dir); err != nil {
		t.Fatalf("Setup(helix) failed: %v", err)
	}
}
