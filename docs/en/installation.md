# Installation

## Preferred: install from GitHub release

No source build is required for normal usage.

Install the latest release:

```bash
curl -fsSL https://github.com/ckken/ralphx/releases/latest/download/install.sh | bash
```

Install a specific version:

```bash
curl -fsSL https://github.com/ckken/ralphx/releases/download/v0.1.0/install.sh | VERSION=v0.1.0 bash
```

## Runtime requirements

Required:
- `codex`
- `curl` or `wget`

Recommended:
- `git`
- `gh`
- `bash`
- `python3`

Optional:
- `jq` (legacy-only helper; not required by the Go-native main path)

## Installed layout

Wrappers:
- `~/.local/bin/ralphx`
- `~/.local/bin/ralphx-doctor`

Persistent execution state:
- `~/.config/ralphx/current.env`

Downloaded release binaries:
- `~/.local/share/ralphx/releases/`

If your shell cannot find `ralphx`, add:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

## Verify

```bash
ralphx-doctor
ralphx --help
ralphx version
```

## Uninstall

```bash
curl -fsSL https://github.com/ckken/ralphx/releases/latest/download/uninstall.sh | bash
```
