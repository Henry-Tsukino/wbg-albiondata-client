#!/usr/bin/env bash
set -eo pipefail
cd "$(dirname "$0")/.."
source scripts/version.sh "$@"

# ---- Build dependencies (installed only if missing) --------------------------
# Arch: libpcap gtk4 webkitgtk-6.0 patchelf | Debian/Ubuntu | Fedora
missing_deps() {
  command -v patchelf >/dev/null 2>&1 || return 0
  command -v pkg-config >/dev/null 2>&1 || return 0
  pkg-config --exists libpcap gtk4 webkitgtk-6.0 || return 0
  return 1
}

if missing_deps; then
  echo "==> Installing build dependencies (sudo)"
  if command -v pacman >/dev/null 2>&1; then
    sudo pacman -S --needed --noconfirm pkgconf libpcap gtk4 webkitgtk-6.0 patchelf
  elif command -v apt-get >/dev/null 2>&1; then
    sudo apt-get update
    sudo apt-get install -y pkg-config libpcap-dev patchelf libgtk-4-dev libwebkitgtk-6.0-dev
  elif command -v dnf >/dev/null 2>&1; then
    sudo dnf install -y pkgconf libpcap-devel patchelf gtk4-devel webkitgtk6.0-devel
  else
    echo "ERROR: unknown package manager. Install: libpcap, gtk4, webkitgtk-6.0, patchelf, pkg-config" >&2
    exit 1
  fi
fi

source scripts/legacy.sh
archive_old_builds   # previous dist/ -> legacy/dist/ (nothing is deleted)
(cd frontend && npm ci && npm run build)
go build -ldflags "$LDFLAGS" -o "dist/${APP_BIN}" wbg-client.go
patchelf --replace-needed libpcap.so.0.8 libpcap.so "dist/${APP_BIN}"

"./dist/${APP_BIN}" -version

gzip -9 -c "dist/${APP_BIN}" > dist/update-linux-amd64.gz
