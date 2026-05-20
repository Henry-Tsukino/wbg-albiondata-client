
# =============================================================================
#  WBG Albion Data Client — Build Script
#  Supports: native Linux build  OR  cross-compile → Windows (.exe)
#
#  Usage:
#    ./build.sh           # auto-detects: linux by default
#    ./build.sh --windows # cross-compile for Windows (.exe)
#    ./build.sh --linux   # native Linux binary
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
TARGET="linux"   # по умолчанию

for arg in "$@"; do
    case "$arg" in
        --windows) TARGET="windows" ;;
        --linux)   TARGET="linux"   ;;
        --help|-h)
            echo "Usage: $0 [--windows|--linux]"
            echo "  --windows  Cross-compile for Windows (.exe)"
            echo "  --linux    Build native Linux binary (default)"
            exit 0
            ;;
        *) error "Unknown argument: $arg" ;;
    esac
done

# ── Параметры таргета ─────────────────────────────────────────────────────────
if [[ "$TARGET" == "windows" ]]; then
    OUTPUT_NAME="${BASE_NAME}.exe"
    GOOS_VAL="windows"
    GOARCH_VAL="amd64"
    USE_WINRES=true
else
    OUTPUT_NAME="${BASE_NAME}"
    GOOS_VAL="linux"
    GOARCH_VAL="amd64"
    USE_WINRES=false
fi

echo ""
echo -e "${BOLD}${GREEN}=== WBG Albion Data Client Build ===${RESET}"
echo -e "${CYAN}Target:  ${YELLOW}${GOOS_VAL}/${GOARCH_VAL}${RESET}"
echo -e "${CYAN}Version: ${YELLOW}${APP_VERSION}${RESET}"
echo -e "${CYAN}Output:  ${YELLOW}${OUTPUT_NAME}${RESET}"
echo ""

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

# ── go-winres (только для Windows-таргета) ────────────────────────────────────
if [[ "$USE_WINRES" == true ]]; then
    info "Installing go-winres..."
    go install github.com/tc-hib/go-winres@v0.3.2 \
        || error "Failed to install go-winres"

    if ! command -v go-winres &>/dev/null; then
        error "go-winres not found after install (check \$GOPATH/bin is in PATH)"
    fi
fi

# ── Чистка ────────────────────────────────────────────────────────────────────
info "Cleaning old files..."
rm -f rsrc_windows_*.syso "${OUTPUT_NAME}" ./*.bak

# ── Windows-ресурсы (.syso) ───────────────────────────────────────────────────
if [[ "$USE_WINRES" == true ]]; then
    info "Generating Windows resources..."
    # go-winres make читает winres/winres.json и создаёт rsrc_windows_*.syso
    # Эти файлы go build подхватит автоматически — отдельный patch не нужен
    GOOS=windows go-winres make \
        || error "Failed to generate Windows resources"
fi

# ── Сборка ────────────────────────────────────────────────────────────────────
info "Building application..."
GOOS="${GOOS_VAL}" GOARCH="${GOARCH_VAL}" \
    go build \
        -ldflags "-s -w -X main.version=${APP_VERSION}" \
        -o "${OUTPUT_NAME}" \
        . \
    || error "Build failed"

# ── Итог ──────────────────────────────────────────────────────────────────────
echo ""
success "=== Build Successful! ==="

if [[ -f "$OUTPUT_NAME" ]]; then
    SIZE_BYTES=$(stat -c%s "$OUTPUT_NAME")
    SIZE_MB=$(awk "BEGIN { printf \"%.2f\", $SIZE_BYTES / 1048576 }")
    warn "Output:   ${OUTPUT_NAME}"
    warn "Size:     ${SIZE_MB} MB"
    warn "Location: $(realpath "$OUTPUT_NAME")"
fi
echo ""
