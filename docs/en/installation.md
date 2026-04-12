# Installation

## Required tools

- `bash`
- `jq`
- `python3`
- `codex`

## Recommended tools

- `git`
- `gh`
- `timeout` or `gtimeout`

## Install

```bash
git clone https://github.com/ckken/codex-ralph.git
cd codex-ralph
./install.sh
```

By default the command is installed into `~/.local/bin`.

If your shell cannot find `codex-ralph`, add:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

## Verify

```bash
codex-ralph-doctor
codex-ralph --help
```

## Uninstall

```bash
./uninstall.sh
```
