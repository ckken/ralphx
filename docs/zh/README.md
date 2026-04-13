     1|# ralphx
     2|
     3|[English](../../README.md) | [中文](README.md)
     4|
     5|快速入口：
     6|- [安装说明](installation.md)
     7|- [方法论](methodology.md)
     8|- [生产 SOP](production-sop.md)
     9|- [并行协议 v0](../en/go-parallel-protocol-v0.md)
    10|- [流程图 / 链路图](architecture.md)
    11|
    12|`ralphx` 是一个基于 Go 的 Codex / coding agent 外层执行器。
    13|
    14|核心目标：
    15|- 让 agent 用当前工具持续推进，直到真实任务完成
    16|- 用 checklist / validation / leader gate 阻止过早完成
    17|- 在任务可拆分时支持本地多 worker 并行执行
    18|
    19|## Release 安装（推荐）
    20|
    21|正常使用不需要源码安装。
    22|
    23|安装最新 release：
    24|
    25|```bash
    26|curl -fsSL https://github.com/ckken/ralphx/releases/latest/download/install.sh | bash
    27|ralphx-doctor
    28|```
    29|
    30|安装指定版本：
    31|
    32|```bash
    33|curl -fsSL https://github.com/ckken/ralphx/releases/download/v0.1.0/install.sh | VERSION=v0.1.0 bash
    34|ralphx-doctor
    35|```
    36|
    37|如果需要：
    38|
    39|```bash
    40|export PATH="$HOME/.local/bin:$PATH"
    41|```
    42|
    43|## 执行路径持久化
    44|
    45|安装器会把当前激活的执行路径持久化到：
    46|
    47|```bash
    48|~/.config/ralphx/current.env
    49|```
    50|
    51|下载下来的版本二进制会放在：
    52|
    53|```bash
    54|~/.local/share/ralphx/releases/
    55|```
    56|
    57|所以 `ralphx` / `ralphx-doctor` 始终是稳定入口，实际执行版本由持久化路径决定。

查看当前持久化执行状态：

```bash
ralphx current
```
    58|
    59|## 依赖
    60|
    61|必需：
    62|- `codex`
    63|- `curl` 或 `wget`
    64|
    65|建议：
    66|- `git`
    67|- `gh`
    68|- `bash`
    69|- `python3`
    70|
    71|可选：
    72|- `jq`（仅 legacy 辅助；Go 主链路不依赖）
    73|
    74|## 快速开始
    75|
    76|单 worker：
    77|
    78|```bash
    79|ralphx --task ./examples/sample-task.md --workdir /path/to/repo
    80|```
    81|
    82|显式指定 checklist：
    83|
    84|```bash
    85|ralphx   --task ./examples/sample-task.md   --checklist ./examples/sample-task.checklist.md   --workdir /path/to/repo
    86|```
    87|
    88|并行 checklist：
    89|
    90|```bash
    91|ralphx   --task ./examples/sample-task.md   --checklist ./examples/sample-task.checklist.md   --workdir /path/to/repo   --workers 3
    92|```
    93|
    94|## 常用环境变量
    95|
    96|```bash
    97|export CODEX_CMD=codex
    98|export CODEX_ARGS='-m gpt-5.4'
    99|export TESTS_CMD='go test ./...'
   100|export MAX_ITERATIONS=0
   101|export MAX_NO_PROGRESS=0
   102|export ROUND_TIMEOUT_SECONDS=1800
   103|export RALPHX_WORKERS=3
   104|```
   105|
   106|## 生产建议
   107|
   108|1. 先跑 `ralphx-doctor`
   109|2. 准备 task + checklist
   110|3. 默认先用 `--workers 1`
   111|4. checklist 明确可拆时再开并行
   112|5. 配置 `TESTS_CMD`
   113|6. 运行后检查 `.ralphx/last-result.json`、`.ralphx/state.json`、`.ralphx/results/`
   114|
   115|完整操作请看：[生产 SOP](production-sop.md)
   116|