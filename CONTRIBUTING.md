# Contributing to gdt

## Getting Started

1. Fork the repository
2. Clone your fork
3. Create a feature branch

## Development

### Prerequisites

- Go 1.23+
- Git

### Building

```sh
go build -ldflags "-X main.Version=dev" -o gdt ./cmd/gdt
```

### Testing

```sh
go test ./...
```

### Running

```sh
go run ./cmd/gdt
```

## Code Guidelines

- Follow standard Go conventions (`gofmt`, `go vet`)
- Keep the dependency flow: `cli/ -> services -> infrastructure`
- No global state — pass config structs through functions
- Return structured errors with actionable hints
- Write table-driven tests
- No external test dependencies — use stdlib `testing` package

## Pull Requests

- One feature or fix per PR
- Include tests for new functionality
- Keep commits focused and well-described
- Ensure `go test ./...` passes
- Ensure `go vet ./...` is clean

## Architecture

```
cli/ → services (engine, plugins, project) → infrastructure (config, metadata, download, platform)
```

Key packages:
- `engine` — version management, install/remove, binary resolution, export templates
- `plugins` — discovery, manifest, contributions, hooks, namespace, scaffold
- `project` — project detection, scaffolding, preset parsing
- `ci` — CI pipeline generation (GitHub Actions, GitLab CI, generic shell)
- `proxy` — LSP/DAP TCP-to-stdio bridge
- `config`, `download`, `metadata`, `platform`, `selfupdate`

## Plugin Development

Plugins use the V2 protocol with a `[contributions]` table in `plugin.toml`. Scaffold a new plugin with:

```sh
gdt plugin new myplugin
```

See `_plugins/` for example plugin sources.

## Reporting Issues

- Use GitHub Issues
- Include: Go version, OS, gdt version, steps to reproduce
