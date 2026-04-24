package hooks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ckken/ralphx/internal/contracts"
	"github.com/ckken/ralphx/internal/state"
)

func TestNormalizeNativeHookEvent(t *testing.T) {
	cases := []struct {
		raw  string
		want NativeHookEvent
		ok   bool
	}{
		{raw: "UserPromptSubmit", want: NativeHookEventUserPromptSubmit, ok: true},
		{raw: "prompt-submit", want: NativeHookEventUserPromptSubmit, ok: true},
		{raw: "Stop", want: NativeHookEventStop, ok: true},
		{raw: "stop-guard", want: NativeHookEventStop, ok: true},
		{raw: "unknown", ok: false},
	}
	for _, tc := range cases {
		got, ok := NormalizeNativeHookEvent(tc.raw)
		if ok != tc.ok {
			t.Fatalf("NormalizeNativeHookEvent(%q) ok = %v, want %v", tc.raw, ok, tc.ok)
		}
		if got != tc.want {
			t.Fatalf("NormalizeNativeHookEvent(%q) = %q, want %q", tc.raw, got, tc.want)
		}
	}
}

func TestDispatchNativeHookPromptSubmitActivatesRalphx(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	workdir := t.TempDir()
	payloadPath := filepath.Join(workdir, "payload.json")
	mustWriteJSON(t, payloadPath, map[string]any{
		"hook_event_name": "UserPromptSubmit",
		"cwd":             workdir,
		"prompt":          "$ralphx",
	})

	resp, err := DispatchNativeHook(NativeHookRequest{
		Event:       NativeHookEventUserPromptSubmit,
		Workdir:     workdir,
		PayloadPath: payloadPath,
	})
	if err != nil {
		t.Fatalf("DispatchNativeHook() error = %v", err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("ExitCode = %d, want 0", resp.ExitCode)
	}
	if resp.Decision.Reason != "prompt_submit" {
		t.Fatalf("Decision.Reason = %q", resp.Decision.Reason)
	}
	if !strings.Contains(resp.Stderr, "[hook prompt-submit] mode=ralphx active (ralphx active)") {
		t.Fatalf("Stderr = %q", resp.Stderr)
	}
	output, ok := resp.Stdout.(map[string]any)
	if !ok {
		t.Fatalf("Stdout type = %T, want map[string]any", resp.Stdout)
	}
	hookSpecificOutput, ok := output["hookSpecificOutput"].(map[string]any)
	if !ok {
		t.Fatalf("hookSpecificOutput type = %T", output["hookSpecificOutput"])
	}
	if got := hookSpecificOutput["hookEventName"]; got != string(NativeHookEventUserPromptSubmit) {
		t.Fatalf("hookEventName = %v, want %q", got, NativeHookEventUserPromptSubmit)
	}
	if got := hookSpecificOutput["additionalContext"]; got != "ralphx active." {
		t.Fatalf("additionalContext = %v", got)
	}
	if _, err := os.Stat(filepath.Join(workdir, ".ralphx", "ralphx-active.json")); err != nil {
		t.Fatalf("expected active state file: %v", err)
	}
	runState, err := state.LoadRunState(state.DerivePaths(workdir, filepath.Join(workdir, ".ralphx")))
	if err != nil {
		t.Fatalf("LoadRunState() error = %v", err)
	}
	if runState.Hook == nil || runState.Hook.Event != string(EventPromptSubmit) {
		t.Fatalf("hook state = %#v", runState.Hook)
	}
	if runState.Hook.Reason != "prompt_submit" {
		t.Fatalf("hook reason = %q", runState.Hook.Reason)
	}
}

func TestDispatchNativeHookPromptSubmitPreservesActiveStateAcrossInactivePrompt(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	workdir := t.TempDir()
	stateDir := filepath.Join(workdir, ".ralphx")
	payloadPath := filepath.Join(workdir, "payload.json")

	mustWriteJSON(t, payloadPath, map[string]any{
		"hook_event_name": "UserPromptSubmit",
		"cwd":             workdir,
		"prompt":          "$ralphx",
	})
	if _, err := DispatchNativeHook(NativeHookRequest{
		Event:       NativeHookEventUserPromptSubmit,
		Workdir:     workdir,
		PayloadPath: payloadPath,
	}); err != nil {
		t.Fatalf("DispatchNativeHook(active) error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(stateDir, "ralphx-active.json")); err != nil {
		t.Fatalf("expected active state file: %v", err)
	}

	mustWriteJSON(t, payloadPath, map[string]any{
		"hook_event_name": "UserPromptSubmit",
		"cwd":             workdir,
		"prompt":          "doctor only",
	})
	resp, err := DispatchNativeHook(NativeHookRequest{
		Event:       NativeHookEventUserPromptSubmit,
		Workdir:     workdir,
		PayloadPath: payloadPath,
	})
	if err != nil {
		t.Fatalf("DispatchNativeHook(inactive) error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(stateDir, "ralphx-active.json")); err != nil {
		t.Fatalf("expected active state file preserved, err=%v", err)
	}
	if !resp.Silent {
		t.Fatalf("resp.Silent = false, want true")
	}
	if resp.Decision.Reason != "prompt_submit" {
		t.Fatalf("Decision.Reason = %q, want prompt_submit", resp.Decision.Reason)
	}
	if resp.Stderr != "" {
		t.Fatalf("Stderr = %q, want empty", resp.Stderr)
	}
}

func TestDispatchNativeHookPromptSubmitStopsRalphx(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	workdir := t.TempDir()
	stateDir := filepath.Join(workdir, ".ralphx")
	payloadPath := filepath.Join(workdir, "payload.json")

	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatalf("mkdir state dir: %v", err)
	}
	if err := WriteActiveState(stateDir, "$ralphx"); err != nil {
		t.Fatalf("WriteActiveState() error = %v", err)
	}
	mustWriteJSON(t, payloadPath, map[string]any{
		"hook_event_name": "UserPromptSubmit",
		"cwd":             workdir,
		"prompt":          "stop",
	})

	resp, err := DispatchNativeHook(NativeHookRequest{
		Event:       NativeHookEventUserPromptSubmit,
		Workdir:     workdir,
		PayloadPath: payloadPath,
	})
	if err != nil {
		t.Fatalf("DispatchNativeHook() error = %v", err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("ExitCode = %d, want 0", resp.ExitCode)
	}
	if resp.Decision.Reason != "prompt_stop" {
		t.Fatalf("Decision.Reason = %q, want prompt_stop", resp.Decision.Reason)
	}
	if _, err := os.Stat(filepath.Join(stateDir, "ralphx-active.json")); !os.IsNotExist(err) {
		t.Fatalf("expected active state file removed, err=%v", err)
	}
	if strings.Contains(resp.Stderr, "mode active") {
		t.Fatalf("Stderr = %q, want stop handling", resp.Stderr)
	}
	runState, err := state.LoadRunState(state.DerivePaths(workdir, stateDir))
	if err != nil {
		t.Fatalf("LoadRunState() error = %v", err)
	}
	if runState.Hook == nil || runState.Hook.Event != string(EventPromptSubmit) || runState.Hook.Reason != "prompt_stop" {
		t.Fatalf("hook state = %#v", runState.Hook)
	}
}

func TestDispatchNativeHookStopGuardUsesNativeBlockShape(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	workdir := t.TempDir()
	stateDir := filepath.Join(workdir, ".ralphx")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatalf("mkdir state dir: %v", err)
	}
	taskPath := filepath.Join(workdir, "task.md")
	checklistPath := filepath.Join(workdir, "task.checklist.md")
	summaryPath := filepath.Join(stateDir, "summary.txt")
	statePath := filepath.Join(stateDir, "state.json")
	lastResultPath := filepath.Join(stateDir, "last-result.json")

	mustWriteFile(t, taskPath, "# Task\n\nContinue.\n")
	mustWriteFile(t, checklistPath, "- [ ] first\n")
	mustWriteFile(t, summaryPath, "summary")
	mustWriteJSON(t, statePath, state.RunState{
		Iteration: 1,
		Result: contracts.RoundResult{
			Status: contracts.StatusInProgress,
			Mode:   contracts.ModeProducePlan,
		},
	})
	mustWriteJSON(t, lastResultPath, contracts.RoundResult{
		Status: contracts.StatusInProgress,
		Mode:   contracts.ModeProducePlan,
	})

	resp, err := DispatchNativeHook(NativeHookRequest{
		Event:          NativeHookEventStop,
		Workdir:        workdir,
		StateDir:       stateDir,
		TaskPath:       taskPath,
		ChecklistPath:  checklistPath,
		SummaryPath:    summaryPath,
		StatePath:      statePath,
		LastResultPath: lastResultPath,
		NativeJSON:     true,
	})
	if err != nil {
		t.Fatalf("DispatchNativeHook() error = %v", err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("ExitCode = %d, want 0", resp.ExitCode)
	}
	if !strings.Contains(resp.Stderr, "[hook stop-guard] block (task_incomplete)") {
		t.Fatalf("Stderr = %q", resp.Stderr)
	}
	output, ok := resp.Stdout.(map[string]any)
	if !ok {
		t.Fatalf("Stdout type = %T, want map[string]any", resp.Stdout)
	}
	if output["decision"] != "block" {
		t.Fatalf("decision = %v, want block", output["decision"])
	}
	if output["reason"] != "task_incomplete" {
		t.Fatalf("reason = %v, want task_incomplete", output["reason"])
	}
	if _, ok := output["feedback"]; ok {
		t.Fatalf("feedback field should be omitted: %v", output["feedback"])
	}
	if _, ok := output["systemMessage"]; ok {
		t.Fatalf("systemMessage field should be omitted: %v", output["systemMessage"])
	}
	runState, err := state.LoadRunState(state.DerivePaths(workdir, stateDir))
	if err != nil {
		t.Fatalf("LoadRunState() error = %v", err)
	}
	if runState.Hook == nil || runState.Hook.Event != string(EventStop) || runState.Hook.Reason != "task_incomplete" {
		t.Fatalf("hook state = %#v", runState.Hook)
	}
}

func TestDispatchNativeHookStopGuardBlocksAndMarksActiveWorkflow(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	workdir := t.TempDir()
	stateDir := filepath.Join(workdir, ".ralphx")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatalf("mkdir state dir: %v", err)
	}
	if err := WriteActiveState(stateDir, "$ralphx"); err != nil {
		t.Fatalf("WriteActiveState() error = %v", err)
	}

	resp, err := DispatchNativeHook(NativeHookRequest{
		Event:      NativeHookEventStop,
		Workdir:    workdir,
		StateDir:   stateDir,
		NativeJSON: true,
	})
	if err != nil {
		t.Fatalf("DispatchNativeHook() error = %v", err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("ExitCode = %d, want 0", resp.ExitCode)
	}
	if resp.Decision.Allow {
		t.Fatalf("Decision.Allow = true, want false")
	}
	if resp.Decision.Reason != "active_workflow" {
		t.Fatalf("Decision.Reason = %q, want active_workflow", resp.Decision.Reason)
	}
	if !strings.Contains(resp.Stderr, "[hook stop-guard] block (active_workflow)") {
		t.Fatalf("Stderr = %q, want block feedback", resp.Stderr)
	}
	marked, err := ReadActiveState(stateDir)
	if err != nil {
		t.Fatalf("ReadActiveState() error = %v", err)
	}
	if !marked.Active || !marked.StopHookActive || marked.StopReason != "active_workflow" {
		t.Fatalf("marked = %#v", marked)
	}
	output, ok := resp.Stdout.(map[string]any)
	if !ok {
		t.Fatalf("Stdout type = %T, want map[string]any", resp.Stdout)
	}
	if output["decision"] != "block" {
		t.Fatalf("decision = %v, want block", output["decision"])
	}
	if output["reason"] != "active_workflow" {
		t.Fatalf("reason = %v, want active_workflow", output["reason"])
	}
	if _, ok := output["feedback"]; ok {
		t.Fatalf("feedback field should be omitted: %v", output["feedback"])
	}

	repeatResp, err := DispatchNativeHook(NativeHookRequest{
		Event:      NativeHookEventStop,
		Workdir:    workdir,
		StateDir:   stateDir,
		NativeJSON: true,
	})
	if err != nil {
		t.Fatalf("DispatchNativeHook(repeat) error = %v", err)
	}
	if !repeatResp.Silent {
		t.Fatalf("repeatResp.Silent = false, want true")
	}
	if repeatResp.ExitCode != 0 {
		t.Fatalf("repeat ExitCode = %d, want 0", repeatResp.ExitCode)
	}
}

func TestDispatchNativeHookStopGuardClearsActiveStateWhenAllowed(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	workdir := t.TempDir()
	stateDir := filepath.Join(workdir, ".ralphx")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatalf("mkdir state dir: %v", err)
	}

	taskPath := filepath.Join(workdir, "task.md")
	checklistPath := filepath.Join(workdir, "task.checklist.md")
	summaryPath := filepath.Join(stateDir, "summary.txt")
	statePath := filepath.Join(stateDir, "state.json")
	lastResultPath := filepath.Join(stateDir, "last-result.json")

	mustWriteFile(t, taskPath, "# Task\n\nDone.\n")
	mustWriteFile(t, checklistPath, "- [x] first\n")
	mustWriteFile(t, summaryPath, "summary")
	mustWriteJSON(t, statePath, state.RunState{
		Iteration: 1,
		Result: contracts.RoundResult{
			Status:     contracts.StatusComplete,
			Mode:       contracts.ModeComplete,
			ExitSignal: true,
		},
	})
	mustWriteJSON(t, lastResultPath, contracts.RoundResult{
		Status:     contracts.StatusComplete,
		Mode:       contracts.ModeComplete,
		ExitSignal: true,
	})

	resp, err := DispatchNativeHook(NativeHookRequest{
		Event:          NativeHookEventStop,
		Workdir:        workdir,
		StateDir:       stateDir,
		TaskPath:       taskPath,
		ChecklistPath:  checklistPath,
		SummaryPath:    summaryPath,
		StatePath:      statePath,
		LastResultPath: lastResultPath,
		NativeJSON:     true,
	})
	if err != nil {
		t.Fatalf("DispatchNativeHook() error = %v", err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("ExitCode = %d, want 0", resp.ExitCode)
	}
	if _, err := os.Stat(filepath.Join(stateDir, "ralphx-active.json")); !os.IsNotExist(err) {
		t.Fatalf("expected no active state file, err=%v", err)
	}
	runState, err := state.LoadRunState(state.DerivePaths(workdir, stateDir))
	if err != nil {
		t.Fatalf("LoadRunState() error = %v", err)
	}
	if runState.Hook == nil || runState.Hook.Event != string(EventStop) || !runState.Hook.Allow {
		t.Fatalf("hook state = %#v", runState.Hook)
	}
}

func mustWriteJSON(t *testing.T, path string, value any) {
	t.Helper()
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}
	mustWriteFile(t, path, string(data))
}
