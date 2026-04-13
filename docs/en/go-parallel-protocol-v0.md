# Go parallel protocol v0

This is a minimal leader/worker protocol for a local-first `ralphx` Go rewrite.

## Goals

- durable local state under `.ralphx/`
- one leader process owns scheduling
- many short-lived worker subprocesses do bounded work
- workers can later wrap `codex exec` without changing the state model
- leader remains the only authority for completion, retries, and checklist gates

## Design constraints

1. Single leader lock per run.
2. Workers never mutate global queue state directly.
3. Every state file is JSON, written atomically with temp file + rename.
4. Append-only `events.jsonl` is optional for observability but useful now.
5. A worker handles exactly one assignment at a time.
6. A worker result is advisory. Only the leader can mark the run complete.

## State layout

```text
.ralphx/
  run.json
  leader.lock
  events.jsonl
  summary.txt
  tasks/
    task-0001.json
    task-0002.json
  workers/
    worker-01.json
    worker-02.json
  logs/
    worker-01.stdout.log
    worker-01.stderr.log
    worker-02.stdout.log
    worker-02.stderr.log
  results/
    task-0001.result.json
    task-0002.result.json
```

## Ownership model

- `run.json`: leader writes
- `tasks/*.json`: leader writes
- `workers/*.json`: worker writes its own file, leader may initialize or mark terminal status after process exit
- `results/*.result.json`: worker writes once, leader reads then folds into task state
- `events.jsonl`: leader appends; workers may optionally emit events through stdout and let leader append them

This keeps concurrency simple: leader owns queue state, workers own only heartbeat/result files.

## File schemas

### 1. `run.json`

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://ralphx.dev/schemas/run-state-v0.json",
  "type": "object",
  "additionalProperties": false,
  "required": [
    "version",
    "run_id",
    "status",
    "workdir",
    "task_file",
    "checklist_file",
    "max_parallel",
    "created_at",
    "updated_at",
    "leader_pid",
    "task_counts"
  ],
  "properties": {
    "version": { "const": "v0" },
    "run_id": { "type": "string", "minLength": 1 },
    "status": {
      "type": "string",
      "enum": ["starting", "running", "paused", "blocked", "complete", "failed", "canceled"]
    },
    "workdir": { "type": "string", "minLength": 1 },
    "task_file": { "type": "string", "minLength": 1 },
    "checklist_file": { "type": ["string", "null"] },
    "max_parallel": { "type": "integer", "minimum": 1 },
    "leader_pid": { "type": "integer", "minimum": 1 },
    "created_at": { "type": "string", "format": "date-time" },
    "updated_at": { "type": "string", "format": "date-time" },
    "current_summary": { "type": "string" },
    "checklist_open_items": { "type": "integer", "minimum": 0 },
    "last_task_seq": { "type": "integer", "minimum": 0 },
    "task_counts": {
      "type": "object",
      "additionalProperties": false,
      "required": ["pending", "running", "succeeded", "blocked", "failed"],
      "properties": {
        "pending": { "type": "integer", "minimum": 0 },
        "running": { "type": "integer", "minimum": 0 },
        "succeeded": { "type": "integer", "minimum": 0 },
        "blocked": { "type": "integer", "minimum": 0 },
        "failed": { "type": "integer", "minimum": 0 }
      }
    }
  }
}
```

### 2. `tasks/task-XXXX.json`

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://ralphx.dev/schemas/task-state-v0.json",
  "type": "object",
  "additionalProperties": false,
  "required": [
    "version",
    "task_id",
    "kind",
    "status",
    "title",
    "attempt",
    "max_attempts",
    "created_at",
    "updated_at",
    "input"
  ],
  "properties": {
    "version": { "const": "v0" },
    "task_id": {
      "type": "string",
      "pattern": "^task-[0-9]{4,}$"
    },
    "kind": {
      "type": "string",
      "enum": ["implement", "validate", "repair", "summarize"]
    },
    "status": {
      "type": "string",
      "enum": ["pending", "assigned", "running", "succeeded", "blocked", "failed", "canceled"]
    },
    "title": { "type": "string", "minLength": 1 },
    "assigned_worker_id": { "type": ["string", "null"] },
    "attempt": { "type": "integer", "minimum": 0 },
    "max_attempts": { "type": "integer", "minimum": 1 },
    "created_at": { "type": "string", "format": "date-time" },
    "updated_at": { "type": "string", "format": "date-time" },
    "started_at": { "type": ["string", "null"], "format": "date-time" },
    "finished_at": { "type": ["string", "null"], "format": "date-time" },
    "depends_on": {
      "type": "array",
      "items": { "type": "string" },
      "uniqueItems": true
    },
    "input": {
      "type": "object",
      "additionalProperties": false,
      "required": ["prompt", "bounds"],
      "properties": {
        "prompt": { "type": "string", "minLength": 1 },
        "checklist_item": { "type": ["string", "null"] },
        "validation_cmd": { "type": ["string", "null"] },
        "bounds": {
          "type": "object",
          "additionalProperties": false,
          "required": ["max_runtime_seconds", "max_files_to_touch"],
          "properties": {
            "max_runtime_seconds": { "type": "integer", "minimum": 1 },
            "max_files_to_touch": { "type": "integer", "minimum": 1 }
          }
        }
      }
    },
    "result_ref": { "type": ["string", "null"] },
    "leader_notes": {
      "type": "array",
      "items": { "type": "string" }
    }
  }
}
```

