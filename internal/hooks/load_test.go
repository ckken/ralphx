package hooks

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/ckken/ralphx/internal/contracts"
	"github.com/ckken/ralphx/internal/state"
)

func TestLoadStopGuardInputUsesChecklistAndLastResult(t *testing.T) {
	dir := t.TempDir()
	taskPath := filepath.Join(dir, "task.md")
	checklistPath := filepath.Join(dir, "task.checklist.md")
	summaryPath := filepath.Join(dir, "summary.txt")
	statePath := filepath.Join(dir, "state.json")
	lastResultPath := filepath.Join(dir, "last-result.json")

	mustWriteFile(t, taskPath, "# Task\n\nKeep going.\n")
	mustWriteFile(t, checklistPath, "- [ ] first\n- [x] done\n")
	mustWriteFile(t, summaryPath, "summary")

	runState := state.RunState{
		Iteration: 1,
		UpdatedAt: "2026-01-02 15:04:05",
		Result: contracts.RoundResult{
			Status:        contracts.StatusComplete,
			Mode:          contracts.ModeComplete,
			ExitSignal:    true,
			FilesModified: 1,
			TestsPassed:   true,
			Summary:       "done",
		},
	}
	data, _ := json.Marshal(runState)
	mustWriteFile(t, statePath, string(data))

	lastResult := contracts.RoundResult{
		Status:        contracts.StatusInProgress,
		Mode:          contracts.ModeProducePlan,
		ExitSignal:    false,
		FilesModified: 0,
		TestsPassed:   false,
		Summary:       "plan",
		NextStep:      "Continue with the next bounded step",
	}
	data, _ = json.Marshal(lastResult)
	mustWriteFile(t, lastResultPath, string(data))

	input, err := LoadStopGuardInput(taskPath, checklistPath, summaryPath, statePath, lastResultPath, true, false)
	if err != nil {
		t.Fatalf("LoadStopGuardInput() error = %v", err)
	}
	if input.ChecklistOpen != 1 {
		t.Fatalf("checklist open = %d", input.ChecklistOpen)
	}
	if input.Result.Mode != contracts.ModeProducePlan {
		t.Fatalf("mode = %q", input.Result.Mode)
	}
}

func TestLoadStopGuardInputReturnsNoTaskContextWithoutTaskOrState(t *testing.T) {
	dir := t.TempDir()
	_, err := LoadStopGuardInput("", "", filepath.Join(dir, "summary.txt"), filepath.Join(dir, "state.json"), filepath.Join(dir, "last-result.json"), false, false)
	if !errors.Is(err, ErrNoTaskContext) {
		t.Fatalf("expected ErrNoTaskContext, got %v", err)
	}
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
