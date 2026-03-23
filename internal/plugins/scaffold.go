package plugins

import (
	"fmt"
	"os"
	"path/filepath"
)

// ScaffoldOptions configures plugin scaffold generation.
type ScaffoldOptions struct {
	Name string
	Lang string // "shell" (default), "go", "rust", "python"
}

// ScaffoldV2 creates a new plugin directory with V2 structure and language choice.
func (s *Service) ScaffoldV2(opts ScaffoldOptions) (string, error) {
	dir := filepath.Join(".", "gdt-"+opts.Name)

	if _, err := os.Stat(dir); err == nil {
		return "", fmt.Errorf("directory %s already exists", dir)
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	// Create contribution directories
	for _, subdir := range []string{"templates", "presets", "ci"} {
		os.MkdirAll(filepath.Join(dir, subdir), 0755)
	}

	manifest := fmt.Sprintf(`name = %q
version = "0.1.0"
protocol = 2
commands = [%q]
requires_gdt = ">=1.0"
description = ""

# [build]
# command = ""

# [contributions]
# templates = []
# presets = []
# ci_providers = []
# hooks = []
# doctor = false
# completions = false
`, opts.Name, opts.Name)

	if err := os.WriteFile(filepath.Join(dir, ManifestFile), []byte(manifest), 0644); err != nil {
		return "", err
	}

	var writeErr error
	switch opts.Lang {
	case "go":
		writeErr = writeGoScaffold(dir, opts.Name)
	case "rust":
		writeErr = writeRustScaffold(dir, opts.Name)
	case "python":
		writeErr = writePythonScaffold(dir, opts.Name)
	default:
		writeErr = writeShellScaffold(dir, opts.Name)
	}
	if writeErr != nil {
		return "", writeErr
	}

	readme := fmt.Sprintf("# gdt-%s\n\nA gdt plugin.\n\n## Usage\n\n```sh\ngdt %s\n```\n", opts.Name, opts.Name)
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte(readme), 0644); err != nil {
		return "", err
	}

	return dir, nil
}

func writeShellScaffold(dir, name string) error {
	script := fmt.Sprintf(`#!/usr/bin/env bash
set -euo pipefail

case "${1:-}" in
  doctor)
    echo "OK %s ready"
    ;;
  hook)
    event="${2:-}"
    echo "OK $event completed"
    ;;
  completions)
    # shell="${2:-}"
    ;;
  *)
    echo "gdt-%s plugin"
    echo "Usage: gdt %s <command>"
    ;;
esac
`, name, name, name)
	return os.WriteFile(filepath.Join(dir, name), []byte(script), 0755)
}

func writeGoScaffold(dir, name string) error {
	main := `package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "gdt-` + name + ` plugin\nUsage: gdt ` + name + ` <command>\n")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "doctor":
		fmt.Println("OK ` + name + ` ready")
	case "hook":
		if len(os.Args) > 2 {
			fmt.Printf("OK %s completed\n", os.Args[2])
		}
	case "completions":
		// Generate shell completions for os.Args[2] (bash|zsh|fish)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
`

	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(main), 0644); err != nil {
		return err
	}

	gomod := fmt.Sprintf("module github.com/example/gdt-%s\n\ngo 1.22\n", name)
	return os.WriteFile(filepath.Join(dir, "go.mod"), []byte(gomod), 0644)
}

func writeRustScaffold(dir, name string) error {
	srcDir := filepath.Join(dir, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		return err
	}

	main := fmt.Sprintf(`use std::env;
use std::process;

fn main() {
    let args: Vec<String> = env::args().collect();
    if args.len() < 2 {
        eprintln!("gdt-%s plugin\nUsage: gdt %s <command>");
        process::exit(1);
    }

    match args[1].as_str() {
        "doctor" => {
            if args.len() > 2 && args[2] == "check" {
                println!("OK %s ready");
            }
        }
        "hook" => {
            if args.len() > 2 {
                println!("OK {} completed", args[2]);
            }
        }
        "completions" => {
            // Generate shell completions for args[2] (bash|zsh|fish)
        }
        _ => {
            eprintln!("Unknown command: {}", args[1]);
            process::exit(1);
        }
    }
}
`, name, name, name)

	if err := os.WriteFile(filepath.Join(srcDir, "main.rs"), []byte(main), 0644); err != nil {
		return err
	}

	cargo := fmt.Sprintf(`[package]
name = "%s"
version = "0.1.0"
edition = "2021"

[[bin]]
name = "%s"
path = "src/main.rs"
`, name, name)
	return os.WriteFile(filepath.Join(dir, "Cargo.toml"), []byte(cargo), 0644)
}

func writePythonScaffold(dir, name string) error {
	script := fmt.Sprintf(`#!/usr/bin/env python3
import sys

def main():
    if len(sys.argv) < 2:
        print(f"gdt-%s plugin", file=sys.stderr)
        print(f"Usage: gdt %s <command>", file=sys.stderr)
        sys.exit(1)

    cmd = sys.argv[1]

    if cmd == "doctor" and len(sys.argv) > 2 and sys.argv[2] == "check":
        print("OK %s ready")
    elif cmd == "hook" and len(sys.argv) > 2:
        print(f"OK {sys.argv[2]} completed")
    elif cmd == "completions":
        pass  # Generate shell completions for sys.argv[2]
    else:
        print(f"Unknown command: {cmd}", file=sys.stderr)
        sys.exit(1)

if __name__ == "__main__":
    main()
`, name, name, name)
	return os.WriteFile(filepath.Join(dir, name), []byte(script), 0755)
}
