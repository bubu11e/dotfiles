#!/usr/bin/env bash
# Syntax-check the shared shell files as chezmoi actually renders them.
#
# `chezmoi apply --dry-run` proves the Go templates render; it says nothing about
# whether what they render is valid shell. ~/.shell.env is the sharp case: it is
# sourced by every invocation of both shells, login included, so a syntax error
# there costs a working shell. It is also the one shell file no linter reads in
# place — shellcheck sees Go template syntax rather than shell, which is why
# .pre-commit-config.yaml excludes it.
#
# The three files are POSIX sh by contract and are parsed by zsh and bash as well,
# so all three parsers have to agree. A missing shell is reported and skipped
# rather than failing: CI installs all three, a laptop may not have bash 5.
#
# Requires an initialised chezmoi config (CI runs `chezmoi init --promptDefaults`).
# Run locally with: bash scripts/check-rendered-shell.sh
set -euo pipefail

SOURCE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
RENDERED="$(mktemp -d)"
trap 'rm -rf "$RENDERED"' EXIT

FILES=(.shell.env .shell.aliases .shell.rc)

status=0
for f in "${FILES[@]}"; do
  chezmoi cat --source="$SOURCE_DIR" "$HOME/$f" >"${RENDERED}/${f}"
  for shell in sh bash zsh; do
    if ! command -v "$shell" >/dev/null 2>&1; then
      echo "skip: $shell not installed"
      continue
    fi
    if "$shell" -n "${RENDERED}/${f}"; then
      echo "ok: $f under $shell"
    else
      echo "FAIL: $f under $shell" >&2
      status=1
    fi
  done
done

exit "$status"
