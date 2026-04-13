#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if command -v go >/dev/null 2>&1; then
  cd "$SCRIPT_DIR"
  exec go run ./cmd/ralphx-doctor "$@"
fi

echo "ralphx doctor"
echo "[missing] go"
echo "Install Go and retry so the Go-native doctor can run."
exit 1
