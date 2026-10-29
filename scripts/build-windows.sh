#!/usr/bin/env bash
# Builds the Windows client, the self-update asset and the Inno Setup installer.
#   VERSION=1.5.0 scripts/build-windows.sh      (or: scripts/build-windows.sh 1.5.0)
#
# Outputs (dist/):
#   WBG-albion-data-client.exe          the client
#   update-windows-amd64.exe.gz         asset the self-updater downloads (upload to the GitHub release)
#   installer-WBG-tool.exe              installer built from installer/wbg.iss
#
# Inno Setup compiler (ISCC.exe, NOT the ISIDE.exe GUI) is located automatically
# (Inno Setup 6 and 7). Override with environment variables, e.g.:
#   ISCC="$HOME/PortProton/prefixes/DEFAULT/drive_c/Inno Setup 7/ISCC.exe" \
#   WINEPREFIX="$HOME/PortProton/prefixes/DEFAULT" scripts/build-windows.sh 1.5.0
# Or put those variables once into scripts/local.env (git-ignored).
# WINE selects the wine binary (default: wine). Do NOT put "set ISCC=..." in this
# file: in bash `set` overwrites the script arguments ($1) and breaks the version.

set -eo pipefail
cd "$(dirname "$0")/.."

# Optional machine-specific settings (git-ignored). Example scripts/local.env:
#   ISCC="$HOME/PortProton/prefixes/DEFAULT/drive_c/Inno Setup 7/ISCC.exe"
#   WINE=/path/to/wine
# Loaded BEFORE version.sh and without touching the script arguments.
if [ -f scripts/local.env ]; then
  # shellcheck disable=SC1091
  source scripts/local.env
fi

# Inno Setup compiler path. Empty by default => find_iscc() below searches the usual
# places (Windows runner, Wine/PortProton/Bottles prefixes). Do NOT hard-code a
# machine-specific path here: it breaks CI. Override with ISCC=... or scripts/local.env.
# Plain assignment on purpose: `set ISCC=...` would overwrite the script arguments.
ISCC="${ISCC:-}"

source scripts/version.sh "$@"

rm -rf dist rsrc_windows_*
mkdir -p dist

go install github.com/tc-hib/go-winres@v0.3.1
export PATH="$PATH:$(go env GOPATH)/bin"

(cd frontend && npm ci && npm run build)

CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "$LDFLAGS" -o "dist/${APP_BIN}.exe" wbg-client.go

# Icon + version info inside the exe
go-winres patch --product-version "${APP_VERSION}.0" --file-version "${APP_VERSION}.0" "dist/${APP_BIN}.exe"

# Self-update asset (exact name required by the updater)
gzip -9 -c "dist/${APP_BIN}.exe" > dist/update-windows-amd64.exe.gz

# ---- Installer -------------------------------------------------------------
# SKIP_INSTALLER=1 builds only the exe + update asset (no Inno Setup needed).
if [ "${SKIP_INSTALLER:-0}" = "1" ]; then
  echo "==> SKIP_INSTALLER=1: installer step skipped"
  ls -la dist
  exit 0
fi

WINE_BIN="${WINE:-wine}"

find_iscc() {
  if [ -n "$ISCC" ]; then echo "$ISCC"; return; fi
  if command -v iscc >/dev/null 2>&1; then command -v iscc; return; fi
  if command -v ISCC.exe >/dev/null 2>&1; then command -v ISCC.exe; return; fi
  local drive v c
  # "/c" = native Windows (Git Bash); the rest are Wine/PortProton/Bottles/Lutris prefixes.
  for drive in "/c" \
      ${WINEPREFIX:+"$WINEPREFIX/drive_c"} \
      "$HOME/.wine/drive_c" \
      "$HOME"/PortProton/prefixes/*/drive_c \
      "$HOME"/.local/share/PortProton/prefixes/*/drive_c \
      "$HOME"/.var/app/com.usebottles.bottles/data/bottles/bottles/*/drive_c \
      "$HOME"/Games/*/drive_c; do
    for v in 7 6; do
      for c in "$drive/Inno Setup $v/ISCC.exe" \
               "$drive/Program Files/Inno Setup $v/ISCC.exe" \
               "$drive/Program Files (x86)/Inno Setup $v/ISCC.exe"; do
        [ -f "$c" ] && { echo "$c"; return; }
      done
    done
  done
}

ISCC_PATH="$(find_iscc || true)"
if [ -z "$ISCC_PATH" ]; then
  echo "ERROR: Inno Setup compiler (ISCC.exe) not found. Set ISCC=/path/to/ISCC.exe" >&2
  exit 1
fi
if [[ "$(basename "$ISCC_PATH")" != [Ii][Ss][Cc][Cc]* ]]; then
  echo "ERROR: ISCC points to '$ISCC_PATH' - that is not the compiler." >&2
  echo "       Use ISCC.exe from the same folder (ISIDE.exe is the GUI editor)." >&2
  exit 1
fi

# ISCC lives inside a Wine prefix (".../drive_c/...") -> use that prefix automatically.
if [ -z "$WINEPREFIX" ] && [[ "$ISCC_PATH" == */drive_c/* ]]; then
  export WINEPREFIX="${ISCC_PATH%%/drive_c/*}"
fi
echo "==> Inno Setup: $ISCC_PATH${WINEPREFIX:+ (WINEPREFIX=$WINEPREFIX)}"

case "$(uname -s)" in
  MINGW*|MSYS*|CYGWIN*)
    "$ISCC_PATH" "//DMyAppVersion=${APP_VERSION}" installer/wbg.iss ;;
  *)
    if [[ "$ISCC_PATH" == *.exe ]]; then
      "$WINE_BIN" "$ISCC_PATH" "/DMyAppVersion=${APP_VERSION}" "$("$WINE_BIN" winepath -w "$PWD/installer/wbg.iss")"
    else
      "$ISCC_PATH" "/DMyAppVersion=${APP_VERSION}" installer/wbg.iss
    fi ;;
esac

ls -la dist
