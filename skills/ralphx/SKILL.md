---
name: ralphx
description: Use when you want Codex to keep pushing a repo task to completion with an outer-loop mindset. Covers task truth, checklist gating, validation discipline, recovery, and ralphx installation or extension when working on the ralphx project itself.
---

# ralphx

## When To Use

Use this skill whenever you want persistent outer-loop execution in the current repo:

- continue a multi-step task instead of stopping after one patch
- treat the task file as the source of truth
- keep a checklist as real remaining work
- validate after meaningful progress
- recover from blocked or partial runs

Use the repo-specific parts of this skill when the task also involves the `ralphx` repository itself:

- install or upgrade `ralphx`
- initialize Codex with the repo's expected workflow
- add or debug task/checklist/validation files
- extend the loop, prompt, or installer behavior

## Core Model

`ralphx` is a leader-controlled outer loop around Codex.
The task file is the source of truth, checklist items are hard remaining work, and completion is accepted only when the loop output, validation, and state all line up.

## Operating Mode

When invoked in a non-`ralphx` repository:

- do not refuse just because the current repo is not the `ralphx` codebase
- apply the same outer-loop discipline manually if the `ralphx` binary is not installed
- keep moving on the user task until the real objective is done, blocked, or needs clarification
- prefer low-risk, well-bounded progress when the task is large or ambiguous

## Model Routing

Use the strongest reasoning model for coordination, not routine code writing:

- `gpt-5.4 high`: task decomposition, logic-heavy reasoning, scheduling, conflict resolution, and final review
- `gpt-5.4-mini`: default code-writing and patch generation
- `gpt-5.3-codex` or `gpt-5.2`: narrower implementation passes when token cost matters and the change is well-scoped

Policy:

- keep `gpt-5.4 high` on the critical path for planning and decisions
- prefer smaller models for direct edits, repetitive transformations, and mechanical fixes
- only escalate to a larger model when the code path is ambiguous, high-risk, or needs broader context

## Subagent Routing

Only use subagents when the user explicitly asks for delegation, parallel work, or a coordinated multi-agent pass.

When subagents are allowed:

- coordinator / planner / reviewer: `gpt-5.4` with `high` reasoning
- default coding worker: `gpt-5.4-mini`
- bounded code-edit worker where token cost matters: `gpt-5.3-codex` first, `gpt-5.2-codex` if available in the environment
- verification or test-repair worker: `gpt-5.4-mini`, escalate to `gpt-5.3-codex` only when failure analysis is unclear

Execution rules:

- keep the immediate blocking step local unless delegation clearly shortens the critical path
- give each worker a narrow, disjoint write scope
- do not spawn a high-cost reviewer for straightforward mechanical edits
- report the chosen subagent model in the commentary when delegation is used

## Quick Start

For any repo:

1. Read the task statement and infer the total objective.
2. If a checklist exists, treat unchecked items as hard remaining work.
3. Make one bounded step of real progress.
4. Re-validate and continue until the full objective is done.

For the `ralphx` project itself:

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

## Task Execution With The Binary

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

If the binary is not available in the current repo, keep the same recovery logic manually:

1. restate the task boundary
2. re-check unfinished checklist items or implied remaining work
3. verify the current patch state
4. continue with the next bounded step instead of declaring done early

## Editing The ralphx Repo

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
