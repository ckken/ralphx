# ralphx 重构图谱总览

本文收集下一阶段 `ralphx` runtime 重构所需的关键图谱与链路图，方便快速分析、讨论和落地。

## 1. 当前模块依赖图

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

## 2. 当前 runtime 控制链路

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
    I --> J[state.Store 写入 artifacts]
    J --> K{停止? 进入下一轮?}
    K -->|下一轮| E
    K -->|停止| L[final summary / exit]
```

## 3. 当前状态与产物流

```mermaid
flowchart LR
    TaskFile[task 文件] --> Runner[runner loop]
    Checklist[checklist 文件] --> Runner
    PromptAsset[prompt assets] --> Runner
    Runner --> Agent[agent backend]
    Agent --> RoundResult[结构化 round result]
    RoundResult --> Validation[validation chain]
    Validation --> StateDir[.ralphx/ 状态文件]
    StateDir --> Summary[summary / last result / stats]
    StateDir --> Runner
```

## 4. 重构后的目标分层

```mermaid
flowchart TD
    subgraph Entry[入口 / UX]
        CLI[CLI commands]
        App[app actions: run resume status inspect]
    end

    subgraph Runtime[Runtime / 编排层]
        Runner[runner loop]
        Policy[stop retry validation progress policies]
        Planner[parallel planner / merger]
    end

    subgraph Adapters[Backend / Tool 适配层]
        AgentFactory[agent factory]
        Codex[Codex adapter]
        Claude[Claude Code adapter]
        Hermes[Hermes adapter]
        Git[VCS adapter]
        Validate[validation executors]
    end

    subgraph State[状态 / 报告层]
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

## 5. 建议的 run/round 状态模型

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

## 6. 计划中的 validation pipeline

```mermaid
flowchart LR
    Start[round result] --> Format[format/lint]
    Format --> Unit[unit tests]
    Unit --> Smoke[smoke/integration]
    Smoke --> Policy[runner validation policy]
    Policy --> Pass[允许 complete 或记为 progress]
    Policy --> Fail[将失败摘要回灌到下一轮]
```

## 7. 计划中的 backend 调用契约

```mermaid
flowchart LR
    Runner[runner] --> Request[AgentRequest]
    Request --> Backend[backend adapter]
    Backend --> Response[AgentResponse]
    Response --> Normalize[runner normalization]
    Normalize --> Outcome[RoundOutcome]
    Outcome --> State[state store]
```

## 8. 安全并行执行目标图

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

## 9. 文档与代码映射建议

| 关注点 | 当前文件 | 后续更合适的位置 |
| --- | --- | --- |
| CLI 分发 | `internal/cli/app.go` | `internal/app/*` + 更薄的 `internal/cli/app.go` |
| Agent backend | `internal/agent/codex.go` | `internal/agent/{factory,codex,claudecode,hermes}.go` |
| Runner loop | `internal/runner/loop.go` | `internal/runner/{loop,policies,stop,progress}.go` |
| Validation | `internal/validate/validate.go` | `internal/validate/{pipeline,steps}.go` |
| State | `internal/state/*` | `internal/state/*` + `internal/report/*` |
| Parallel | `internal/parallel/*` | `internal/parallel/{scheduler,planner,merger}.go` |
| 图谱/导出 | 当前仅 docs | 后续可加 `internal/report/graph.go` |

## 10. 使用方式

- 修改 package 边界前，先看第 1 节。
- 要验证当前行为是否被保留，先看第 2–3 节。
- 做 staged runtime refactor 时，重点参考第 4–8 节。
- 之后只要 package 边界或 artifact 布局有变化，都应同步更新这份图谱。
