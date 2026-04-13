#!/usr/bin/env bash
set -euo pipefail

REPO="${GITHUB_REPO:-ckken/ralphx}"
TOOL_NAME="ralphx"
DOCTOR_NAME="ralphx-doctor"
SKILL_NAME="ralphx"
LEGACY_SKILL_NAME="ralphx-drive"
VERSION="${VERSION:-latest}"
BIN_DIR="${PREFIX:-${HOME}/.local/bin}"
DATA_DIR="${RALPHX_DATA_DIR:-${HOME}/.local/share/ralphx}"
CONFIG_DIR="${RALPHX_CONFIG_DIR:-${HOME}/.config/ralphx}"
CURRENT_ENV="$CONFIG_DIR/current.env"
RELEASES_DIR="$DATA_DIR/releases"
BACKUP_SUFFIX="$(date +%Y%m%d%H%M%S)"
CODEX_HOME_DIR="${CODEX_HOME:-${HOME}/.codex}"
SKILLS_DIR="$CODEX_HOME_DIR/skills"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]:-$0}")" && pwd)"

info() {
  printf '[install] %s
' "$*"
}

die() {
  printf '%s
' "$*" >&2
  exit 1
}

detect_platform() {
  local os arch
  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  arch="$(uname -m)"
  case "$os" in
    linux|darwin) ;;
    msys*|mingw*|cygwin*) os="windows" ;;
    *) die "Unsupported OS: $os" ;;
  esac
  case "$arch" in
    x86_64|amd64) arch="amd64" ;;
    arm64|aarch64) arch="arm64" ;;
    *) die "Unsupported architecture: $arch" ;;
  esac
  printf '%s %s
' "$os" "$arch"
}

release_base_url() {
  if [[ "$VERSION" == "latest" ]]; then
    printf 'https://github.com/%s/releases/latest/download
' "$REPO"
  else
    printf 'https://github.com/%s/releases/download/%s
' "$REPO" "$VERSION"
  fi
}

download() {
  local url="$1"
  local out="$2"
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$url" -o "$out"
  elif command -v wget >/dev/null 2>&1; then
    wget -qO "$out" "$url"
  else
    die "Need curl or wget to download release assets"
  fi
}

verify_checksum() {
  local sums_file="$1"
  local asset_name="$2"
  local asset_path="$3"
  local expected actual
  expected="$(grep "  $asset_name$" "$sums_file" | awk '{print $1}')"
  [[ -n "$expected" ]] || die "Checksum for $asset_name not found in SHA256SUMS"
  if command -v sha256sum >/dev/null 2>&1; then
    actual="$(sha256sum "$asset_path" | awk '{print $1}')"
  elif command -v shasum >/dev/null 2>&1; then
    actual="$(shasum -a 256 "$asset_path" | awk '{print $1}')"
  else
    die "Need sha256sum or shasum to verify release assets"
  fi
  [[ "$expected" == "$actual" ]] || die "Checksum mismatch for $asset_name"
}

write_wrapper() {
  local install_path="$1"
  local key="$2"

  mkdir -p "$BIN_DIR"
  if [[ -e "$install_path" && ! -L "$install_path" ]]; then
    cp "$install_path" "${install_path}.bak.${BACKUP_SUFFIX}"
    info "Backed up existing file to ${install_path}.bak.${BACKUP_SUFFIX}"
  fi

  cat > "$install_path" <<EOF
#!/usr/bin/env bash
set -euo pipefail
CURRENT_ENV="${CURRENT_ENV}"
[[ -f "\$CURRENT_ENV" ]] || { echo "Missing ralphx install state: \$CURRENT_ENV" >&2; exit 1; }
# shellcheck disable=SC1090
source "\$CURRENT_ENV"
TARGET="\${$key:-}"
[[ -n "\$TARGET" ]] || { echo "Missing target for $key in \$CURRENT_ENV" >&2; exit 1; }
exec "\$TARGET" "\$@"
EOF
  chmod +x "$install_path"
}

