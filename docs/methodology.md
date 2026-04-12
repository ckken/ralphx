# Methodology

`codex-ralph` is a Bash orchestration layer for Codex. The method is simple:

1. Treat the task file as the source of truth.
2. Call Codex in non-interactive mode.
3. Require a strict JSON result.
4. Refuse to stop on weak completion signals.
5. Run validation between iterations.
6. Continue until the real task is done or a true blocker appears.

## Core principles

### 1. Outer-loop control

The model does not decide alone when work is complete. The Bash loop owns:

- iteration count
- timeout
- validation
- no-progress detection
- checklist gating

### 2. Checklist gating

A markdown checklist can be attached to the task. Any unchecked item is treated as hard remaining work.

That means:

- a local slice can finish without ending the total task
- `complete` is rejected while checklist items remain

### 3. Validation-first progression

Each loop can run a low-cost validation chain before moving on.

Typical validation:

- `bun src/index.ts --help`
- `bash scripts/verify-golden.sh --skip-build`
- `bash scripts/verify-batch.sh --skip-build`

### 4. Premature-complete defense

`codex-ralph` rejects completion when:

- Codex returns `complete`
- but no new changes were actually made

This avoids the common failure mode where the model “concludes” instead of progressing.

## Recommended workflow

1. Write a total-task file.
2. Add a checklist if the task spans multiple milestones.
3. Attach the lowest-cost useful validation command chain.
4. Let the loop run.
5. Only accept final completion when the checklist is empty and validation is green.
