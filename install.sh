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
BACKUP_SUFFIX="$(date +%Y%m%d%H%M%S)"

die() {
  printf '%s\n' "$*" >&2
  exit 1
}

info() {
  printf '[install] %s\n' "$*"
}

ensure_executable() {
  if [[ ! -f "$TARGET_SCRIPT" ]]; then
    die "Missing target script: $TARGET_SCRIPT"
  fi
  chmod +x "$TARGET_SCRIPT"
  [[ -f "$DOCTOR_SCRIPT" ]] || die "Missing doctor script: $DOCTOR_SCRIPT"
  chmod +x "$DOCTOR_SCRIPT"
}

write_wrapper() {
  local install_path="$1"
  local target_script="$2"

  if [[ -L "$install_path" ]]; then
    local linked
    linked="$(readlink "$install_path")"
    if [[ "$linked" == "$target_script" ]]; then
      info "Already installed: $install_path -> $target_script"
      return 0
    fi
    rm -f "$install_path"
  fi

  if [[ -e "$install_path" && ! -L "$install_path" ]]; then
    local backup_path="${install_path}.bak.${BACKUP_SUFFIX}"
    cp "$install_path" "$backup_path"
    info "Backed up existing file to $backup_path"
  fi

  cat > "$install_path" <<EOF
#!/usr/bin/env bash
set -euo pipefail
exec "$target_script" "\$@"
EOF
  chmod +x "$install_path"
}

main() {
  mkdir -p "$BIN_DIR"
  ensure_executable
  write_wrapper "$INSTALL_PATH" "$TARGET_SCRIPT"
  write_wrapper "$DOCTOR_INSTALL_PATH" "$DOCTOR_SCRIPT"

  cat <<EOF
Installed codex-ralph:
  $INSTALL_PATH

Installed doctor command:
  $DOCTOR_INSTALL_PATH

Repository:
  $SCRIPT_DIR

If $BIN_DIR is not on PATH, add it:
  export PATH="$BIN_DIR:\$PATH"
EOF
}

main "$@"
