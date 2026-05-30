#!/usr/bin/env bash
# Builds macOS binaries via Docker (uses osxcross inside multiarch/crossbuild).
# Run from the project root: bash scripts/build-and-copy-darwin.sh
#
# Optional: set GITHUB_REF_NAME to embed a version string, e.g.:
#   GITHUB_REF_NAME=1.3.12 bash scripts/build-and-copy-darwin.sh

set -euo pipefail

VERSION="${GITHUB_REF_NAME:-dev}"

echo "==> Building macOS Docker image (version: ${VERSION})..."
docker build \
    --build-arg GITHUB_REF_NAME="${VERSION}" \
    -f ./Dockerfile.build.darwin \
    -t albiondataclient-darwin \
    .

echo "==> Running builder container..."
docker rm -f builder 2>/dev/null || true
docker run --name builder albiondataclient-darwin

echo "==> Copying artifacts..."
docker cp builder:/usr/src/app/update-darwin-amd64.gz ./update-darwin-amd64.gz
docker cp builder:/usr/src/app/albiondata-client-amd64-mac.zip ./albiondata-client-amd64-mac.zip

echo "==> Cleaning up container..."
docker rm builder

echo ""
echo "Done! Artifacts:"
ls -lh update-darwin-amd64.gz albiondata-client-amd64-mac.zip