write_skill_bundle() {
  local skill_dir="$1"
  mkdir -p "$skill_dir/agents"

  cat > "$skill_dir/SKILL.md" <<'EOF'
---
name: ralphx
description: Use when working in the ralphx repo and you need a complete workflow for installation, skill setup, running tasks, validation, recovery, or extending the outer loop with Codex.
---

# ralphx

## When To Use

Use this skill whenever the task involves the `ralphx` repository itself:

- install or upgrade `ralphx`
- initialize Codex with the repo's expected workflow
- run a task through `ralphx`
- add or debug task/checklist/validation files
- recover a stalled or partial run
- extend the loop, prompt, or installer behavior

## Core Model

`ralphx` is a leader-controlled outer loop around Codex.
The task file is the source of truth, checklist items are hard remaining work, and completion is accepted only when the loop output, validation, and state all line up.

## Quick Start

1. Run `ralphx doctor`.
2. Confirm the active binary with `ralphx current`.
3. Run the task with `ralphx run --task <task-file> --checklist <checklist-file> --workdir .`.
4. Keep `TESTS_CMD` set when validation matters.

## Installation

Preferred install path:

```bash
curl -fsSL https://github.com/ckken/ralphx/releases/latest/download/install.sh | bash
```

The installer:

- verifies release checksums
- installs `ralphx` and `ralphx-doctor`
- installs the Codex skill to `~/.codex/skills/ralphx`

If you need a pinned version, pass `VERSION=vX.Y.Z`.

## Task Execution

Use this shape for most runs:

```bash
ralphx run --task tasks/<name>.md --checklist tasks/<name>.checklist.md --workdir .
```

Prefer these defaults unless the repo state says otherwise:

- `--task` is required.
- `--checklist` is optional, but use it when the task can be split.
- `--workdir .` is usually correct inside the repo.
- `--tests-cmd` or `TESTS_CMD` should define the validation chain.
- `--prompt` and `--schema` are for custom loop surfaces.

If a task is not decomposable, run without a checklist.
If a task has a checklist, treat unchecked items as unfinished work even if a partial slice succeeds.

## Validation

Keep validation close to the change.

- Use `TESTS_CMD` for the normal validation command.
- Keep the command deterministic and repo-local.
- Prefer a fast smoke check before a slower full suite when both exist.

Common examples:

```bash
go test ./...
```

```bash
bash scripts/verify-golden.sh --skip-build
```

## Recovery

If the loop stops early or reports blocked:

1. Check `ralphx current`.
2. Inspect the `.ralphx/` state under the working directory.
3. Re-run `ralphx doctor` if the wrapper or binary path looks stale.
4. Re-read the task file, checklist, and validation command before continuing.

## Editing This Repo

When changing `ralphx` itself:

- keep diffs small
- update docs if the execution contract changes
- preserve the strict JSON output schema for the loop
- keep the installer and the skill in sync

## Output Contract

The loop should not declare success prematurely.

- `complete` means the total task is done
- `blocked` means a real blocker exists
- `in_progress` means more work remains
- checklist items are not cosmetic; they gate completion
EOF

  cat > "$skill_dir/agents/openai.yaml" <<'EOF'
interface:
  display_name: "ralphx"
  short_description: "Complete workflow for the ralphx repo"
  default_prompt: "Use $ralphx to install, run, debug, or extend this repo with Codex."
EOF
}

install_skill() {
  local skill_dst="$SKILLS_DIR/$SKILL_NAME"
  local legacy_skill_dst="$SKILLS_DIR/$LEGACY_SKILL_NAME"
  local repo_skill_dir="$SCRIPT_DIR/skills/$SKILL_NAME"

  mkdir -p "$SKILLS_DIR"
  rm -rf "$legacy_skill_dst"
  if [[ -d "$repo_skill_dir" ]]; then
    rm -rf "$skill_dst"
    cp -R "$repo_skill_dir" "$skill_dst"
    info "Installed skill bundle from repository: $skill_dst"
  else
    rm -rf "$skill_dst"
    write_skill_bundle "$skill_dst"
    info "Installed embedded skill bundle: $skill_dst"
  fi
}

main() {
  read -r os arch < <(detect_platform)
  local version_dir target_dir base_url main_asset doctor_asset sums_asset main_target doctor_target sums_target
  version_dir="$VERSION"
  [[ "$version_dir" == "latest" ]] && version_dir="latest"
  target_dir="$RELEASES_DIR/$version_dir/$os-$arch"
  mkdir -p "$target_dir" "$CONFIG_DIR" "$BIN_DIR"

  base_url="$(release_base_url)"
  main_asset="$TOOL_NAME-$os-$arch"
  doctor_asset="$DOCTOR_NAME-$os-$arch"
  sums_asset="SHA256SUMS"
  if [[ "$os" == "windows" ]]; then
    main_asset+='.exe'
    doctor_asset+='.exe'
  fi

  main_target="$target_dir/$TOOL_NAME"
  doctor_target="$target_dir/$DOCTOR_NAME"
  sums_target="$target_dir/$sums_asset"
  [[ "$os" == "windows" ]] && main_target+='.exe' && doctor_target+='.exe'

  info "Downloading $sums_asset"
  download "$base_url/$sums_asset" "$sums_target"
  info "Downloading $main_asset"
  download "$base_url/$main_asset" "$main_target"
  info "Downloading $doctor_asset"
  download "$base_url/$doctor_asset" "$doctor_target"

  verify_checksum "$sums_target" "$main_asset" "$main_target"
  verify_checksum "$sums_target" "$doctor_asset" "$doctor_target"
  chmod +x "$main_target" "$doctor_target" || true

  cat > "$CURRENT_ENV" <<EOF
RALPHX_VERSION="$VERSION"
RALPHX_BINARY="$main_target"
RALPHX_DOCTOR_BINARY="$doctor_target"
EOF

  write_wrapper "$BIN_DIR/$TOOL_NAME" "RALPHX_BINARY"
  write_wrapper "$BIN_DIR/$DOCTOR_NAME" "RALPHX_DOCTOR_BINARY"
  install_skill

  cat <<EOF
Installed ralphx from GitHub release:
  repo:    $REPO
  version: $VERSION
  os/arch: $os/$arch

Wrappers:
  $BIN_DIR/$TOOL_NAME
  $BIN_DIR/$DOCTOR_NAME

Installed skill:
  $SKILLS_DIR/$SKILL_NAME

Persistent execution state:
  $CURRENT_ENV

Downloaded binaries:
  $main_target
  $doctor_target

Checksums file:
  $sums_target

If $BIN_DIR is not on PATH, add it:
  export PATH="$BIN_DIR:\$PATH"
EOF
}

main "$@"
