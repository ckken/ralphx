# Go 并行协议 v0

这是一个面向本地优先 `ralphx` Go 重构的最小 leader / worker 协议。

## 目标

- 在 `.ralphx/` 下维护可持久化的本地状态
- 由单个 leader 进程负责调度
- 由多个短生命周期 worker 子进程执行有边界的工作
- 后续可让 worker 包装 `codex exec`，而不改变状态模型
- leader 始终是完成判定、重试与 checklist gate 的唯一权威

## 设计约束

1. 每次 run 只允许一个 leader 锁。
2. worker 不能直接修改全局队列状态。
3. 所有状态文件都使用 JSON，并通过临时文件 + rename 原子写入。
4. `events.jsonl` 可以是可选的可观测层，但当前就很有价值。
5. 每个 worker 同一时刻只处理一个 assignment。
6. worker 结果只是建议；只有 leader 可以把 run 标记为 complete。

## 状态目录布局

```text
.ralphx/
  run.json
  leader.lock
  events.jsonl
  summary.txt
  tasks/
    task-0001.json
    task-0002.json
  workers/
    worker-01.json
    worker-02.json
  logs/
    worker-01.stdout.log
    worker-01.stderr.log
    worker-02.stdout.log
    worker-02.stderr.log
  results/
    task-0001.result.json
    task-0002.result.json
```

## 归属模型

- `run.json`：leader 写入
- `tasks/*.json`：leader 写入
- `workers/*.json`：各 worker 写自己的文件；leader 可初始化，也可在进程退出后标记终态
- `results/*.result.json`：worker 单次写入；leader 读取后回填进 task 状态
- `events.jsonl`：leader 追加；worker 可通过 stdout 输出事件，再由 leader 追加落盘

这样可以把并发模型保持得足够简单：leader 拥有队列状态，worker 只拥有心跳与结果文件。

## 文件 schema

### 1. `run.json`

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://ralphx.dev/schemas/run-state-v0.json",
  "type": "object",
  "additionalProperties": false,
  "required": [
    "version",
    "run_id",
    "status",
    "workdir",
    "task_file",
    "checklist_file",
    "max_parallel",
    "created_at",
    "updated_at",
    "leader_pid",
    "task_counts"
  ],
  "properties": {
    "version": { "const": "v0" },
    "run_id": { "type": "string", "minLength": 1 },
    "status": {
      "type": "string",
      "enum": ["starting", "running", "paused", "blocked", "complete", "failed", "canceled"]
    },
    "workdir": { "type": "string", "minLength": 1 },
    "task_file": { "type": "string", "minLength": 1 },
    "checklist_file": { "type": ["string", "null"] },
    "max_parallel": { "type": "integer", "minimum": 1 },
    "leader_pid": { "type": "integer", "minimum": 1 },
    "created_at": { "type": "string", "format": "date-time" },
    "updated_at": { "type": "string", "format": "date-time" },
    "current_summary": { "type": "string" },
    "checklist_open_items": { "type": "integer", "minimum": 0 },
    "last_task_seq": { "type": "integer", "minimum": 0 },
    "task_counts": {
      "type": "object",
      "additionalProperties": false,
      "required": ["pending", "running", "succeeded", "blocked", "failed"],
      "properties": {
        "pending": { "type": "integer", "minimum": 0 },
        "running": { "type": "integer", "minimum": 0 },
        "succeeded": { "type": "integer", "minimum": 0 },
        "blocked": { "type": "integer", "minimum": 0 },
        "failed": { "type": "integer", "minimum": 0 }
      }
    }
  }
}
```

### 2. `tasks/task-XXXX.json`

后续字段约束与 `docs/en/go-parallel-protocol-v0.md` 原始版本一致；本中文版作为当前仓库内的主文档，后续新增字段应优先在此处维护。

## 运行原则

- leader 负责切分 checklist、分配任务、汇总结果、执行最终验证。
- worker 只拿到受边界约束的一段工作，不拥有总任务完成权。
- 所有结果都必须可落盘、可重放、可审计。
- 并行只是提速手段，不能削弱 completion gate 与 validation gate。

## 为什么先定义协议

在真正启用多 worker 之前，先把状态布局、文件归属与完成权边界定义清楚，能显著降低后续并行实现时的返工与数据不兼容风险。
