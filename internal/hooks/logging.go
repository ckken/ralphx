package hooks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type LogEntry struct {
	Timestamp     string   `json:"timestamp"`
	Event         Event    `json:"event"`
	TaskPath      string   `json:"task_path,omitempty"`
	ChecklistPath string   `json:"checklist_path,omitempty"`
	Decision      Decision `json:"decision"`
	Result        any      `json:"result,omitempty"`
}

func AppendLog(logDir string, entry LogEntry) error {
	if entry.Timestamp == "" {
		entry.Timestamp = time.Now().Format(time.RFC3339)
	}
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(logDir, "hooks-"+time.Now().Format("2006-01-02")+".jsonl")
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(data)
	return err
}

func AppendUserLog(entry LogEntry) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	return AppendLog(filepath.Join(home, ".codex", "log"), entry)
}

func WriteLatest(path string, entry LogEntry) error {
	if entry.Timestamp == "" {
		entry.Timestamp = time.Now().Format(time.RFC3339)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

func WriteUserLatest(entry LogEntry) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	return WriteLatest(filepath.Join(home, ".codex", "log", "ralphx-last-hook-event.json"), entry)
}

func ReadLatest(path string) (LogEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return LogEntry{}, err
	}
	var entry LogEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return LogEntry{}, err
	}
	return entry, nil
}
