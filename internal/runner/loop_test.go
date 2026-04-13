package runner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ckken/ralphx/internal/config"
	"github.com/ckken/ralphx/internal/contracts"
	"github.com/ckken/ralphx/internal/state"
	"github.com/ckken/ralphx/internal/task"
)

func TestApplyProducePlanWritesTaskAndChecklist(t *testing.T) {
	dir := t.TempDir()
	taskPath := filepath.Join(dir, "task.md")
	checklistPath := filepath.Join(dir, "task.checklist.md")
	if err := os.WriteFile(taskPath, []byte("# Task\n\nShip it.\n"), 0o644); err != nil {
		t.Fatalf("write task: %v", err)
	}
	if err := os.WriteFile(checklistPath, []byte("# Task checklist\n\n- [x] done one\n"), 0o644); err != nil {
		t.Fatalf("write checklist: %v", err)
	}

	loop := Loop{Config: config.RunConfig{TaskFile: taskPath}}
	paths := state.DerivePaths(dir, filepath.Join(dir, ".ralphx"))
	if err := paths.Ensure(); err != nil {
		t.Fatalf("ensure state paths: %v", err)
	}
	bundle := task.Bundle{
		Task: task.Document{Path: taskPath, Content: "# Task\n\nShip it.\n"},
		Checklist: task.Checklist{
			Path:    checklistPath,
			Content: "# Task checklist\n\n- [x] done one\n",
		},
	}
	result := contracts.RoundResult{
		Status:          contracts.StatusInProgress,
		Mode:            contracts.ModeProducePlan,
		ExitSignal:      false,
		FilesModified:   0,
		TestsPassed:     false,
		Summary:         "planned next slice",
		NextStep:        "Refactor registry access behind session-local state.",
		ChecklistUpdate: []string{"Move component registry access behind session handle"},
	}

	guidance, err := loop.applyProducePlan(bundle, paths, result)
	if err != nil {
		t.Fatalf("applyProducePlan: %v", err)
	}
	if guidance == nil {
		t.Fatal("expected guidance")
	}

	taskData, err := os.ReadFile(taskPath)
	if err != nil {
		t.Fatalf("read task: %v", err)
	}
	if !strings.Contains(string(taskData), "## Planned Next Step") {
		t.Fatalf("task missing planned next step section: %q", string(taskData))
	}

	checklistData, err := os.ReadFile(checklistPath)
	if err != nil {
		t.Fatalf("read checklist: %v", err)
	}
	if !strings.Contains(string(checklistData), "- [ ] Move component registry access behind session handle") {
		t.Fatalf("checklist missing new item: %q", string(checklistData))
	}
	if !strings.Contains(string(checklistData), "done one") {
		t.Fatalf("checklist lost completed item: %q", string(checklistData))
	}
}
