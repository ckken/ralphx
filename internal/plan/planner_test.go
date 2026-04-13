package plan

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractOutputFindsEmbeddedJSON(t *testing.T) {
	raw := []byte("noise\n{\"title\":\"Ship task\",\"task_markdown\":\"# Task\\n\\nDo work.\",\"checklist\":[\"step one\",\"step two\"],\"tests_cmd\":\"go test ./...\"}\n")
	out, err := ExtractOutput(raw)
	if err != nil {
		t.Fatalf("ExtractOutput() error = %v", err)
	}
	if out.Title != "Ship task" {
		t.Fatalf("title = %q", out.Title)
	}
	if len(out.Checklist) != 2 {
		t.Fatalf("checklist len = %d", len(out.Checklist))
	}
}

func TestWriteFilesCreatesTaskAndChecklist(t *testing.T) {
	dir := t.TempDir()
	taskPath := filepath.Join(dir, "tasks", "sample.md")
	taskFile, checklistFile, err := WriteFiles(taskPath, Output{
		Title:        "Sample",
		TaskMarkdown: "# Task\n\nShip it.",
		Checklist:    []string{"first", "second"},
	})
	if err != nil {
		t.Fatalf("WriteFiles() error = %v", err)
	}

	taskData, err := os.ReadFile(taskFile)
	if err != nil {
		t.Fatalf("read task: %v", err)
	}
	if !strings.Contains(string(taskData), "Ship it.") {
		t.Fatalf("task file missing content: %q", string(taskData))
	}

	checklistData, err := os.ReadFile(checklistFile)
	if err != nil {
		t.Fatalf("read checklist: %v", err)
	}
	if !strings.Contains(string(checklistData), "- [ ] first") {
		t.Fatalf("checklist missing item: %q", string(checklistData))
	}
}
