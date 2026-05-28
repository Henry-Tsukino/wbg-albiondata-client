#!/usr/bin/env bash
# =============================================================================
#  WBG Albion Data Client — Build Script
#  Supports: native Linux build  OR  cross-compile → Windows (.exe)  OR  both
#
#  Usage:
#    ./build.sh                        # linux по умолчанию
#    ./build.sh --windows              # cross-compile для Windows (.exe)
#    ./build.sh --linux                # native Linux binary
#    ./build.sh --all                  # собрать Linux и Windows
#    ./build.sh --linux -u             # собрать + создать .gz для обновления
#    ./build.sh --windows -u           # собрать + создать .gz для обновления
#    ./build.sh --all -u               # собрать оба + создать .gz для обоих
#    ./build.sh --all -u -v1.3.11      # собрать оба + gz + обновить версию
#    ./build.sh -v1.4.0 --linux -u     # версия + linux + gz
# =============================================================================

set -euo pipefail

# ── Цвета ─────────────────────────────────────────────────────────────────────
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'
CYAN='\033[0;36m'; BOLD='\033[1m'; RESET='\033[0m'

info()    { echo -e "${CYAN}$*${RESET}"; }
success() { echo -e "${GREEN}$*${RESET}"; }
warn()    { echo -e "${YELLOW}$*${RESET}"; }
error()   { echo -e "${RED}ERROR: $*${RESET}" >&2; exit 1; }

# ── Константы ─────────────────────────────────────────────────────────────────
BASE_NAME="WBG-albion-data-client"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# ── Текущая версия из файла (fallback если нет флага -v) ──────────────────────
VERSION_FILE="${SCRIPT_DIR}/version.txt"
if [[ -f "$VERSION_FILE" ]]; then
    APP_VERSION="$(cat "$VERSION_FILE")"
else
    APP_VERSION="1.3.10"
fi

# ── Разбор аргументов ─────────────────────────────────────────────────────────
TARGET="linux"
WITH_UPDATE=false
NEW_VERSION=""

for arg in "$@"; do
    case "$arg" in
        --windows)  TARGET="windows" ;;
        --linux)    TARGET="linux"   ;;
        --all)      TARGET="all"     ;;
        -u)         WITH_UPDATE=true ;;
        -v*)        NEW_VERSION="${arg#-v}" ;;
        --help|-h)
            echo "Usage: $0 [--windows|--linux|--all] [-u] [-v<version>]"
            echo "  --windows      Cross-compile for Windows (.exe)"
            echo "  --linux        Build native Linux binary (default)"
            echo "  --all          Build both Linux and Windows"
            echo "  -u             Also create .gz archive for update deployment"
            echo "  -v<version>    Set version (e.g. -v1.3.11), saves to version.txt"
            exit 0
            ;;
        *) error "Unknown argument: $arg" ;;
    esac
done

