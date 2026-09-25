package ci

import (
	"strings"
	"testing"
)

const defaultInstallScriptURLForTests = "https://raw.githubusercontent.com/monkeymonk/gdt/main/scripts/install.sh"

func TestGenerateGitHub(t *testing.T) {
	content := GenerateGitHub(defaultInstallScriptURLForTests)
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
	content := GenerateGitLab(defaultInstallScriptURLForTests)
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
	content := GenerateGeneric(defaultInstallScriptURLForTests)
	if !strings.Contains(content, "#!/") {
		t.Error("should be a shell script")
	}
	if !strings.Contains(content, "gdt install") {
		t.Error("should contain gdt install")
	}
}

// TestGenerate_InstallScriptURLSubstitution proves the installScriptURL
// parameter is substituted into the generated output for every provider:
// the default URL (mirroring config.Config.InstallScriptURL()'s fallback)
// produces the historical hardcoded output, and a custom URL (e.g. for a
// fork or mirror) replaces it exactly, with no trace of the default left
// behind.
func TestGenerate_InstallScriptURLSubstitution(t *testing.T) {
	const custom = "https://example.com/fork/install.sh"
	tests := []struct {
		provider string
		gen      func(string) string
	}{
		{"github", GenerateGitHub},
		{"gitlab", GenerateGitLab},
		{"generic", GenerateGeneric},
	}
	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			defaultContent := tt.gen(defaultInstallScriptURLForTests)
			if !strings.Contains(defaultContent, defaultInstallScriptURLForTests) {
				t.Errorf("expected default URL %q in output", defaultInstallScriptURLForTests)
			}

			customContent := tt.gen(custom)
			if !strings.Contains(customContent, custom) {
				t.Errorf("expected custom URL %q in output", custom)
			}
			if strings.Contains(customContent, defaultInstallScriptURLForTests) {
				t.Errorf("did not expect default URL %q in output when custom URL given", defaultInstallScriptURLForTests)
			}
		})
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
