# Production SOP

This SOP is the recommended operating path for using `ralphx` in a real repository.

## 1. Preflight

Run:

```bash
ralphx-doctor
```

Minimum expected:
- `codex` is available
- `go` is available if building/installing from source
- `git` is available if you want git-aware completion checks

## 2. Prepare inputs

Create:
- one task file describing the total objective
- one checklist file for bounded milestones

Rules:
- task file = overall objective
- checklist = hard remaining work
- checklist items should be short, independently understandable, and verification-friendly

Good checklist items:
- add command X
- update config loader
- add smoke test
- update README section

Bad checklist items:
- finish everything
- fix all remaining issues
- make it production ready

## 3. Choose execution mode

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

## 4. Add validation

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

Without validation, `ralphx` can still run, but production confidence is lower.

## 5. Run

Typical production invocation:

```bash
export CODEX_CMD=codex
export CODEX_ARGS='-m gpt-5.4'
export TESTS_CMD='go test ./...'

ralphx   --task docs/tasks/release-task.md   --checklist docs/tasks/release-task.checklist.md   --workdir /path/to/repo   --workers 3
```

## 6. Inspect outputs

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

## 7. Production rollout checklist

Before pushing to GitHub:
- run `go build ./...`
- run `go test ./...`
- run `ralphx-doctor`
- run one single-worker smoke path
- run one parallel smoke path if you intend to use `--workers`
- inspect `.ralphx/last-result.json`
- confirm checklist is fully checked if the run ended `complete`

## 8. GitHub promotion SOP

Recommended sequence:

```bash
git status
go build ./...
go test ./...
git add .
git commit -m "feat: productionize Go ralphx workflow"
git push origin main
```

If creating the first public-facing rollout, also verify:
- README install instructions are current
- doctor output matches reality
- install/uninstall scripts work on a clean prefix

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
