package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/monkeymonk/gdt/internal/editor"
	"github.com/spf13/cobra"
)

func newEditorCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "editor",
		Short: "Editor integration helpers",
	}

	cmd.AddCommand(newEditorSetupCmd(app))
	cmd.AddCommand(newEditorSnippetCmd())
	return cmd
}

func newEditorSetupCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "setup <editor>",
		Short: "Set up editor LSP integration",
		Long:  fmt.Sprintf("Supported editors: %s", strings.Join(editor.SupportedEditors(), ", ")),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			return editor.Setup(args[0], cwd)
		},
	}
}

func newEditorSnippetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "snippet <editor>",
		Short: "Print LSP config snippet for an editor",
		Long:  fmt.Sprintf("Supported editors: %s", strings.Join(editor.SupportedEditors(), ", ")),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := editor.Snippet(args[0])
			if s == "" {
				return fmt.Errorf("unsupported editor: %s (supported: %s)", args[0], strings.Join(editor.SupportedEditors(), ", "))
			}
			fmt.Println(s)
			return nil
		},
	}
}
