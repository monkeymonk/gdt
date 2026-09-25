---
name: gdt
description: "Use when working in a Godot project that manages its engine version, exports, or plugins via `gdt` (present if a `.godot-version` file exists, or `gdt` is on PATH). Covers non-interactive invocation, version resolution, plugin protocol, and troubleshooting — avoids the interactive-prompt hang that a naive invocation can trigger."
---

# gdt — Godot Developer Toolchain

`gdt` is a cross-platform CLI that manages Godot Engine installations,
scaffolds projects, proxies LSP/DAP for editors, automates exports, and
generates CI pipelines. Full reference: `gdt --help`, `gdt <command>
--help`, or the project README (https://github.com/monkeymonk/gdt).

This skill exists for one reason: an agent that invokes `gdt` the way a
human would at an interactive terminal can hang or misbehave. Read
"Non-interactive invocation" before running any `gdt` command.

## Non-interactive invocation (read this first)

Most `gdt` subcommands prompt interactively when a required
argument/flag is missing **and stdin looks like a terminal**. As an
agent, you are not a human at a terminal — but some sandboxed
environments present a pty-like stdin that still satisfies that check,
so **do not rely on non-interactivity being detected automatically**.
Always pass every required argument and flag explicitly.

Commands with an interactive fallback, and what to pass instead:

| Command | Always pass |
|---|---|
| `gdt install [version]` | the version (or `latest`/`stable`) |
| `gdt use [version]` | the version |
| `gdt local [version]` | the version |
| `gdt remove [version]` | the version — and expect it to proceed **without confirmation** when stdin isn't a real interactive terminal (the confirm-prompt step is skipped, not defaulted to "no") |
| `gdt templates install\|remove [version]` | the version — same no-confirmation behavior as `remove` |
| `gdt export [preset]` | the preset name (`gdt export --list` to discover presets first) |
| `gdt plugin install [repo]` | the repo (`owner/repo` or URL) |
| `gdt plugin remove [name]` | the name — same no-confirmation behavior |
| `gdt plugin new [name]` | the name |
| `gdt new [name]` | `name` **and** `--version` **and** `--renderer` (this one is *not* gated behind a TTY check the way the others are — it will attempt to prompt for any of these left unset regardless) |

If you must discover a value before proceeding (e.g. which versions are
installed), query it first (`gdt list`, `gdt ls-remote`,
`gdt templates list`) and pass the result explicitly — never invoke a
command bare and hope it degrades gracefully.

`gdt doctor` and `gdt list`/`gdt ls-remote` never prompt; safe to run
freely for discovery.

## Version resolution

When a command needs to know "which Godot version," it checks, in
order:

1. `.godot-version` in the current directory
2. `.godot-version` in a parent directory (walks up)
3. `GDT_GODOT_VERSION` environment variable
4. The global default (`gdt use <version>`)
5. The latest installed version

Before assuming which engine a command will use, check for a
`.godot-version` file in the project directory — it silently overrides
everything except an explicit CLI argument. `gdt local <version>`
writes/updates that file; `gdt use <version>` changes the global
default instead (no file write, affects every project with no
`.godot-version` of its own).

`gdt list` and `gdt ls-remote` print versions newest to oldest.

## Common tasks

```sh
# Discover, then install explicitly (never bare `gdt install`)
gdt ls-remote
gdt install 4.3

# Pin a project (writes .godot-version)
gdt local 4.3

# Scaffold a new project — pass every field
gdt new mygame --version 4.3 --renderer forward_plus
gdt new mygame --version 4.3 --renderer forward_plus --csharp   # C#
gdt new mygame --template 2d --version 4.3                     # built-in template

# Export (presets must already exist in export_presets.cfg — created via the Godot editor, not gdt)
gdt export --list
gdt export "Linux/X11"

# CI pipeline generation
gdt ci setup --provider github   # or gitlab, generic

# Diagnose a broken install/config
gdt doctor
```

## Environment variables

| Variable | Effect |
|---|---|
| `GDT_HOME` | Override base directory (default `~/.gdt`) |
| `GDT_GODOT_VERSION` | Override resolved engine version |
| `GITHUB_TOKEN` | Avoids GitHub API rate limits |
| `GDT_DEBUG=1` | Verbose diagnostic logging on stderr — set this first when a command's behavior is unclear, before guessing |

Add `--refresh` to `install`/`ls-remote`/`templates install` to bypass
the 24h metadata cache when you need current release data immediately.

## Plugins

Plugins are Git-repo-distributed, with prebuilt binaries. Two manifest
protocols coexist (`plugin.toml`'s `protocol` field: `1` legacy,
`2` structured `[contributions]`) — don't assume one or the other.
Plugin contributions (templates, presets, CI providers) are namespaced
`<plugin>:<name>`; an unqualified name only resolves if exactly one
installed plugin provides it.

```sh
gdt plugin list                    # see what's installed and what it contributes
gdt plugin install owner/repo
gdt new mygame --template pluginname:templatename --version 4.3
```

## LSP/DAP for code intelligence

`gdt lsp` / `gdt dap` bridge stdio to Godot's built-in language
server/debugger — this is *code intelligence* (GDScript completions,
diagnostics), not project/build control. If you're an AI coding tool
that spawns LSP servers directly (Zed, Neovim, Helix, Claude Code via
its plugin system), point your LSP config at `gdt lsp`. If your tool
only supports MCP and not spawning LSP servers directly (some Codex/
Gemini CLI configurations), use a generic LSP→MCP bridge in front of
`gdt lsp` rather than expecting `gdt` to speak MCP itself — `gdt` does
not ship an MCP server by design (see the project's `DEFERRED.md`: a
dedicated third-party Godot MCP server already covers project-state
control better than a thin wrapper around this CLI would).

Full per-editor setup snippets are in the project README's "LSP and DAP
Proxy" section — read that before hand-writing a config block.

## Troubleshooting

- Command fails with no clear reason → `GDT_DEBUG=1 gdt <command> ...`
  first, `gdt doctor` second.
- `gdt export` fails immediately → check `export_presets.cfg` exists
  (created via the Godot editor's Project → Export dialog, not by
  `gdt`) and that export templates are installed
  (`gdt templates install <version>`).
- A version-dependent command behaves unexpectedly → check for a
  `.godot-version` in the current or a parent directory before assuming
  the global default or `GDT_GODOT_VERSION` applies.
