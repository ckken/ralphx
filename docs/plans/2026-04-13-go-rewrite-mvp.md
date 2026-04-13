# ralphx Go Rewrite MVP Plan

> For Hermes: execute this plan with subagent-driven-development where tasks are independent, but keep shared-file edits serialized.

Goal: start the Go rewrite in-place without breaking the current Bash fallback, and land a compiling single-binary CLI with a minimal multi-agent-ready architecture.

Architecture: keep the current repo and assets, add a Go front door now, preserve `.ralphx` as the durable local state contract, and introduce clean seams for `Agent`, `Runner`, and future `Scheduler`/`Worker`. The first Go milestone should compile, expose `run`, `doctor`, and `version`, and delegate `run` to the legacy Bash loop while the Go-native runtime grows behind the same CLI.

Tech Stack: Go 1.19+, stdlib-first, local filesystem state, shell fallback for legacy execution, JSON-first contracts.

---

## MVP scope

1. Add `go.mod` and a compiling `cmd/ralphx` binary.
2. Add a tiny `cmd/ralphx-doctor` compatible binary.
3. Implement minimal internal packages for CLI dispatch, config, version, doctor, and legacy execution.
4. Preserve current shell runtime and assets temporarily.
5. Add initial Go-native state/contract types for the upcoming runner and worker protocol.
6. Add initial scheduler/worker scaffolding without enabling true parallel execution yet.
7. Verify build, help output, doctor output, and legacy delegation.

## Non-goals for this pass

- Full parity rewrite of `ralphx-loop.sh`
- TUI
- Network services / daemons
- Plugin system
- Rich planner or task decomposition logic
- True parallel worker execution

## Target file layout for this pass

- `go.mod`
- `cmd/ralphx/main.go`
- `cmd/ralphx-doctor/main.go`
- `internal/cli/app.go`
- `internal/config/config.go`
- `internal/version/version.go`
- `internal/doctor/doctor.go`
- `internal/legacy/exec.go`
- `internal/contracts/result.go`
- `internal/state/types.go`
- `internal/parallel/types.go`
- `docs/plans/2026-04-13-go-rewrite-mvp.md`

## Task 1: Bootstrap the Go module

Objective: create a compiling Go module rooted at the final GitHub path.

Files:
- Create: `go.mod`
- Create: `cmd/ralphx/main.go`
- Create: `cmd/ralphx-doctor/main.go`

Steps:
1. Create `go.mod` with module path `github.com/ckken/ralphx`.
2. Add a minimal `main.go` for `ralphx` calling into `internal/cli`.
3. Add a minimal `main.go` for `ralphx-doctor` calling into `internal/doctor`.
4. Verify with:
   - `go build ./cmd/ralphx`
   - `go build ./cmd/ralphx-doctor`

## Task 2: Add CLI/config/version plumbing

Objective: give the binary a stable command surface and environment/flag handling.

Files:
- Create: `internal/cli/app.go`
- Create: `internal/config/config.go`
- Create: `internal/version/version.go`

Steps:
1. Define the initial commands:
   - `ralphx run`
   - `ralphx doctor`
   - `ralphx version`
2. Make bare `ralphx --task ...` map to `run` for compatibility.
3. Support key env vars from the legacy Bash flow.
4. Add a simple build info/version string.
5. Verify with:
   - `go run ./cmd/ralphx --help`
   - `go run ./cmd/ralphx version`

## Task 3: Add doctor and legacy execution bridge

Objective: get a useful binary into users’ hands immediately, even before Go parity exists.

Files:
- Create: `internal/doctor/doctor.go`
- Create: `internal/legacy/exec.go`

Steps:
1. Implement Go-native doctor checks for `bash`, `python3`, `git`, `gh`, `codex`, optional `jq`.
2. Add legacy script execution helpers that can locate repo-root scripts and run them with passthrough args.
3. Make `ralphx run ...` delegate to `./ralphx-loop.sh` for now.
4. Verify with:
   - `go run ./cmd/ralphx-doctor`
   - `go run ./cmd/ralphx run --help` or argument validation
   - `go run ./cmd/ralphx --task ./examples/sample-task.md --workdir .` (smoke path)

## Task 4: Add Go-native contracts and state types

Objective: land the stable types that the Bash runtime currently implies.

Files:
- Create: `internal/contracts/result.go`
- Create: `internal/state/types.go`

Steps:
1. Define `RoundResult` matching the current JSON schema.
2. Define `RunState` and `Stats` types matching current `.ralphx` files closely enough for a low-risk migration.
3. Add JSON tags and helper constructors/validators as needed.
4. Verify with `go test ./...` once tests exist and `go build ./...` now.

## Task 5: Add multi-agent foundation types

Objective: make parallelism a first-class future mode without building the whole scheduler now.

Files:
- Create: `internal/parallel/types.go`

Steps:
1. Define `Job`, `WorkerState`, `WorkerResult`, and `Scheduler`/`Worker` interfaces.
2. Keep these types local-only and file-based.
3. Ensure the design preserves:
   - leader owns completion
   - workers only own bounded slices
   - result files are append/replace-safe per worker
4. Verify by compiling the package and referencing types from the CLI package where appropriate.

## Task 6: Verify and document the bootstrap state

Objective: confirm the Go front door works and the repo has a clear migration direction.

Files:
- Modify later if needed: `README.md`

Steps:
1. Run:
   - `gofmt -w ./cmd ./internal`
   - `go build ./...`
   - `go run ./cmd/ralphx --help`
   - `go run ./cmd/ralphx version`
   - `go run ./cmd/ralphx-doctor`
2. Capture any gaps.
3. Keep docs updates minimal until the Go runtime replaces the Bash core.

## Acceptance criteria for this pass

- The repo contains a valid Go module.
- `ralphx` and `ralphx-doctor` compile.
- `ralphx doctor` works natively in Go.
- `ralphx run` exists and can bridge to the legacy Bash runtime.
- Core Go contract/state/parallel types are present for the next implementation pass.
- The current Bash implementation remains available as a fallback.
