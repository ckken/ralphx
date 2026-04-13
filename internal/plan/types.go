package plan

import (
	"fmt"
	"strings"
)

type Output struct {
	Title        string   `json:"title"`
	TaskMarkdown string   `json:"task_markdown"`
	Checklist    []string `json:"checklist"`
	TestsCmd     string   `json:"tests_cmd"`
}

func (o Output) Validate() error {
	if strings.TrimSpace(o.Title) == "" {
		return fmt.Errorf("missing title")
	}
	if strings.TrimSpace(o.TaskMarkdown) == "" {
		return fmt.Errorf("missing task_markdown")
	}
	for i, item := range o.Checklist {
		if strings.TrimSpace(item) == "" {
			return fmt.Errorf("checklist item %d is empty", i)
		}
	}
	return nil
}
