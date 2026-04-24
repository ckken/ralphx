# ralphx

[English](../../README.md) | [中文](README.md)

快速链接：
- [安装说明](installation.md)
- [方法论](methodology.md)
- [生产 SOP](production-sop.md)
- [并行协议 v0](../en/go-parallel-protocol-v0.md)
- [架构与链路图](architecture.md)
- [Codex 自循环机制](../en/codex-self-loop.md)

`ralphx` 是一个基于 Go 的 Codex / coding agent 外层执行器。

## GPT-5.5 下的定位

GPT-5.5 提升的是推理和编码质量，但不会自动替代执行纪律。
`ralphx` 负责模型外侧的运行时约束：本地状态、checklist gate、验证 gate、Stop hook 续跑、session resume 和停滞后的 replan。

长任务、发布任务、多轮修复、必须“做完才能停”的任务适合用 `ralphx`。
普通一次性改动直接用 Codex 通常足够。

## Release 安装（推荐）

正常使用不需要源码安装。

安装最新 release：

```bash
curl -fsSL https://github.com/ckken/ralphx/releases/latest/download/install.sh | bash
ralphx doctor
```

安装指定版本：

```bash
curl -fsSL https://github.com/ckken/ralphx/releases/download/v0.2.3/install.sh | VERSION=v0.2.3 bash
ralphx doctor
```

安装器会下载 `SHA256SUMS` 并校验二进制后再激活。
它还会把 `ralphx` Codex skill 安装到 `~/.codex/skills/ralphx`。
它还会把受管的 Codex hooks 安装到 `~/.codex/hooks.json`。

你也可以通过 CLI 重新安装或刷新 skill：

```bash
ralphx skill install
ralphx skill install --project
```

可选：只有在明确需要 delegation 时，才发现或安装精选 subagent 集：

```bash
ralphx agents discover
ralphx agents install
ralphx agents install --project
```

你也可以通过 CLI 安装或卸载 hooks：

```bash
ralphx hook install
ralphx hook uninstall
```

## 执行路径持久化

当前激活执行路径持久化到：

```bash
~/.config/ralphx/current.env
```

查看当前持久化执行状态：

```bash
ralphx current
```

从目标生成 task 和 checklist：

```bash
ralphx plan --goal "finish the current migration batch" --out tasks/migration.md
```

基于当前状态重生成下一版 task/checklist：

```bash
ralphx replan --task tasks/migration.md
```

恢复仍然有效的 Codex session：

```bash
ralphx run --task tasks/migration.md --resume --session-expiry 24h
```

在新的 Codex session 里激活工作流：

```text
$ralphx
```

如果刚修改过 hooks，先开一个新 session，再提交 `$ralphx`，这样 `UserPromptSubmit` 才会触发。
