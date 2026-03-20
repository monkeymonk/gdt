package editor

import (
	"fmt"
	"os"
	"path/filepath"
)

// SupportedEditors returns the list of supported editor identifiers.
func SupportedEditors() []string {
	return []string{"nvim", "helix", "vscode"}
}

// Snippet returns the LSP configuration snippet for the given editor.
// Returns an empty string if the editor is not supported.
func Snippet(editor string) string {
	switch editor {
	case "nvim":
		return neovimSnippet()
	case "helix":
		return helixSnippet()
	case "vscode":
		return vscodeSnippet()
	}
	return ""
}

// Setup writes editor-specific configuration files to the project root.
// For vscode, it creates .vscode/settings.json if it does not already exist.
// For other editors, it prints the snippet to stdout.
func Setup(editor string, projectRoot string) error {
	snippet := Snippet(editor)
	if snippet == "" {
		return fmt.Errorf("unsupported editor: %s (supported: nvim, helix, vscode)", editor)
	}

	if editor == "vscode" {
		dir := filepath.Join(projectRoot, ".vscode")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		path := filepath.Join(dir, "settings.json")
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("%s already exists; snippet:\n%s", path, snippet)
		}
		return os.WriteFile(path, []byte(snippet), 0o644)
	}

	// For non-vscode editors, print the snippet for the user to add manually.
	fmt.Println(snippet)
	return nil
}
