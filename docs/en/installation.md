# Installation

## Required tools

Required:
- `go`
- `codex`

Recommended:
- `git`
- `gh`
- `bash`
- `python3`

Optional:
- `jq` (legacy-only helper; not required by the Go-native main path)

## Install from source

```bash
git clone https://github.com/ckken/ralphx.git
cd ralphx
./install.sh
```

By default the commands are installed into `~/.local/bin`.

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

## Install to a custom prefix

```bash
PREFIX=/custom/bin ./install.sh
```

## Uninstall

```bash
./uninstall.sh
```
