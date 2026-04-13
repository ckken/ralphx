#!/usr/bin/env bash
set -euo pipefail

BIN_DIR="${PREFIX:-${HOME}/.local/bin}"
DATA_DIR="${RALPHX_DATA_DIR:-${HOME}/.local/share/ralphx}"
CONFIG_DIR="${RALPHX_CONFIG_DIR:-${HOME}/.config/ralphx}"
ALIAS_NAME="codex-ralph"
CODEX_HOME_DIR="${CODEX_HOME:-${HOME}/.codex}"
SKILLS_DIR="$CODEX_HOME_DIR/skills"
SKILL_NAME="ralphx"
LEGACY_SKILL_NAME="ralphx-drive"

info() {
  printf '[uninstall] %s\n' "$*"
}

remove_if_exists() {
  local path="$1"
  if [[ -e "$path" || -L "$path" ]]; then
    rm -rf "$path"
    info "Removed $path"
  else
    info "Nothing to remove: $path"
  fi
}

main() {
  remove_if_exists "$BIN_DIR/ralphx"
  remove_if_exists "$BIN_DIR/ralphx-doctor"
  remove_if_exists "$BIN_DIR/$ALIAS_NAME"
  remove_if_exists "$CONFIG_DIR/current.env"
  remove_if_exists "$DATA_DIR/releases"
  remove_if_exists "$SKILLS_DIR/$SKILL_NAME"
  remove_if_exists "$SKILLS_DIR/$LEGACY_SKILL_NAME"
}

main "$@"
