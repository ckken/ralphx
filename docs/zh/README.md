# codex-ralph 中文说明

`codex-ralph` 是一个基于 Bash 的 Codex 外层控制器。

它解决的问题不是“怎么调用一次 Codex”，而是“怎么让 Codex 在一个明确约束下持续推进，直到总任务真正完成”。

## 适合什么场景

- 多轮重构
- 迁移任务
- 需要验收门的自动推进
- 不希望模型做完一个局部点就误判完成

## 核心方法论

### 1. 总任务驱动

任务文件描述的是总目标，不是一个局部子任务。

模型每轮都只能在这个总目标下推进，而不是自己随意缩窄范围。

### 2. Checklist 作为硬门

如果提供了 checklist：

- 未勾选项 = 还没完成的硬任务
- 即使模型返回 `complete`
- 只要 checklist 还有未完成项，执行器就不会停

### 3. Bash 外层掌控停止条件

停止权不交给模型，而交给外层执行器。

执行器负责：

- 循环次数
- 超时
- Git 状态快照
- 无进展判断
- checklist gating
- 验证命令执行

### 4. 验证优先

每一轮成功改动后，可以跑低成本验证链，例如：

```bash
bun src/index.ts --help
bash scripts/verify-golden.sh --skip-build
bash scripts/verify-batch.sh --skip-build
```

如果验证失败，就不能继续把这一轮当作有效进展。

### 5. 防止“假完成”

`codex-ralph` 会拒绝两类常见误停：

1. 模型说完成了，但没有任何真实代码变化
2. 模型说完成了，但 checklist 还有未完成项

## 需要安装的内容

必需：

- `bash`
- `jq`
- `python3`
- `codex`

建议：

- `git`
- `gh`
- `timeout` 或 `gtimeout`

## 安装

```bash
git clone https://github.com/ckken/codex-ralph.git
cd codex-ralph
./install.sh
codex-ralph-doctor
```

如果命令找不到：

```bash
export PATH="$HOME/.local/bin:$PATH"
```

## 最常见用法

```bash
codex-ralph --task ./examples/sample-task.md --workdir /path/to/repo
```

带 checklist：

```bash
codex-ralph \
  --task ./examples/sample-task.md \
  --checklist ./examples/sample-task.checklist.md \
  --workdir /path/to/repo
```

## 推荐阅读顺序

1. `README.md`
2. `docs/installation.md`
3. `docs/methodology.md`
4. `docs/architecture.md`

这样可以先装起来，再理解背后的控制逻辑。
