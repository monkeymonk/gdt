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
- Transparent shim system — just run `godot`
- Project scaffolding with interactive prompts
- LSP and DAP proxy for Neovim, Helix, and other editors
- Export automation with preset management
- CI pipeline generation (GitHub Actions, GitLab CI, shell script)
- Export template management
- Plugin ecosystem for extensibility
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

Add the shim directory to your PATH:

```sh
# Add to your .bashrc, .zshrc, etc.
eval "$(gdt shell init)"
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

# Open the editor
cd mygame
gdt edit

# Run the game
gdt run

# Pin a project to a specific version
gdt local 4.2

# Run Godot through the shim (resolves the right version automatically)
godot .
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
| `gdt templates list` | List installed templates |

### Plugins

| Command | Description |
|---|---|
| `gdt plugin install <repo>` | Install a plugin |
| `gdt plugin list` | List installed plugins |
| `gdt plugin update [name]` | Update plugins (all or by name) |
| `gdt plugin remove <name>` | Remove a plugin |
| `gdt plugin new <name>` | Scaffold a new plugin |

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
| `--template` | Clone from a template repository (`user/repo` or URL) |

### gdt completion

```sh
gdt completion bash
gdt completion zsh
gdt completion fish
gdt completion powershell
```

Generate shell completion scripts. Example setup:

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

**Template mode** (`--template`) clones the repository, removes `.git/`, and overwrites `.godot-version`.

### Renderers

| Value | Description |
|---|---|
| `forward_plus` | Best quality, desktop GPUs |
| `mobile` | Balanced, mobile-friendly |
| `gl_compatibility` | Widest support, OpenGL |

---

## LSP and DAP Proxy

gdt provides stdin/stdout to TCP proxies for editors that need them (Neovim, Helix, etc.). The proxy automatically starts Godot headless, connects to its TCP server, and bridges stdin/stdout.

### Neovim LSP

```lua
require('lspconfig').gdscript.setup({
  cmd = { 'gdt', 'lsp' },
})

-- Custom port
require('lspconfig').gdscript.setup({
  cmd = { 'gdt', 'lsp', '--port', '6010' },
})
```

### Neovim DAP

```lua
require('dap').adapters.godot = {
  type = 'pipe',
  pipe = { 'gdt', 'dap' },
}

-- Custom port
require('dap').adapters.godot = {
  type = 'pipe',
  pipe = { 'gdt', 'dap', '--port', '6011' },
}

require('dap').configurations.gdscript = {
  {
    type = 'godot',
    request = 'launch',
    name = 'Launch Godot',
  },
}
```

### Helix

In `languages.toml`:

```toml
[language-server.gdscript]
command = "gdt"
args = ["lsp"]

# Custom port
# args = ["lsp", "--port", "6010"]
```

### VS Code

VS Code connects directly to Godot's TCP server — no proxy needed:

```json
{
  "gdscript.lsp.serverPort": 6005
}
```

### Default Ports

| Service | Default Port | Flag |
|---|---|---|
| LSP | 6005 | `--port` |
| DAP | 6006 | `--port` |

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
```

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
```

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
