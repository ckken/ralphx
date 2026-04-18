# ralphx Runtime 重构计划

> 给 Hermes：这是仓库内跟踪的规划文档，用于将 `ralphx` 从一个以 Codex 为中心的外层循环执行器，升级为一个可插拔、可检查、由验证门控的编排运行时。

**目标：** 通过分阶段重构，在保留当前可用 CLI 的前提下，将 `ralphx` 升级为一个多代理、策略驱动、可观测的运行时，具备持久化状态、更强的验证门控，以及可图谱化的架构。

**架构：** 保留当前 Go 代码库和 `.ralphx` 本地状态模型，但将代码拆分为四个更清晰的层次：CLI / 应用入口、编排 / 运行时策略、后端适配器，以及状态 / 报告。Runner 仍然是完成判定、重试、验证门控以及未来并行协调的权威主体。

**技术栈：** Go 1.19+，优先标准库，后续采用基于 Git worktree 的本地并行，JSON 状态文件存放在 `.ralphx/` 下，使用 Mermaid 作为仓库原生图表方案，GitHub 用于审查与分发。

---

## 1. 当前已具备的基础

当前仓库已经具备了正确的骨架：

- CLI 入口：`cmd/ralphx/main.go`、`cmd/ralphx-doctor/main.go`
- 分发：`internal/cli/app.go`
- Runner 循环：`internal/runner/loop.go`
- 单一后端适配器：`internal/agent/codex.go`
- Prompt 组装：`internal/prompt/builder.go`
- 任务加载：`internal/task/load.go`
- 验证：`internal/validate/validate.go`
- VCS 快照：`internal/vcs/git.go`
- 状态持久化：`internal/state/*.go`
- 并行脚手架：`internal/parallel/*.go`、`internal/state/parallel.go`

这意味着此次重构应该是渐进式的，而不是一次推倒重写。

## 2. 目标结果

在分阶段重构结束时，`ralphx` 应当支持：

1. **可插拔的代理后端**
   - Codex
   - Claude Code
   - Hermes / ACP 风格后端
   - 未来的本地 / 远程编码代理，无需重写 runner

2. **策略驱动的外层循环**
   - 显式停止策略
   - 验证策略
   - 重试策略
   - 进度策略
   - 未来的并行 / 合并策略

3. **持久化且可观测的运行模型**
   - run / round / worker / validation 工件
   - 便于恢复的状态
   - 清晰的停止原因
   - 可导出的摘要与图谱

4. **将验证作为真正的门控**
   - 多步骤验证流水线
   - 将验证反馈注入后续轮次
   - 完成判定绝不只委托给后端

5. **安全的并行路径**
   - 基于 worktree 的 worker 隔离
   - 结果打包
   - 合并 / 拒绝 / 回退决策由 leader 掌控

---

## 3. 架构原则

### 3.1 Runner 掌握真相
只有 runner 决定：
- 是否发生了进展
- 任务是否完成
- 是否需要重试
- 是否需要阻塞 / 停止
- 是否接受并行结果

### 3.2 后端是适配器，不是大脑
每个后端只应当：
- 接收标准化请求
- 执行有边界的工作
- 返回结构化输出和元数据

所有编排语义都应保留在适配器之外。

### 3.3 状态本身就是产品界面
`.ralphx/` 下的文件不是调试垃圾。它们是以下能力的持久化契约：
- 恢复运行
- 检查失败原因
- 导出报告
- 构建图谱视图

### 3.4 验证优先于自报完成
后端声称 `complete` 只能作为参考。只有当 runner 的门控检查通过时，完成判定才成立。

### 3.5 并行必须先实现隔离
不要在共享 workdir 上进行并行写入。必须先使用隔离的 worktree，再启用真正的并行执行。

---

## 4. 推荐的目标包结构

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

## 5. 分阶段实施计划

## Phase 0 — 在不改变行为的前提下稳定契约

**目标：** 保持当前 CLI 行为不变，但降低耦合，使未来改动成本更低。

