# ralphx

[English](../../README.md) | [中文](README.md)

快速入口：
- [安装说明](installation.md)
- [方法论](methodology.md)
- [生产 SOP](production-sop.md)
- [流程图 / 链路图](architecture.md)
- [Runtime 重构计划](../plans/2026-04-18-ralphx-runtime-refactor-plan.md)
- [重构图谱总览](refactor-graph-atlas.md)

`ralphx` 是一个基于 Go 的 Codex / coding agent 外层执行器。

## Release 安装（推荐）

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

安装器会下载 `SHA256SUMS` 并校验二进制后再激活。

## 执行路径持久化

当前激活执行路径持久化到：

```bash
~/.config/ralphx/current.env
```

查看当前持久化执行状态：

```bash
ralphx current
```
