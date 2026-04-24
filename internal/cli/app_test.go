package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ckken/ralphx/internal/hooks"
)

func TestHookStatusIncludesActiveState(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	workdir := t.TempDir()
	stateDir := filepath.Join(workdir, ".ralphx")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatalf("mkdir state dir: %v", err)
	}
	if err := hooks.WriteActiveState(stateDir, "$ralphx continue"); err != nil {
		t.Fatalf("WriteActiveState() error = %v", err)
	}

	stdout, stderr := captureCLIOutput(t, func() int {
		return hookStatus([]string{"--workdir", workdir})
	})
	if !strings.Contains(stderr, "[hook status] active=true") {
		t.Fatalf("stderr = %q, want active summary", stderr)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("unmarshal stdout: %v\noutput: %s", err, stdout)
	}
	if _, ok := got["active"].(map[string]any); !ok {
		t.Fatalf("active = %T, want map[string]any; output: %s", got["active"], stdout)
	}
}

func TestHookStatusIncludesRepoAndUserSummaries(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	workdir := t.TempDir()
	stateDir := filepath.Join(workdir, ".ralphx")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatalf("mkdir state dir: %v", err)
	}
	if err := hooks.WriteLatest(filepath.Join(stateDir, "last-hook-event.json"), hooks.LogEntry{
		Event: hooks.EventStop,
		Decision: hooks.Decision{
			Allow:   false,
			Reason:  "task_incomplete",
			Message: "Continue the current branch of work.",
		},
	}); err != nil {
		t.Fatalf("WriteLatest(repo) error = %v", err)
	}
	if err := hooks.WriteLatest(filepath.Join(home, ".codex", "log", "ralphx-last-hook-event.json"), hooks.LogEntry{
		Event: hooks.EventPromptSubmit,
		Decision: hooks.Decision{
			Allow:   true,
			Reason:  "prompt_submit",
			Message: "ralphx mode active",
		},
	}); err != nil {
		t.Fatalf("WriteLatest(user) error = %v", err)
	}

	stdout, stderr := captureCLIOutput(t, func() int {
		return hookStatus([]string{"--workdir", workdir})
	})
	if !strings.Contains(stderr, "[hook status] repo=stop repoReason=task_incomplete") {
		t.Fatalf("stderr = %q, want repo summary", stderr)
	}
	if !strings.Contains(stderr, "user=prompt-submit userReason=prompt_submit") {
		t.Fatalf("stderr = %q, want user summary", stderr)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("unmarshal stdout: %v\noutput: %s", err, stdout)
	}
	if _, ok := got["repo"].(map[string]any); !ok {
		t.Fatalf("repo = %T, want map[string]any; output: %s", got["repo"], stdout)
	}
	if _, ok := got["user"].(map[string]any); !ok {
		t.Fatalf("user = %T, want map[string]any; output: %s", got["user"], stdout)
	}
}

func TestHookStatusIncludesInstalledHooksWithoutRuntimeEvents(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	workdir := t.TempDir()
	if _, err := hooks.InstallUserStopHook(); err != nil {
		t.Fatalf("InstallUserStopHook() error = %v", err)
	}

	stdout, stderr := captureCLIOutput(t, func() int {
		return hookStatus([]string{"--workdir", workdir})
	})
	if !strings.Contains(stderr, "[hook status] installed=true stopHook=true promptHook=true") {
		t.Fatalf("stderr = %q, want installed summary", stderr)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("unmarshal stdout: %v\noutput: %s", err, stdout)
	}
	installed, ok := got["installed"].(map[string]any)
	if !ok {
		t.Fatalf("installed = %T, want map[string]any; output: %s", got["installed"], stdout)
	}
	if installed["managed_installed"] != true {
		t.Fatalf("managed_installed = %v, want true", installed["managed_installed"])
	}
}

func TestEmitNativeHookResponseKeepsActiveWorkflowStdoutOnly(t *testing.T) {
	var exitCode int
	stdout, stderr := captureCLIOutput(t, func() int {
		exitCode = emitNativeHookResponse(hooks.NativeHookResponse{
			ExitCode: 3,
			Decision: hooks.Decision{
				Allow:   false,
				Reason:  "active_workflow",
				Message: "ralphx workflow is still active in this workspace.",
			},
			Stdout: map[string]any{
				"decision": "block",
				"reason":   "active_workflow",
			},
		}, true)
		return exitCode
	})
	if exitCode != 3 {
		t.Fatalf("exitCode = %d, want 3", exitCode)
	}
	if !strings.Contains(stdout, `"reason": "active_workflow"`) {
		t.Fatalf("stdout = %q, want reason JSON", stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
}

func TestExecutionConfigKeepsAutoReplanDefault(t *testing.T) {
	t.Setenv("RALPHX_AUTO_REPLAN", "")
	cfg := executionConfig("task.md", "task.checklist.md", "/work", "/work/.ralphx", "go test ./...", "codex", []string{"--model", "gpt-5.5"})
	if !cfg.AutoReplan {
		t.Fatalf("AutoReplan = false, want true")
	}
	if cfg.TaskFile != "task.md" || cfg.ChecklistFile != "task.checklist.md" || cfg.Workdir != "/work" {
		t.Fatalf("unexpected execution config: %#v", cfg)
	}
	if got := strings.Join(cfg.CodexArgs, " "); got != "--model gpt-5.5" {
		t.Fatalf("CodexArgs = %q", got)
	}
}

func TestExecutionConfigCanDisableAutoReplan(t *testing.T) {
	t.Setenv("RALPHX_AUTO_REPLAN", "0")
	cfg := executionConfig("task.md", "", "/work", "/work/.ralphx", "", "codex", nil)
	if cfg.AutoReplan {
		t.Fatalf("AutoReplan = true, want false")
	}
}

func captureCLIOutput(t *testing.T, fn func() int) (stdout string, stderr string) {
	t.Helper()

	origStdout := os.Stdout
	origStderr := os.Stderr
	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe stdout: %v", err)
	}
	stderrR, stderrW, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe stderr: %v", err)
	}
	os.Stdout = stdoutW
	os.Stderr = stderrW
	defer func() {
		os.Stdout = origStdout
		os.Stderr = origStderr
	}()

	exitCode := fn()
	_ = exitCode
	_ = stdoutW.Close()
	_ = stderrW.Close()

	var outBuf, errBuf bytes.Buffer
	_, _ = outBuf.ReadFrom(stdoutR)
	_, _ = errBuf.ReadFrom(stderrR)

	return outBuf.String(), errBuf.String()
}
