# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| latest  | Yes       |

## Reporting a Vulnerability

If you discover a security vulnerability in gdt, please open a [GitHub Issue](https://github.com/monkeymonk/gdt/issues/new) with:

- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

## Scope

gdt downloads binaries from GitHub releases (godotengine/godot-builds). Security concerns include:

- Binary integrity during download (SHA-512 checksum verification)
- Path traversal in archive extraction
- Plugin execution (plugins run arbitrary binaries)
- Shell injection in shim or proxy execution

## Disclosure

Once a fix is released, we will:

1. Publish a GitHub Security Advisory
2. Release a patched version
3. Credit the reporter (unless anonymity is requested)
