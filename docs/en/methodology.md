# Methodology

`ralphx` is a Go-based outer-loop runner for Codex and coding agents.

Its method is simple:

1. Treat the task file as the source of truth.
2. Optionally attach a checklist and treat unchecked items as hard remaining work.
3. Invoke the agent with a strict JSON result contract.
4. Persist local runtime state under `.ralphx/`.
5. Reject premature completion on the leader side.
6. Run validation between iterations when configured.
7. In parallel mode, let workers execute bounded checklist slices while the leader remains the only authority for total completion.

## Core principles

### 1. Outer-loop control

The agent does not decide alone when the work is complete. `ralphx` owns:
- iteration count
- timeout
- checklist gating
- no-progress detection
- validation
- final completion acceptance

### 2. Checklist gating

Unchecked checklist items are treated as real remaining work.

That means:
- a local slice can finish without ending the total task
- `complete` is rejected while checklist items remain
- in parallel mode, workers can finish slices but only the leader can end the run

### 3. Validation-first progression

Use `TESTS_CMD` to keep progress honest.

Typical validation:
- `go test ./...`
- `bun src/index.ts --help`
- `bash scripts/verify-golden.sh --skip-build`

### 4. Premature-complete defense

`ralphx` rejects completion when:
- the agent returns `complete`
- but no meaningful progress exists, or
- checklist items still remain

### 5. Parallel mode discipline

Parallel mode is for bounded checklist decomposition, not unconstrained swarming.

Use `--workers N` only when:
- checklist items are separable
- items do not require conflicting writes
- the leader can aggregate results and run final validation
