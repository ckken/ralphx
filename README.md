# ralphx

[English](README.md) | [中文](docs/zh/README.md)

Quick links:
- [Installation](docs/en/installation.md)
- [Methodology](docs/en/methodology.md)
- [Production SOP](docs/en/production-sop.md)
- [Parallel protocol v0](docs/en/go-parallel-protocol-v0.md)
- [Flowcharts / Chain Diagram](docs/en/architecture.md)

`ralphx` is a Go-based outer-loop runner for Codex and coding agents.

It is designed for one core goal:
- let the agent keep working with the current tools until the real task is done
- keep completion gated by checklist / validation / leader-side rules
- support local multi-worker execution when the task is checklist-decomposable

## What it does

`ralphx` gives you a local-first execution loop that can:
- read a task file
- optionally read a checklist
- invoke `codex exec` with a strict JSON contract
- persist run state under `.ralphx/`
- reject premature completion
- run validation commands
- split checklist items into parallel worker jobs with `--workers N`

## Current shape

Production-relevant surfaces:
- `ralphx`: main CLI
- `ralphx-doctor`: environment/self-check command
- `install.sh`: build + install wrappers into `~/.local/bin`
- `.ralphx/`: local runtime state, logs, results, runtime schema

Key runtime behavior:
- single-run mode works with `--workers 1` (default)
- checklist-driven parallel mode works with `--workers N`
- leader owns overall completion
- worker results are advisory; final completion is leader-gated

## Install

```bash
git clone https://github.com/ckken/ralphx.git
cd ralphx
./install.sh
ralphx-doctor
```

If needed:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

## Runtime dependencies

Required:
- `go` (for local build/install from source)
- `codex`

Recommended:
- `git`
- `gh`
- `bash`
- `python3`

Optional:
- `jq` (legacy-only helper; not required by the Go-native main path)

## Quick start

Run a single-worker task:

```bash
ralphx --task ./examples/sample-task.md --workdir /path/to/repo
```

Run with an explicit checklist:

```bash
ralphx   --task ./examples/sample-task.md   --checklist ./examples/sample-task.checklist.md   --workdir /path/to/repo
```

Run checklist items in parallel:

```bash
ralphx   --task ./examples/sample-task.md   --checklist ./examples/sample-task.checklist.md   --workdir /path/to/repo   --workers 3
```

## Common environment variables

```bash
export CODEX_CMD=codex
export CODEX_ARGS='-m gpt-5.4'
export TESTS_CMD='go test ./...'
export MAX_ITERATIONS=0
export MAX_NO_PROGRESS=0
export ROUND_TIMEOUT_SECONDS=1800
export RALPHX_WORKERS=3
```

Meaning:
- `CODEX_CMD`: agent executable to invoke
- `CODEX_ARGS`: extra args passed to the agent
- `TESTS_CMD`: post-round validation command
- `MAX_ITERATIONS=0`: no hard iteration cap
- `MAX_NO_PROGRESS=0`: no no-progress stop gate
- `ROUND_TIMEOUT_SECONDS`: per-round timeout in seconds
- `RALPHX_WORKERS`: default worker count for parallel mode

## Recommended production path

1. Run `ralphx-doctor`
2. Prepare a task file and checklist
3. Start with `--workers 1` unless the checklist items are clearly independent
4. Use `--workers N` only when checklist items are bounded and separable
5. Set `TESTS_CMD` so every successful round is validated
6. Review `.ralphx/last-result.json`, `.ralphx/state.json`, and `.ralphx/results/` after runs

See the full rollout guide in [Production SOP](docs/en/production-sop.md).

## Output contract

Each agent round must return one JSON object:

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

- `ralphx`
- `ralphx-doctor`
