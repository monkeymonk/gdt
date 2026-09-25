# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed

- `gdt plugin install` now actually enforces a plugin's `requires_gdt`
  version constraint instead of silently ignoring it — a plugin
  declaring a constraint the running `gdt` version doesn't satisfy is
  now rejected with an actionable error (`gdt self update`) instead of
  installing anyway. Only the `>=X.Y[.Z]` form is supported. An
  unversioned (`dev`) build always skips the check. This affects new
  installs only — a plugin already installed under the old, unenforced
  behavior is unaffected until its next `gdt plugin install`/reinstall.

## [0.3.0] - 2026-09-25

### Fixed

- CLI commands no longer silently ignore errors from working-directory
  detection, project-root detection, plugin discovery, engine-version/
  template listing, and plugin hook execution — failures are now reported
  (or, for best-effort operations like the desktop launcher and doctor
  diagnostics, logged/surfaced) instead of proceeding with empty or
  zero-value state. In particular, a fatal `after_install`/`after_export`/
  `after_use`/`after_ci_setup` plugin hook failure now correctly aborts
  the command with a non-zero exit, matching the documented plugin hook
  contract, instead of being silently discarded.
- `mirrors` configured in `~/.gdt/config.toml` now actually take effect:
  `gdt install`/`gdt templates install` fall back to a configured mirror
  when the primary GitHub download URL is unavailable, per the
  documented behavior — previously this config key was silently ignored.
- `gdt install`'s and `gdt update`'s GitHub API URL resolution now
  honors a configured `godot_api` (forks/mirrors), matching `gdt
  install`'s own version-download behavior — previously only the actual
  download used it, while version listing/refresh always hit the
  default GitHub API regardless of configuration.
- `gdt self update` no longer silently loses a file-copy error during
  the binary swap when the copy's data itself transferred but the final
  flush/close failed (e.g. disk full) — this could previously report
  success on a truncated binary.
- `gdt ci setup` no longer fails the whole command (despite having
  already written the CI configuration file) when run outside an
  existing Godot project directory — project detection is only needed
  for the optional `after_ci_setup` plugin hook's context, which is now
  skipped rather than treated as a fatal error when no project is found.
- `gdt shell init` no longer silently falls back to an empty PATH entry
  when it can't determine the running binary's own path (a rare,
  sandboxed-environment edge case) — it now keeps the already-resolved
  fallback directory instead.
- The Godot engine download (`gdt install`), the self-update release
  check (`FetchLatestRelease`), and plugin binary resolution
  (`gdt plugin install`) no longer use an unbounded HTTP client — all
  three now share the same 30s-timeout-configured client already used
  elsewhere, closing a hang risk on a stalled connection that an
  earlier round's timeout fix missed on these three call sites.

### Added

- `skills/gdt/SKILL.md` — an agent-facing skill for AI coding tools
  driving `gdt` itself, covering non-interactive invocation, version
  resolution order, and common task recipes.
- `install_script_url` config key (`~/.gdt/config.toml`) — overrides
  the install-script URL baked into `gdt ci setup`-generated CI output
  (GitHub Actions, GitLab CI, generic shell), for forks, internal
  mirrors, or air-gapped CI environments.

### Changed

- `gdt list` and `gdt ls-remote` now order versions newest to oldest
  (numeric comparison, e.g. `4.10` above `4.9`) instead of an
  alphabetical or API-incidental order — including when the result
  comes from a cached `gdt ls-remote` response, not just a fresh fetch.
- Plugin discovery now runs at most once per `gdt` invocation instead
  of once per command handler / hook call — internal change, no
  user-visible behavior difference beyond commands that invoke
  multiple plugin hooks (e.g. `gdt new`) doing noticeably less
  filesystem I/O.
- Metadata cache-write failures (`gdt update`, `gdt ls-remote`, `gdt
  install`) are now observable under `GDT_DEBUG=1` instead of being
  fully silent; behavior is unchanged (a cache-write failure still never
  fails the command, since the fetched data is already in hand).
