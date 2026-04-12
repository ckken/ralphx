# codex-ralph

[English](../../README.md) | [中文](README.md)

快速入口：
- [安装说明](installation.md)
- [方法论](methodology.md)
- [流程图 / 链路图](architecture.md)

`codex-ralph` 是一个基于 Bash 的 Codex 外层控制器。

它把 Codex 变成一个受控的外层循环系统：

- 任务驱动
- checklist gate
- 验证优先
- 抵抗过早完成

## 它解决什么问题

Codex 很容易在完成一个局部切片后过早宣布成功。`codex-ralph` 提供一个外层循环来：

- 读取任务文件
- 可选读取 checklist
- 以非交互方式调用 `codex exec`
- 强制要求严格 JSON 输出
- 在每轮成功后执行验证
- 拒绝弱完成信号

结果是一套可重复的“直到真实任务完成才停”的工作流。

## 项目结构

- `codex-loop.sh`：主执行器
- `doctor.sh`：依赖和环境自检
- `install.sh`：安装命令包装器到 `~/.local/bin`
- `uninstall.sh`：移除已安装命令
- `prompts/loop-system-prompt.md`：循环系统提示词
- `schemas/loop-output.schema.json`：严格输出 schema
- `docs/en/`：英文文档
- `docs/zh/`：中文文档
- `examples/`：任务与 checklist 示例
- `tasks/`：项目开发中使用的真实任务示例

## 安装

```bash
git clone https://github.com/ckken/codex-ralph.git
cd codex-ralph
./install.sh
codex-ralph-doctor
```

如果需要：

```bash
export PATH="$HOME/.local/bin:$PATH"
```

## 必需依赖

- `bash`
- `jq`
- `python3`
- `codex`

建议安装：

- `git`
- `gh`
- `timeout` 或 `gtimeout`

## 快速开始

使用任务文件：

```bash
codex-ralph --task ./examples/sample-task.md --workdir /path/to/repo
```

显式指定 checklist：

```bash
codex-ralph \
  --task ./examples/sample-task.md \
  --checklist ./examples/sample-task.checklist.md \
  --workdir /path/to/repo
```

## 运行参数

常用环境变量：

```bash
export CODEX_CMD=codex
export CODEX_ARGS='-m gpt-5.4-mini'
export TESTS_CMD='bun src/index.ts --help && bash scripts/verify-golden.sh --skip-build'
export MAX_ITERATIONS=0
export MAX_NO_PROGRESS=0
export ROUND_TIMEOUT_SECONDS=1800
```

含义：

- `MAX_ITERATIONS=0`：不设迭代上限
- `MAX_NO_PROGRESS=0`：不设无进展停止阈值
- `CHECKLIST_FILE`：强制指定 checklist 路径
- `TESTS_CMD`：每轮成功后执行的验证链

## 方法

相关文档：

- [方法论](methodology.md)
- [安装说明](installation.md)
- [流程图 / 链路图](architecture.md)

简版方法：

1. Codex 负责局部实现
2. Bash 负责外层控制循环
3. Checklist 未完成项视为硬剩余工作
4. 验证 gate 用于阻止坏进展
5. 只有总任务真正完成时才接受完成

## 输出契约

每一轮 Codex 必须返回一个 JSON 对象：

```json
{
  "status": "in_progress|blocked|complete",
  "exit_signal": true,
  "files_modified": 0,
  "tests_passed": false,
  "blockers": [],
  "summary": ""
}
```

## 已安装命令

- `codex-ralph`
- `codex-ralph-doctor`
