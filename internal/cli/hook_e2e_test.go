package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ckken/ralphx/internal/hooks"
	"github.com/ckken/ralphx/internal/state"
)

func TestHookFlowEndToEnd(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	workdir := t.TempDir()
	stateDir := filepath.Join(workdir, ".ralphx")
	repoLatestPath := filepath.Join(stateDir, "last-hook-event.json")
	userLatestPath := filepath.Join(home, ".codex", "log", "ralphx-last-hook-event.json")
	repoLogPath := filepath.Join(stateDir, "logs", "hooks-"+time.Now().Format("2006-01-02")+".jsonl")

	stdout, stderr := captureCLIOutput(t, func() int {
		return hookMain([]string{"install"})
	})
	if !strings.Contains(stdout, "installed hooks:") {
		t.Fatalf("install stdout = %q", stdout)
	}
	if stderr != "" {
		t.Fatalf("install stderr = %q, want empty", stderr)
	}

	hooksPath := filepath.Join(home, ".codex", "hooks.json")
	hooksText := mustReadFile(t, hooksPath)
	if !strings.Contains(hooksText, "ralphx hook native --event Stop") {
		t.Fatalf("hooks.json missing managed stop hook: %s", hooksText)
	}
	if !strings.Contains(hooksText, "ralphx hook native --event UserPromptSubmit") {
		t.Fatalf("hooks.json missing managed prompt-submit hook: %s", hooksText)
	}

	activatePayloadPath := filepath.Join(workdir, "activate-payload.json")
	mustWriteJSON(t, activatePayloadPath, map[string]any{
		"hook_event_name": "UserPromptSubmit",
		"cwd":             workdir,
		"prompt":          "$ralphx",
	})
	stdout, stderr = captureCLIOutput(t, func() int {
		return hookMain([]string{
			"native",
			"--event", "UserPromptSubmit",
			"--workdir", workdir,
			"--payload", activatePayloadPath,
		})
	})
	if !strings.Contains(stderr, "[hook prompt-submit] mode=ralphx active (ralphx active)") {
		t.Fatalf("activate stderr = %q", stderr)
	}
	var activateOutput map[string]any
	if err := json.Unmarshal([]byte(stdout), &activateOutput); err != nil {
		t.Fatalf("unmarshal activate stdout: %v\noutput: %s", err, stdout)
	}
	hookSpecificOutput, ok := activateOutput["hookSpecificOutput"].(map[string]any)
	if !ok {
		t.Fatalf("activate hookSpecificOutput = %T", activateOutput["hookSpecificOutput"])
	}
	if got := hookSpecificOutput["hookEventName"]; got != string(hooks.NativeHookEventUserPromptSubmit) {
		t.Fatalf("activate hookEventName = %v", got)
	}

	activeState, err := hooks.ReadActiveState(stateDir)
	if err != nil {
		t.Fatalf("ReadActiveState() after activate error = %v", err)
	}
	if !activeState.Active || activeState.Mode != "ralphx" || activeState.StopHookActive {
		t.Fatalf("active state after activate = %#v", activeState)
	}
	repoLatest, err := hooks.ReadLatest(repoLatestPath)
	if err != nil {
		t.Fatalf("ReadLatest(repo) after activate error = %v", err)
	}
	if repoLatest.Event != hooks.EventPromptSubmit || repoLatest.Decision.Reason != "prompt_submit" {
		t.Fatalf("repo latest after activate = %#v", repoLatest)
	}
	userLatest, err := hooks.ReadLatest(userLatestPath)
	if err != nil {
		t.Fatalf("ReadLatest(user) after activate error = %v", err)
	}
	if userLatest.Event != hooks.EventPromptSubmit || userLatest.Decision.Reason != "prompt_submit" {
		t.Fatalf("user latest after activate = %#v", userLatest)
	}
	runState := mustLoadRunState(t, workdir, stateDir)
	if runState.Hook == nil || runState.Hook.Event != string(hooks.EventPromptSubmit) || runState.Hook.Reason != "prompt_submit" {
		t.Fatalf("run state after activate = %#v", runState.Hook)
	}
	if got := jsonlLineCount(t, repoLogPath); got != 1 {
		t.Fatalf("repo log line count after activate = %d, want 1", got)
	}

	stdout, stderr = captureCLIOutput(t, func() int {
		return hookMain([]string{
			"native",
			"--event", "Stop",
			"--workdir", workdir,
			"--native-json",
		})
	})
	if !strings.Contains(stderr, "[hook stop-guard] block (active_workflow) $ralphx") {
		t.Fatalf("stop stderr = %q", stderr)
	}
	var stopOutput map[string]any
	if err := json.Unmarshal([]byte(stdout), &stopOutput); err != nil {
		t.Fatalf("unmarshal stop stdout: %v\noutput: %s", err, stdout)
	}
	if stopOutput["decision"] != "block" || stopOutput["reason"] != "active_workflow" {
		t.Fatalf("stop output = %#v", stopOutput)
	}

	activeState, err = hooks.ReadActiveState(stateDir)
	if err != nil {
		t.Fatalf("ReadActiveState() after stop error = %v", err)
	}
	if !activeState.StopHookActive || activeState.StopReason != "active_workflow" {
		t.Fatalf("active state after stop = %#v", activeState)
	}
	repoLatest, err = hooks.ReadLatest(repoLatestPath)
	if err != nil {
		t.Fatalf("ReadLatest(repo) after stop error = %v", err)
	}
	if repoLatest.Event != hooks.EventStop || repoLatest.Decision.Reason != "active_workflow" {
		t.Fatalf("repo latest after stop = %#v", repoLatest)
	}
	runState = mustLoadRunState(t, workdir, stateDir)
	if runState.Hook == nil || runState.Hook.Event != string(hooks.EventStop) || runState.Hook.Reason != "active_workflow" {
		t.Fatalf("run state after stop = %#v", runState.Hook)
	}
	if got := jsonlLineCount(t, repoLogPath); got != 2 {
		t.Fatalf("repo log line count after first stop = %d, want 2", got)
	}

	stdout, stderr = captureCLIOutput(t, func() int {
		return hookMain([]string{
			"native",
			"--event", "Stop",
			"--workdir", workdir,
			"--native-json",
		})
	})
	if stdout != "" || stderr != "" {
		t.Fatalf("repeat stop stdout/stderr = %q / %q, want both empty", stdout, stderr)
	}
	if got := jsonlLineCount(t, repoLogPath); got != 2 {
		t.Fatalf("repo log line count after repeat stop = %d, want 2", got)
	}

	stdout, stderr = captureCLIOutput(t, func() int {
		return hookMain([]string{"status", "--workdir", workdir})
	})
	if !strings.Contains(stderr, "[hook status] active=true") {
		t.Fatalf("status stderr missing active summary: %q", stderr)
	}
	if !strings.Contains(stderr, "repo=stop repoReason=active_workflow") {
		t.Fatalf("status stderr missing repo summary: %q", stderr)
	}
	if !strings.Contains(stderr, "user=stop userReason=active_workflow") {
		t.Fatalf("status stderr missing user summary: %q", stderr)
	}
	var statusOutput map[string]any
	if err := json.Unmarshal([]byte(stdout), &statusOutput); err != nil {
		t.Fatalf("unmarshal status stdout: %v\noutput: %s", err, stdout)
	}
	activeMap, ok := statusOutput["active"].(map[string]any)
	if !ok {
		t.Fatalf("status active = %T", statusOutput["active"])
	}
	if activeMap["stop_hook_active"] != true {
		t.Fatalf("status active.stop_hook_active = %v", activeMap["stop_hook_active"])
	}

	stopPayloadPath := filepath.Join(workdir, "stop-payload.json")
	mustWriteJSON(t, stopPayloadPath, map[string]any{
		"hook_event_name": "UserPromptSubmit",
		"cwd":             workdir,
		"prompt":          "stop",
	})
	stdout, stderr = captureCLIOutput(t, func() int {
		return hookMain([]string{
			"native",
			"--event", "UserPromptSubmit",
			"--workdir", workdir,
			"--payload", stopPayloadPath,
		})
	})
	if !strings.Contains(stderr, "[hook prompt-submit] ralphx workflow stopped") {
		t.Fatalf("stop prompt stderr = %q", stderr)
	}
	if _, err := os.Stat(filepath.Join(stateDir, "ralphx-active.json")); !os.IsNotExist(err) {
		t.Fatalf("expected active state cleared, err=%v", err)
	}
	repoLatest, err = hooks.ReadLatest(repoLatestPath)
	if err != nil {
		t.Fatalf("ReadLatest(repo) after prompt stop error = %v", err)
	}
	if repoLatest.Event != hooks.EventPromptSubmit || repoLatest.Decision.Reason != "prompt_stop" {
		t.Fatalf("repo latest after prompt stop = %#v", repoLatest)
	}
	runState = mustLoadRunState(t, workdir, stateDir)
	if runState.Hook == nil || runState.Hook.Event != string(hooks.EventPromptSubmit) || runState.Hook.Reason != "prompt_stop" {
		t.Fatalf("run state after prompt stop = %#v", runState.Hook)
	}
	if got := jsonlLineCount(t, repoLogPath); got != 3 {
		t.Fatalf("repo log line count after prompt stop = %d, want 3", got)
	}

	stdout, stderr = captureCLIOutput(t, func() int {
		return hookMain([]string{"uninstall"})
	})
	if !strings.Contains(stdout, "removed managed hooks from:") {
		t.Fatalf("uninstall stdout = %q", stdout)
	}
	if stderr != "" {
		t.Fatalf("uninstall stderr = %q, want empty", stderr)
	}
	hooksText = mustReadFile(t, hooksPath)
	if strings.Contains(hooksText, "ralphx hook native --event Stop") {
		t.Fatalf("hooks.json still contains managed stop hook after uninstall: %s", hooksText)
	}
	if strings.Contains(hooksText, "ralphx hook native --event UserPromptSubmit") {
		t.Fatalf("hooks.json still contains managed prompt-submit hook after uninstall: %s", hooksText)
	}
}

func mustWriteJSON(t *testing.T, path string, value any) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatalf("marshal %s: %v", path, err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func mustReadFile(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

func mustLoadRunState(t *testing.T, workdir, stateDir string) state.RunState {
	t.Helper()

	runState, err := state.LoadRunState(state.DerivePaths(workdir, stateDir))
	if err != nil {
		t.Fatalf("LoadRunState(%s) error = %v", stateDir, err)
	}
	return runState
}

func jsonlLineCount(t *testing.T, path string) int {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read jsonl %s: %v", path, err)
	}
	return strings.Count(string(data), "\n")
}
