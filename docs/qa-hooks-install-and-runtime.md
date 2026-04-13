# QA: Hooks Install And Runtime

Goal: verify `ralphx` hook installation, uninstall, and runtime visibility are reliable enough for everyday Codex sessions.

## Scope

- global `~/.codex/hooks.json` install/uninstall
- `UserPromptSubmit` visibility
- `Stop` guard behavior
- repo-local `.ralphx` logs
- user-level `~/.codex/log` logs

## Preconditions

- `ralphx` is on `PATH`
- `~/.codex/config.toml` contains:

```toml
[features]
codex_hooks = true
```

- start a **new Codex session** after changing native hook config

## Install

1. Run:

```bash
curl -fsSL https://github.com/ckken/ralphx/releases/latest/download/install.sh | bash
```

or:

```bash
ralphx hook install
```

2. Verify `~/.codex/hooks.json` contains both commands:

```bash
ralphx hook prompt-submit
ralphx hook stop-guard
```

3. Verify each managed entry has a `statusMessage`:

- `Activating ralphx workflow hooks`
- `Running ralphx stop guard`

## Prompt-Submit Runtime

1. Open a fresh Codex session in a test repository.
2. Enter:

```text
$ralphx 开始
```

3. Expected outcomes:

- terminal may show:
  - `[hook prompt-submit] ralphx mode active`
- user log must contain a new `prompt-submit` entry:

```bash
tail -n 20 ~/.codex/log/hooks-$(date +%F).jsonl
```

- if the repository has `.ralphx/`, repo-local log should also gain a new `prompt-submit` entry:

```bash
tail -n 20 .ralphx/logs/hooks-$(date +%F).jsonl
```

4. Required fields in the JSONL line:

- `event: "prompt-submit"`
- `decision.Reason: "prompt_submit"`
- `result.active: true`

## Stop Guard Runtime

### A. No task context

1. In a repository with no `.ralphx/state.json` and no task/checklist arguments, run:

```bash
ralphx hook stop-guard --workdir "$PWD" --json
```

2. Expected:

- exit code `0`
- `Allow: true`
- `Reason: "no_task_context"`

### B. Incomplete task context

1. Prepare:

- task file
- checklist with at least one unchecked item
- `.ralphx/last-result.json` with `status=in_progress`

2. Run:

```bash
ralphx hook stop-guard --task tasks/demo.md --checklist tasks/demo.checklist.md --workdir "$PWD" --json
```

3. Expected:

- non-zero exit
- `Allow: false`
- `Reason: "task_incomplete"`

## Uninstall

1. Run:

```bash
./uninstall.sh
```

or:

```bash
ralphx hook uninstall
```

2. Verify `~/.codex/hooks.json` no longer contains:

```bash
ralphx hook prompt-submit
ralphx hook stop-guard
```

3. Verify unrelated user hooks remain unchanged.

## Regression Checklist

- [ ] `ralphx hook install` succeeds
- [ ] managed `UserPromptSubmit` hook is present
- [ ] managed `Stop` hook is present
- [ ] prompt-submit writes user-level hook logs
- [ ] stop-guard allows when no task context exists
- [ ] stop-guard blocks when incomplete work remains
- [ ] `ralphx hook uninstall` removes only managed entries
