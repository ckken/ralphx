# ralphx

[中文文档入口](docs/zh/README.md)

快速入口：
- [安装说明](docs/zh/installation.md)
- [方法论](docs/zh/methodology.md)
- [生产 SOP](docs/zh/production-sop.md)
- [并行协议 v0](docs/zh/go-parallel-protocol-v0.md)
- [流程图 / 链路图](docs/zh/architecture.md)
- [Runtime 重构计划](docs/plans/2026-04-18-ralphx-runtime-refactor-plan.md)
- [Go 重写 MVP 计划](docs/plans/2026-04-13-go-rewrite-mvp.md)
- [重构图谱总览](docs/zh/refactor-graph-atlas.md)

`ralphx` 是一个基于 Go 的 Codex / coding agent 外层执行器。

它的目标很直接：
- 让 agent 在现有工具链里持续工作，直到真实任务完成
- 用 checklist / validation / leader 侧规则约束完成条件
- 当任务适合拆分时，支持本地多 worker 执行

## 通过 GitHub Release 安装

正常使用不需要源码安装。

安装最新 release：

```bash
curl -fsSL https://github.com/ckken/ralphx/releases/latest/download/install.sh | bash
ralphx-doctor
```

安装指定版本：

```bash
curl -fsSL https://github.com/ckken/ralphx/releases/download/v0.1.2/install.sh | VERSION=v0.1.2 bash
ralphx-doctor
```

安装器会下载 `SHA256SUMS` 并在激活前校验 release 二进制。

## 执行路径持久化

当前激活二进制路径持久化在：

```bash
~/.config/ralphx/current.env
```

下载的 release 二进制存放在：

```bash
~/.local/share/ralphx/releases/
```

`ralphx` 和 `ralphx-doctor` 是位于 `~/.local/bin` 的稳定包装命令，始终读取当前持久化执行路径。

查看当前持久化执行状态：

```bash
ralphx current
```
