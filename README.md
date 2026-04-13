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

## Install from GitHub release

No source build is required for normal use.

Latest release:

```bash
curl -fsSL https://github.com/ckken/ralphx/releases/latest/download/install.sh | bash
ralphx doctor
```

Install a specific version:

```bash
curl -fsSL https://github.com/ckken/ralphx/releases/download/v0.1.2/install.sh | VERSION=v0.1.2 bash
ralphx doctor
```

The installer downloads `SHA256SUMS` and verifies the release binaries before activation.
It also installs the `ralphx` Codex skill into `~/.codex/skills/ralphx`.
It also installs managed Codex hooks into `~/.codex/hooks.json`.

You can also install or refresh the skill from the CLI:

```bash
ralphx skill install
ralphx skill install --project
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
