# ralphx

[English](README.md) | [中文](docs/zh/README.md)

Quick links:
- [Installation](docs/en/installation.md)
- [Methodology](docs/en/methodology.md)
- [Production SOP](docs/en/production-sop.md)
- [Parallel protocol v0](docs/en/go-parallel-protocol-v0.md)
- [Flowcharts / Chain Diagram](docs/en/architecture.md)
- [Codex Self-Loop Mechanism](docs/en/codex-self-loop.md)

`ralphx` is a Go-based outer-loop runner for Codex and coding agents.

It is designed for one core goal:
- let the agent keep working with the current tools until the real task is done
- keep completion gated by checklist / validation / leader-side rules
- make state, hooks, validation, and recovery explicit enough for daily use

## Why ralphx still matters with GPT-5.5

GPT-5.5 improves reasoning quality, but it does not replace execution discipline.
`ralphx` owns the runtime controls around the model: repo-local state, checklist gates, validation gates, Stop-hook continuation, session resume, and replanning when progress stalls.

Use `ralphx` for long-running repo work, release tasks, multi-round repairs, and any task where "almost done" is not an acceptable stop condition.
For one-off edits, regular Codex is usually enough.

## Install from GitHub release

No source build is required for normal use.

Latest release:

```bash
curl -fsSL https://github.com/ckken/ralphx/releases/latest/download/install.sh | bash
ralphx doctor
ralphx doctor --json
```

Install a specific version:

```bash
curl -fsSL https://github.com/ckken/ralphx/releases/download/v0.2.3/install.sh | VERSION=v0.2.3 bash
ralphx doctor
ralphx doctor --json
```

The installer downloads `SHA256SUMS` and verifies the release binaries before activation.
It also installs the `ralphx` Codex skill into `~/.codex/skills/ralphx`.
It also installs managed Codex hooks into `~/.codex/hooks.json`.

You can also install or refresh the skill from the CLI:

```bash
ralphx skill install
ralphx skill install --project
```

Optional: discover or install the curated subagent set when a task is explicitly delegated:

```bash
ralphx agents discover
ralphx agents install
ralphx agents install --project
```

You can also install or refresh the hooks from the CLI:

```bash
ralphx hook install
ralphx hook uninstall
```

## Persistent execution path

The installer persists the active binary paths in:

```bash
~/.config/ralphx/current.env
```

Downloaded release binaries are stored under:

```bash
~/.local/share/ralphx/releases/
```

The `ralphx` command is the stable wrapper in `~/.local/bin` that always reads the current persisted execution path.

Inspect the active persisted execution state:

```bash
ralphx current
```

Create a task and checklist from a goal:

```bash
ralphx plan --goal "finish the current migration batch" --out tasks/migration.md
```

Regenerate the next task/checklist from current state:

```bash
ralphx replan --task tasks/migration.md
```

Resume the previous Codex session when it is still fresh:

```bash
ralphx run --task tasks/migration.md --resume --session-expiry 24h
```

Activate the workflow in a fresh Codex session:

```text
$ralphx
```

If you just changed hooks, start a new session first so `UserPromptSubmit` can fire.
