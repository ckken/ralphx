# 生产 SOP

这是 `ralphx` 在真实仓库中的推荐操作路径。

## 1. 预检

先执行：

```bash
ralphx-doctor
```

最低要求：
- `codex` 可用
- 如果从源码安装，需要 `go`
- 如果要用 git 感知 gate，建议安装 `git`

## 2. 准备输入

准备两类文件：
- task file：描述总目标
- checklist：描述可验证、可拆分的剩余工作

规则：
- task = 总目标
- checklist = 硬剩余工作
- checklist item 要短、边界清晰、便于验证

好的 checklist 项：
- 增加 command X
- 更新 config loader
- 加 smoke test
- 更新 README 某节

不好的 checklist 项：
- 把所有事情做完
- 修完所有问题
- 变成生产可用

## 3. 选择执行模式

### 单 worker
适合：
- 工作强耦合
- 文件冲突可能大
- 先走最稳路径

```bash
ralphx --task task.md --checklist task.checklist.md --workdir /repo
```

### 并行模式
适合：
- checklist 可拆分
- 每项是有边界的 slice
- 需要加速本地执行

```bash
ralphx --task task.md --checklist task.checklist.md --workdir /repo --workers 3
```

推荐规则：
- 默认从 `--workers 1` 开始
- 只有 checklist 确实可拆时才增加 worker 数

## 4. 配置验证

尽量总是配置 `TESTS_CMD`。

例如：

```bash
export TESTS_CMD='go test ./...'
```

```bash
export TESTS_CMD='bun test && bun run lint'
```

```bash
export TESTS_CMD='pytest -q'
```

## 5. 运行

典型生产调用：

```bash
export CODEX_CMD=codex
export CODEX_ARGS='-m gpt-5.4'
export TESTS_CMD='go test ./...'

ralphx   --task docs/tasks/release-task.md   --checklist docs/tasks/release-task.checklist.md   --workdir /path/to/repo   --workers 3
```

## 6. 查看输出

关键文件：
- `.ralphx/last-result.json`
- `.ralphx/state.json`
- `.ralphx/stats.json`
- `.ralphx/logs/`
- `.ralphx/results/`（并行模式）
- `.ralphx/runtime/loop-output.schema.json`

判断方式：
- `complete`：所有 gate 通过，leader 接受完成
- `in_progress`：工作仍剩余，或 complete 被降级
- `blocked`：真实阻塞、无效输出、或验证失败

## 7. 推进到 GitHub 前检查

- `go build ./...`
- `go test ./...`
- `ralphx-doctor`
- 跑一次单 worker smoke
- 如果要用并行，再跑一次 `--workers` smoke
- 检查 `.ralphx/last-result.json`
- 如果 run 最终是 `complete`，确认 checklist 已全部打勾

## 8. GitHub 推进 SOP

推荐顺序：

```bash
git status
go build ./...
go test ./...
git add .
git commit -m "feat: productionize Go ralphx workflow"
git push origin main
```

## 9. 运行边界

要做：
- 用边界清晰的 checklist
- 验证命令尽量便宜但有意义
- blocked 时看结果文件和 logs
- 先用较小 worker 数

不要做：
- 没 checklist 就上高并行
- 把 worker 完成等同于总任务完成
- 生产变更时完全跳过验证