- A failing `before_run`/`before_new`/`after_new`/`before_export` plugin
  hook now returns the same actionable error shape (with a "check the
  plugin's hook script" suggestion) as the other plugin hooks already
  did — previously these four returned a bare error with no recovery
  guidance.
- `gdt export`'s "preset name required" message now shows the command's
  usage form (`gdt export <preset>`) alongside the existing `--list`
  hint, matching every other command's required-argument message
  format.

## [0.2.2] - 2026-07-21

### Changed

- Release signing now uses the cosign/Sigstore bundle format: releases publish a
  single `checksums.txt.bundle` instead of separate `.sig` and `.pem` files.
  Verify with `cosign verify-blob --bundle checksums.txt.bundle ...` (see README)
- `cosign-installer` upgraded to v4 (off deprecated Node 20)
- `.goreleaser.yml` uses the `formats:` list syntax (replaces deprecated
  `archives.format` / `format_overrides.format`)

## [0.2.1] - 2026-07-21

### Fixed

- Release binaries are now built with the latest Go 1.25.x patch release
  (`go.mod` uses `go 1.25` plus `check-latest`), picking up standard-library
  security fixes that the pinned `1.25.0` toolchain missed

### Changed

- CI `lint` job uses golangci-lint v2 (via `golangci-lint-action` v9) with a
  curated linter set in `.golangci.yml`
- Bumped GitHub Actions off the deprecated Node 20 runtime (`checkout`,
  `setup-go`, `goreleaser-action`); `cosign-installer` stays on v3 for
  compatibility with the current signing recipe

## [0.2.0] - 2026-07-21

### Fixed

- `gdt self update` now works: the release asset is matched against the published
  (`v`-stripped) name, the downloaded archive is extracted, and the `gdt` binary
  inside it is swapped in atomically (previously the archive was written over the
  binary and the asset name never matched)

### Added

- SHA-256 checksum verification of the downloaded archive during `gdt self update`
  when the release publishes `checksums.txt`
- Keyless cosign signing of release `checksums.txt` (Sigstore), producing `.sig`
  and `.pem` for verifiable release provenance

### Changed

- Build with Go 1.25 to pick up standard-library security fixes; CI and release
  workflows now derive the Go version from `go.mod`
- CI runs `golangci-lint` and `govulncheck`; added Dependabot for Go modules and
  GitHub Actions
- Upgraded `charmbracelet/huh` v0.8.0 → v1.0.0

## [0.1.3] - 2026-03-23

### Added

- Auto-resolve plugin binary on install and update (download release or build from source)
- GitHub Release download with OS/arch detection for pre-built plugin binaries
- Auto-detect build system (Makefile, go.mod, Cargo.toml, build.sh) for source builds
- Optional `[build]` section in plugin.toml for custom build commands
- Rust and Python plugin scaffolds (`gdt plugin new --lang rust|python`)

### Fixed

- Plugin command dispatch now registers as proper cobra subcommands instead of help function override
- `gdt plugin remove` now accepts repo slugs (`owner/repo`) and manifest names in addition to directory names

## [0.1.2] - 2026-03-22

### Fixed

- Double `v` prefix in help output (`vv0.1.0` → `v0.1.0`)
- Reference demo.gif from release assets in README
- Windows: use forward slashes for CI output paths
- Windows: close file handle before rename in download resume
- Windows: add `.exe` suffix to test binary in integration tests
- Windows: use `file://` URLs for local git clone in scaffold tests
- Accept `file://` URLs in `CloneTemplate`

## [0.1.1] - 2026-03-22

### Added

- Configurable GitHub API URLs via `config.toml` (`godot_api`, `selfupdate_api`) for Godot forks and mirrors

### Changed

- Consolidated path resolution into `engine.Service` (single source of truth for `VersionsDir`, `TemplatesDir`, `CacheDir`, `CachePath`)
- Moved `resolveProjectVersion` logic into `engine.Service.ResolveProject()`
- Extracted shared `downloadAndInstall()` helper, eliminating duplication between `Install()` and `InstallTemplates()`
- Extracted `listDirectories()` helper, reducing duplication between `List()` and `ListTemplates()`
- Extracted generic `discoverContributions[T]()` helper for plugin discovery methods
- Simplified template resolution logic in `new` command via extracted `resolveTemplate()`
- Moved plugin dispatch logic closer to `plugins.Service`
- Optimized `ResolveVersion()` from 3-pass to 2-pass with combined alias+exact matching
- Unexported internal GitHub API types (`githubRelease`, `githubAsset`)
- Renamed `GDT_GITHUB_TOKEN` to `GITHUB_TOKEN` for consistency

### Fixed

- Windows compilation: use build tags for platform-specific process group handling
- CI: use bash shell for test step on Windows

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
