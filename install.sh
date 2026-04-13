#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEFAULT_BIN_DIR="${HOME}/.local/bin"
BIN_DIR="${PREFIX:-$DEFAULT_BIN_DIR}"
BUILD_DIR="$SCRIPT_DIR/bin"
COMMAND_NAME="ralphx"
DOCTOR_NAME="ralphx-doctor"
TARGET_BINARY="$BUILD_DIR/$COMMAND_NAME"
DOCTOR_BINARY="$BUILD_DIR/$DOCTOR_NAME"
INSTALL_PATH="$BIN_DIR/$COMMAND_NAME"
DOCTOR_INSTALL_PATH="$BIN_DIR/$DOCTOR_NAME"
BACKUP_SUFFIX="$(date +%Y%m%d%H%M%S)"

info() {
  printf '[install] %s\n' "$*"
}

die() {
  printf '%s\n' "$*" >&2
  exit 1
}

ensure_prereqs() {
  command -v go >/dev/null 2>&1 || die "Missing required command: go"
}

build_binaries() {
  mkdir -p "$BUILD_DIR"
  info "Building Go binaries into $BUILD_DIR"
  (cd "$SCRIPT_DIR" && go build -o "$TARGET_BINARY" ./cmd/ralphx)
  (cd "$SCRIPT_DIR" && go build -o "$DOCTOR_BINARY" ./cmd/ralphx-doctor)
}

write_wrapper() {
  local install_path="$1"
  local target_binary="$2"
  local mode="$3"

  if [[ -e "$install_path" && ! -L "$install_path" ]]; then
    local backup_path="${install_path}.bak.${BACKUP_SUFFIX}"
    cp "$install_path" "$backup_path"
    info "Backed up existing file to $backup_path"
  fi

  cat > "$install_path" <<EOF
#!/usr/bin/env bash
set -euo pipefail
exec "$target_binary" "\$@"
EOF
  chmod +x "$install_path"
  info "Installed $mode wrapper: $install_path"
}

main() {
  ensure_prereqs
  mkdir -p "$BIN_DIR"
  build_binaries
  write_wrapper "$INSTALL_PATH" "$TARGET_BINARY" "$COMMAND_NAME"
  write_wrapper "$DOCTOR_INSTALL_PATH" "$DOCTOR_BINARY" "$DOCTOR_NAME"

  cat <<EOF
Installed ralphx:
  $INSTALL_PATH

Installed doctor command:
  $DOCTOR_INSTALL_PATH

Built binaries:
  $TARGET_BINARY
  $DOCTOR_BINARY

Repository:
  $SCRIPT_DIR

If $BIN_DIR is not on PATH, add it:
  export PATH="$BIN_DIR:\$PATH"
EOF
}

main "$@"
