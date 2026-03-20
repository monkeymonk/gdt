package ci

import (
	"strings"
	"testing"
)

func TestGenerateGitHub(t *testing.T) {
	content := GenerateGitHub()
	if !strings.Contains(content, "actions/checkout") {
		t.Error("should contain checkout action")
	}
	if !strings.Contains(content, "gdt install") {
		t.Error("should contain gdt install")
	}
	if !strings.Contains(content, "gdt export") {
		t.Error("should contain gdt export")
	}
}

func TestGenerateGitLab(t *testing.T) {
	content := GenerateGitLab()
	if !strings.Contains(content, "gdt install") {
		t.Error("should contain gdt install")
	}
	if !strings.Contains(content, "gdt export") {
		t.Error("should contain gdt export")
	}
	if !strings.Contains(content, "artifacts") {
		t.Error("should contain artifacts section")
	}
}

func TestGenerateGeneric(t *testing.T) {
	content := GenerateGeneric()
	if !strings.Contains(content, "#!/") {
		t.Error("should be a shell script")
	}
	if !strings.Contains(content, "gdt install") {
		t.Error("should contain gdt install")
	}
}

func TestProviders(t *testing.T) {
	providers := Providers()
	if len(providers) != 3 {
		t.Errorf("expected 3 providers, got %d", len(providers))
	}
}

func TestOutputPath(t *testing.T) {
	tests := []struct {
		provider string
		want     string
	}{
		{"github", ".github/workflows/export.yml"},
		{"gitlab", ".gitlab-ci.yml"},
		{"generic", "ci/export.sh"},
	}
	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			got := OutputPath(tt.provider)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
