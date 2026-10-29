#!/usr/bin/env bash

set -euo pipefail

RED='\033[0;31m'; GREEN='\033[0;32m'; CYAN='\033[0;36m'; RESET='\033[0m'

info()    { echo -e "${CYAN}$*${RESET}"; }
success() { echo -e "${GREEN}$*${RESET}"; }
error()   { echo -e "${RED}ERROR: $*${RESET}" >&2; exit 1; }

BINARY="WBG-albion-data-client"
INSTALL_DIR="/usr/local/bin"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BINARY_PATH="${SCRIPT_DIR}/${BINARY}"

[[ -f "$BINARY_PATH" ]] || error "Binary not found: ${BINARY_PATH}"

# sudo нужен, только если мы не root
if [[ $EUID -eq 0 ]]; then
    SUDO=""
else
    command -v sudo >/dev/null 2>&1 || error "Нужен sudo (или запустите скрипт от root)."
    SUDO="sudo"
fi

LDCONFIG="$(command -v ldconfig || true)"
if [[ -z "$LDCONFIG" && -x /sbin/ldconfig ]]; then LDCONFIG=/sbin/ldconfig; fi

# pkg_install <arch> <debian/ubuntu> <fedora> <opensuse>
pkg_install() {
    if command -v pacman >/dev/null 2>&1; then
        $SUDO pacman -S --needed --noconfirm $1
    elif command -v apt-get >/dev/null 2>&1; then
        $SUDO apt-get update -qq && $SUDO apt-get install -y $2
    elif command -v dnf >/dev/null 2>&1; then
        $SUDO dnf install -y $3
    elif command -v zypper >/dev/null 2>&1; then
        $SUDO zypper --non-interactive install $4
    else
        return 1
    fi
}

# Путь к любой установленной libpcap.so.* (кроме имени $1), либо пусто
find_real_libpcap() {
    local skip="$1" p="" d f
    if [[ -n "$LDCONFIG" ]]; then
        p="$("$LDCONFIG" -p 2>/dev/null | awk -v skip="$skip" \
            '$1 ~ /^libpcap\.so\.[0-9.]+$/ && $1 != skip && /x86-64/ {print $NF; exit}' || true)"
    fi
    if [[ -z "$p" ]]; then
        for d in /usr/lib /usr/lib64 /lib /lib64 /usr/lib/x86_64-linux-gnu /lib/x86_64-linux-gnu; do
            for f in "$d"/libpcap.so.*; do
                if [[ -e "$f" && "$(basename "$f")" != "$skip" ]]; then p="$f"; break 2; fi
            done
        done
    fi
    echo "$p"
}

# Имя libpcap, которое бинарнику нужно, но которого в системе нет (пусто = всё хорошо)
missing_libpcap() {
    command -v ldd >/dev/null 2>&1 || return 0
    ldd "$BINARY_PATH" 2>/dev/null | awk '/libpcap/ && /not found/ {print $1; exit}' || true
}

# Бинарник может быть собран с libpcap.so / libpcap.so.1 / libpcap.so.0.8 - а в системе
# лежит другое имя (в разных дистрибутивах оно разное). Делаем симлинк рядом с настоящей
# библиотекой (именно там, а не в /usr/local/lib: бинарник с setcap игнорирует LD_LIBRARY_PATH).
ensure_libpcap() {
    local needed real dir
    needed="$(missing_libpcap)"
    if [[ -z "$needed" ]]; then return 0; fi

    info "Нужна ${needed}, но под таким именем её нет - исправляю..."
    real="$(find_real_libpcap "$needed")"
    if [[ -z "$real" ]]; then
        info "libpcap не установлена - ставлю пакет..."
        pkg_install "libpcap" "libpcap0.8" "libpcap" "libpcap1" \
            || error "Не удалось поставить libpcap автоматически. Установите пакет libpcap вручную и запустите скрипт снова."
        if [[ -n "$LDCONFIG" ]]; then $SUDO "$LDCONFIG"; fi
        needed="$(missing_libpcap)"
        if [[ -z "$needed" ]]; then return 0; fi
        real="$(find_real_libpcap "$needed")"
    fi
    [[ -n "$real" ]] || error "libpcap так и не найдена. Установите пакет libpcap вручную."

    dir="$(dirname "$real")"
    $SUDO ln -sf "$(basename "$real")" "${dir}/${needed}"
    if [[ -n "$LDCONFIG" ]]; then $SUDO "$LDCONFIG"; fi
    [[ -z "$(missing_libpcap)" ]] || error "Не удалось подключить libpcap (${dir}/${needed})."
    success "Создан симлинк ${dir}/${needed} -> $(basename "$real")"
}

info "Installing ${BINARY}..."

ensure_libpcap

# setcap входит в libcap (в Debian/Ubuntu - libcap2-bin)
if ! $SUDO sh -c 'command -v setcap >/dev/null 2>&1'; then
    info "Ставлю setcap..."
    pkg_install "libcap" "libcap2-bin" "libcap" "libcap-progs" \
        || error "Не удалось поставить setcap. Установите libcap (libcap2-bin) вручную."
fi

# Запрашивает пароль один раз, дальше sudo кешируется
$SUDO install -Dm755 "$BINARY_PATH" "${INSTALL_DIR}/${BINARY}"
$SUDO setcap cap_net_raw+eip "${INSTALL_DIR}/${BINARY}"
info "Creating desktop shortcut..."

# Определяем терминал
if command -v konsole &>/dev/null; then
    TERM_CMD="konsole -e"
elif command -v alacritty &>/dev/null; then
    TERM_CMD="alacritty -e"
elif command -v kitty &>/dev/null; then
    TERM_CMD="kitty"
else
    TERM_CMD="xterm -e"
fi

mkdir -p ~/.local/share/applications
cat > ~/.local/share/applications/wbg-albion.desktop << EOF
[Desktop Entry]
Name=WBG Albion Client
Exec=${TERM_CMD} WBG-albion-data-client
Icon=applications-games
Terminal=false
Type=Application
Categories=Game;
EOF

chmod +x ~/.local/share/applications/wbg-albion.desktop
success "Done! Запускай без sudo:"
success "  ${BINARY}"
