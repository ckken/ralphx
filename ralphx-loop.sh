#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(pwd)"

CODEX_CMD="${CODEX_CMD:-codex}"
TASK_FILE=""
CHECKLIST_FILE="${CHECKLIST_FILE:-}"
PROMPT_FILE="${PROMPT_FILE:-$SCRIPT_DIR/prompts/loop-system-prompt.md}"
OUTPUT_SCHEMA_FILE="${OUTPUT_SCHEMA_FILE:-$SCRIPT_DIR/schemas/loop-output.schema.json}"
MAX_ITERATIONS="${MAX_ITERATIONS:-30}"
MAX_NO_PROGRESS="${MAX_NO_PROGRESS:-3}"
ROUND_TIMEOUT_SECONDS="${ROUND_TIMEOUT_SECONDS:-1800}"
CODEX_ARGS="${CODEX_ARGS:-}"
TESTS_CMD="${TESTS_CMD:-}"
WORKDIR="${WORKDIR:-$ROOT_DIR}"
STATE_DIR="${STATE_DIR:-$WORKDIR/.ralphx}"
LOG_DIR="$STATE_DIR/logs"
STATE_FILE="$STATE_DIR/state.json"
LAST_OUTPUT_FILE="$STATE_DIR/last-output.txt"
LAST_JSON_FILE="$STATE_DIR/last-result.json"
SUMMARY_FILE="$STATE_DIR/summary.txt"
STATS_FILE="$STATE_DIR/stats.json"

mkdir -p "$STATE_DIR" "$LOG_DIR"

usage() {
  cat <<'EOF'
Usage:
  ralphx-loop.sh --task FILE [--checklist FILE] [--workdir DIR]

Environment:
  CODEX_CMD            Codex CLI command, default: codex
  CODEX_ARGS           Extra Codex CLI args, space-separated
  CHECKLIST_FILE       Optional markdown checklist gate; defaults to TASK basename + .checklist.md if present
  PROMPT_FILE          System prompt template, default: prompts/loop-system-prompt.md
  TESTS_CMD            Optional test command run after each successful iteration
  MAX_ITERATIONS       Safety cap, default: 30; set to 0 for no limit
  MAX_NO_PROGRESS      Stop after N iterations with no file changes, default: 3; set to 0 for no limit
  ROUND_TIMEOUT_SECONDS Per-round timeout, default: 1800

Protocol:
  Codex must return a single JSON object with:
    status, exit_signal, files_modified, tests_passed, blockers, summary
EOF
}

die() {
  printf '%s\n' "$*" >&2
  exit 1
}

log() {
  printf '[%s] %s\n' "$(date '+%Y-%m-%d %H:%M:%S')" "$*"
}

trim_json() {
  jq -r 'if type == "object" then . else empty end' 2>/dev/null
}

read_task() {
  local task_file="$1"
  [[ -f "$task_file" ]] || die "Task file not found: $task_file"
  cat "$task_file"
}

get_git_status_snapshot() {
  if command -v git >/dev/null 2>&1 && git -C "$WORKDIR" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    git -C "$WORKDIR" status --short | LC_ALL=C sort
  else
    printf '\n'
  fi
}

resolve_checklist_file() {
  local task_file="$1"

  if [[ -n "$CHECKLIST_FILE" ]]; then
    [[ -f "$CHECKLIST_FILE" ]] || die "Checklist file not found: $CHECKLIST_FILE"
    printf '%s\n' "$CHECKLIST_FILE"
    return
  fi

  local auto_file="${task_file%.md}.checklist.md"
  if [[ -f "$auto_file" ]]; then
    printf '%s\n' "$auto_file"
    return
  fi

  printf '\n'
}

read_checklist() {
  local checklist_file="$1"
  [[ -f "$checklist_file" ]] || return 0
  cat "$checklist_file"
}

count_open_checklist_items() {
  local checklist_file="$1"
  [[ -f "$checklist_file" ]] || {
    printf '0\n'
    return
  }

  local count
  count="$(grep -cE '^[[:space:]]*[-*][[:space:]]+\[ \]' "$checklist_file" || true)"
  printf '%s\n' "${count:-0}"
}

