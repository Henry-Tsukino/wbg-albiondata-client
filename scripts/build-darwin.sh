#!/usr/bin/env bash

set -eo pipefail
cd "$(dirname "$0")/.."
source scripts/version.sh "$@"

rm -rf dist; mkdir -p dist

(cd frontend && npm ci && npm run build)

# Native macOS build. GitHub's macos-latest runners are arm64 hardware, so
# amd64 output requires explicit CC/CGO_LDFLAGS arch flags. CGO stays on
# since gopacket links against libpcap.
export CGO_ENABLED=1
export GOARCH=amd64
export CC="clang -arch x86_64"
export CGO_LDFLAGS="-arch x86_64"
go build -ldflags "$LDFLAGS" -o "dist/${APP_BIN}" wbg-client.go

gzip -9 -c "dist/${APP_BIN}" > dist/update-darwin-amd64.gz

# Zipped folder with a run.command file that runs the client under sudo
TEMP="WBG-albion-data-client"
rm -rf "./scripts/$TEMP"
mkdir "./scripts/$TEMP"
cp "dist/${APP_BIN}" "./scripts/$TEMP/WBG-albion-data-client-executable"
cp ./scripts/run.command "./scripts/$TEMP/run.command"
chmod 777 ./scripts/$TEMP/*
(cd scripts && zip -r ../dist/WBG-albion-data-client-amd64-mac.zip ./"$TEMP")
