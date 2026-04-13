package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/ckken/ralphx/internal/contracts"
	"github.com/ckken/ralphx/internal/execx"
)

type CodexAgent struct {
	Command string
}

func NewCodex(command string) CodexAgent {
	if strings.TrimSpace(command) == "" {
		command = "codex"
	}
	return CodexAgent{Command: command}
}

func (a CodexAgent) Run(ctx context.Context, req Request) (Response, error) {
	args := append([]string{}, req.ExtraArgs...)
	if a.Command == "codex" {
		args = append([]string{"exec", "--skip-git-repo-check", "--dangerously-bypass-approvals-and-sandbox", "-C", req.Workdir, "--output-schema", req.OutputSchemaPath, "-o", req.RawLogPath, "-"}, args...)
	}
	res, err := execx.Run(ctx, a.Command, args, []byte(req.Prompt), req.Workdir)
	raw := res.Output
	if req.RawLogPath != "" {
		if data, readErr := os.ReadFile(req.RawLogPath); readErr == nil && len(data) > 0 {
			raw = data
		} else if len(raw) > 0 {
			_ = os.WriteFile(req.RawLogPath, raw, 0o644)
		}
	}
	parsed, parseErr := ExtractRoundResult(raw)
	if parseErr != nil {
		if err != nil {
			return Response{RawOutput: raw}, fmt.Errorf("command error: %w; parse error: %v", err, parseErr)
		}
		return Response{RawOutput: raw}, parseErr
	}
	if err != nil {
		return Response{RawOutput: raw, Parsed: parsed}, err
	}
	return Response{RawOutput: raw, Parsed: parsed}, nil
}

func ExtractRoundResult(raw []byte) (contracts.RoundResult, error) {
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
		var result contracts.RoundResult
		if err := json.Unmarshal(data, &result); err != nil {
			continue
		}
		result.Blockers = contracts.NormalizeBlockers(result.Blockers)
		if err := result.Validate(); err != nil {
			continue
		}
		return result, nil
	}

	// fallback scan for first object substring
	for start := 0; start < len(text); start++ {
		if text[start] != '{' {
			continue
		}
		for end := len(text); end > start; end-- {
			if text[end-1] != '}' {
				continue
			}
			var result contracts.RoundResult
			if err := json.Unmarshal([]byte(text[start:end]), &result); err == nil {
				result.Blockers = contracts.NormalizeBlockers(result.Blockers)
				if err := result.Validate(); err == nil {
					return result, nil
				}
			}
		}
	}
	return contracts.RoundResult{}, errors.New("could not parse JSON result")
}
