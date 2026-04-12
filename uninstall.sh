#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEFAULT_BIN_DIR="${HOME}/.local/bin"
BIN_DIR="${PREFIX:-$DEFAULT_BIN_DIR}"
COMMAND_NAME="codex-ralph"
TARGET_SCRIPT="$SCRIPT_DIR/codex-loop.sh"
INSTALL_PATH="$BIN_DIR/$COMMAND_NAME"
DOCTOR_NAME="codex-ralph-doctor"
DOCTOR_SCRIPT="$SCRIPT_DIR/doctor.sh"
DOCTOR_INSTALL_PATH="$BIN_DIR/$DOCTOR_NAME"

die() {
  printf '%s\n' "$*" >&2
  exit 1
}

info() {
  printf '[uninstall] %s\n' "$*"
}

main() {
  for pair in \
    "$INSTALL_PATH:$TARGET_SCRIPT" \
    "$DOCTOR_INSTALL_PATH:$DOCTOR_SCRIPT"
  do
    local install_path="${pair%%:*}"
    local target_script="${pair##*:}"

    if [[ -L "$install_path" ]]; then
      local linked
      linked="$(readlink "$install_path")"
      if [[ "$linked" == "$target_script" ]]; then
        rm -f "$install_path"
        info "Removed wrapper: $install_path"
        continue
      fi
      die "Refusing to remove symlink that points elsewhere: $install_path -> $linked"
    fi

    if [[ -f "$install_path" ]]; then
      if grep -qF "$target_script" "$install_path" 2>/dev/null; then
        rm -f "$install_path"
        info "Removed wrapper: $install_path"
        continue
      fi

      die "Refusing to remove non-codex-ralph file: $install_path"
    fi

    info "Nothing to uninstall: $install_path"
  done
}

main "$@"
