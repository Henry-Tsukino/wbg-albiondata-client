#!/usr/bin/env bash
set -eo pipefail
sudo apt-get update && sudo apt-get install -y libpcap-dev patchelf
env | sort
go build -ldflags "-s -w -X main.version=$GITHUB_REF_NAME" -o WBG-albion-data-client albiondata-client.go
patchelf --replace-needed libpcap.so.0.8 libpcap.so WBG-albion-data-client
./WBG-albion-data-client -version
cp WBG-albion-data-client WBG-albion-data-client.old
gzip -9 WBG-albion-data-client
mv WBG-albion-data-client.gz update-linux-amd64.gz
mv WBG-albion-data-client.old WBG-albion-data-client