### 3. `workers/worker-XX.json`

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://ralphx.dev/schemas/worker-state-v0.json",
  "type": "object",
  "additionalProperties": false,
  "required": [
    "version",
    "worker_id",
    "status",
    "created_at",
    "updated_at"
  ],
  "properties": {
    "version": { "const": "v0" },
    "worker_id": {
      "type": "string",
      "pattern": "^worker-[0-9]{2,}$"
    },
    "status": {
      "type": "string",
      "enum": ["starting", "idle", "running", "stopping", "exited", "lost"]
    },
    "pid": { "type": ["integer", "null"], "minimum": 1 },
    "task_id": { "type": ["string", "null"] },
    "attempt_key": { "type": ["string", "null"] },
    "started_at": { "type": ["string", "null"], "format": "date-time" },
    "heartbeat_at": { "type": ["string", "null"], "format": "date-time" },
    "updated_at": { "type": "string", "format": "date-time" },
    "exit_code": { "type": ["integer", "null"] },
    "stdout_log": { "type": ["string", "null"] },
    "stderr_log": { "type": ["string", "null"] },
    "last_error": { "type": ["string", "null"] }
  }
}
```

### 4. `results/task-XXXX.result.json`

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://ralphx.dev/schemas/worker-result-v0.json",
  "type": "object",
  "additionalProperties": false,
  "required": [
    "version",
    "task_id",
    "worker_id",
    "attempt_key",
    "status",
    "summary",
    "files_modified",
    "tests_passed",
    "blockers",
    "exit_signal",
    "started_at",
    "finished_at"
  ],
  "properties": {
    "version": { "const": "v0" },
    "task_id": { "type": "string" },
    "worker_id": { "type": "string" },
    "attempt_key": { "type": "string", "minLength": 1 },
    "status": {
      "type": "string",
      "enum": ["in_progress", "blocked", "complete", "failed", "retryable"]
    },
    "summary": { "type": "string" },
    "files_modified": { "type": "integer", "minimum": 0 },
    "tests_passed": { "type": "boolean" },
    "blockers": {
      "type": "array",
      "items": { "type": "string" }
    },
    "exit_signal": { "type": "boolean" },
    "raw_output_path": { "type": ["string", "null"] },
    "started_at": { "type": "string", "format": "date-time" },
    "finished_at": { "type": "string", "format": "date-time" }
  }
}
```

