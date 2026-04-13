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

## Verify

```bash
ralphx-doctor
ralphx --help
ralphx version
ralphx current
```
