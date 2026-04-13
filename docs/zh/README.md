# ralphx

[English](../../README.md) | [中文](README.md)

`ralphx` 是一个基于 Go 的 Codex / coding agent 外层执行器。

## Release 安装（推荐）

正常使用不需要源码安装。

安装最新 release：

```bash
curl -fsSL https://github.com/ckken/ralphx/releases/latest/download/install.sh | bash
ralphx doctor
```

安装指定版本：

```bash
curl -fsSL https://github.com/ckken/ralphx/releases/download/v0.1.2/install.sh | VERSION=v0.1.2 bash
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
