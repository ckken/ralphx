# Production SOP

This SOP covers the full `ralphx` execution loop:

1. install the release binary
2. verify the active wrapper
3. prepare hooks and runtime state
4. generate or load a task and checklist
5. run the loop with validation
6. replan automatically when blocked or stale
7. resume until the task is actually done

With GPT-5.5, use `ralphx` as the execution-discipline layer, not as a substitute for model reasoning.
GPT-5.5 handles better planning and coding; `ralphx` keeps state, checklist gates, validation gates, Stop-hook continuation, and recovery honest.

## 1. Install

Use release installation, not source installation.

Latest:

```bash
curl -fsSL https://github.com/ckken/ralphx/releases/latest/download/install.sh | bash
```

Specific version:

```bash
curl -fsSL https://github.com/ckken/ralphx/releases/download/v0.2.3/install.sh | VERSION=v0.2.3 bash
```

The installer verifies release binaries against `SHA256SUMS` before activating them.
The installer persists the active execution path in:

```bash
~/.config/ralphx/current.env
```

Confirm the active wrapper and binary:

```bash
ralphx doctor
ralphx current
```

## 2. Prepare Hooks

Install the managed stop hook so the runtime can keep control until the task is genuinely complete:

```bash
ralphx hook install
ralphx hook status --workdir "$PWD"
```

If you need to remove the managed hook later:

```bash
ralphx hook uninstall
```

## 3. Start From a Goal

When you only have a goal statement, generate both the task and checklist and immediately hand off to the runner:

```bash
ralphx plan --goal "finish the current migration batch" --out tasks/migration.md --execute
```

If you already have a task and checklist, start the loop directly:

```bash
TESTS_CMD="go test ./..." ralphx run --task tasks/migration.md --checklist tasks/migration.checklist.md --resume --session-expiry 24h
```

Recommended defaults:

- keep `TESTS_CMD` set for real validation
- leave `RALPHX_AUTO_REPLAN=1` enabled unless you are intentionally debugging
- use `--resume` only when the prior Codex session is still fresh enough to reuse

## 4. Keep the Loop Running

`ralphx run` handles the outer loop for you:

- it invokes Codex
- it writes `.ralphx/` state
- it runs validation after meaningful progress
- it preserves checklist gating as the completion authority
- it automatically triggers replanning when the run becomes blocked, stale, or no-progress

If the run returns with regenerated files and a replan warning, review the new task/checklist and rerun:

```bash
ralphx replan --task tasks/migration.md --execute
```

## 5. Recover and Resume

Use these commands when the workflow needs inspection or recovery:

```bash
ralphx current
ralphx hook status --workdir "$PWD"
```

Inspect the local runtime state under:

```bash
.ralphx/
```

If the session is still fresh, reuse it:

```bash
ralphx run --task tasks/migration.md --checklist tasks/migration.checklist.md --resume --session-expiry 24h
```

If the session is stale, let `ralphx` start a fresh one and continue from the task/checklist plus summary state.

## 6. Finish Cleanly

The loop should stop only when all of the following are true:

- the task is actually complete
- the checklist is closed
- required validation has passed
- the stop hook no longer has a reason to block exit

At the end, keep the final inspection simple:

```bash
ralphx current
ralphx hook status --workdir "$PWD"
```
