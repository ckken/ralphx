package task

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAutoChecklistAndStateTexts(t *testing.T) {
	dir := t.TempDir()
	taskPath := filepath.Join(dir, "sample-task.md")
	checklistPath := filepath.Join(dir, "sample-task.checklist.md")
	summaryPath := filepath.Join(dir, "summary.txt")
	statePath := filepath.Join(dir, "state.json")

	writeFile(t, taskPath, "# Task\n\nShip it.\n")
	writeFile(t, checklistPath, "- [ ] first\n- [x] done\n* [ ] second\n")
	writeFile(t, summaryPath, "previous summary")
	writeFile(t, statePath, "{\"iteration\":1}")

	bundle, err := Load(taskPath, LoadOptions{
		SummaryPath: summaryPath,
		StatePath:   statePath,
	})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if bundle.Task.Content == "" {
		t.Fatal("expected task content")
	}
	if bundle.Checklist.Path != checklistPath {
		t.Fatalf("checklist path = %q, want %q", bundle.Checklist.Path, checklistPath)
	}
	if !bundle.Checklist.AutoDiscovered {
		t.Fatal("expected checklist to be auto-discovered")
	}
	if bundle.Checklist.OpenItems != 2 {
		t.Fatalf("open items = %d, want 2", bundle.Checklist.OpenItems)
	}
	if bundle.State.Summary != "previous summary" {
		t.Fatalf("summary = %q", bundle.State.Summary)
	}
	if bundle.State.State != "{\"iteration\":1}" {
		t.Fatalf("state = %q", bundle.State.State)
	}
}

func TestResolveChecklistPathExplicitMissing(t *testing.T) {
	dir := t.TempDir()
	taskPath := filepath.Join(dir, "task.md")
	writeFile(t, taskPath, "task")

	_, _, err := ResolveChecklistPath(taskPath, filepath.Join(dir, "missing.md"))
	if err == nil {
		t.Fatal("expected error for missing explicit checklist")
	}
}

func TestCountOpenChecklistItems(t *testing.T) {
	content := "- [ ] one\n  - [ ] nested\n* [x] closed\ntext\n"
	if got := CountOpenChecklistItems(content); got != 2 {
		t.Fatalf("CountOpenChecklistItems() = %d, want 2", got)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
