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
