# ralphx Refactor Graph Atlas

This document collects the key graphs needed to analyze and execute the next `ralphx` runtime refactor.

## 1. Current module dependency graph

```mermaid
flowchart LR
    cmd_ralphx[cmd/ralphx] --> cli[internal/cli]
    cmd_doctor[cmd/ralphx-doctor] --> doctor[internal/doctor]

    cli --> config[internal/config]
    cli --> current[internal/current]
    cli --> doctor
    cli --> runner[internal/runner]
    cli --> version[internal/version]

    runner --> agent[internal/agent]
    runner --> assets[internal/assets]
    runner --> config
    runner --> contracts[internal/contracts]
    runner --> parallel[internal/parallel]
    runner --> prompt[internal/prompt]
    runner --> state[internal/state]
    runner --> task[internal/task]
    runner --> validate[internal/validate]
    runner --> vcs[internal/vcs]

    prompt --> assets
    prompt --> task
    agent --> contracts
    agent --> execx[internal/execx]
    validate --> execx
    state --> contracts
```

## 2. Current runtime control path

```mermaid
flowchart TD
    A[CLI: ralphx run] --> B[config.ParseRunArgs]
    B --> C[runner.New]
    C --> D[task.Load]
    D --> E[prompt.Build]
    E --> F[agent.Codex.Invoke]
    F --> G[contracts.RoundResult.Validate]
    G --> H[vcs snapshot / diff]
    H --> I[validate.Run]
    I --> J[state.Store writes artifacts]
    J --> K{stop? next round?}
    K -->|next round| E
    K -->|stop| L[final summary / exit]
```

## 3. Current state and artifact flow

```mermaid
flowchart LR
    TaskFile[task file] --> Runner[runner loop]
    Checklist[checklist file] --> Runner
    PromptAsset[prompt assets] --> Runner
    Runner --> Agent[agent backend]
    Agent --> RoundResult[structured round result]
    RoundResult --> Validation[validation chain]
    Validation --> StateDir[.ralphx/ state files]
    StateDir --> Summary[summary / last result / stats]
    StateDir --> Runner
```

## 4. Refactor target layers

```mermaid
flowchart TD
    subgraph Entry[Entry / UX]
        CLI[CLI commands]
        App[app actions: run resume status inspect]
    end

    subgraph Runtime[Runtime / Orchestration]
        Runner[runner loop]
        Policy[stop retry validation progress policies]
        Planner[parallel planner / merger]
    end

    subgraph Adapters[Backend / Tool Adapters]
        AgentFactory[agent factory]
        Codex[Codex adapter]
        Claude[Claude Code adapter]
        Hermes[Hermes adapter]
        Git[VCS adapter]
        Validate[validation executors]
    end

    subgraph State[State / Reports]
        Store[state store]
        Reports[summary export graph]
        Artifacts[run round worker artifacts]
    end

    CLI --> App
    App --> Runner
    Runner --> Policy
    Runner --> Planner
    Runner --> AgentFactory
    Runner --> Git
    Runner --> Validate
    Runner --> Store
    Store --> Reports
    Store --> Artifacts
    AgentFactory --> Codex
    AgentFactory --> Claude
    AgentFactory --> Hermes
```

## 5. Proposed run/round state model

```mermaid
flowchart TD
    Run[Run]
    Run --> Round1[Round N]
    Round1 --> Prompt[prompt.txt]
    Round1 --> BackendOut[agent-output.txt]
    Round1 --> Parsed[result.json]
    Round1 --> Validation[validation.json]
    Round1 --> Diff[git.diff]
    Round1 --> Outcome[outcome.json]
    Run --> Summary[summary.md]
    Run --> Stats[stats.json]
    Run --> Events[events.jsonl]
```

## 6. Planned validation pipeline

```mermaid
flowchart LR
    Start[round result] --> Format[format/lint]
    Format --> Unit[unit tests]
    Unit --> Smoke[smoke/integration]
    Smoke --> Policy[runner validation policy]
    Policy --> Pass[allow completion or progress]
    Policy --> Fail[feed failures into next round]
```

## 7. Planned backend invocation contract

```mermaid
flowchart LR
    Runner[runner] --> Request[AgentRequest]
    Request --> Backend[backend adapter]
    Backend --> Response[AgentResponse]
    Response --> Normalize[runner normalization]
    Normalize --> Outcome[RoundOutcome]
    Outcome --> State[state store]
```

## 8. Safe parallel execution target

```mermaid
flowchart TD
    Leader[leader runner] --> Planner[task splitter]
    Planner --> WorkerA[worker A worktree]
    Planner --> WorkerB[worker B worktree]
    WorkerA --> ResultA[result bundle A]
    WorkerB --> ResultB[result bundle B]
    ResultA --> Merger[leader merger]
    ResultB --> Merger
    Merger --> Accept[accept / cherry-pick]
    Merger --> Reject[reject / fallback serial]
```

## 9. Recommended doc-to-code mapping

| Concern | Current files | Likely future home |
| --- | --- | --- |
| CLI dispatch | `internal/cli/app.go` | `internal/app/*` + thinner `internal/cli/app.go` |
| Agent backend | `internal/agent/codex.go` | `internal/agent/{factory,codex,claudecode,hermes}.go` |
| Runner loop | `internal/runner/loop.go` | `internal/runner/{loop,policies,stop,progress}.go` |
| Validation | `internal/validate/validate.go` | `internal/validate/{pipeline,steps}.go` |
| State | `internal/state/*` | `internal/state/*` + `internal/report/*` |
| Parallel | `internal/parallel/*` | `internal/parallel/{scheduler,planner,merger}.go` |
| Graph/export | docs only today | `internal/report/graph.go` later |

## 10. How to use this atlas

- Use section 1 before changing package boundaries.
- Use sections 2–3 when validating current behavior preservation.
- Use sections 4–8 when implementing the staged runtime refactor.
- Keep this atlas updated whenever package boundaries or artifact layout change.