build_prompt() {
  local iteration="$1"
  local task_content="$2"
  local checklist_file="$3"
  local template=""
  local previous_summary=""
  local current_state=""
  local current_diff=""
  local checklist_content=""
  local checklist_open_items=0

  if [[ -f "$PROMPT_FILE" ]]; then
    template="$(cat "$PROMPT_FILE")"
  fi
  if [[ -f "$SUMMARY_FILE" ]]; then
    previous_summary="$(cat "$SUMMARY_FILE")"
  fi
  if [[ -f "$STATE_FILE" ]]; then
    current_state="$(cat "$STATE_FILE")"
  fi
  if command -v git >/dev/null 2>&1 && git -C "$WORKDIR" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    current_diff="$(git -C "$WORKDIR" status --short)"
  fi
  if [[ -n "$checklist_file" ]]; then
    checklist_content="$(read_checklist "$checklist_file")"
    checklist_open_items="$(count_open_checklist_items "$checklist_file")"
  fi

  cat <<EOF
${template}

You are running inside an autonomous Bash loop.

Task:
$task_content

Iteration:
$iteration

Workspace:
$WORKDIR

Checklist file:
${checklist_file:-none}

Open checklist items:
$checklist_open_items

Checklist content:
$checklist_content

Previous summary:
$previous_summary

Current state:
$current_state

Current git status:
$current_diff

Rules:
- Make the smallest correct change.
- If a checklist file is provided, you must treat unchecked items as hard remaining work.
- Update the checklist file when you complete a milestone.
- If the task is not complete, return status="in_progress".
- If blocked, return status="blocked" and include blockers.
- If done, return status="complete" and set exit_signal=true.
- Always return exactly one JSON object and no extra text.
- Use this schema:
  {
    "status": "in_progress|blocked|complete",
    "exit_signal": true|false,
    "files_modified": 0,
    "tests_passed": true|false,
    "blockers": [],
    "summary": "short summary"
  }
EOF
}

run_codex() {
  local prompt="$1"
  local log_file="$2"
  local stdout_file="${log_file}.stdout"

  if [[ -n "$CODEX_ARGS" ]]; then
    # shellcheck disable=SC2206
    local args=( $CODEX_ARGS )
  else
    local args=()
  fi

  local -a cmd_args=()
  if [[ "$CODEX_CMD" == "codex" ]]; then
    [[ -f "$OUTPUT_SCHEMA_FILE" ]] || die "Output schema file not found: $OUTPUT_SCHEMA_FILE"
    cmd_args=(
      "$CODEX_CMD"
      exec
      --skip-git-repo-check
      --dangerously-bypass-approvals-and-sandbox
      -C "$WORKDIR"
      --output-schema "$OUTPUT_SCHEMA_FILE"
      -o "$log_file"
      -
    )
  else
    cmd_args=("$CODEX_CMD")
  fi

  cmd_args+=("${args[@]}")

  if command -v timeout >/dev/null 2>&1; then
    timeout "$ROUND_TIMEOUT_SECONDS" "${cmd_args[@]}" <<<"$prompt" >"$stdout_file" 2>&1
  elif command -v gtimeout >/dev/null 2>&1; then
    gtimeout "$ROUND_TIMEOUT_SECONDS" "${cmd_args[@]}" <<<"$prompt" >"$stdout_file" 2>&1
  else
    "${cmd_args[@]}" <<<"$prompt" >"$stdout_file" 2>&1
  fi

  if [[ "$CODEX_CMD" != "codex" ]]; then
    cp "$stdout_file" "$log_file"
  elif [[ ! -s "$log_file" ]]; then
    cp "$stdout_file" "$log_file"
  fi
}

extract_json() {
  local file="$1"
  python3 - "$file" <<'PY'
import json, re, sys
from pathlib import Path

text = Path(sys.argv[1]).read_text(errors="ignore")
decoder = json.JSONDecoder()
for match in re.finditer(r'[\{\[]', text):
    try:
        obj, end = decoder.raw_decode(text[match.start():])
        if isinstance(obj, dict):
            print(json.dumps(obj))
            raise SystemExit(0)
    except Exception:
        continue
raise SystemExit(1)
PY
}

record_state() {
  local iteration="$1"
  local payload="$2"
  jq -n \
    --argjson iteration "$iteration" \
    --argjson payload "$payload" \
    --arg ts "$(date '+%Y-%m-%d %H:%M:%S')" \
    '{iteration:$iteration, updated_at:$ts, result:$payload}' > "$STATE_FILE"
}