### 范围
- 保持 `ralphx run` 按现状工作
- 保留当前 Codex 路径
- 保留当前 `.ralphx` 输出
- 引入更清晰的领域类型和停止原因

### 交付物
- 在 `internal/agent/interface.go` 中强化 `Agent` 接口
- 为 run / round / progress / validation 定义显式运行时 / 领域类型
- 基于后端结果 + runner 检查得出标准化的 `RoundOutcome`
- 标准的停止原因分类体系

### 可能变更的文件
- `internal/agent/interface.go`
- `internal/contracts/result.go`
- `internal/runner/loop.go`
- `internal/state/types.go`
- `internal/state/store.go`
- `internal/config/config.go`

### 验证
```bash
go test ./...
go build ./...
```

### 验收标准
- 当前命令仍然可用
- 当前 Codex 路径无行为回归
- 停止原因是显式的，并且可序列化

---

## Phase 1 — 可插拔后端运行时

**目标：** 将 `ralphx` 从一个 Codex 专用 runner，转变为后端无关的编排运行时。

### 范围
- 后端注册表 / 工厂
- 后端能力模型
- 将后端专有适配器隐藏在共享接口之后
- 模块化 Prompt 组装

### 交付物
- `internal/agent/factory.go`
- `internal/agent/claudecode.go`
- `internal/agent/hermes.go`
- 后端能力结构体
- 模块化 Prompt section

### 可能变更的文件
- `internal/agent/*`
- `internal/prompt/builder.go`
- `internal/config/config.go`
- `internal/runner/loop.go`

### 验证
```bash
go test ./...
go build ./...
ralphx --help
ralphx doctor
```

### 验收标准
- 可通过配置 / 参数选择后端
- runner 逻辑不再基于后端细节分支
- prompt builder 能够清晰地组装共享 section 与后端专属 section

---

## Phase 2 — 将验证提升为一等门控

**目标：** 防止虚假进展与伪完成。

### 范围
- 多步骤验证流水线
- 每轮的验证报告工件
- 从真实信号推导进度评分
- 将失败反馈注入下一轮

### 交付物
- `internal/validate/pipeline.go`
- `internal/validate/steps.go`
- `internal/runner/progress.go`
- 验证报告 schema
- 将反馈注入 prompt 组装流程

### 可能变更的文件
- `internal/validate/*`
- `internal/runner/loop.go`
- `internal/prompt/*`
- `internal/state/*`

### 验证
```bash
go test ./...
go build ./...
```

### 验收标准
- 可顺序执行多个验证步骤
- 每个步骤都会记录状态 / 输出 / 时长
- 下一轮可以消费结构化失败摘要
- 完成判定必须要求 runner 侧验证成功

---

## Phase 3 — 可检查性、恢复与导出

**目标：** 使运行过程可理解、可操作。

### 范围
- `status`、`resume`、`inspect`、`logs` 命令
- `.ralphx/rounds/` 下的轮次工件
- 人类可读摘要 + 机器可读导出
- 面向仓库分析和文档的图谱 / 导出层

### 交付物
- `internal/app/status.go`
- `internal/app/resume.go`
- `internal/report/*`
- 轮次工件布局
- 图谱导出命令或报告生成器

### 可能变更的文件
- `internal/cli/app.go`
- `internal/state/*`
- `internal/report/*`
- `README.md`
- `docs/zh/` 下的文档

### 验证
```bash
go test ./...
go build ./...
ralphx status --help
ralphx resume --help
```

### 验收标准
- 用户无需手动读取原始 JSON 也能检查当前 / 过往运行
- 工件足够稳定，可驱动图表与未来 UI

---

## Phase 4 — 安全的本地并行

**目标：** 将当前的并行脚手架升级为生产级本地并行模式。

### 范围
- worktree 隔离
- worker 分配协议
- 结果打包
- 合并 / 拒绝 / 回退决策
- 冲突处理与降级路径

