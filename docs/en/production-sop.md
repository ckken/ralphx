# Production SOP

## Installation

Use release installation, not source installation.

Latest:

```bash
curl -fsSL https://github.com/ckken/ralphx/releases/latest/download/install.sh | bash
```

Specific version:

```bash
curl -fsSL https://github.com/ckken/ralphx/releases/download/v0.1.2/install.sh | VERSION=v0.1.2 bash
```

The installer verifies release binaries against `SHA256SUMS` before activating them.

The installer persists the active execution path in:

```bash
~/.config/ralphx/current.env
```

Inspect the current persisted execution target with:

```bash
ralphx current
```
