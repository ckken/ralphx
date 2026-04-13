package prompt

import (
	"fmt"
	"os"
	"strings"

	"github.com/ckken/ralphx/internal/assets"
	"github.com/ckken/ralphx/internal/task"
)

type BuildInput struct {
	Iteration int
	Workdir   string
	Bundle    task.Bundle
	Template  string
	GitStatus string
}

func LoadTemplate(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return assets.DefaultPrompt(), nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return assets.DefaultPrompt(), nil
		}
		return "", err
	}
	return string(data), nil
}

func Build(in BuildInput) string {
	checklistPath := "none"
	checklistContent := ""
	checklistOpenItems := 0
	if strings.TrimSpace(in.Bundle.Checklist.Path) != "" {
		checklistPath = in.Bundle.Checklist.Path
		checklistContent = in.Bundle.Checklist.Content
		checklistOpenItems = in.Bundle.Checklist.OpenItems
	}

	return fmt.Sprintf(`%s

You are running inside an autonomous Go loop.

Task:
%s

Iteration:
%d

Workspace:
%s

Checklist file:
%s

Open checklist items:
%d

Checklist content:
%s

Previous summary:
%s

Current state:
%s

Current git status:
%s

Rules:
- Make the smallest correct change.
- If a checklist file is provided, you must treat unchecked items as hard remaining work.
- Update the checklist file when you complete a milestone.
- Once you identify the next highest-value edge, do not stop at advice alone.
- If you are not executing a bounded step right now, you must return a concrete next-step plan.
- If the task is not complete, return status="in_progress".
- If blocked, return status="blocked" and include blockers.
- If done, return status="complete" and set exit_signal=true.
- Always return exactly one JSON object and no extra text.
- Use this schema:
  {
    "status": "in_progress|blocked|complete",
    "mode": "execute_next_step|produce_plan|blocked|complete",
    "exit_signal": true|false,
    "files_modified": 0,
    "tests_passed": true|false,
    "blockers": [],
    "summary": "short summary",
    "next_step": "required when mode=produce_plan",
    "checklist_update": ["optional checklist items for the next slice"]
  }
`, strings.TrimSpace(in.Template), in.Bundle.Task.Content, in.Iteration, in.Workdir, checklistPath, checklistOpenItems, checklistContent, in.Bundle.State.Summary, in.Bundle.State.State, in.GitStatus)
}
