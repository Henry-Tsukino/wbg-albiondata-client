#!/usr/bin/env bash
# Resolves and validates the application version. Source this file.
#
# The version is REQUIRED: a binary built without one is a "dev" build and
# never self-updates. Provide it as:
#     VERSION=1.5.0 make build-windows          (or ./scripts/build-windows.sh 1.5.0)
# In CI the git tag (GITHUB_REF_NAME) is used automatically.
#
# IMPORTANT: the updater compares versions with strict semver, which does NOT
# accept a leading "v". The GitHub release tag must therefore be "1.5.0",
# not "v1.5.0" - otherwise the comparison silently fails and no update is
# ever offered. A leading "v" typed here is stripped for the binary.

RAW_VERSION="${1:-${VERSION:-${GITHUB_REF_NAME:-}}}"
APP_VERSION="${RAW_VERSION#v}"

if [[ ! "$APP_VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "ERROR: application version is required, in the form 1.5.0" >&2
  echo "       got: '${RAW_VERSION}'" >&2
  echo "       usage: VERSION=1.5.0 make build-windows   (or: $0 1.5.0)" >&2
  exit 1
fi

export APP_VERSION
export APP_BIN="WBG-albion-data-client"
export LDFLAGS="-s -w -X main.version=${APP_VERSION}"
echo "==> Building WBG Client version ${APP_VERSION}"
