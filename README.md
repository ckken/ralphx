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
ralphx-doctor
```

Install a specific version:

```bash
curl -fsSL https://github.com/ckken/ralphx/releases/download/v0.1.2/install.sh | VERSION=v0.1.2 bash
ralphx-doctor
```

The installer downloads `SHA256SUMS` and verifies the release binaries before activation.

## Persistent execution path

The installer persists the active binary paths in:

```bash
~/.config/ralphx/current.env
```

Downloaded release binaries are stored under:

```bash
~/.local/share/ralphx/releases/
```

The `ralphx` and `ralphx-doctor` commands are stable wrappers in `~/.local/bin` that always read the current persisted execution path.

Inspect the active persisted execution state:

```bash
ralphx current
```
