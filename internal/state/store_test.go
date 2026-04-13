package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ckken/ralphx/internal/contracts"
)

func TestSessionFresh(t *testing.T) {
	now := time.Date(2026, 1, 2, 15, 4, 5, 0, time.UTC)
	fresh := SessionMeta{ThreadID: "thread-1", UpdatedAt: now.Add(-30 * time.Minute).Format(timestampLayout)}
	if !SessionFresh(fresh, time.Hour, now) {
		t.Fatal("expected session to be fresh")
	}

	stale := SessionMeta{ThreadID: "thread-1", UpdatedAt: now.Add(-2 * time.Hour).Format(timestampLayout)}
	if SessionFresh(stale, time.Hour, now) {
		t.Fatal("expected session to be stale")
	}
}

func TestWriteStateWithGuidance(t *testing.T) {
	dir := t.TempDir()
	paths := DerivePaths(dir, filepath.Join(dir, ".ralphx"))
	if err := paths.Ensure(); err != nil {
		t.Fatalf("Ensure() error = %v", err)
	}
	result := contracts.RoundResult{Status: contracts.StatusBlocked, ExitSignal: false, FilesModified: 0, TestsPassed: false, Summary: "blocked"}
	guidance := &Guidance{
		Reason:        "no_progress",
		Message:       "auto replan happened",
		TaskFile:      "tasks/demo.md",
		ChecklistFile: "tasks/demo.checklist.md",
		GeneratedAt:   "2026-01-02 15:04:05",
	}
	if err := WriteStateWithGuidance(paths, 3, result, guidance); err != nil {
		t.Fatalf("WriteStateWithGuidance() error = %v", err)
	}
	data, err := os.ReadFile(paths.StateFile)
	if err != nil {
		t.Fatalf("read state file: %v", err)
	}
	var runState RunState
	if err := json.Unmarshal(data, &runState); err != nil {
		t.Fatalf("unmarshal state: %v", err)
	}
	if runState.Guidance == nil || runState.Guidance.Reason != "no_progress" {
		t.Fatalf("guidance = %#v", runState.Guidance)
	}
}
