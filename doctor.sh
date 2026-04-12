#!/usr/bin/env bash
set -euo pipefail

DEFAULT_BIN_DIR="${HOME}/.local/bin"
BIN_DIR="${PREFIX:-$DEFAULT_BIN_DIR}"

check_cmd() {
  local name="$1"
  if command -v "$name" >/dev/null 2>&1; then
    printf '[ok] %s -> %s\n' "$name" "$(command -v "$name")"
  else
    printf '[missing] %s\n' "$name"
  fi
}

echo "codex-ralph doctor"
echo "BIN_DIR=$BIN_DIR"
echo

check_cmd bash
check_cmd jq
check_cmd python3
check_cmd git
check_cmd gh
check_cmd codex

echo
if [[ ":$PATH:" == *":$BIN_DIR:"* ]]; then
  echo "[ok] PATH contains $BIN_DIR"
else
  echo "[missing] PATH does not contain $BIN_DIR"
  echo "Add it with: export PATH=\"$BIN_DIR:\$PATH\""
fi
