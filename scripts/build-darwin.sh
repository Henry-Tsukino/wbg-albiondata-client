#!/usr/bin/env bash

set -eo pipefail

export OSXCROSS_NO_INCLUDE_PATH_WARNINGS=1
export MACOSX_DEPLOYMENT_TARGET=12.0
export CC=/opt/osxcross/target/bin/o64-clang
export CXX=/opt/osxcross/target/bin/o64-clang++
export GOOS=darwin
export GOARCH=amd64
export CGO_ENABLED=1

go build -ldflags "-s -w -X main.version=$GITHUB_REF_NAME" -o albiondata-client .

gzip -k9 albiondata-client
mv albiondata-client.gz update-darwin-amd64.gz


# Creates a zipped folder with a run.command file that runs the client under sudo
TEMP="albiondata-client"
ZIPNAME="albiondata-client-amd64-mac.zip"
rm -rfv ./scripts/$TEMP
rm -rfv ./$ZIPNAME
rm -rfv ./scripts/update-darwin-amd64.zip
mkdir -v ./scripts/$TEMP
cp -v albiondata-client ./scripts/$TEMP/albiondata-client-executable
cd scripts
cp -v run.command ./$TEMP/run.command
chown -Rv ${USER}:${USER} ./$TEMP 2>/dev/null || true
chmod -v 777 ./$TEMP/*
zip -v ../$ZIPNAME -r ./"$TEMP"
