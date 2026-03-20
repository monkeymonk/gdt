package metadata

import (
	"fmt"
	"strings"

	"github.com/monkeymonk/gdt/internal/platform"
)

// platformArtifactString returns the platform-specific artifact suffix.
func platformArtifactString(plat platform.Info) (string, error) {
	switch plat.OS {
	case "linux":
		switch plat.Arch {
		case "amd64":
			return "linux.x86_64", nil
		}
	case "darwin":
		return "macos.universal", nil
	case "windows":
		switch plat.Arch {
		case "amd64":
			return "win64.exe", nil
		}
	}
	return "", fmt.Errorf("unsupported platform: %s/%s", plat.OS, plat.Arch)
}

// ArtifactName returns the expected zip filename for a Godot engine build.
func ArtifactName(plat platform.Info, version string, mono bool) (string, error) {
	platStr, err := platformArtifactString(plat)
	if err != nil {
		return "", err
	}
	if mono {
		return fmt.Sprintf("Godot_v%s-stable_mono_%s.zip", version, platStr), nil
	}
	return fmt.Sprintf("Godot_v%s-stable_%s.zip", version, platStr), nil
}

// TemplateArtifactName returns the expected filename for export templates.
func TemplateArtifactName(version string, mono bool) string {
	if mono {
		return fmt.Sprintf("Godot_v%s-stable_mono_export_templates.tpz", version)
	}
	return fmt.Sprintf("Godot_v%s-stable_export_templates.tpz", version)
}

// ResolveEngineArtifact searches the release assets for the matching engine artifact.
// Falls back to a constructed name if no exact match is found.
func ResolveEngineArtifact(release *Release, plat platform.Info, mono bool) (string, error) {
	platStr, err := platformArtifactString(plat)
	if err != nil {
		return "", err
	}

	prefix := fmt.Sprintf("Godot_v%s-stable_", release.Version)
	if mono {
		prefix = fmt.Sprintf("Godot_v%s-stable_mono_", release.Version)
	}

	for name := range release.Assets {
		if strings.HasPrefix(name, prefix) && strings.Contains(name, platStr) && !strings.Contains(name, "export_templates") {
			return name, nil
		}
	}

	return prefix + platStr + ".zip", nil
}
