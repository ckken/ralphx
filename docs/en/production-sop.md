# Production SOP

This SOP is the recommended operating path for using `ralphx` in a real repository.

## 1. Installation

Use release installation, not source installation.

Latest:

```bash
curl -fsSL https://github.com/ckken/ralphx/releases/latest/download/install.sh | bash
```

Specific version:

```bash
curl -fsSL https://github.com/ckken/ralphx/releases/download/v0.1.0/install.sh | VERSION=v0.1.0 bash
```

The installer persists the active execution path in:

```bash
~/.config/ralphx/current.env
```

This allows the wrapper commands to remain stable while the underlying versioned binaries live under `~/.local/share/ralphx/releases/`.

## 2. Preflight

Run:

```bash
ralphx-doctor
```

Minimum expected:
- `codex` is available
- `git` is available if you want git-aware completion checks

## 3. Prepare inputs

Create:
- one task file describing the total objective
- one checklist file for bounded milestones

Rules:
- task file = overall objective
- checklist = hard remaining work
- checklist items should be short, independently understandable, and verification-friendly

## 4. Choose execution mode

### Single-worker mode
Use when:
- work is tightly coupled
- file conflicts are likely
- you want the safest baseline path

```bash
ralphx --task task.md --checklist task.checklist.md --workdir /repo
```

### Parallel mode
Use when:
- checklist items are separable
- each item is a bounded slice
- you want faster local execution

```bash
ralphx --task task.md --checklist task.checklist.md --workdir /repo --workers 3
```

Recommended production rule:
- default to `--workers 1`
- only increase workers after the checklist is clearly decomposed

## 5. Add validation

Set `TESTS_CMD` whenever possible.

Examples:

```bash
export TESTS_CMD='go test ./...'
```

```bash
export TESTS_CMD='bun test && bun run lint'
```

```bash
export TESTS_CMD='pytest -q'
```

## 6. Run

Typical production invocation:

```bash
export CODEX_CMD=codex
export CODEX_ARGS='-m gpt-5.4'
export TESTS_CMD='go test ./...'

ralphx   --task docs/tasks/release-task.md   --checklist docs/tasks/release-task.checklist.md   --workdir /path/to/repo   --workers 3
```

## 7. Inspect outputs

Key files after a run:
- `.ralphx/last-result.json`
- `.ralphx/state.json`
- `.ralphx/stats.json`
- `.ralphx/logs/`
- `.ralphx/results/` (parallel mode)
- `.ralphx/runtime/loop-output.schema.json`

Interpretation:
- `complete`: all gates passed and leader accepted completion
- `in_progress`: work remains or a complete signal was downgraded
- `blocked`: real blocker, invalid output, or validation failure

## 8. GitHub release promotion SOP

Before tagging a release:
- run `go build ./...`
- run `go test ./...`
- run `ralphx-doctor`
- run one single-worker smoke path
- run one parallel smoke path if you intend to use `--workers`
- inspect `.ralphx/last-result.json`

Tag and push:

```bash
git status
git add .
git commit -m "feat: release prep"
git push origin main
git tag v0.1.0
git push origin v0.1.0
```

## 9. Operating boundaries

Do:
- use bounded checklist items
- keep validation cheap but meaningful
- review result files after blocked runs
- start with lower worker counts

Do not:
- use parallel mode for overlapping risky edits without a checklist
- assume worker completion equals total completion
- skip validation on production changes unless absolutely necessary
