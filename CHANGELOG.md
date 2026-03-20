# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-03-19

### Added

- Version management: install, remove, list, use, local
- Remote version discovery with `ls-remote`
- Shim system for transparent version resolution (busybox-style symlink)
- Export template management
- Plugin system with manifest-based discovery and dispatch
- Project scaffolding (`gdt new`) with interactive prompts
- C# project support (`--csharp` flag generates .csproj and .sln)
- Template-based scaffolding (`--template` flag clones Git repos)
- LSP proxy for Neovim, Helix, and other stdin/stdout editors
- DAP proxy for debugger integration
- Export automation wrapping Godot's headless export
- Headless export captures stdout/stderr; included in error output on failure
- `--verbose` / `-v` flag on `gdt export` to stream engine output in real-time
- CI pipeline generation (GitHub Actions, GitLab CI, shell script)
- `gdt edit` command and `--editor` flag on `gdt run`
- `gdt doctor` diagnostics
- Self-update mechanism
- Shell integration (`gdt shell init`)
- Interactive mode for all commands via charmbracelet/huh
- Desktop launcher integration on Linux
- Cross-platform install scripts (shell + PowerShell)
- Cross-platform support (Linux, macOS, Windows)
