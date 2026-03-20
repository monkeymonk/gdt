#!/bin/sh
set -e

REPO="monkeymonk/gdt"
INSTALL_DIR="${GDT_HOME:-$HOME/.gdt}"
BIN_DIR="${INSTALL_DIR}/bin"

main() {
    platform="$(detect_platform)"
    arch="$(detect_arch)"
    version="$(get_latest_version)"

    if [ -z "$version" ]; then
        echo "Error: could not determine latest version"
        exit 1
    fi

    echo "Installing gdt ${version} for ${platform}/${arch}..."

    artifact="gdt-${version}-${platform}-${arch}.tar.gz"
    url="https://github.com/${REPO}/releases/download/v${version}/${artifact}"

    tmpdir="$(mktemp -d)"
    trap 'rm -rf "$tmpdir"' EXIT

    echo "Downloading ${url}..."
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "$url" -o "${tmpdir}/${artifact}"
    elif command -v wget >/dev/null 2>&1; then
        wget -q "$url" -O "${tmpdir}/${artifact}"
    else
        echo "Error: curl or wget required"
        exit 1
    fi

    mkdir -p "$BIN_DIR"
    tar -xzf "${tmpdir}/${artifact}" -C "$BIN_DIR"
    chmod +x "${BIN_DIR}/gdt"

    # Create shims directory
    mkdir -p "${INSTALL_DIR}/shims"

    # Create godot shim symlink
    ln -sf "${BIN_DIR}/gdt" "${INSTALL_DIR}/shims/godot"

    echo ""
    echo "gdt ${version} installed to ${BIN_DIR}/gdt"
    echo ""
    echo "Add gdt to your PATH by adding this to your shell profile:"
    echo ""
    echo "  export PATH=\"${BIN_DIR}:\${INSTALL_DIR}/shims:\$PATH\""
    echo ""
    echo "Or run:"
    echo ""
    echo "  eval \"\$(${BIN_DIR}/gdt shell init)\""
    echo ""
}

detect_platform() {
    case "$(uname -s)" in
        Linux*)  echo "linux" ;;
        Darwin*) echo "darwin" ;;
        *)       echo "Error: unsupported platform $(uname -s)" >&2; exit 1 ;;
    esac
}

detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64)  echo "amd64" ;;
        aarch64|arm64) echo "arm64" ;;
        *)             echo "Error: unsupported architecture $(uname -m)" >&2; exit 1 ;;
    esac
}

get_latest_version() {
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed 's/.*"v\(.*\)".*/\1/'
    elif command -v wget >/dev/null 2>&1; then
        wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed 's/.*"v\(.*\)".*/\1/'
    fi
}

main "$@"
