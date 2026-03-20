# gdt-assets

Asset optimization and auditing plugin for [gdt](https://github.com/monkeymonk/gdt) (Godot Developer Toolchain).

## Installation

```bash
gdt plugin install assets
```

Or build manually:

```bash
make build
make install
```

## Usage

### Audit

Scan all asset files in the project and report sizes:

```bash
gdt assets audit
```

Output:

```
  OK    player.png (245 KB)
  WARN  level.glb (5.2 MB)
  FAIL  world.glb (15.3 MB)

Total: 3 files, 20.7 MB
Warnings: 1  Errors: 1
```

- **OK** - Under 1 MB
- **WARN** - Between 1 MB and 10 MB
- **FAIL** - Over 10 MB (exits with code 1)

### Optimize

Show which files would benefit from optimization:

```bash
gdt assets optimize
```

This lists all files over 1 MB that could be reduced with external tools.

## Supported file types

Images: `.png`, `.jpg`, `.svg`
Audio: `.wav`, `.ogg`, `.mp3`
3D: `.glb`, `.gltf`, `.obj`, `.fbx`
Godot: `.tres`, `.tscn`

## Environment

- `GDT_PROJECT_ROOT` - Project directory to scan (defaults to current directory)

## License

MIT
