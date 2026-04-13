     1|# Production SOP
     2|
     3|This SOP is the recommended operating path for using `ralphx` in a real repository.
     4|
     5|## 1. Installation
     6|
     7|Use release installation, not source installation.
     8|
     9|Latest:
    10|
    11|```bash
    12|curl -fsSL https://github.com/ckken/ralphx/releases/latest/download/install.sh | bash
    13|```
    14|
    15|Specific version:
    16|
    17|```bash
    18|curl -fsSL https://github.com/ckken/ralphx/releases/download/v0.1.0/install.sh | VERSION=v0.1.0 bash
    19|```
    20|
    21|The installer persists the active execution path in:
    22|
    23|```bash
    24|~/.config/ralphx/current.env
    25|```
    26|
    27|This allows the wrapper commands to remain stable while the underlying versioned binaries live under `~/.local/share/ralphx/releases/`.

Inspect the current persisted execution target with:

```bash
ralphx current
```
    28|
    29|## 2. Preflight
    30|
    31|Run:
    32|
    33|```bash
    34|ralphx-doctor
    35|```
    36|
    37|Minimum expected:
    38|- `codex` is available
    39|- `git` is available if you want git-aware completion checks
    40|
    41|## 3. Prepare inputs
    42|
    43|Create:
    44|- one task file describing the total objective
    45|- one checklist file for bounded milestones
    46|
    47|Rules:
    48|- task file = overall objective
    49|- checklist = hard remaining work
    50|- checklist items should be short, independently understandable, and verification-friendly
    51|
    52|## 4. Choose execution mode
    53|
    54|### Single-worker mode
    55|Use when:
    56|- work is tightly coupled
    57|- file conflicts are likely
    58|- you want the safest baseline path
    59|
    60|```bash
    61|ralphx --task task.md --checklist task.checklist.md --workdir /repo
    62|```
    63|
    64|### Parallel mode
    65|Use when:
    66|- checklist items are separable
    67|- each item is a bounded slice
    68|- you want faster local execution
    69|
    70|```bash
    71|ralphx --task task.md --checklist task.checklist.md --workdir /repo --workers 3
    72|```
    73|
    74|Recommended production rule:
    75|- default to `--workers 1`
    76|- only increase workers after the checklist is clearly decomposed
    77|
    78|## 5. Add validation
    79|
    80|Set `TESTS_CMD` whenever possible.
    81|
    82|Examples:
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
    96|## 6. Run
    97|
    98|Typical production invocation:
    99|
   100|```bash
   101|export CODEX_CMD=codex
   102|export CODEX_ARGS='-m gpt-5.4'
   103|export TESTS_CMD='go test ./...'
   104|
   105|ralphx   --task docs/tasks/release-task.md   --checklist docs/tasks/release-task.checklist.md   --workdir /path/to/repo   --workers 3
   106|```
   107|
   108|## 7. Inspect outputs
   109|
   110|Key files after a run:
   111|- `.ralphx/last-result.json`
   112|- `.ralphx/state.json`
   113|- `.ralphx/stats.json`
   114|- `.ralphx/logs/`
   115|- `.ralphx/results/` (parallel mode)
   116|- `.ralphx/runtime/loop-output.schema.json`
   117|
   118|Interpretation:
   119|- `complete`: all gates passed and leader accepted completion
   120|- `in_progress`: work remains or a complete signal was downgraded
   121|- `blocked`: real blocker, invalid output, or validation failure
   122|
   123|## 8. GitHub release promotion SOP
   124|
   125|Before tagging a release:
   126|- run `go build ./...`
   127|- run `go test ./...`
   128|- run `ralphx-doctor`
   129|- run one single-worker smoke path
   130|- run one parallel smoke path if you intend to use `--workers`
   131|- inspect `.ralphx/last-result.json`
   132|
   133|Tag and push:
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
   144|## 9. Operating boundaries
   145|
   146|Do:
   147|- use bounded checklist items
   148|- keep validation cheap but meaningful
   149|- review result files after blocked runs
   150|- start with lower worker counts
   151|
   152|Do not:
   153|- use parallel mode for overlapping risky edits without a checklist
   154|- assume worker completion equals total completion
   155|- skip validation on production changes unless absolutely necessary
   156|