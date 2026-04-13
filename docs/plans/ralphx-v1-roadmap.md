# ralphx v1.0 Roadmap

Goal: ship `ralphx` as a complete outer-loop runtime with planning, replanning, session rollover, hook guards, and repo-local state strong enough for broad daily use.

## v1 Themes

- one binary, one install surface
- planning and execution as first-class commands
- deterministic stop/continue control
- durable repo-local state
- enough hooks and tests to trust the runtime under real sessions

## Milestone 1: Runtime Stability

- finish wiring native hook entrypoints into real Codex lifecycle use
- make prompt-submit and stop hooks observable in both repo-local and user-level logs
- remove remaining local-only install assumptions from the user workflow
- close the remaining guidance/state drift gaps

## Milestone 2: Full Autodrive

- strengthen `produce_plan` application semantics
- preserve completed checklist items more robustly across replans
- trigger replanning with clearer next-step state instead of generic summaries
- support cleaner task/checklist regeneration for long-running work

## Milestone 3: Session-Scoped Runtime

- expand session metadata beyond thread id and timestamp
- make resume/expiry/fresh-session transitions explicit in state
- ensure hook guards understand current session ownership
- eliminate hidden fallback behavior in session rollover paths

## Milestone 4: Hook Platform

- stabilize `hook install` / `hook uninstall`
- add repo-local and user-level hook telemetry by default
- add end-to-end hook tests that simulate Codex native payloads
- document native hook expectations and failure modes

## Milestone 5: Architecture Refactors

- keep shrinking process-wide mutable state in downstream consumers
- favor session-local handles over singleton registries
- make task/plan/registry boundaries easier to test in isolation

## Release Gate for v1.0

- `plan`, `run`, `replan`, `resume`, and hook guard flows all have E2E coverage
- install/uninstall and hook install/uninstall are one-command reliable
- no critical state-loss or silent-stop paths remain
- docs reflect the shipped control model, not just the intended one
