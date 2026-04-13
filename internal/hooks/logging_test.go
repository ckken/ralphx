package hooks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAppendLogWritesJSONL(t *testing.T) {
	dir := t.TempDir()
	err := AppendLog(dir, LogEntry{
		Event:         EventStop,
		TaskPath:      "tasks/demo.md",
		ChecklistPath: "tasks/demo.checklist.md",
		Decision: Decision{
			Allow:   false,
			Reason:  "task_incomplete",
			Message: "Continue",
		},
	})
	if err != nil {
		t.Fatalf("AppendLog() error = %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 log file, got %d", len(entries))
	}
	data, err := os.ReadFile(filepath.Join(dir, entries[0].Name()))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := string(data)
	if !strings.Contains(text, "\"event\":\"stop\"") {
		t.Fatalf("log missing event: %q", text)
	}
	if !strings.Contains(text, "\"Reason\":\"task_incomplete\"") {
		t.Fatalf("log missing decision: %q", text)
	}
}

func TestAppendUserLogWritesUnderCodexLog(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	err := AppendUserLog(LogEntry{
		Event: EventPromptSubmit,
		Decision: Decision{
			Allow:   true,
			Reason:  "prompt_submit",
			Message: "ralphx mode active",
		},
	})
	if err != nil {
		t.Fatalf("AppendUserLog() error = %v", err)
	}
	entries, err := os.ReadDir(filepath.Join(home, ".codex", "log"))
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 user log file, got %d", len(entries))
	}
}