### 5. Optional `events.jsonl` line schema

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://ralphx.dev/schemas/event-v0.json",
  "type": "object",
  "additionalProperties": false,
  "required": ["ts", "seq", "type", "run_id"],
  "properties": {
    "ts": { "type": "string", "format": "date-time" },
    "seq": { "type": "integer", "minimum": 1 },
    "type": {
      "type": "string",
      "enum": [
        "run_started",
        "task_created",
        "task_assigned",
        "worker_started",
        "worker_heartbeat",
        "worker_exited",
        "result_recorded",
        "task_requeued",
        "run_completed",
        "run_blocked",
        "run_failed"
      ]
    },
    "run_id": { "type": "string" },
    "task_id": { "type": ["string", "null"] },
    "worker_id": { "type": ["string", "null"] },
    "message": { "type": "string" }
  }
}
```

## Protocol

### Single-leader rule

The leader acquires `.ralphx/leader.lock`. If the lock is already held by a live process, a second leader exits or offers `resume --force`.

### Assignment rule

The leader is the only scheduler:

1. create or load pending tasks
2. pick a ready task with all dependencies satisfied
3. write `tasks/task-XXXX.json` with `status=assigned`, `assigned_worker_id`, incremented `attempt`
4. write or update `workers/worker-XX.json`
5. spawn `ralphx worker --run-id <run_id> --worker-id <id> --task-id <task_id>`

Workers do not self-claim from a shared queue in v0. That avoids lease races and keeps implementation small.

### Worker contract

Inputs are read-only except:

- `workers/<worker_id>.json`
- `results/<task_id>.result.json`
- worker stdout/stderr log files

A worker must not rewrite `run.json` or any `tasks/*.json` file.

### Result folding rule

When a worker exits or a result file appears, the leader:

1. validates the result schema
2. checks worker pid/exit code if available
3. compares git status before/after if that gate is enabled
4. applies current checklist gate
5. updates `tasks/task-XXXX.json`
6. updates `run.json`
7. decides requeue / next task / terminal state

## Minimal worker lifecycle

```text
starting
  -> running
  -> stopping
  -> exited
```

More complete state semantics:

1. `starting`
   - worker file created
   - pid known
   - task already assigned by leader

2. `running`
   - worker updates `heartbeat_at` every 5-15s
   - worker executes one bounded slice
   - expected to write exactly one result file before exit on normal path

3. `stopping`
   - worker received cancel or is flushing final output
   - optional in v0; leader may jump directly to `exited`

4. `exited`
   - worker process ended
   - leader has observed exit
   - task is already folded or marked for retry

5. `lost`
   - heartbeat stale and process no longer exists
   - leader marks current task `failed` or `pending` for retry

## Bounded task rules

Each worker assignment should be small enough to retry cheaply.

Recommended v0 bounds:

- one checklist item or one narrow implementation slice per task
- default `max_runtime_seconds`: 900 to 1800
- default `max_files_to_touch`: 3 to 10
- default `max_attempts`: 2 or 3

Good examples:

- implement a single checklist item
- repair one failing validation command
- summarize current repo state for replanning

Avoid in v0:

- one worker handling the entire top-level task
- workers spawning nested workers
- workers deciding global completion

## Leader reconciliation rules

The leader should preserve the current Bash safety behavior:

1. Reject `complete` if checklist items remain.
2. Reject `complete` if no real file changes occurred when changes were expected.
3. If validation fails, mark the task `blocked` or create a repair task.
4. Only the leader can set `run.status=complete`.
5. A worker result with `exit_signal=true` means “this slice thinks it is done”, not “the run is done”.

Suggested mapping from worker result to task state:

- `status=complete` + gates green -> task `succeeded`
- `status=blocked` -> task `blocked`
- `status=failed` -> task `failed`
- `status=retryable` -> task back to `pending` if attempts remain
- `status=in_progress` -> usually task `succeeded` for that slice, then leader creates follow-up tasks if needed

## Crash recovery

On leader startup:

1. load `run.json`
2. scan workers
3. scan results
4. for each task with `status=assigned|running`:
   - if matching worker pid is alive, keep it running
   - if pid is gone and result exists, fold result
   - if pid is gone and no result exists, requeue as `pending` if attempts remain, else `failed`
5. recompute `task_counts`
6. continue scheduling

This makes the protocol durable without needing a daemon.

## Why this is minimal enough now

- no distributed queue
- no leases across machines
- no RPC requirement
- only file-based coordination plus subprocess pids
- single writer for scheduling state
- clean path to future `codex exec` workers: the worker binary only needs to translate assigned task input into the existing strict JSON output

## Recommended CLI shape

```text
ralphx run --task TASK.md --workdir REPO --max-parallel 4
ralphx worker --run-id RUN --worker-id worker-01 --task-id task-0001
ralphx ps --workdir REPO
ralphx resume --workdir REPO
```

## Suggested first implementation cut

1. leader process + `run.json`
2. task file schema + sequential execution with `max_parallel=1`
3. worker file + result file
4. process monitoring and heartbeats
5. parallel scheduling with fixed worker pool
6. validation/repair task types after the basic loop is stable
