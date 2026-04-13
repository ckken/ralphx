# ralphx

[English](../../README.md) | [中文](README.md)

快速入口：
- [安装说明](installation.md)
- [方法论](methodology.md)
- [生产 SOP](production-sop.md)
- [并行协议 v0](../en/go-parallel-protocol-v0.md)
- [流程图 / 链路图](architecture.md)

`ralphx` 是一个基于 Go 的 Codex / coding agent 外层执行器。

核心目标很简单：
- 让 agent 用当前工具持续推进，直到真实任务完成
- 用 checklist / validation / leader gate 阻止过早完成
- 在任务可拆分时支持本地多 worker 并行执行

## 它现在能做什么

- 读取任务文件
- 可选读取 checklist
- 调用 `codex exec` 并要求严格 JSON 输出
- 把运行状态落到 `.ralphx/`
- 拒绝弱完成信号
- 执行验证命令
- 用 `--workers N` 把 checklist 项拆成并行 worker 任务

## 安装

```bash
git clone https://github.com/ckken/ralphx.git
cd ralphx
./install.sh
ralphx-doctor
```

如果需要：

```bash
export PATH="$HOME/.local/bin:$PATH"
```

## 依赖

必需：
- `go`
- `codex`

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
