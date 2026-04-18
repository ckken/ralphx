# ralphx Runtime Refactor Plan

> For Hermes: this is the repo-tracked planning document for upgrading `ralphx` from a Codex-centric outer-loop runner into a pluggable, inspectable, validation-gated orchestration runtime.

**Goal:** land a staged refactor that preserves the current working CLI while upgrading `ralphx` into a multi-agent, policy-driven, observable runtime with durable state, stronger validation gates, and graphable architecture.

**Architecture:** keep the current Go codebase and `.ralphx` local-state model, but separate the code into four clearer layers: CLI/app entrypoints, orchestration/runtime policy, backend adapters, and state/reporting. The runner remains the authority for completion, retries, validation gates, and future parallel coordination.

**Tech Stack:** Go 1.19+, stdlib-first, Git worktree-based local parallelism later, JSON state files under `.ralphx/`, Mermaid for repo-native diagrams, GitHub for review/distribution.

---

## 1. Current baseline already present

The current repo already has the right bones:

- CLI entrypoints: `cmd/ralphx/main.go`, `cmd/ralphx-doctor/main.go`
- Dispatch: `internal/cli/app.go`
- Runner loop: `internal/runner/loop.go`
- Single backend adapter: `internal/agent/codex.go`
- Prompt assembly: `internal/prompt/builder.go`
- Task loading: `internal/task/load.go`
- Validation: `internal/validate/validate.go`
- VCS snapshotting: `internal/vcs/git.go`
- State persistence: `internal/state/*.go`
- Parallel scaffolding: `internal/parallel/*.go`, `internal/state/parallel.go`

This means the refactor should be incremental, not a rewrite-from-scratch.

## 2. Target outcomes

At the end of the staged refactor, `ralphx` should support:

1. **Pluggable agent backends**
   - Codex
   - Claude Code
   - Hermes / ACP-style backend
   - future local/remote coding agents without runner rewrite

2. **Policy-driven outer loop**
   - explicit stop policy
   - validation policy
   - retry policy
   - progress policy
   - future parallel/merge policy

3. **Durable observable run model**
   - run / round / worker / validation artifacts
   - resume-friendly state
   - clear stop reasons
   - exportable summaries and graphs

4. **Validation as a real gate**
   - multi-step validation pipeline
   - validation feedback injected into future rounds
   - completion never delegated to the backend alone

5. **Safe parallel path**
   - worktree-based worker isolation
   - result bundles
   - merge / reject / fallback decisions owned by leader

---

## 3. Architectural principles

### 3.1 Runner owns truth
Only the runner decides:
- whether progress happened
- whether the task is complete
- whether to retry
- whether to block/stop
- whether parallel results are accepted

### 3.2 Backends are adapters, not brains
Each backend should only:
- accept a normalized request
- execute bounded work
- return structured output + metadata

All orchestration semantics stay outside the adapter.

### 3.3 State is a product surface
Files under `.ralphx/` are not debug junk. They are a durable contract for:
- resuming runs
- inspecting failures
- exporting reports
- building graph views

### 3.4 Validation beats self-reported completion
A backend saying `complete` is advisory. Completion only stands if runner gates pass.

### 3.5 Parallelism must be isolated first
No shared-workdir parallel writes. Use isolated worktrees before enabling real parallel execution.

---

## 4. Recommended target package shape

```text
cmd/
  ralphx/
  ralphx-doctor/

internal/
  agent/
    interface.go
    factory.go
    codex.go
    claudecode.go
    hermes.go

  app/
    run.go
    resume.go
    status.go
    inspect.go

  config/
    config.go
    profile.go

  domain/
    run.go
    round.go
    validation.go
    progress.go
    report.go

  runner/
    loop.go
    policies.go
    stop.go
    progress.go

  prompt/
    builder.go
    sections.go

  validate/
    pipeline.go
    steps.go

  vcs/
    git.go
    worktree.go
    diff.go

  parallel/
    scheduler.go
    planner.go
    merger.go

  state/
    store.go
    paths.go
    writer.go

  report/
    summary.go
    export.go
    graph.go
```

---

## 5. Staged implementation plan

## Phase 0 — stabilize contracts without changing behavior

**Objective:** keep current CLI behavior, but reduce coupling and make future changes cheap.

### Scope
- keep `ralphx run` working as-is
- preserve current Codex path
- preserve current `.ralphx` outputs
- introduce clearer domain types and stop reasons

### Deliverables
- stronger `Agent` interface in `internal/agent/interface.go`
- explicit runtime/domain types for run/round/progress/validation
- normalized `RoundOutcome` derived from backend result + runner checks
- standard stop reason taxonomy

### Files likely to change
- `internal/agent/interface.go`
- `internal/contracts/result.go`
- `internal/runner/loop.go`
- `internal/state/types.go`
- `internal/state/store.go`
- `internal/config/config.go`

### Validation
```bash
go test ./...
go build ./...
```

### Acceptance
- current commands still work
- no behavior regression in current Codex path
- stop reasons are explicit and serialized

---

## Phase 1 — pluggable backend runtime

**Objective:** turn `ralphx` from a Codex-specific runner into a backend-neutral orchestration runtime.

### Scope
- backend registry / factory
- backend capabilities model
- backend-specific adapters hidden behind shared interface
- modular prompt assembly

### Deliverables
- `internal/agent/factory.go`
- `internal/agent/claudecode.go`
- `internal/agent/hermes.go`
- backend capabilities struct
- modular prompt sections

### Files likely to change
- `internal/agent/*`
- `internal/prompt/builder.go`
- `internal/config/config.go`
- `internal/runner/loop.go`

