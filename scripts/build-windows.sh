#!/usr/bin/env bash

set -eo pipefail

# usage helper
print_usage() {
    echo "Usage: $0 [--gzip-only] <version>" >&2
    echo "If --gzip-only is given, the script skips compilation and only creates" >&2
    echo "an update package from an existing WBG-albion-data-client-<version>.exe." >&2
    echo "Version may also be supplied via GITHUB_REF_NAME when building." >&2
}

# decide mode and version
if [ "$1" = "--gzip-only" ]; then
    mode="gzip"
    shift
else
    mode="build"
fi

version="${1:-$GITHUB_REF_NAME}"
if [ -z "$version" ]; then
    print_usage
    exit 1
fi

# cleanup previous artefacts
rm -f rsrc_windows_*
rm -f albiondata-client.exe
rm -f WBG-albion-data-client*.exe
rm -f albiondata-client.*.bak
rm -f .albiondata-client.*.old

rm -f albiondata-client-amd64-installer.exe

# if gzip-only mode, skip compile/patсh
if [ "$mode" = "gzip" ]; then
    exe_name="WBG-albion-data-client-${version}.exe"
    if [ ! -f "$exe_name" ]; then
        echo "error: executable $exe_name not found; build it first or run without --gzip-only" >&2
        exit 1
    fi
    if ! command -v gzip >/dev/null 2>&1; then
        echo "error: gzip not available, cannot create update package" >&2
        exit 1
    fi
    cp "$exe_name" "$exe_name.copy"
    gzip -9 "$exe_name"
    mv "$exe_name.gz" "update-windows-amd64-${version}.exe.gz"
    mv "$exe_name.copy" "$exe_name"
    echo "created update-windows-amd64-${version}.exe.gz from $exe_name"
    exit 0
fi

go install github.com/tc-hib/go-winres@v0.3.1

export PATH="$PATH:/root/go/bin"

go-winres make

exe_name="WBG-albion-data-client-${version}.exe"

env GOOS=windows GOARCH=amd64 \
    go build -ldflags "-s -w -X main.version=$version" \
    -o "$exe_name" -v -x albiondata-client.go
	# patch the executable produced by go-winres
	go-winres patch "$exe_name"

cd pkg/nsis
make nsis

cd ../..
ls -la WBG-albion-data-client*

# make a versioned gzip update package (and a generic copy for the auto-updater)
cp "$exe_name" "$exe_name.copy"
if command -v gzip >/dev/null 2>&1; then
    gzip -9 "$exe_name"
    mv "$exe_name.gz" "update-windows-amd64-${version}.exe.gz"
    cp "update-windows-amd64-${version}.exe.gz" update-windows-amd64.exe.gz
elif command -v python >/dev/null 2>&1 || command -v python3 >/dev/null 2>&1; then
    py="$(command -v python 2>/dev/null || command -v python3)"
    # python can compress raw gzip
    "$py" - <<'PY' <<'PY'
import sys, gzip, shutil
inp=sys.argv[1]
out=sys.argv[2]
with open(inp,'rb') as f_in, gzip.open(out,'wb') as f_out:
    shutil.copyfileobj(f_in, f_out)
PY
    mv "$exe_name.gz" "update-windows-amd64-${version}.exe.gz"
    cp "update-windows-amd64-${version}.exe.gz" update-windows-amd64.exe.gz
else
    echo "warning: gzip not found and no python available, skipping update package creation" >&2
fi
mv "$exe_name.copy" "$exe_name"
