package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func newShellCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shell",
		Short: "Shell integration",
	}

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Print shell PATH configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			shell := detectShell()
			binDir := filepath.Dir(os.Args[0])
			if binDir == "." {
				exe, _ := os.Executable()
				binDir = filepath.Dir(exe)
			}
			shimsDir := app.ShimsDir()

			switch shell {
			case "fish":
				fmt.Printf("fish_add_path %s\n", shimsDir)
				fmt.Printf("fish_add_path %s\n", binDir)
			default:
				fmt.Printf("export PATH=\"%s:%s:$PATH\"\n", shimsDir, binDir)
			}
			return nil
		},
	}

	cmd.AddCommand(initCmd)
	return cmd
}

func detectShell() string {
	shell := os.Getenv("SHELL")
	if strings.Contains(shell, "fish") {
		return "fish"
	}
	if strings.Contains(shell, "zsh") {
		return "zsh"
	}
	return "bash"
}
