package runner

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ckken/ralphx/internal/agent"
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

func TestTryAutoReplanUsesRoundTimeout(t *testing.T) {
	dir := t.TempDir()
	taskPath := filepath.Join(dir, "task.md")
	checklistPath := filepath.Join(dir, "task.checklist.md")
	if err := os.WriteFile(taskPath, []byte("# Task\n\nShip it.\n"), 0o644); err != nil {
		t.Fatalf("write task: %v", err)
	}
	if err := os.WriteFile(checklistPath, []byte("- [ ] one\n"), 0o644); err != nil {
		t.Fatalf("write checklist: %v", err)
	}
	fakeCodex := filepath.Join(dir, "fake-codex.sh")
	if err := os.WriteFile(fakeCodex, []byte("#!/usr/bin/env bash\nsleep 5\n"), 0o755); err != nil {
		t.Fatalf("write fake codex: %v", err)
	}
	paths := state.DerivePaths(dir, filepath.Join(dir, ".ralphx"))
	if err := paths.Ensure(); err != nil {
		t.Fatalf("ensure state paths: %v", err)
	}

	loop := Loop{Config: config.RunConfig{
		TaskFile:      taskPath,
		ChecklistFile: checklistPath,
		Workdir:       dir,
		StateDir:      paths.Root,
		CodexCmd:      fakeCodex,
		AutoReplan:    true,
		RoundTimeout:  20 * time.Millisecond,
	}}
	start := time.Now()
	replanned, guidance, err := loop.tryAutoReplan(context.Background(), paths, "timeout_test")
	if err == nil {
		t.Fatal("tryAutoReplan() error = nil, want timeout")
	}
	if replanned || guidance != nil {
		t.Fatalf("replanned=%t guidance=%#v, want no result", replanned, guidance)
	}
	if time.Since(start) > time.Second {
		t.Fatalf("tryAutoReplan ignored RoundTimeout; elapsed=%s", time.Since(start))
	}
}

func TestRunReturnsErrorWhenMaxIterationsReachedBeforeCompletion(t *testing.T) {
	dir := t.TempDir()
	taskPath := filepath.Join(dir, "task.md")
	if err := os.WriteFile(taskPath, []byte("# Task\n\nKeep working.\n"), 0o644); err != nil {
		t.Fatalf("write task: %v", err)
	}
	loop := Loop{
		Config: config.RunConfig{
			TaskFile:      taskPath,
			Workdir:       dir,
			StateDir:      filepath.Join(dir, ".ralphx"),
			MaxIterations: 1,
			MaxNoProgress: 0,
			RoundTimeout:  time.Second,
			AutoReplan:    false,
		},
		Agent: staticAgent{result: contracts.RoundResult{
			Status:        contracts.StatusInProgress,
			Mode:          contracts.ModeExecuteNextStep,
			ExitSignal:    false,
			FilesModified: 1,
			TestsPassed:   false,
			Blockers:      nil,
			Summary:       "one bounded step remains",
		}},
	}
	err := loop.Run(context.Background())
	if err == nil || !strings.Contains(err.Error(), "MAX_ITERATIONS=1") {
		t.Fatalf("Run() error = %v, want MAX_ITERATIONS error", err)
	}
}

type staticAgent struct {
	result contracts.RoundResult
}

func (a staticAgent) Run(context.Context, agent.Request) (agent.Response, error) {
	return agent.Response{Parsed: a.result}, nil
}
