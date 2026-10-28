#!/usr/bin/env bash
# Source this file. Build scripts never delete old output: it is moved to
# legacy/dist/ instead, replacing a file of the same name that is already there.

archive_old_builds() {
  local item base
  mkdir -p legacy/dist
  shopt -s nullglob dotglob
  for item in dist/* rsrc_windows_*.syso; do
    base="$(basename "$item")"
    rm -rf "legacy/dist/$base"          # replace the previously archived copy
    mv "$item" "legacy/dist/$base"
  done
  shopt -u nullglob dotglob
  mkdir -p dist
}