### Validation
```bash
go test ./...
go build ./...
ralphx --help
ralphx doctor
```

### Acceptance
- backend can be selected by config/flag
- runner logic does not branch on backend-specific details
- prompt builder can assemble shared and backend-specific sections cleanly

---

## Phase 2 — validation as a first-class gate

**Objective:** prevent false progress and fake completion.

### Scope
- multi-step validation pipeline
- validation report artifacts per round
- progress scoring derived from real signals
- failure feedback injected into next round

### Deliverables
- `internal/validate/pipeline.go`
- `internal/validate/steps.go`
- `internal/runner/progress.go`
- validation report schema
- feedback injection into prompt assembly

### Files likely to change
- `internal/validate/*`
- `internal/runner/loop.go`
- `internal/prompt/*`
- `internal/state/*`

### Validation
```bash
go test ./...
go build ./...
```

### Acceptance
- multiple validation steps can run sequentially
- each step is recorded with status/output/duration
- next round can consume structured failure summary
- completion requires runner-side validation success

---

## Phase 3 — inspectability, resume, and exports

**Objective:** make runs understandable and operable.

### Scope
- `status`, `resume`, `inspect`, `logs` commands
- round artifacts under `.ralphx/rounds/`
- human-readable summary + machine-readable export
- graph/export layer for repo analysis and docs

### Deliverables
- `internal/app/status.go`
- `internal/app/resume.go`
- `internal/report/*`
- round artifact layout
- graph export commands or report generators

### Files likely to change
- `internal/cli/app.go`
- `internal/state/*`
- `internal/report/*`
- `README.md`
- docs under `docs/en/` and `docs/zh/`

### Validation
```bash
go test ./...
go build ./...
ralphx status --help
ralphx resume --help
```

### Acceptance
- users can inspect current/previous runs without reading raw JSON manually
- artifacts are stable enough to drive diagrams and future UI

---

## Phase 4 — safe local parallelism

**Objective:** turn current parallel scaffolding into a production-grade local parallel mode.

### Scope
- worktree isolation
- worker assignment protocol
- result bundles
- merge / reject / fallback decisions
- conflict handling and downgrade path

### Deliverables
- `internal/vcs/worktree.go`
- `internal/parallel/planner.go`
- `internal/parallel/merger.go`
- worker result bundle schema
- conflict report format

### Files likely to change
- `internal/parallel/*`
- `internal/state/parallel.go`
- `internal/runner/loop.go`
- `internal/vcs/*`

### Validation
```bash
go test ./...
go build ./...
```

### Acceptance
- at least two workers can run in isolated worktrees
- leader remains sole authority for completion
- conflicting worker results can be rejected or downgraded safely

---

## 6. Priority order (ROI first)

1. agent interface + backend factory
2. validation pipeline
3. run/round/result schema cleanup
4. status / resume / inspect / logs
5. prompt feedback loop
6. worktree-based parallelism

If scope must be cut, cut parallelism first, not contracts/validation.

---

## 7. Concrete migration checklist

### Milestone A: contracts and boundaries
- [ ] normalize `AgentRequest` / `AgentResponse`
- [ ] define stop reasons enum/constants
- [ ] split backend output from runner outcome
- [ ] add run/round identifiers where missing

### Milestone B: backend neutrality
- [ ] add backend factory
- [ ] keep Codex as first adapter through new interface
- [ ] add capability flags
- [ ] make prompt builder backend-neutral

### Milestone C: stronger gates
- [ ] support multi-step validation
- [ ] persist validation report per round
- [ ] compute progress from checklist + diff + validation, not backend claim alone
- [ ] inject failures into next round prompt

### Milestone D: inspectability
- [ ] add round artifact folders
- [ ] add `status`/`resume`/`inspect`
- [ ] add exportable summaries
- [ ] generate diagrams from live structure/artifacts

### Milestone E: safe parallel execution
- [ ] create worktree per worker
- [ ] define worker result bundle
- [ ] implement merge/reject/fallback flow
- [ ] add conflict report output

---

## 8. Risks and tradeoffs

### Risk: over-design before backend parity
Mitigation: keep Phase 0 small and behavior-preserving.

### Risk: parallel mode destabilizes the repo
Mitigation: require isolated worktrees before enabling multi-worker writes.

### Risk: state schema drifts too fast
Mitigation: version JSON artifacts and document them.

### Risk: validation becomes too heavy
Mitigation: support required vs optional steps and fail-fast configuration.

### Tradeoff: more artifacts vs more disk usage
Worth it. Inspectability and replayability are core features, not extras.

---

## 9. Suggested repo deliverables for this planning pass

This planning pass should land repo-tracked documentation, not code changes to the runtime itself:

- plan doc: `docs/plans/2026-04-18-ralphx-runtime-refactor-plan.md`
- graph atlas (English): `docs/en/refactor-graph-atlas.md`
- graph atlas (中文): `docs/zh/refactor-graph-atlas.md`
- README quick links updated to point to the new docs

---

## 10. Verification for this docs-only pass

Run after landing the docs:

```bash
git diff --stat
```

Review manually that:
- the plan is concrete enough to implement against
- every graph renders as Mermaid on GitHub
- English and Chinese docs point to the same conceptual artifacts
- README quick links expose the new planning/graph docs

---

## 11. Definition of done for this planning pass

This pass is complete when:
- the repo contains a staged refactor plan checked into `docs/plans/`
- the repo contains graph/chain docs checked into `docs/en/` and `docs/zh/`
- quick links surface the new docs
- all changes are committed and pushed to the user’s GitHub repository