# ── Обновление версии (если передан флаг -v) ──────────────────────────────────
if [[ -n "$NEW_VERSION" ]]; then
    if [[ ! "$NEW_VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        error "Invalid version format: '${NEW_VERSION}'. Expected: X.Y.Z (e.g. 1.3.11)"
    fi
    info "Updating version: ${APP_VERSION} → ${NEW_VERSION}"
    APP_VERSION="$NEW_VERSION"
    echo "$APP_VERSION" > "$VERSION_FILE"
    success "Version saved to version.txt: ${APP_VERSION}"
    echo ""
fi

# ── Проверка Go ───────────────────────────────────────────────────────────────
info "Checking Go installation..."
if ! command -v go &>/dev/null; then
    error "Go not found in PATH. Install via: sudo pacman -S go"
fi
go version

# ── GOPATH / bin ──────────────────────────────────────────────────────────────
GOPATH_DIR="${GOPATH:-$HOME/go}"
GO_BIN="${GOPATH_DIR}/bin"
export PATH="${GO_BIN}:${PATH}"

# ── Функция создания gz-архива ─────────────────────────────────────────────────
make_update_gz() {
    local binary="$1"
    local output_gz="$2"

    if ! command -v gzip &>/dev/null; then
        error "gzip not found. Install via: sudo pacman -S gzip"
    fi

    if [[ ! -f "$binary" ]]; then
        error "Binary not found for update packaging: ${binary}"
    fi

    info "Compressing ${binary} → ${output_gz}..."
    rm -f "$output_gz"
    gzip -9 -k -c "$binary" > "$output_gz" || error "Compression failed"

    local orig_bytes gz_bytes orig_mb gz_mb saved
    orig_bytes=$(stat -c%s "$binary")
    gz_bytes=$(stat -c%s "$output_gz")
    orig_mb=$(awk  "BEGIN { printf \"%.2f\", $orig_bytes / 1048576 }")
    gz_mb=$(awk    "BEGIN { printf \"%.2f\", $gz_bytes   / 1048576 }")
    saved=$(awk    "BEGIN { printf \"%.1f\", (1 - $gz_bytes / $orig_bytes) * 100 }")

    success "Update archive ready: ${output_gz}"
    warn "  Original:   ${orig_mb} MB"
    warn "  Compressed: ${gz_mb} MB  (saved ${saved}%)"
    warn "  Location:   $(realpath "$output_gz")"
    echo ""
}

# ── Функция сборки одного таргета ─────────────────────────────────────────────
build_target() {
    local target="$1"
    local output output_gz goos goarch use_winres

    if [[ "$target" == "windows" ]]; then
        output="${BASE_NAME}.exe"
        output_gz="update-windows-amd64.exe.gz"
        goos="windows"
        goarch="amd64"
        use_winres=true
    else
        output="${BASE_NAME}"
        output_gz="update-linux-amd64.gz"
        goos="linux"
        goarch="amd64"
        use_winres=false
    fi

    echo ""
    echo -e "${BOLD}${GREEN}=== WBG Albion Data Client Build ===${RESET}"
    echo -e "${CYAN}Target:  ${YELLOW}${goos}/${goarch}${RESET}"
    echo -e "${CYAN}Version: ${YELLOW}${APP_VERSION}${RESET}"
    echo -e "${CYAN}Output:  ${YELLOW}${output}${RESET}"
    [[ "$WITH_UPDATE" == true ]] && \
    echo -e "${CYAN}Archive: ${YELLOW}${output_gz}${RESET}"
    echo ""

    # ── go-winres (только для Windows) ────────────────────────────────────────
    if [[ "$use_winres" == true ]]; then
        info "Installing go-winres..."
        go install github.com/tc-hib/go-winres@v0.3.1 \
            || error "Failed to install go-winres"

        if ! command -v go-winres &>/dev/null; then
            error "go-winres not found after install (check \$GOPATH/bin is in PATH)"
        fi
    fi

    # ── Чистка ────────────────────────────────────────────────────────────────
    info "Cleaning old files..."
    rm -f rsrc_windows_*.syso "${output}" ./*.bak

    # ── Windows-ресурсы (.syso) ───────────────────────────────────────────────
    if [[ "$use_winres" == true ]]; then
        info "Generating Windows resources..."
        GOOS=windows go-winres make \
            || error "Failed to generate Windows resources"
    fi

    # ── Сборка ────────────────────────────────────────────────────────────────
    info "Building application..."
    GOOS="${goos}" GOARCH="${goarch}" \
        go build \
            -ldflags "-s -w -X main.version=${APP_VERSION}" \
            -o "${output}" \
            . \
        || error "Build failed"

    # ── patchelf (только для Linux) ───────────────────────────────────────────
    if [[ "$goos" == "linux" ]]; then
        if command -v patchelf &>/dev/null; then
            info "Patching libpcap dependency..."
            patchelf --replace-needed libpcap.so.0.8 libpcap.so "${output}" \
                || warn "patchelf failed, skipping patch"
        else
            warn "patchelf not found, skipping (install via: sudo pacman -S patchelf)"
        fi
    fi

    # ── Итог сборки ───────────────────────────────────────────────────────────
    echo ""
    success "=== Build Successful: ${target} ==="

    if [[ -f "$output" ]]; then
        local size_bytes size_mb
        size_bytes=$(stat -c%s "$output")
        size_mb=$(awk "BEGIN { printf \"%.2f\", $size_bytes / 1048576 }")
        warn "Output:   ${output}"
        warn "Size:     ${size_mb} MB"
        warn "Location: $(realpath "$output")"
    fi
    echo ""

    # ── gz-архив для обновления (если -u) ─────────────────────────────────────
    if [[ "$WITH_UPDATE" == true ]]; then
        make_update_gz "${output}" "${output_gz}"
    fi
}

# ── Запуск ────────────────────────────────────────────────────────────────────
if [[ "$TARGET" == "all" ]]; then
    build_target "linux"
    build_target "windows"
else
    build_target "$TARGET"
fi

echo ""
success "=== All done! ==="
[[ -n "$NEW_VERSION" ]] && warn "Version: ${APP_VERSION}"
echo ""
