     1|# ralphx
     2|
     3|[English](README.md) | [中文](docs/zh/README.md)
     4|
     5|Quick links:
     6|- [Installation](docs/en/installation.md)
     7|- [Methodology](docs/en/methodology.md)
     8|- [Production SOP](docs/en/production-sop.md)
     9|- [Parallel protocol v0](docs/en/go-parallel-protocol-v0.md)
    10|- [Flowcharts / Chain Diagram](docs/en/architecture.md)
    11|
    12|`ralphx` is a Go-based outer-loop runner for Codex and coding agents.
    13|
    14|It is designed for one core goal:
    15|- let the agent keep working with the current tools until the real task is done
    16|- keep completion gated by checklist / validation / leader-side rules
    17|- support local multi-worker execution when the task is checklist-decomposable
    18|
    19|## What it does
    20|
    21|`ralphx` gives you a local-first execution loop that can:
    22|- read a task file
    23|- optionally read a checklist
    24|- invoke `codex exec` with a strict JSON contract
    25|- persist run state under `.ralphx/`
    26|- reject premature completion
    27|- run validation commands
    28|- split checklist items into parallel worker jobs with `--workers N`
    29|
    30|## Install from GitHub release
    31|
    32|No source build is required for normal use.
    33|
    34|Latest release:
    35|
    36|```bash
    37|curl -fsSL https://github.com/ckken/ralphx/releases/latest/download/install.sh | bash
    38|ralphx-doctor
    39|```
    40|
    41|Install a specific version:
    42|
    43|```bash
    44|curl -fsSL https://github.com/ckken/ralphx/releases/download/v0.1.0/install.sh | VERSION=v0.1.0 bash
    45|ralphx-doctor
    46|```
    47|
    48|If needed:
    49|
    50|```bash
    51|export PATH="$HOME/.local/bin:$PATH"
    52|```
    53|
    54|## Persistent execution path
    55|
    56|The installer persists the active binary paths in:
    57|
    58|```bash
    59|~/.config/ralphx/current.env
    60|```
    61|
    62|Downloaded release binaries are stored under:
    63|
    64|```bash
    65|~/.local/share/ralphx/releases/
    66|```
    67|
    68|The `ralphx` and `ralphx-doctor` commands are stable wrappers in `~/.local/bin` that always read the current persisted execution path.

Inspect the active persisted execution state:

```bash
ralphx current
```
    69|
    70|## Runtime dependencies
    71|
    72|Required:
    73|- `codex`
    74|- `curl` or `wget` for release installation
    75|
    76|Recommended:
    77|- `git`
    78|- `gh`
    79|- `bash`
    80|- `python3`
    81|
    82|Optional:
    83|- `jq` (legacy-only helper; not required by the Go-native main path)
    84|
    85|## Quick start
    86|
    87|Run a single-worker task:
    88|
    89|```bash
    90|ralphx --task ./examples/sample-task.md --workdir /path/to/repo
    91|```
    92|
    93|Run with an explicit checklist:
    94|
    95|```bash
    96|ralphx   --task ./examples/sample-task.md   --checklist ./examples/sample-task.checklist.md   --workdir /path/to/repo
    97|```
    98|
    99|Run checklist items in parallel:
   100|
   101|```bash
   102|ralphx   --task ./examples/sample-task.md   --checklist ./examples/sample-task.checklist.md   --workdir /path/to/repo   --workers 3
   103|```
   104|
   105|## Common environment variables
   106|
   107|```bash
   108|export CODEX_CMD=codex
   109|export CODEX_ARGS='-m gpt-5.4'
   110|export TESTS_CMD='go test ./...'
   111|export MAX_ITERATIONS=0
   112|export MAX_NO_PROGRESS=0
   113|export ROUND_TIMEOUT_SECONDS=1800
   114|export RALPHX_WORKERS=3
   115|```
   116|
   117|## Recommended production path
   118|
   119|1. Run `ralphx-doctor`
   120|2. Prepare a task file and checklist
   121|3. Start with `--workers 1` unless the checklist items are clearly independent
   122|4. Use `--workers N` only when checklist items are bounded and separable
   123|5. Set `TESTS_CMD` so every successful round is validated
   124|6. Review `.ralphx/last-result.json`, `.ralphx/state.json`, and `.ralphx/results/` after runs
   125|
   126|See the full rollout guide in [Production SOP](docs/en/production-sop.md).
   127|
   128|## Output contract
   129|
   130|Each agent round must return one JSON object:
   131|
   132|```json
   133|{
   134|  "status": "in_progress|blocked|complete",
   135|  "exit_signal": true,
   136|  "files_modified": 0,
   137|  "tests_passed": false,
   138|  "blockers": [],
   139|  "summary": ""
   140|}
   141|```
   142|
   143|## Installed commands
   144|
   145|- `ralphx`
   146|- `ralphx-doctor`
- `ralphx current`
   147|