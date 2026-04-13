package task

import (
	"fmt"
	"os"
	"strings"
)

type ChecklistItem struct {
	Index      int
	LineNumber int
	Text       string
	RawLine    string
}

func OpenChecklistItems(content string) []ChecklistItem {
	lines := strings.Split(content, "\n")
	items := make([]ChecklistItem, 0)
	itemIndex := 0
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- [ ] ") || strings.HasPrefix(trimmed, "* [ ] ") {
			text := strings.TrimSpace(trimmed[6:])
			items = append(items, ChecklistItem{
				Index:      itemIndex,
				LineNumber: i,
				Text:       text,
				RawLine:    line,
			})
			itemIndex++
		}
	}
	return items
}

func MarkChecklistItemsDone(path string, indexes []int) error {
	if strings.TrimSpace(path) == "" || len(indexes) == 0 {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	updated, err := MarkChecklistContentDone(string(data), indexes)
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(updated), 0o644)
}

func MarkChecklistContentDone(content string, indexes []int) (string, error) {
	if len(indexes) == 0 {
		return content, nil
	}
	wanted := map[int]bool{}
	for _, idx := range indexes {
		wanted[idx] = true
	}
	lines := strings.Split(content, "\n")
	openIndex := 0
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		prefix := ""
		switch {
		case strings.HasPrefix(trimmed, "- [ ] "):
			prefix = "- [ ] "
		case strings.HasPrefix(trimmed, "* [ ] "):
			prefix = "* [ ] "
		default:
			continue
		}
		if wanted[openIndex] {
			indent := line[:strings.Index(line, prefix)]
			replacement := strings.Replace(prefix, "[ ]", "[x]", 1)
			lines[i] = indent + replacement + strings.TrimSpace(trimmed[len(prefix):])
		}
		openIndex++
	}
	for idx := range wanted {
		if idx >= openIndex {
			return "", fmt.Errorf("checklist index %d out of range", idx)
		}
	}
	return strings.Join(lines, "\n"), nil
}
