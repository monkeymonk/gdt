# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-03-22

### Added

- Version management: install, remove, list, use, local
- Remote version discovery with `ls-remote`
- Export template management (install, remove, list)
- Plugin system with V2 protocol and contributions model
  - Plugin commands, templates, presets, CI providers, hooks, doctor checks, completions
  - Namespace resolution with ambiguity detection
  - Timeout-aware subprocess execution with process group cleanup
  - Line protocol (`OK`, `WARN`, `FAIL`) for structured plugin output
  - Scaffold new plugins with `gdt plugin new` (shell and Go templates)
- Project scaffolding (`gdt new`) with interactive prompts
  - Built-in 2D and 3D templates
  - C# project support (`--csharp` generates .csproj and .sln)
  - Template-based scaffolding from Git repos or plugin templates
- LSP proxy for editors (Neovim, Helix, Zed, Emacs, VS Code)
- DAP proxy for debugger integration
- Export automation wrapping Godot's headless export
  - Plugin lifecycle hooks (`before_export`, `after_export`)
  - `--verbose` flag to stream engine output in real-time
- CI pipeline generation (GitHub Actions, GitLab CI, shell script)
- `gdt edit` command to open the Godot editor
- `gdt doctor` diagnostics with plugin-contributed checks
- `gdt run` with version resolution and `before_run` hook
- Self-update mechanism (`gdt self update`)
- Shell integration (`gdt shell init`)
- Shell completion (bash, zsh, fish, powershell) with plugin aggregation
- Interactive mode for all commands via charmbracelet/huh
- Desktop launcher integration on Linux
- Download resume via HTTP Range requests
- Mirror fallback for downloads
- SHA-512 checksum verification for all downloads
- Cross-platform install scripts (shell + PowerShell)
- Cross-platform support (Linux, macOS, Windows)
