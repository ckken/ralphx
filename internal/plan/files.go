package plan

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func WriteFiles(taskPath string, out Output) (string, string, error) {
	if strings.TrimSpace(taskPath) == "" {
		return "", "", fmt.Errorf("task path is required")
	}
	if err := out.Validate(); err != nil {
		return "", "", err
	}

	taskAbs, err := filepath.Abs(taskPath)
	if err != nil {
		return "", "", err
	}
	checklistAbs := ChecklistPath(taskAbs)

	if err := os.MkdirAll(filepath.Dir(taskAbs), 0o755); err != nil {
		return "", "", err
	}
	if err := os.MkdirAll(filepath.Dir(checklistAbs), 0o755); err != nil {
		return "", "", err
	}

	taskBody := strings.TrimSpace(out.TaskMarkdown) + "\n"
	checklistBody := renderChecklist(out.Title, out.Checklist)
	if err := os.WriteFile(taskAbs, []byte(taskBody), 0o644); err != nil {
		return "", "", err
	}
	if err := os.WriteFile(checklistAbs, []byte(checklistBody), 0o644); err != nil {
		return "", "", err
	}
	return taskAbs, checklistAbs, nil
}

func renderChecklist(title string, items []string) string {
	lines := []string{fmt.Sprintf("# %s checklist", strings.TrimSpace(title)), ""}
	for _, item := range items {
		lines = append(lines, "- [ ] "+strings.TrimSpace(item))
	}
	return strings.Join(lines, "\n") + "\n"
}
