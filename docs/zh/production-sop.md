# 生产 SOP

这是 `ralphx` 在真实仓库中的推荐操作路径。

## 1. 安装

使用 release 安装，不走源码安装。

最新版本：

```bash
curl -fsSL https://github.com/ckken/ralphx/releases/latest/download/install.sh | bash
```

指定版本：

```bash
curl -fsSL https://github.com/ckken/ralphx/releases/download/v0.1.0/install.sh | VERSION=v0.1.0 bash
```

安装器会把当前激活的执行路径持久化到：

```bash
~/.config/ralphx/current.env
```

这样 wrapper 命令保持稳定，而底层实际执行的版本化二进制放在 `~/.local/share/ralphx/releases/`。

## 2. 预检

先执行：

```bash
ralphx-doctor
```

最低要求：
- `codex` 可用
- 如果要用 git 感知 gate，建议安装 `git`

## 3. 准备输入

准备两类文件：
- task file：描述总目标
- checklist：描述可验证、可拆分的剩余工作

规则：
- task = 总目标
- checklist = 硬剩余工作
- checklist item 要短、边界清晰、便于验证

## 4. 选择执行模式

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

## 5. 配置验证

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

## 6. 运行

典型生产调用：

```bash
export CODEX_CMD=codex
export CODEX_ARGS='-m gpt-5.4'
export TESTS_CMD='go test ./...'

ralphx   --task docs/tasks/release-task.md   --checklist docs/tasks/release-task.checklist.md   --workdir /path/to/repo   --workers 3
```

## 7. 查看输出

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

## 8. GitHub Release 推进 SOP

打 tag 前检查：
- `go build ./...`
- `go test ./...`
- `ralphx-doctor`
- 跑一次单 worker smoke
- 如果要用并行，再跑一次 `--workers` smoke
- 检查 `.ralphx/last-result.json`

打 tag 并推送：

```bash
git status
git add .
git commit -m "feat: release prep"
git push origin main
git tag v0.1.0
git push origin v0.1.0
```
