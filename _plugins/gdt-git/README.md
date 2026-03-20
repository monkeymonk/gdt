# gdt-git

Git workflow utilities plugin for [gdt](https://github.com/monkeymonk/gdt) (Godot Developer Toolchain).

## Installation

```bash
gdt plugin install git
```

Or build manually:

```bash
make build
make install
```

## Usage

### Setup Git configuration

Creates a recommended `.gitignore` and `.gitattributes` for Godot projects:

```bash
gdt git setup
gdt git setup --force  # overwrite existing files
```

### Install Git hooks

Installs a pre-commit hook that warns about oversized files (>10MB) and validates `project.godot` exists:

```bash
gdt git hooks
```

## Environment

- `GDT_PROJECT_ROOT` - path to the Godot project root (set automatically by gdt)