### 交付物
- `internal/vcs/worktree.go`
- `internal/parallel/planner.go`
- `internal/parallel/merger.go`
- worker 结果 bundle schema
- 冲突报告格式

### 可能变更的文件
- `internal/parallel/*`
- `internal/state/parallel.go`
- `internal/runner/loop.go`
- `internal/vcs/*`

### 验证
```bash
go test ./...
go build ./...
```

### 验收标准
- 至少两个 worker 能在隔离的 worktree 中运行
- leader 仍是完成判定的唯一权威
- 冲突的 worker 结果可以被安全地拒绝或降级

---

## 6. 优先级顺序（先看 ROI）

1. agent 接口 + 后端工厂
2. 验证流水线
3. run / round / result schema 清理
4. status / resume / inspect / logs
5. prompt 反馈回路
6. 基于 worktree 的并行

如果必须砍范围，优先砍并行，不要砍契约 / 验证。

---

## 7. 具体迁移清单

### Milestone A：契约与边界
- [ ] 标准化 `AgentRequest` / `AgentResponse`
- [ ] 定义停止原因 enum / 常量
- [ ] 将后端输出与 runner 结果拆分
- [ ] 在缺失处补充 run / round 标识符

### Milestone B：后端中立
- [ ] 添加后端工厂
- [ ] 通过新接口保留 Codex 作为首个适配器
- [ ] 添加能力标记
- [ ] 使 prompt builder 与后端无关

### Milestone C：更强的门控
- [ ] 支持多步骤验证
- [ ] 为每轮持久化验证报告
- [ ] 根据 checklist + diff + validation 计算进度，而不是只依赖后端声明
- [ ] 将失败信息注入下一轮 prompt

### Milestone D：可检查性
- [ ] 添加轮次工件目录
- [ ] 添加 `status` / `resume` / `inspect`
- [ ] 添加可导出的摘要
- [ ] 从实时结构 / 工件生成图表

### Milestone E：安全并行执行
- [ ] 为每个 worker 创建 worktree
- [ ] 定义 worker 结果 bundle
- [ ] 实现合并 / 拒绝 / 回退流程
- [ ] 添加冲突报告输出

---

## 8. 风险与权衡

### 风险：在实现后端等价能力前过度设计
缓解：保持 Phase 0 小而精，并确保行为不变。

### 风险：并行模式破坏仓库稳定性
缓解：在启用多 worker 写入前，强制要求隔离 worktree。

### 风险：状态 schema 漂移过快
缓解：为 JSON 工件加版本，并做好文档说明。

### 风险：验证变得过于沉重
缓解：支持必需步骤与可选步骤，并提供 fail-fast 配置。

### 权衡：更多工件 vs 更多磁盘占用
值得。可检查性与可重放性是核心特性，不是附加功能。

---

## 9. 本轮规划建议提交到仓库的交付物

本轮规划应提交仓库跟踪文档，而不是直接修改运行时代码：

- 计划文档：`docs/plans/2026-04-18-ralphx-runtime-refactor-plan.md`
- 图谱总览：`docs/zh/refactor-graph-atlas.md`
- 中文主入口：`docs/zh/README.md`
- 更新 README 快速链接，指向新的文档

---

## 10. 本轮仅文档变更的验证

文档落库后运行：

```bash
git diff --stat
```

手动检查以下事项：
- 该计划足够具体，可以据此实施
- 每个图都能在 GitHub 上作为 Mermaid 正常渲染
- 仓库文档统一以中文主入口承载同一套概念与图谱
- README 快速链接已暴露新的规划 / 图谱文档

---

## 11. 本轮规划完成定义

当满足以下条件时，本轮工作即视为完成：
- 仓库中包含提交到 `docs/plans/` 的分阶段重构计划
- 仓库中包含提交到 `docs/zh/` 的图谱 / 链路文档
- 快速链接已展示这些新文档
- 所有变更都已提交并推送到用户的 GitHub 仓库
