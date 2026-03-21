# gdt — Godot Developer Toolchain

[![CI](https://github.com/monkeymonk/gdt/actions/workflows/ci.yml/badge.svg)](https://github.com/monkeymonk/gdt/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/github/go-mod-go-version/monkeymonk/gdt)](https://go.dev/)
[![License](https://img.shields.io/github/license/monkeymonk/gdt)](LICENSE)
[![Release](https://img.shields.io/github/v/release/monkeymonk/gdt)](https://github.com/monkeymonk/gdt/releases/latest)

A cross-platform CLI to manage Godot Engine installations, scaffold projects, proxy LSP/DAP for editors, automate exports, and generate CI pipelines.

## Features

- Install and manage multiple Godot engine versions
- Standard and Mono/C# build support
- Deterministic per-project version pinning (`.godot-version`)
- Optional `godot` alias — `alias godot="gdt run"`
- Project scaffolding with built-in templates (2D, 3D) and interactive prompts
- LSP and DAP proxy for editors and AI coding tools
- Export automation with preset management and plugin hooks
- CI pipeline generation (GitHub Actions, GitLab CI, shell script)
- Export template management
- Plugin ecosystem with lifecycle hooks
- Download resume and mirror fallback support
- SHA-256 checksum verification for all downloads
- Diagnostic tooling (`gdt doctor`)
- Shell completion (bash, zsh, fish, powershell)
- Desktop launcher integration (Linux)

## Installation

### Linux / macOS

```sh
curl -fsSL https://raw.githubusercontent.com/monkeymonk/gdt/main/scripts/install.sh | sh
```

### Windows (PowerShell)

```powershell
irm https://raw.githubusercontent.com/monkeymonk/gdt/main/scripts/install.ps1 | iex
```

### From source

```sh
go install github.com/monkeymonk/gdt/cmd/gdt@latest
```

## Shell Setup

Add gdt to your PATH:

```sh
# Add to your .bashrc, .zshrc, etc.
eval "$(gdt shell init)"

# Optional: create a godot alias
alias godot="gdt run"
```

## Quick Start

```sh
# Install a Godot version
gdt install 4.3

# Set it as global default
gdt use 4.3

# Create a new project
gdt new mygame --version 4.3 --renderer forward_plus

# Create a C# project
gdt new mygame --version 4.3 --renderer forward_plus --csharp

# Create from a built-in template
gdt new mygame --template 2d --version 4.3
gdt new mygame --template 3d --version 4.3

# Open the editor
cd mygame
gdt edit

# Run the game
gdt run

# Pin a project to a specific version
gdt local 4.2

# With the alias, just run godot directly
godot
```

All commands with required parameters support interactive mode — run without arguments and gdt will prompt you.

## Commands

### Project

| Command | Description |
|---|---|
| `gdt new [name]` | Create a new Godot project |
| `gdt edit [version] [-- <args>]` | Open the Godot editor |
| `gdt run [version] [-- <args>]` | Run the game |
| `gdt export [preset]` | Export project for a platform |
| `gdt ci setup` | Generate CI pipeline configuration |

### Version Management

| Command | Description |
|---|---|
| `gdt install [version]` | Install a Godot engine version |
| `gdt remove [version]` | Remove an installed version |
| `gdt list` | List installed versions |
| `gdt ls-remote` | List available remote versions |
| `gdt use [version]` | Set global default version |
| `gdt local [version]` | Pin version for current project |

Aliases: `remove` → `rm`, `uninstall`; `list` → `ls`

### Templates

| Command | Description |
|---|---|
| `gdt templates install [version]` | Install export templates |
| `gdt templates remove [version]` | Remove installed export templates |
| `gdt templates list` | List installed templates |

Aliases: `remove` → `rm`

### Plugins

| Command | Description |
|---|---|
| `gdt plugin install [repo]` | Install a plugin |
| `gdt plugin list` | List installed plugins |
| `gdt plugin update [name]` | Update plugins (all or by name) |
| `gdt plugin remove [name]` | Remove a plugin |
| `gdt plugin new [name]` | Scaffold a new plugin |

### Utilities

| Command | Description |
|---|---|
| `gdt doctor` | Diagnose installation problems |
| `gdt update` | Refresh release metadata cache |
| `gdt self update` | Update gdt itself |
| `gdt shell init` | Print shell PATH configuration |
| `gdt lsp` | Start LSP proxy for editors |
| `gdt dap` | Start DAP proxy for debuggers |
| `gdt completion <shell>` | Generate shell completion script |

---

## Command Details

### gdt install

```sh
gdt install [version]
```

| Flag | Description |
|---|---|
| `--mono` | Install Mono/C# build |
| `--force` | Force reinstall even if already installed |
| `--refresh` | Refresh metadata cache before resolving |

Downloads are verified against SHA-256 checksums. Interrupted downloads are automatically resumed via HTTP Range requests. If the primary download URL is unavailable, configured mirrors are tried as fallbacks.

When run without arguments: reads `.godot-version` if present, otherwise prompts interactively. On Linux, a desktop launcher is created on first install.

### gdt remove

```sh
gdt remove [version]
```

Aliases: `rm`, `uninstall`. Prompts for confirmation interactively. Warns if removing the global default. Removes the desktop launcher when the last version is uninstalled.

### gdt run

```sh
gdt run [version] [-- <args>]
```

| Flag | Description |
|---|---|
| `-e`, `--editor` | Open the editor instead of running the game |

The first argument is tried as a version (prefix match, alias like `latest`). Falls back to `.godot-version`, `GDT_GODOT_VERSION`, or the global default.

### gdt edit

```sh
gdt edit [version] [-- <args>]
```

Shortcut for `gdt run --editor`. Opens the Godot editor with version resolution.

### gdt ls-remote

```sh
gdt ls-remote
```

| Flag | Description |
|---|---|
| `--refresh` | Force refresh metadata cache |

### gdt export

```sh
gdt export [preset]
```

| Flag | Description |
|---|---|
| `--output` | Output directory (default: `dist/<preset>`) |
| `--debug` | Use debug export instead of release |
| `-v`, `--verbose` | Show Godot engine output |
| `--list` | List available export presets |

Plugin hooks (`before_export`, `after_export`) are executed automatically if any installed plugins define them.

### gdt ci setup

```sh
gdt ci setup
```

| Flag | Description |
|---|---|
| `--provider` | CI provider: `github`, `gitlab`, `generic` |

### gdt templates install

```sh
gdt templates install [version]
```

| Flag | Description |
|---|---|
| `--mono` | Install Mono templates |
| `--refresh` | Refresh metadata cache |

### gdt templates remove

```sh
gdt templates remove [version]
```

Aliases: `rm`. Prompts for confirmation interactively.

### gdt lsp / dap

```sh
gdt lsp [--port PORT] [-C PATH]
gdt dap [--port PORT] [-C PATH]
```

| Flag | Default | Description |
|---|---|---|
| `--port` | 6005 (LSP) / 6006 (DAP) | Godot TCP port |
| `-C`, `--path` | current directory | Path to Godot project |

### gdt new

```sh
gdt new [name]
```

| Flag | Description |
|---|---|
| `--version` | Engine version to pin |
| `--renderer` | `forward_plus`, `mobile`, `gl_compatibility` |
| `--csharp` | Create a C# project (pins to mono build) |
| `--template` | Built-in template (`2d`, `3d`) or git repo (`user/repo` or URL) |

### gdt completion

```sh
gdt completion bash
gdt completion zsh
gdt completion fish
gdt completion powershell
```

Generate shell completion scripts. The install script offers to set up completions automatically.

Manual setup:

```sh
# Bash
gdt completion bash > /etc/bash_completion.d/gdt

# Zsh
gdt completion zsh > "${fpath[1]}/_gdt"

# Fish
gdt completion fish > ~/.config/fish/completions/gdt.fish
```

---

## Project Scaffolding

`gdt new` creates a new Godot project with all standard files.

```sh
# Interactive mode (prompts for all options)
gdt new

# Fully specified
gdt new mygame --version 4.3 --renderer forward_plus

# C# project (pins to mono build)
gdt new mygame --version 4.3 --renderer mobile --csharp

# From a built-in template
gdt new mygame --template 2d --version 4.3
gdt new mygame --template 3d --version 4.3

# From a template repository
gdt new mygame --template user/repo --version 4.3
gdt new mygame --template https://github.com/user/repo --version 4.3
```

### Generated Files

**Default mode** creates:

```
mygame/
  project.godot       # Godot project file with renderer config
  .godot-version      # Pinned engine version
  .gitignore           # Godot + build ignores
  .editorconfig        # Tab/space settings for Godot files
```

**C# mode** (`--csharp`) also creates:

```
  mygame.csproj       # Godot.NET.Sdk project
  mygame.sln          # Visual Studio solution
```

The `.godot-version` file is set to `<version>-mono` for C# projects.

**Built-in templates** (`--template 2d` or `--template 3d`) scaffold a starter project with pre-configured scene files.

**Git template mode** (`--template user/repo`) clones the repository, removes `.git/`, and overwrites `.godot-version`.

### Renderers

| Value | Description |
|---|---|
| `forward_plus` | Best quality, desktop GPUs |
| `mobile` | Balanced, mobile-friendly |
| `gl_compatibility` | Widest support, OpenGL |

---

## LSP and DAP Proxy

gdt provides stdin/stdout to TCP proxies that bridge to Godot's built-in language server and debugger. The proxy starts Godot headless, connects to its TCP port, and exposes a standard stdio interface — compatible with any tool that can spawn an LSP command.

### Default Ports

| Service | Default Port | Flag |
|---|---|---|
| LSP | 6005 | `--port` |
| DAP | 6006 | `--port` |

### Editors

#### Neovim

```lua
-- LSP
require('lspconfig').gdscript.setup({
  cmd = { 'gdt', 'lsp' },
  filetypes = { 'gdscript', 'gd' },
  root_dir = require('lspconfig.util').root_pattern('project.godot'),
})

-- DAP
require('dap').adapters.godot = {
  type = 'pipe',
  pipe = { 'gdt', 'dap' },
}
require('dap').configurations.gdscript = {
  { type = 'godot', request = 'launch', name = 'Launch Godot' },
}
```

#### Helix

In `~/.config/helix/languages.toml`:

```toml
[[language]]
name = "gdscript"
language-servers = ["gdscript"]

[language-server.gdscript]
command = "gdt"
args = ["lsp"]
```

#### VS Code

Install the [godot-tools](https://marketplace.visualstudio.com/items?itemName=geequlim.godot-tools) extension, then in `.vscode/settings.json`:

```json
{
  "gdscript.lsp.serverPort": 6005
}
```

VS Code connects directly to Godot's TCP server — no proxy needed. Start Godot headless first: `gdt run --headless` or open the editor.

#### Zed

In Zed settings (`~/.config/zed/settings.json`):

```json
{
  "lsp": {
    "gdscript": {
      "binary": { "path": "gdt", "arguments": ["lsp"] }
    }
  }
}
```

#### Emacs (lsp-mode)

```elisp
(with-eval-after-load 'lsp-mode
  (lsp-register-client
   (make-lsp-client
    :new-connection (lsp-stdio-connection '("gdt" "lsp"))
    :major-modes '(gdscript-mode)
    :server-id 'gdscript)))
```

### AI Coding Tools

`gdt lsp` works as a standard stdio LSP server, making it compatible with AI coding assistants.

#### Claude Code

```sh
claude mcp add godot-lsp --command gdt --args lsp
```

Or in `.mcp.json`:

```json
{
  "mcpServers": {
    "godot-lsp": {
      "command": "gdt",
      "args": ["lsp"]
    }
  }
}
```

#### Codex

In `codex.json` or equivalent config:

```json
{
  "lsp": {
    "gdscript": {
      "command": ["gdt", "lsp"]
    }
  }
}
```

#### Gemini CLI

```json
{
  "lsp": {
    "gdscript": {
      "command": "gdt",
      "args": ["lsp"]
    }
  }
}
```

For any tool that supports spawning an LSP server via stdio, use `gdt lsp` as the command. Add `--port <N>` to change the Godot TCP port, or `-C <path>` to specify the project directory.

---

## Export Automation

`gdt export` wraps Godot's headless export with version resolution and template validation.

```sh
# List configured presets
gdt export --list

# Export a preset
gdt export "Linux/X11"

# Custom output directory
gdt export "Linux/X11" --output ./build

# Debug export
gdt export "Windows Desktop" --debug

# Show Godot engine output during export
gdt export "Linux/X11" --verbose
```

Export presets must be configured in the Godot editor first (`Project > Export`). gdt reads `export_presets.cfg` and delegates to `godot --headless --export-release`.

By default, engine output is captured silently. On failure, captured stdout/stderr is included in the error message. Use `--verbose` (`-v`) to stream engine output in real-time.

The default output directory is `dist/<preset-name>/`.

### Plugin Hooks

Plugins can hook into the export lifecycle:

| Hook | Timing |
|---|---|
| `before_export` | Before Godot export starts |
| `after_export` | After successful export |
| `before_build` | Before a build step |

See [Plugins](#plugins) for hook configuration.

---

## CI Setup

Generate CI pipeline files for automated exports.

```sh
# Interactive provider selection
gdt ci setup

# Specify provider directly
gdt ci setup --provider github
gdt ci setup --provider gitlab
gdt ci setup --provider generic
```

### Providers

| Provider | Output File |
|---|---|
| `github` | `.github/workflows/export.yml` |
| `gitlab` | `.gitlab-ci.yml` |
| `generic` | `ci/export.sh` |

All templates follow the same flow: install gdt, install engine from `.godot-version`, install export templates, run export.

### GitHub Actions Example

```yaml
name: Export Game

on:
  push:
    branches: [main]

jobs:
  export:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Install gdt
        run: curl -fsSL https://raw.githubusercontent.com/monkeymonk/gdt/main/scripts/install.sh | sh
      - name: Install engine
        run: gdt install
      - name: Install templates
        run: gdt templates install $(cat .godot-version)
      - name: Export
        run: gdt export Linux/X11
```

Note: `gdt install` with no arguments reads `.godot-version` from the project.

---

## Version Resolution

When resolving which Godot version to use, gdt checks (in order):

1. `.godot-version` file in the current directory
2. `.godot-version` in parent directories
3. `GDT_GODOT_VERSION` environment variable
4. Global default set by `gdt use`
5. Latest installed version

## Configuration

Configuration file: `~/.gdt/config.toml`

```toml
default_version = "4.3"
mirrors = [
  "https://mirror.example.com/godot/releases",
]
```

### Mirrors

When the primary GitHub download URL is unavailable, gdt tries configured mirror URLs as fallbacks. Mirrors are checked with HEAD requests before downloading.

### Environment Variables

| Variable | Description |
|---|---|
| `GDT_HOME` | Override base directory (default: `~/.gdt`) |
| `GDT_GODOT_VERSION` | Override resolved engine version |
| `GDT_GITHUB_TOKEN` | GitHub API token (avoids rate limits) |
| `GDT_DEBUG` | Enable debug logging (`1` to enable) |

---

## Plugins

gdt supports plugins distributed as Git repositories with prebuilt binaries.

```sh
# Install a plugin
gdt plugin install user/repo

# Use the plugin command
gdt assets optimize

# Update all plugins
gdt plugin update

# Create your own plugin
gdt plugin new mytools
```

Plugins receive context via environment variables:

| Variable | Description |
|---|---|
| `GDT_HOME` | gdt base directory |
| `GDT_PROJECT_ROOT` | Detected Godot project root |
| `GDT_GODOT_VERSION` | Resolved engine version |
| `GDT_ENGINE_PATH` | Absolute path to engine binary |

### Plugin Manifest

Plugins must include a `plugin.toml` in their root:

```toml
name = "mytools"
version = "0.1.0"
commands = ["mytools"]
requires_gdt = ">=1.0"
description = "My custom tools"

[hooks]
before_export = "scripts/pre-export.sh"
after_export = "scripts/post-export.sh"
```

### Hooks

Plugins can register shell commands for lifecycle events:

| Event | Description |
|---|---|
| `before_export` | Runs before `gdt export` starts |
| `after_export` | Runs after a successful export |
| `before_build` | Runs before a build step |

Hook exit codes:
- **Exit 0**: Success
- **Exit 2**: Fatal error — stops the operation
- **Other non-zero**: Warning — logged but does not stop the operation

---

## Desktop Launcher (Linux)

On Linux, gdt creates a `.desktop` launcher so Godot appears in your system application menu (GNOME, KDE, etc.).

The launcher is created automatically on first `gdt install`. It uses `gdt run` under the hood, so version resolution works as usual — opening a `project.godot` file from your file manager will launch the correct engine version.

```sh
# Launcher is created automatically on first install
gdt install 4.3

# Launcher is removed when the last version is uninstalled
gdt remove 4.3
```

The desktop file is installed to `~/.local/share/applications/gdt-godot.desktop`.

---

## Platform Support

| OS | Architecture | Status |
|---|---|---|
| Linux | x86_64 | Supported |
| macOS | x86_64 / arm64 | Supported |
| Windows | x86_64 | Supported |

## License

MIT License. See [LICENSE](LICENSE) for details.
