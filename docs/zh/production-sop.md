# 生产 SOP

这份 SOP 覆盖 `ralphx` 的完整执行链路：

1. 安装 release 二进制
2. 确认当前激活的 wrapper
3. 准备 hooks 和运行时状态
4. 从目标生成或加载 task / checklist
5. 带验证运行外层循环
6. 在阻塞或失去进展时自动重规划
7. 持续恢复，直到任务真正完成

在 GPT-5.5 下，`ralphx` 是执行纪律层，不是替代模型推理的层。
GPT-5.5 负责更好的规划和编码；`ralphx` 负责状态、checklist gate、验证 gate、Stop hook 续跑和恢复链路。

## 1. 安装

使用 release 安装，不走源码安装。

最新版本：

```bash
curl -fsSL https://github.com/ckken/ralphx/releases/latest/download/install.sh | bash
```

指定版本：

```bash
curl -fsSL https://github.com/ckken/ralphx/releases/download/v0.2.3/install.sh | VERSION=v0.2.3 bash
```

安装器会先校验 `SHA256SUMS`，再激活二进制。
当前激活执行路径持久化在：

```bash
~/.config/ralphx/current.env
```

确认当前 wrapper 和实际执行二进制：

```bash
ralphx doctor
ralphx current
```

## 2. 准备 Hooks

安装受管的 stop hook，让运行时能在任务真正完成前持续接管循环：

```bash
ralphx hook install
ralphx hook status --workdir "$PWD"
```

如果需要移除受管 hook：

```bash
ralphx hook uninstall
```

## 3. 从目标开始

如果只有一个目标描述，先生成 task 和 checklist，再直接交给 runner：

```bash
ralphx plan --goal "完成当前迁移批次" --out tasks/migration.md --execute
```

如果已经有 task 和 checklist，直接启动循环：

```bash
TESTS_CMD="go test ./..." ralphx run --task tasks/migration.md --checklist tasks/migration.checklist.md --resume --session-expiry 24h
```

推荐默认值：

- 始终配置 `TESTS_CMD` 作为真实验证
- 默认保留 `RALPHX_AUTO_REPLAN=1`
- 只有在前一个 Codex session 还足够新鲜时才使用 `--resume`

## 4. 持续跑完整个循环

`ralphx run` 会负责外层循环：

- 调用 Codex
- 写入 `.ralphx/` 状态
- 在有意义的进展后执行验证
- 保持 checklist gate 作为最终完成标准
- 在阻塞、失去进展或变旧时自动触发重规划

如果运行后返回了重新生成的文件和重规划提示，先检查新的 task / checklist，再重新执行：

```bash
ralphx replan --task tasks/migration.md --execute
```

## 5. 恢复和续跑

需要检查或恢复时，优先看这些命令：

```bash
ralphx current
ralphx hook status --workdir "$PWD"
```

本地运行时状态在：

```bash
.ralphx/
```

如果 session 还在有效期内，直接续跑：

```bash
ralphx run --task tasks/migration.md --checklist tasks/migration.checklist.md --resume --session-expiry 24h
```

如果 session 已过期，就让 `ralphx` 从 task / checklist / summary 状态重新起一个 fresh session。

## 6. 干净收口

只有同时满足下面条件时，循环才应该停止：

- task 已经真正完成
- checklist 已经关闭
- 必要验证已经通过
- stop hook 没有继续阻止退出的理由

收口时只需要做一个简洁的最终检查：

```bash
ralphx current
ralphx hook status --workdir "$PWD"
```
