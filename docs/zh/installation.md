# 安装说明

## 推荐：通过 GitHub Release 安装

正常使用不需要源码安装。

安装最新 release：

```bash
curl -fsSL https://github.com/ckken/ralphx/releases/latest/download/install.sh | bash
```

安装指定版本：

```bash
curl -fsSL https://github.com/ckken/ralphx/releases/download/v0.1.2/install.sh | VERSION=v0.1.2 bash
```

安装器会下载 `SHA256SUMS`，校验通过后才激活二进制。
它还会把 `ralphx` Codex skill 安装到 `~/.codex/skills/ralphx`。

你也可以通过 CLI 重新安装或刷新 skill：

```bash
ralphx skill install
ralphx skill install --project
```

## 验证

```bash
ralphx doctor
ralphx --help
ralphx version
ralphx current
```

## 常用流程

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
