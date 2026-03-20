package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const gitignoreContent = `# Godot
.godot/
*.import
.mono/
data_*/
mono_crash.*.txt
export_presets.cfg

# Build
dist/
build/

# OS
.DS_Store
Thumbs.db
`

const gitattributesContent = `# LFS
*.png filter=lfs diff=lfs merge=lfs -text
*.wav filter=lfs diff=lfs merge=lfs -text
*.ogg filter=lfs diff=lfs merge=lfs -text
*.glb filter=lfs diff=lfs merge=lfs -text
*.gltf filter=lfs diff=lfs merge=lfs -text

# Linguist
*.tres linguist-language=Godot
*.tscn linguist-language=Godot
`

const preCommitHook = `#!/bin/sh
# gdt-git pre-commit hook

# Check project.godot exists
if [ ! -f "project.godot" ]; then
    echo "WARNING: project.godot not found in repository root"
fi

# Check for oversized files (>10MB)
LIMIT=10485760
oversized=""
for file in $(git diff --cached --name-only --diff-filter=ACM); do
    if [ -f "$file" ]; then
        size=$(wc -c < "$file")
        if [ "$size" -gt "$LIMIT" ]; then
            mb=$(echo "scale=1; $size / 1048576" | bc 2>/dev/null || echo "$(($size / 1048576))")
            oversized="$oversized\n  $file (${mb}MB)"
        fi
    fi
done

if [ -n "$oversized" ]; then
    echo "WARNING: Large files detected (>10MB):$oversized"
    echo "Consider using Git LFS for these files."
fi

exit 0
`

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	root := os.Getenv("GDT_PROJECT_ROOT")
	if root == "" {
		root, _ = os.Getwd()
	}

	switch os.Args[1] {
	case "setup":
		force := hasFlag("--force")
		if err := cmdSetup(root, force); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "hooks":
		if err := cmdHooks(root); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "help", "--help", "-h":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Println(`gdt-git - Git workflow utilities for Godot projects

Usage:
  gdt-git <command> [flags]

Commands:
  setup    Create .gitignore and .gitattributes for Godot projects
  hooks    Install pre-commit hook for file size checks

Flags (setup):
  --force  Overwrite existing files`)
}

func hasFlag(flag string) bool {
	for _, arg := range os.Args[2:] {
		if arg == flag {
			return true
		}
	}
	return false
}

func cmdSetup(root string, force bool) error {
	files := map[string]string{
		".gitignore":      gitignoreContent,
		".gitattributes":  gitattributesContent,
	}

	for name, content := range files {
		path := filepath.Join(root, name)
		if !force {
			if _, err := os.Stat(path); err == nil {
				fmt.Printf("skipped %s (already exists, use --force to overwrite)\n", name)
				continue
			}
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("writing %s: %w", name, err)
		}
		fmt.Printf("created %s\n", name)
	}
	return nil
}

func cmdHooks(root string) error {
	gitDir := filepath.Join(root, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return fmt.Errorf(".git directory not found in %s (is this a git repository?)", root)
	}

	hooksDir := filepath.Join(gitDir, "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return fmt.Errorf("creating hooks directory: %w", err)
	}

	hookPath := filepath.Join(hooksDir, "pre-commit")
	if _, err := os.Stat(hookPath); err == nil {
		existing, readErr := os.ReadFile(hookPath)
		if readErr == nil && !strings.Contains(string(existing), "gdt-git") {
			return fmt.Errorf("pre-commit hook already exists (not managed by gdt-git)")
		}
	}

	if err := os.WriteFile(hookPath, []byte(preCommitHook), 0755); err != nil {
		return fmt.Errorf("writing pre-commit hook: %w", err)
	}

	fmt.Println("installed pre-commit hook")
	return nil
}
