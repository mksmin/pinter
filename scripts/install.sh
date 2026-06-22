#!/bin/sh

set -eu

REPO="mksmin/pinter"
VERSION="${PINTER_VERSION:-latest}"
INSTALL_DIR="${PINTER_INSTALL_DIR:-$HOME/.local/bin}"

os="$(uname -s)"
arch="$(uname -m)"

case "$os" in
  Darwin) os_name="macos" ;;
  Linux) os_name="linux" ;;
  *) echo "Unsupported OS: $os" >&2; exit 1 ;;
esac

case "$arch" in
  x86_64|amd64) arch_name="amd64" ;;
  arm64|aarch64) arch_name="arm64" ;;
  *) echo "Unsupported arch: $arch" >&2; exit 1 ;;
esac

asset="pinter-${os_name}-${arch_name}"

if [ "$VERSION" = "latest" ]; then
  url="https://github.com/${REPO}/releases/latest/download/${asset}"
else
  url="https://github.com/${REPO}/releases/download/${VERSION}/${asset}"
fi

mkdir -p "$INSTALL_DIR"
tmp="$(mktemp)"

echo "Downloading $url"
curl -fsSL "$url" -o "$tmp"
chmod +x "$tmp"
mv "$tmp" "$INSTALL_DIR/pinter"

echo "Installed: $INSTALL_DIR/pinter"
echo "Run: pinter help"