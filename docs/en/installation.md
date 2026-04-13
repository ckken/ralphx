# Installation

## Preferred: install from GitHub release

No source build is required for normal usage.

Install the latest release:

```bash
curl -fsSL https://github.com/ckken/ralphx/releases/latest/download/install.sh | bash
```

Install a specific version:

```bash
curl -fsSL https://github.com/ckken/ralphx/releases/download/v0.1.2/install.sh | VERSION=v0.1.2 bash
```

The installer downloads `SHA256SUMS` and verifies the release binaries before activating them.
It also installs the `ralphx` Codex skill into `~/.codex/skills/ralphx`.

You can also install or refresh the skill from the CLI:

```bash
ralphx skill install
ralphx skill install --project
```

## Verify

```bash
ralphx-doctor
ralphx --help
ralphx version
ralphx current
```

## Common Flows

Create a task and checklist from a goal:

```bash
ralphx plan --goal "finish the current migration batch" --out tasks/migration.md
```

Regenerate the next task/checklist from current state:

```bash
ralphx replan --task tasks/migration.md
```

Resume a fresh-enough Codex session:

```bash
ralphx run --task tasks/migration.md --resume --session-expiry 24h
```
