# 方法论

`ralphx` 是一个基于 Go 的 Codex / coding agent 外层执行器。

核心方法：

1. task file 是总目标真源。
2. checklist 未完成项视为硬剩余工作。
3. agent 必须返回严格 JSON 结果。
4. 运行状态落到 `.ralphx/`。
5. 由 leader 侧拒绝过早完成。
6. 配置了 `TESTS_CMD` 时，每轮进展后执行验证。
7. 并行模式下，worker 只负责局部 slice，总完成只能由 leader 判定。

## 核心原则

### 1. 外层控制权

不是 agent 自己决定什么时候完成，`ralphx` 控制：
- 迭代次数
- 超时
- checklist gate
- no-progress stop
- validation
- 最终完成接受

### 2. Checklist gate

只要 checklist 还有未完成项：
- 单个 slice 做完不代表总任务完成
- `complete` 会被拒绝
- 并行 worker 做完局部也不能直接结束总任务

### 3. 验证优先

建议总是配置 `TESTS_CMD`。

例如：
- `go test ./...`
- `bun test && bun run lint`
- `pytest -q`

### 4. 并行模式纪律

并行模式适合“边界清晰的 checklist 拆分”，不适合无边界乱 swarm。
