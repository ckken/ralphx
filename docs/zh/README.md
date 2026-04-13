# ralphx

[English](../../README.md) | [中文](README.md)

快速入口：
- [安装说明](installation.md)
- [方法论](methodology.md)
- [生产 SOP](production-sop.md)
- [并行协议 v0](../en/go-parallel-protocol-v0.md)
- [流程图 / 链路图](architecture.md)

`ralphx` 是一个基于 Go 的 Codex / coding agent 外层执行器。

核心目标：
- 让 agent 用当前工具持续推进，直到真实任务完成
- 用 checklist / validation / leader gate 阻止过早完成
- 在任务可拆分时支持本地多 worker 并行执行

## Release 安装（推荐）

正常使用不需要源码安装。

安装最新 release：

```bash
curl -fsSL https://github.com/ckken/ralphx/releases/latest/download/install.sh | bash
ralphx-doctor
```

安装指定版本：

```bash
curl -fsSL https://github.com/ckken/ralphx/releases/download/v0.1.0/install.sh | VERSION=v0.1.0 bash
ralphx-doctor
```

如果需要：

```bash
export PATH="$HOME/.local/bin:$PATH"
```

## 执行路径持久化

安装器会把当前激活的执行路径持久化到：

```bash
~/.config/ralphx/current.env
```

下载下来的版本二进制会放在：

```bash
~/.local/share/ralphx/releases/
```

所以 `ralphx` / `ralphx-doctor` 始终是稳定入口，实际执行版本由持久化路径决定。

## 依赖

必需：
- `codex`
- `curl` 或 `wget`

建议：
- `git`
- `gh`
- `bash`
- `python3`

可选：
- `jq`（仅 legacy 辅助；Go 主链路不依赖）

## 快速开始

单 worker：

```bash
ralphx --task ./examples/sample-task.md --workdir /path/to/repo
```

显式指定 checklist：

```bash
ralphx   --task ./examples/sample-task.md   --checklist ./examples/sample-task.checklist.md   --workdir /path/to/repo
```

并行 checklist：

```bash
ralphx   --task ./examples/sample-task.md   --checklist ./examples/sample-task.checklist.md   --workdir /path/to/repo   --workers 3
```

## 常用环境变量

```bash
export CODEX_CMD=codex
export CODEX_ARGS='-m gpt-5.4'
export TESTS_CMD='go test ./...'
export MAX_ITERATIONS=0
export MAX_NO_PROGRESS=0
export ROUND_TIMEOUT_SECONDS=1800
export RALPHX_WORKERS=3
```

## 生产建议

1. 先跑 `ralphx-doctor`
2. 准备 task + checklist
3. 默认先用 `--workers 1`
4. checklist 明确可拆时再开并行
5. 配置 `TESTS_CMD`
6. 运行后检查 `.ralphx/last-result.json`、`.ralphx/state.json`、`.ralphx/results/`

完整操作请看：[生产 SOP](production-sop.md)
