package plan

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/ckken/ralphx/internal/task"
)

type ReplanRequest struct {
	TaskPath          string
	ChecklistPath     string
	SummaryPath       string
	StatePath         string
	Workdir           string
	CodexCmd          string
	CodexArgs         []string
	OutputSchemaPath  string
	RawLogPath        string
	PreserveCompleted bool
}

func Replan(ctx context.Context, req ReplanRequest) (Output, task.Bundle, []byte, error) {
	bundle, err := task.Load(req.TaskPath, task.LoadOptions{
		ChecklistPath: req.ChecklistPath,
		SummaryPath:   req.SummaryPath,
		StatePath:     req.StatePath,
	})
	if err != nil {
		return Output{}, task.Bundle{}, nil, err
	}

	goal := buildReplanGoal(bundle)
	out, raw, err := Run(ctx, Request{
		Goal:             goal,
		Workdir:          req.Workdir,
		CodexCmd:         req.CodexCmd,
		CodexArgs:        req.CodexArgs,
		OutputSchemaPath: req.OutputSchemaPath,
		RawLogPath:       req.RawLogPath,
	})
	if err != nil {
		return Output{}, bundle, raw, err
	}
	if req.PreserveCompleted && strings.TrimSpace(bundle.Checklist.Content) != "" {
		out.Checklist = mergeChecklist(bundle.Checklist.Content, out.Checklist)
	}
	return out, bundle, raw, nil
}

func buildReplanGoal(bundle task.Bundle) string {
	var b strings.Builder
	b.WriteString("Replan the current repository task.\n\n")
	b.WriteString("Current task file:\n")
	b.WriteString(strings.TrimSpace(bundle.Task.Content))
	b.WriteString("\n\n")
	if strings.TrimSpace(bundle.Checklist.Content) != "" {
		b.WriteString("Current checklist:\n")
		b.WriteString(strings.TrimSpace(bundle.Checklist.Content))
		b.WriteString("\n\n")
	}
	if strings.TrimSpace(bundle.State.Summary) != "" {
		b.WriteString("Previous summary:\n")
		b.WriteString(strings.TrimSpace(bundle.State.Summary))
		b.WriteString("\n\n")
	}
	if strings.TrimSpace(bundle.State.State) != "" {
		b.WriteString("Current state snapshot:\n")
		b.WriteString(strings.TrimSpace(bundle.State.State))
		b.WriteString("\n\n")
	}
	b.WriteString("Generate the next best task markdown and checklist for continuing this work. Keep the remaining scope concrete and execution-ready.")
	return b.String()
}

func mergeChecklist(existing string, next []string) []string {
	completed := completedChecklistTexts(existing)
	if len(completed) == 0 {
		return next
	}
	seen := map[string]bool{}
	merged := make([]string, 0, len(completed)+len(next))
	for _, item := range completed {
		key := normalizeChecklistText(item)
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		merged = append(merged, item)
	}
	for _, item := range next {
		key := normalizeChecklistText(item)
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		merged = append(merged, item)
	}
	return merged
}

func completedChecklistTexts(content string) []string {
	lines := strings.Split(content, "\n")
	out := make([]string, 0)
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "- [x] "):
			out = append(out, strings.TrimSpace(trimmed[6:]))
		case strings.HasPrefix(trimmed, "* [x] "):
			out = append(out, strings.TrimSpace(trimmed[6:]))
		case strings.HasPrefix(trimmed, "- [X] "):
			out = append(out, strings.TrimSpace(trimmed[6:]))
		case strings.HasPrefix(trimmed, "* [X] "):
			out = append(out, strings.TrimSpace(trimmed[6:]))
		}
	}
	return out
}

func normalizeChecklistText(text string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(text)), " "))
}

func DefaultReplanPaths(workdir, stateDir string) (string, string) {
	root := stateDir
	if strings.TrimSpace(root) == "" {
		root = filepath.Join(workdir, ".ralphx")
	}
	return filepath.Join(root, "summary.txt"), filepath.Join(root, "state.json")
}

func EnsureLogDir(path string) error {
	return os.MkdirAll(filepath.Dir(path), 0o755)
}