update_stats() {
  local start_epoch="$1"
  local start_time="$2"
  local iteration="$3"
  local round_seconds="$4"
  local last_status="$5"
  local last_exit_signal="$6"
  local last_files_modified="$7"

  local now_epoch elapsed_seconds avg_round_seconds
  now_epoch="$(date +%s)"
  elapsed_seconds=$((now_epoch - start_epoch))

  if [[ "$iteration" -gt 0 ]]; then
    avg_round_seconds=$((elapsed_seconds / iteration))
  else
    avg_round_seconds=0
  fi

  jq -n \
    --argjson loops_completed "$iteration" \
    --argjson total_elapsed_seconds "$elapsed_seconds" \
    --argjson last_round_seconds "$round_seconds" \
    --argjson average_round_seconds "$avg_round_seconds" \
    --arg last_status "$last_status" \
    --argjson last_exit_signal "$last_exit_signal" \
    --argjson last_files_modified "$last_files_modified" \
    --arg started_at "$start_time" \
    --arg updated_at "$(date '+%Y-%m-%d %H:%M:%S')" \
    '{
      started_at: $started_at,
      updated_at: $updated_at,
      loops_completed: $loops_completed,
      total_elapsed_seconds: $total_elapsed_seconds,
      last_round_seconds: $last_round_seconds,
      average_round_seconds: $average_round_seconds,
      last_status: $last_status,
      last_exit_signal: $last_exit_signal,
      last_files_modified: $last_files_modified
    }' > "$STATS_FILE"
}

run_tests() {
  local tests_cmd="$1"
  local test_log="$2"

  if [[ -z "$tests_cmd" ]]; then
    return 0
  fi

  log "Running tests: $tests_cmd"
  if bash -lc "$tests_cmd" >"$test_log" 2>&1; then
    return 0
  fi

  return 1
}

