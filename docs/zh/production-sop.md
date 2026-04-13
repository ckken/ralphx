     1|# 生产 SOP
     2|
     3|这是 `ralphx` 在真实仓库中的推荐操作路径。
     4|
     5|## 1. 安装
     6|
     7|使用 release 安装，不走源码安装。
     8|
     9|最新版本：
    10|
    11|```bash
    12|curl -fsSL https://github.com/ckken/ralphx/releases/latest/download/install.sh | bash
    13|```
    14|
    15|指定版本：
    16|
    17|```bash
    18|curl -fsSL https://github.com/ckken/ralphx/releases/download/v0.1.0/install.sh | VERSION=v0.1.0 bash
    19|```
    20|
    21|安装器会把当前激活的执行路径持久化到：
    22|
    23|```bash
    24|~/.config/ralphx/current.env
    25|```
    26|
    27|这样 wrapper 命令保持稳定，而底层实际执行的版本化二进制放在 `~/.local/share/ralphx/releases/`。

查看当前持久化执行目标：

```bash
ralphx current
```
    28|
    29|## 2. 预检
    30|
    31|先执行：
    32|
    33|```bash
    34|ralphx-doctor
    35|```
    36|
    37|最低要求：
    38|- `codex` 可用
    39|- 如果要用 git 感知 gate，建议安装 `git`
    40|
    41|## 3. 准备输入
    42|
    43|准备两类文件：
    44|- task file：描述总目标
    45|- checklist：描述可验证、可拆分的剩余工作
    46|
    47|规则：
    48|- task = 总目标
    49|- checklist = 硬剩余工作
    50|- checklist item 要短、边界清晰、便于验证
    51|
    52|## 4. 选择执行模式
    53|
    54|### 单 worker
    55|适合：
    56|- 工作强耦合
    57|- 文件冲突可能大
    58|- 先走最稳路径
    59|
    60|```bash
    61|ralphx --task task.md --checklist task.checklist.md --workdir /repo
    62|```
    63|
    64|### 并行模式
    65|适合：
    66|- checklist 可拆分
    67|- 每项是有边界的 slice
    68|- 需要加速本地执行
    69|
    70|```bash
    71|ralphx --task task.md --checklist task.checklist.md --workdir /repo --workers 3
    72|```
    73|
    74|推荐规则：
    75|- 默认从 `--workers 1` 开始
    76|- 只有 checklist 确实可拆时才增加 worker 数
    77|
    78|## 5. 配置验证
    79|
    80|尽量总是配置 `TESTS_CMD`。
    81|
    82|例如：
    83|
    84|```bash
    85|export TESTS_CMD='go test ./...'
    86|```
    87|
    88|```bash
    89|export TESTS_CMD='bun test && bun run lint'
    90|```
    91|
    92|```bash
    93|export TESTS_CMD='pytest -q'
    94|```
    95|
    96|## 6. 运行
    97|
    98|典型生产调用：
    99|
   100|```bash
   101|export CODEX_CMD=codex
   102|export CODEX_ARGS='-m gpt-5.4'
   103|export TESTS_CMD='go test ./...'
   104|
   105|ralphx   --task docs/tasks/release-task.md   --checklist docs/tasks/release-task.checklist.md   --workdir /path/to/repo   --workers 3
   106|```
   107|
   108|## 7. 查看输出
   109|
   110|关键文件：
   111|- `.ralphx/last-result.json`
   112|- `.ralphx/state.json`
   113|- `.ralphx/stats.json`
   114|- `.ralphx/logs/`
   115|- `.ralphx/results/`（并行模式）
   116|- `.ralphx/runtime/loop-output.schema.json`
   117|
   118|判断方式：
   119|- `complete`：所有 gate 通过，leader 接受完成
   120|- `in_progress`：工作仍剩余，或 complete 被降级
   121|- `blocked`：真实阻塞、无效输出、或验证失败
   122|
   123|## 8. GitHub Release 推进 SOP
   124|
   125|打 tag 前检查：
   126|- `go build ./...`
   127|- `go test ./...`
   128|- `ralphx-doctor`
   129|- 跑一次单 worker smoke
   130|- 如果要用并行，再跑一次 `--workers` smoke
   131|- 检查 `.ralphx/last-result.json`
   132|
   133|打 tag 并推送：
   134|
   135|```bash
   136|git status
   137|git add .
   138|git commit -m "feat: release prep"
   139|git push origin main
   140|git tag v0.1.0
   141|git push origin v0.1.0
   142|```
   143|