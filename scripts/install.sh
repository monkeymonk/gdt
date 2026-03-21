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

    echo ""
    echo "gdt ${version} installed to ${BIN_DIR}/gdt"
    echo ""
    echo "Add gdt to your PATH by adding this to your shell profile:"
    echo ""
    echo "  eval \"\$(${BIN_DIR}/gdt shell init)\""
    echo ""
    echo "Optional: create a godot alias:"
    echo ""
    echo "  alias godot=\"gdt run\""
    echo ""

    setup_completions
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

setup_completions() {
    # Skip if not interactive
    if [ ! -t 0 ]; then
        return
    fi

    printf "Install shell completions? [y/N] "
    read -r answer
    case "$answer" in
        [yY]|[yY][eE][sS]) ;;
        *) return ;;
    esac

    shell="$(detect_shell)"
    case "$shell" in
        zsh)
            comp_dir="${ZDOTDIR:-$HOME}/.zfunc"
            mkdir -p "$comp_dir"
            "${BIN_DIR}/gdt" completion zsh > "${comp_dir}/_gdt"
            echo "Zsh completions installed to ${comp_dir}/_gdt"
            echo ""
            echo "  Ensure this is in your .zshrc:"
            echo "    fpath=(${comp_dir} \$fpath)"
            echo "    autoload -Uz compinit && compinit"
            ;;
        bash)
            if [ -d /etc/bash_completion.d ] && [ -w /etc/bash_completion.d ]; then
                comp_dir="/etc/bash_completion.d"
            else
                comp_dir="${XDG_DATA_HOME:-$HOME/.local/share}/bash-completion/completions"
                mkdir -p "$comp_dir"
            fi
            "${BIN_DIR}/gdt" completion bash > "${comp_dir}/gdt"
            echo "Bash completions installed to ${comp_dir}/gdt"
            ;;
        fish)
            comp_dir="${XDG_CONFIG_HOME:-$HOME/.config}/fish/completions"
            mkdir -p "$comp_dir"
            "${BIN_DIR}/gdt" completion fish > "${comp_dir}/gdt.fish"
            echo "Fish completions installed to ${comp_dir}/gdt.fish"
            ;;
        *)
            echo "Unsupported shell for completions: $shell"
            ;;
    esac
}

detect_shell() {
    shell_name="$(basename "${SHELL:-/bin/sh}")"
    case "$shell_name" in
        zsh)  echo "zsh" ;;
        fish) echo "fish" ;;
        bash) echo "bash" ;;
        *)    echo "$shell_name" ;;
    esac
}

main "$@"
