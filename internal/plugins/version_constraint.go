package plugins

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/monkeymonk/gdt/internal/metadata"
)

// checkRequiresGdt validates a plugin manifest's requires_gdt constraint
// against the running gdt version. Only the ">=X.Y[.Z]" constraint form is
// supported — no other operator has ever appeared in this codebase's own
// scaffold, README, or any reference plugin manifest.
//
// An empty constraint (nothing declared), an unversioned/dev build
// (gdtVersion == "" or "dev", which cannot be meaningfully compared), and a
// malformed or unrecognized constraint (wrong operator, unparseable
// version) all pass without blocking install — a malformed constraint is
// logged via slog.Debug rather than rejected, matching ParseManifest's own
// posture of never validating this field.
func checkRequiresGdt(name, constraint, gdtVersion string) error {
	if constraint == "" {
		return nil
	}
	if gdtVersion == "" || gdtVersion == "dev" {
		return nil
	}

	requiredVersion, ok := strings.CutPrefix(constraint, ">=")
	if !ok {
		slog.Debug("unrecognized requires_gdt constraint", "constraint", constraint)
		return nil
	}

	if metadata.CompareVersions(gdtVersion, requiredVersion) >= 0 {
		return nil
	}

	return fmt.Errorf("plugin %s requires gdt %s, running %s\n\n  gdt self update", name, constraint, gdtVersion)
}
