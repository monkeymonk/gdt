package ci

import (
	"fmt"
	"path"
)

type Provider struct {
	Name  string
	Label string
}

func Providers() []Provider {
	return []Provider{
		{Name: "github", Label: "GitHub Actions"},
		{Name: "gitlab", Label: "GitLab CI"},
		{Name: "generic", Label: "Generic (shell script)"},
	}
}

func OutputPath(provider string) string {
	switch provider {
	case "github":
		return path.Join(".github", "workflows", "export.yml")
	case "gitlab":
		return ".gitlab-ci.yml"
	case "generic":
		return path.Join("ci", "export.sh")
	default:
		return ""
	}
}

func Generate(provider string, installScriptURL string) string {
	switch provider {
	case "github":
		return GenerateGitHub(installScriptURL)
	case "gitlab":
		return GenerateGitLab(installScriptURL)
	case "generic":
		return GenerateGeneric(installScriptURL)
	default:
		return ""
	}
}

func GenerateGitHub(installScriptURL string) string {
	return fmt.Sprintf(`name: Export Game

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  export:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v4

      - name: Install gdt
        run: curl -fsSL %s | sh

      - name: Install engine
        run: gdt install

      - name: Install templates
        run: gdt templates install $(cat .godot-version)

      - name: Export
        run: gdt export Linux/X11

      - name: Upload artifact
        uses: actions/upload-artifact@v4
        with:
          name: game-linux
          path: dist/
`, installScriptURL)
}

func GenerateGitLab(installScriptURL string) string {
	return fmt.Sprintf(`stages:
  - export

export:
  stage: export
  image: ubuntu:latest
  before_script:
    - curl -fsSL %s | sh
    - gdt install
    - gdt templates install $(cat .godot-version)
  script:
    - gdt export Linux/X11
  artifacts:
    paths:
      - dist/
`, installScriptURL)
}

func GenerateGeneric(installScriptURL string) string {
	return fmt.Sprintf(`#!/usr/bin/env bash
set -euo pipefail

# Install gdt if not available
if ! command -v gdt &> /dev/null; then
    curl -fsSL %s | sh
fi

# Install engine and templates
gdt install
gdt templates install "$(cat .godot-version)"

# Export
gdt export Linux/X11

echo "Export complete: dist/"
`, installScriptURL)
}
