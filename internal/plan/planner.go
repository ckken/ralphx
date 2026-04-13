package plan

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ckken/ralphx/internal/execx"
)

type Request struct {
	Goal             string
	Workdir          string
	CodexCmd         string
	CodexArgs        []string
	OutputSchemaPath string
	RawLogPath       string
}

func Run(ctx context.Context, req Request) (Output, []byte, error) {
	if strings.TrimSpace(req.Goal) == "" {
		return Output{}, nil, fmt.Errorf("goal is required")
	}
	if strings.TrimSpace(req.Workdir) == "" {
		return Output{}, nil, fmt.Errorf("workdir is required")
	}

	prompt := buildPrompt(req.Goal, req.Workdir)
	command := strings.TrimSpace(req.CodexCmd)
	if command == "" {
		command = "codex"
	}

	args := append([]string{}, req.CodexArgs...)
	if command == "codex" {
		args = append([]string{
			"exec",
			"--skip-git-repo-check",
			"--dangerously-bypass-approvals-and-sandbox",
			"-C", req.Workdir,
			"--output-schema", req.OutputSchemaPath,
			"-o", req.RawLogPath,
			"-",
		}, args...)
	}

	res, err := execx.Run(ctx, command, args, []byte(prompt), req.Workdir)
	raw := res.Output
	if req.RawLogPath != "" {
		if data, readErr := os.ReadFile(req.RawLogPath); readErr == nil && len(data) > 0 {
			raw = data
		} else if len(raw) > 0 {
			_ = os.WriteFile(req.RawLogPath, raw, 0o644)
		}
	}

	parsed, parseErr := ExtractOutput(raw)
	if parseErr != nil {
		if err != nil {
			return Output{}, raw, fmt.Errorf("command error: %w; parse error: %v", err, parseErr)
		}
		return Output{}, raw, parseErr
	}
	if err != nil {
		return parsed, raw, err
	}
	return parsed, raw, nil
}

func ExtractOutput(raw []byte) (Output, error) {
	text := string(raw)
	dec := json.NewDecoder(strings.NewReader(text))
	for {
		var value any
		if err := dec.Decode(&value); err != nil {
			break
		}
		obj, ok := value.(map[string]any)
		if !ok {
			continue
		}
		data, err := json.Marshal(obj)
		if err != nil {
			continue
		}
		var out Output
		if err := json.Unmarshal(data, &out); err != nil {
			continue
		}
		if err := out.Validate(); err != nil {
			continue
		}
		return out, nil
	}

	for start := 0; start < len(text); start++ {
		if text[start] != '{' {
			continue
		}
		for end := len(text); end > start; end-- {
			if text[end-1] != '}' {
				continue
			}
			var out Output
			if err := json.Unmarshal([]byte(text[start:end]), &out); err == nil {
				if err := out.Validate(); err == nil {
					return out, nil
				}
			}
		}
	}

	return Output{}, errors.New("could not parse planner JSON result")
}

func buildPrompt(goal, workdir string) string {
	return fmt.Sprintf(`You are planning a repository task for an outer-loop coding runner.

Goal:
%s

Workspace:
%s

Produce exactly one JSON object that follows the provided schema.

Requirements:
- Convert the goal into a concrete task markdown file.
- Break the work into a short checklist of actionable implementation steps.
- Prefer 3-7 checklist items.
- Keep the task markdown concise and execution-oriented.
- Include a tests_cmd only if there is an obvious repo-local validation command; otherwise use an empty string.
- Do not output markdown fences.
- Do not output commentary outside JSON.
`, strings.TrimSpace(goal), workdir)
}

func ChecklistPath(taskPath string) string {
	ext := filepath.Ext(taskPath)
	if ext != ".md" {
		return taskPath + ".checklist.md"
	}
	return strings.TrimSuffix(taskPath, ext) + ".checklist" + ext
}
