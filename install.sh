#!/bin/sh
# Install the maddock agent.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/MadJlzz/maddock/main/install.sh | sh
#   curl -fsSL https://raw.githubusercontent.com/MadJlzz/maddock/main/install.sh | sh -s -- v1.2.0
#
# Environment variables:
#   INSTALL_DIR   Installation directory (default: /usr/local/bin)

set -eu

REPO="MadJlzz/maddock"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
VERSION="${1:-latest}"

log() {
    printf '==> %s\n' "$*"
}

err() {
    printf 'error: %s\n' "$*" >&2
    exit 1
}

# Detect OS.
os=$(uname -s | tr '[:upper:]' '[:lower:]')
if [ "$os" != "linux" ]; then
    err "unsupported OS: $os (maddock only supports Linux)"
fi

# Detect architecture.
arch=$(uname -m)
case "$arch" in
    x86_64|amd64) arch="amd64" ;;
    aarch64|arm64) arch="arm64" ;;
    *) err "unsupported architecture: $arch" ;;
esac

# Resolve "latest" to a concrete tag via GitHub's redirect.
if [ "$VERSION" = "latest" ]; then
    log "Resolving latest version..."
    VERSION=$(curl -fsSLI -o /dev/null -w '%{url_effective}' "https://github.com/$REPO/releases/latest" | sed 's|.*/tag/||')
    [ -n "$VERSION" ] || err "could not resolve latest version"
fi

binary_name="maddock-agent-${os}-${arch}"
url="https://github.com/${REPO}/releases/download/${VERSION}/${binary_name}"

log "Downloading ${binary_name} (${VERSION})..."
tmpdir=$(mktemp -d)
trap 'rm -rf "$tmpdir"' EXIT

if ! curl -fsSL -o "$tmpdir/maddock-agent" "$url"; then
    err "failed to download $url"
fi

chmod +x "$tmpdir/maddock-agent"

# Choose how to install (sudo only if needed).
target="$INSTALL_DIR/maddock-agent"
if [ -w "$INSTALL_DIR" ]; then
    mv "$tmpdir/maddock-agent" "$target"
else
    log "Need elevated permissions to write to $INSTALL_DIR"
    sudo mv "$tmpdir/maddock-agent" "$target"
fi

log "Installed $target"
"$target" --version
