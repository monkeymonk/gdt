# gdt-ai

AI development tools and project introspection plugin for [gdt](https://github.com/monkeymonk/gdt).

Exposes Godot project structure as JSON for AI development tools.

## Installation

```sh
make install
```

Or build manually:

```sh
go build -o gdt-ai .
cp gdt-ai ~/.local/bin/
```

## Usage

Set environment variables (provided automatically when invoked via gdt):

```sh
export GDT_PROJECT_ROOT=/path/to/godot/project
export GDT_GODOT_VERSION=4.3
```

### Commands

**inspect** - Full project introspection as JSON:

```sh
gdt-ai inspect
```

Output:

```json
{
  "godot_version": "4.3",
  "project_root": "/path/to/project",
  "scenes": ["main.tscn", "levels/level1.tscn"],
  "scripts": ["player.gd", "enemy.gd"],
  "resources": ["theme.tres"],
  "shaders": ["water.gdshader"]
}
```

**scenes** - List all .tscn files:

```sh
gdt-ai scenes
```

**scripts** - List all .gd and .cs files:

```sh
gdt-ai scripts
```

## Future

- MCP (Model Context Protocol) server support for direct AI agent integration
