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
		if strings.TrimSpace(req.SessionID) != "" {
			args = append([]string{"exec", "resume", req.SessionID, "--skip-git-repo-check", "--dangerously-bypass-approvals-and-sandbox", "--json", "-"}, args...)
		} else {
			args = append([]string{"exec", "--skip-git-repo-check", "--dangerously-bypass-approvals-and-sandbox", "-C", req.Workdir, "--json", "--output-schema", req.OutputSchemaPath, "-"}, args...)
		}
	}
	res, err := execx.Run(ctx, a.Command, args, []byte(req.Prompt), req.Workdir)
	raw := res.Output
	if req.RawLogPath != "" {
		_ = os.WriteFile(req.RawLogPath, raw, 0o644)
	}
	messageText, sessionID := extractAgentMessageAndSession(raw)
	if sessionID == "" {
		sessionID = strings.TrimSpace(req.SessionID)
	}
	parsed, parseErr := ExtractRoundResult([]byte(messageText))
	if parseErr != nil {
		if err != nil {
			return Response{RawOutput: raw, SessionID: sessionID}, fmt.Errorf("command error: %w; parse error: %v", err, parseErr)
		}
		return Response{RawOutput: raw, SessionID: sessionID}, parseErr
	}
	if err != nil {
		return Response{RawOutput: raw, Parsed: parsed, SessionID: sessionID}, err
	}
	return Response{RawOutput: raw, Parsed: parsed, SessionID: sessionID}, nil
}

func extractAgentMessageAndSession(raw []byte) (message string, sessionID string) {
	type event struct {
		Type     string `json:"type"`
		ThreadID string `json:"thread_id"`
		Item     struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"item"`
	}

	dec := json.NewDecoder(strings.NewReader(string(raw)))
	for {
		var ev event
		if err := dec.Decode(&ev); err != nil {
			break
		}
		if ev.Type == "thread.started" && strings.TrimSpace(ev.ThreadID) != "" {
			sessionID = ev.ThreadID
		}
		if ev.Type == "item.completed" && ev.Item.Type == "agent_message" && strings.TrimSpace(ev.Item.Text) != "" {
			message = ev.Item.Text
		}
	}
	if strings.TrimSpace(message) == "" {
		message = string(raw)
	}
	return strings.TrimSpace(message), strings.TrimSpace(sessionID)
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