main() {
  local task_arg=""
  local checklist_arg=""

  while [[ $# -gt 0 ]]; do
    case "$1" in
      --task)
        task_arg="${2:-}"
        shift 2
        ;;
      --checklist)
        checklist_arg="${2:-}"
        CHECKLIST_FILE="$checklist_arg"
        shift 2
        ;;
      --workdir)
        WORKDIR="${2:-}"
        STATE_DIR="$WORKDIR/.ralphx"
        LOG_DIR="$STATE_DIR/logs"
        STATE_FILE="$STATE_DIR/state.json"
        LAST_OUTPUT_FILE="$STATE_DIR/last-output.txt"
        LAST_JSON_FILE="$STATE_DIR/last-result.json"
        SUMMARY_FILE="$STATE_DIR/summary.txt"
        mkdir -p "$STATE_DIR" "$LOG_DIR"
        shift 2
        ;;
      -h|--help)
        usage
        exit 0
        ;;
      *)
        die "Unknown argument: $1"
        ;;
    esac
  done

  if [[ -z "$task_arg" ]]; then
    usage
    exit 0
  fi
  [[ -f "$PROMPT_FILE" ]] || die "Prompt file not found: $PROMPT_FILE"

  local task_content
  task_content="$(read_task "$task_arg")"
  local checklist_file
  checklist_file="$(resolve_checklist_file "$task_arg")"
  local iteration=1
  local no_progress=0
  local last_files_modified=-1
  local start_epoch
  start_epoch="$(date +%s)"
  local start_time
  start_time="$(date '+%Y-%m-%d %H:%M:%S')"
  local last_round_seconds=0
  local last_status="not_started"
  local last_exit_signal=false

  log "Starting codex loop in $WORKDIR"
  log "Task: $task_arg"

  while true; do
    if [[ "$MAX_ITERATIONS" -gt 0 && "$iteration" -gt "$MAX_ITERATIONS" ]]; then
      log "Stopping after reaching MAX_ITERATIONS=$MAX_ITERATIONS"
      break
    fi

    local prompt log_file raw_output json_output files_modified exit_signal status tests_passed summary blockers_count
    local pre_round_status post_round_status forced_continue=false
    local checklist_open_items_before=0 checklist_open_items_after=0
    log_file="$LOG_DIR/round-${iteration}.log"
    raw_output="$STATE_DIR/round-${iteration}.txt"

    prompt="$(build_prompt "$iteration" "$task_content" "$checklist_file")"
    pre_round_status="$(get_git_status_snapshot)"
    if [[ -n "$checklist_file" ]]; then
      checklist_open_items_before="$(count_open_checklist_items "$checklist_file")"
    fi

    log "Round $iteration: invoking Codex"
    local round_start_epoch round_end_epoch
    round_start_epoch="$(date +%s)"
    if ! run_codex "$prompt" "$raw_output"; then
      log "Codex command failed or timed out; see $raw_output"
    fi
    round_end_epoch="$(date +%s)"
    last_round_seconds=$((round_end_epoch - round_start_epoch))

    cp "$raw_output" "$LAST_OUTPUT_FILE"

    if ! json_output="$(extract_json "$raw_output")"; then
      log "Could not parse JSON from Codex output"
      jq -n \
        --arg status "blocked" \
        --arg summary "Codex did not return a JSON object" \
        '{status:$status, exit_signal:false, files_modified:0, tests_passed:false, blockers:["invalid_json"], summary:$summary}' \
        > "$LAST_JSON_FILE"
      record_state "$iteration" "$(cat "$LAST_JSON_FILE")"
      break
    fi

    printf '%s\n' "$json_output" > "$LAST_JSON_FILE"

    status="$(jq -r '.status // "blocked"' "$LAST_JSON_FILE")"
    exit_signal="$(jq -r '.exit_signal // false' "$LAST_JSON_FILE")"
    files_modified="$(jq -r '.files_modified // 0' "$LAST_JSON_FILE")"
    tests_passed="$(jq -r '.tests_passed // false' "$LAST_JSON_FILE")"
    summary="$(jq -r '.summary // ""' "$LAST_JSON_FILE")"
    blockers_count="$(jq -r '.blockers // [] | length' "$LAST_JSON_FILE")"
    post_round_status="$(get_git_status_snapshot)"
    if [[ -n "$checklist_file" ]]; then
      checklist_open_items_after="$(count_open_checklist_items "$checklist_file")"
    fi

    if [[ "$status" == "complete" && "$exit_signal" == "true" && "$files_modified" -le 0 && "$pre_round_status" == "$post_round_status" ]]; then
      forced_continue=true
      status="in_progress"
      exit_signal="false"
      tests_passed="false"
      summary="Ignored premature completion because no new changes were detected. ${summary}"
      jq -n \
        --arg status "$status" \
        --arg summary "$summary" \
        '{status:$status, exit_signal:false, files_modified:0, tests_passed:false, blockers:[], summary:$summary}' \
        > "$LAST_JSON_FILE"
      blockers_count=0
      log "Ignoring premature completion with no detected changes"
    fi

    if [[ "$status" == "complete" && "$exit_signal" == "true" && "$checklist_open_items_after" -gt 0 ]]; then
      forced_continue=true
      status="in_progress"
      exit_signal="false"
      tests_passed="false"
      summary="Ignored premature completion because checklist still has ${checklist_open_items_after} open items. ${summary}"
      jq -n \
        --arg status "$status" \
        --arg summary "$summary" \
        '{status:$status, exit_signal:false, files_modified:0, tests_passed:false, blockers:[], summary:$summary}' \
        > "$LAST_JSON_FILE"
      blockers_count=0
      log "Ignoring completion because checklist still has open items (${checklist_open_items_before} -> ${checklist_open_items_after})"
    fi

    last_status="$status"
    last_exit_signal="$exit_signal"

    printf '%s\n' "$summary" > "$SUMMARY_FILE"
    record_state "$iteration" "$(cat "$LAST_JSON_FILE")"
    update_stats "$start_epoch" "$start_time" "$iteration" "$last_round_seconds" "$last_status" "$last_exit_signal" "$files_modified"

    log "Result: status=$status exit_signal=$exit_signal files_modified=$files_modified tests_passed=$tests_passed blockers=$blockers_count"

    if [[ -n "$TESTS_CMD" && "$forced_continue" != "true" ]]; then
      local test_log="$LOG_DIR/tests-${iteration}.log"
      if ! run_tests "$TESTS_CMD" "$test_log"; then
        jq -n \
          --arg status "blocked" \
          --arg summary "Tests failed" \
          '{status:$status, exit_signal:false, files_modified:0, tests_passed:false, blockers:["tests_failed"], summary:$summary}' \
          > "$LAST_JSON_FILE"
        record_state "$iteration" "$(cat "$LAST_JSON_FILE")"
        log "Tests failed; see $test_log"
        break
      fi
    fi

    if [[ "$exit_signal" == "true" && "$status" == "complete" ]]; then
      log "Task complete"
      break
    fi

    if [[ "$status" == "blocked" ]]; then
      log "Codex reported blockers"
      break
    fi

    if [[ "$files_modified" -gt 0 ]]; then
      no_progress=0
    else
      no_progress=$((no_progress + 1))
    fi

    if [[ "$MAX_NO_PROGRESS" -gt 0 && "$no_progress" -ge "$MAX_NO_PROGRESS" ]]; then
      log "Stopping after $no_progress no-progress rounds"
      break
    fi

    if [[ "$files_modified" -eq "$last_files_modified" && "$files_modified" -eq 0 ]]; then
      :
    fi
    last_files_modified="$files_modified"
    iteration=$((iteration + 1))
  done

  log "Finished. State: $STATE_FILE"
  if [[ -f "$STATS_FILE" ]]; then
    log "Stats: $(cat "$STATS_FILE")"
  fi
}

main "$@"
