package cli

import (
	"testing"

	"github.com/monkeymonk/gdt/internal/config"
	"github.com/monkeymonk/gdt/internal/platform"
)

// TestNewDapCmd_FlagDefaults verifies that the DAP command has the correct
// default flag values: --port defaults to 6006 and -C/--path defaults to
// an empty string.
func TestNewDapCmd_FlagDefaults(t *testing.T) {
	app := &App{
		Home:     t.TempDir(),
		Platform: platform.Info{OS: "linux", Arch: "amd64"},
		Config:   &config.Config{},
	}

	cmd := newDapCmd(app)

	// Check --port default
	portFlag := cmd.Flags().Lookup("port")
	if portFlag == nil {
		t.Fatalf("expected --port flag to exist")
	}
	if portFlag.DefValue != "6006" {
		t.Errorf("expected --port default value of 6006, got %q", portFlag.DefValue)
	}

	// Check -C/--path default
	pathFlag := cmd.Flags().Lookup("path")
	if pathFlag == nil {
		t.Fatalf("expected --path flag to exist")
	}
	if pathFlag.DefValue != "" {
		t.Errorf("expected --path default value to be empty string, got %q", pathFlag.DefValue)
	}
}
