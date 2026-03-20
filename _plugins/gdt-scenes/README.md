# gdt-scenes

Scene graph inspection tools for Godot projects. A plugin for [gdt](https://github.com/monkeymonk/gdt).

## Installation

```bash
gdt plugin install scenes
```

Or build manually:

```bash
make build
make install
```

## Usage

### List all scenes

```bash
gdt scenes list
```

Lists all `.tscn` files in the project (skips `.godot/` directory). Requires `GDT_PROJECT_ROOT` to be set.

### Display scene tree

```bash
gdt scenes tree path/to/scene.tscn
```

Parses a `.tscn` file and displays the node hierarchy:

```
MyScene (Node2D)
  Player (CharacterBody2D)
    Sprite (Sprite2D)
    CollisionShape (CollisionShape2D)
  Camera (Camera2D)
```

## Environment Variables

- `GDT_PROJECT_ROOT` - Path to the Godot project root (required for `list` command)
