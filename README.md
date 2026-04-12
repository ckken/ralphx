# codex-ralph

[English](README.md) | [中文](docs/zh/README.md)

Quick links:
- [Docs](docs/README.md)
- [Installation](docs/en/installation.md)
- [Methodology](docs/en/methodology.md)
- [Flowcharts / Chain Diagram](docs/en/architecture.md)

`codex-ralph` is a Bash orchestration layer for Codex.

It turns Codex into a controlled outer-loop system:

- task-driven
- checklist-gated
- validation-first
- resistant to premature completion

## What problem it solves

Codex can implement a slice and then declare success too early. `codex-ralph` adds an outer loop that:

- reads a task file
- optionally reads a checklist
- calls `codex exec` non-interactively
- requires strict JSON output
- validates each successful round
- refuses weak completion

The result is a repeatable “keep going until the real task is done” workflow.

## Project layout

- `codex-loop.sh`: main executor
- `doctor.sh`: dependency and environment check
- `install.sh`: install command wrappers into `~/.local/bin`
- `uninstall.sh`: remove installed wrappers
- `prompts/loop-system-prompt.md`: system prompt for the loop
- `schemas/loop-output.schema.json`: strict response schema
- `docs/en/`: English docs
- `docs/zh/`: Chinese docs
- `examples/`: task and checklist examples
- `tasks/`: real task examples used during development

## Install

```bash
git clone https://github.com/ckken/codex-ralph.git
cd codex-ralph
./install.sh
codex-ralph-doctor
```

If needed:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

## Required dependencies

- `bash`
- `jq`
- `python3`
- `codex`

Recommended:

- `git`
- `gh`
- `timeout` or `gtimeout`

## Quick start

Run with a task file:

```bash
codex-ralph --task ./examples/sample-task.md --workdir /path/to/repo
```

Run with an explicit checklist:

```bash
codex-ralph \
  --task ./examples/sample-task.md \
  --checklist ./examples/sample-task.checklist.md \
  --workdir /path/to/repo
```

## Runtime controls

Common environment variables:

```bash
export CODEX_CMD=codex
export CODEX_ARGS='-m gpt-5.4-mini'
export TESTS_CMD='bun src/index.ts --help && bash scripts/verify-golden.sh --skip-build'
export MAX_ITERATIONS=0
export MAX_NO_PROGRESS=0
export ROUND_TIMEOUT_SECONDS=1800
```

Meaning:

- `MAX_ITERATIONS=0`: no hard iteration cap
- `MAX_NO_PROGRESS=0`: no no-progress stop gate
- `CHECKLIST_FILE`: force a checklist file path
- `TESTS_CMD`: validation chain to run after successful rounds

## Method

The method is documented in:

- [docs/en/methodology.md](docs/en/methodology.md)
- [docs/en/installation.md](docs/en/installation.md)
- [docs/en/architecture.md](docs/en/architecture.md)

Short version:

1. Codex does the local work.
2. Bash owns the control loop.
3. Checklist items are hard remaining work.
4. Validation gates stop bad progress.
5. Completion is accepted only when the total objective is truly done.

## Output contract

Each Codex round must return one JSON object:

```json
{
  "status": "in_progress|blocked|complete",
  "exit_signal": true,
  "files_modified": 0,
  "tests_passed": false,
  "blockers": [],
  "summary": ""
}
```

## Installed commands

- `codex-ralph`
- `codex-ralph-doctor`
