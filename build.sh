# =============================================================================
#  WBG Albion Data Client — Build Script
#  Supports: native Linux build  OR  cross-compile → Windows (.exe)  OR  both
#
#  Usage:
#    ./build.sh                # linux по умолчанию
#    ./build.sh --windows      # cross-compile для Windows (.exe)
#    ./build.sh --linux        # native Linux binary
#    ./build.sh --all          # собрать Linux и Windows
#    ./build.sh --linux -u     # собрать + создать .gz для обновления
#    ./build.sh --windows -u   # собрать + создать .gz для обновления
#    ./build.sh --all -u       # собрать оба + создать .gz для обоих
# =============================================================================

set -euo pipefail

# ── Цвета ────────────────────────────────────────────────────────────────────
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'
CYAN='\033[0;36m'; BOLD='\033[1m'; RESET='\033[0m'

info()    { echo -e "${CYAN}$*${RESET}"; }
success() { echo -e "${GREEN}$*${RESET}"; }
warn()    { echo -e "${YELLOW}$*${RESET}"; }
error()   { echo -e "${RED}ERROR: $*${RESET}" >&2; exit 1; }

# ── Константы ─────────────────────────────────────────────────────────────────
APP_VERSION="1.3.9"
BASE_NAME="WBG-albion-data-client"

# ── Разбор аргументов ─────────────────────────────────────────────────────────
TARGET="linux"
WITH_UPDATE=false

for arg in "$@"; do
    case "$arg" in
        --windows) TARGET="windows" ;;
        --linux)   TARGET="linux"   ;;
        --all)     TARGET="all"     ;;
        -u)        WITH_UPDATE=true ;;
        --help|-h)
            echo "Usage: $0 [--windows|--linux|--all] [-u]"
            echo "  --windows  Cross-compile for Windows (.exe)"
            echo "  --linux    Build native Linux binary (default)"
            echo "  --all      Build both Linux and Windows"
            echo "  -u         Also create .gz archive for update deployment"
            exit 0
            ;;
        *) error "Unknown argument: $arg" ;;
    esac
done

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

# ── Функция создания gz-архива для обновления ─────────────────────────────────
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

    ORIG_BYTES=$(stat -c%s "$binary")
    GZ_BYTES=$(stat -c%s "$output_gz")
    ORIG_MB=$(awk  "BEGIN { printf \"%.2f\", $ORIG_BYTES / 1048576 }")
    GZ_MB=$(awk    "BEGIN { printf \"%.2f\", $GZ_BYTES   / 1048576 }")
    SAVED=$(awk    "BEGIN { printf \"%.1f\", (1 - $GZ_BYTES / $ORIG_BYTES) * 100 }")

    success "Update archive ready: ${output_gz}"
    warn "  Original:   ${ORIG_MB} MB"
    warn "  Compressed: ${GZ_MB} MB  (saved ${SAVED}%)"
    warn "  Location:   $(realpath "$output_gz")"
    echo ""
}

# ── Функция сборки одного таргета ─────────────────────────────────────────────
build_target() {
    local target="$1"

    if [[ "$target" == "windows" ]]; then
        local output="${BASE_NAME}.exe"
        local output_gz="update-windows-amd64.exe.gz"
        local goos="windows"
        local goarch="amd64"
        local use_winres=true
    else
        local output="${BASE_NAME}"
        local output_gz="update-linux-amd64.gz"
        local goos="linux"
        local goarch="amd64"
        local use_winres=false
    fi

    echo ""
    echo -e "${BOLD}${GREEN}=== WBG Albion Data Client Build ===${RESET}"
    echo -e "${CYAN}Target:  ${YELLOW}${goos}/${goarch}${RESET}"
    echo -e "${CYAN}Version: ${YELLOW}${APP_VERSION}${RESET}"
    echo -e "${CYAN}Output:  ${YELLOW}${output}${RESET}"
    echo ""

    # ── go-winres (только для Windows-таргета) ────────────────────────────────
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

    # ── Итог сборки ───────────────────────────────────────────────────────────
    echo ""
    success "=== Build Successful! ==="

    if [[ -f "$output" ]]; then
        SIZE_BYTES=$(stat -c%s "$output")
        SIZE_MB=$(awk "BEGIN { printf \"%.2f\", $SIZE_BYTES / 1048576 }")
        warn "Output:   ${output}"
        warn "Size:     ${SIZE_MB} MB"
        warn "Location: $(realpath "$output")"
    fi
    echo ""

    # ── Создание gz-архива (если передан флаг -u) ─────────────────────────────
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
