package hooks

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"
)

type PromptSubmitPayload struct {
	HookEventName string `json:"hook_event_name"`
	Cwd           string `json:"cwd"`
	Prompt        string `json:"prompt"`
	Input         string `json:"input"`
	UserPrompt    string `json:"user_prompt"`
	Text          string `json:"text"`
}

func LoadPromptSubmitPayload(path string) (PromptSubmitPayload, error) {
	var data []byte
	var err error
	if strings.TrimSpace(path) == "" {
		data, err = os.ReadFile("/dev/stdin")
	} else {
		data, err = os.ReadFile(path)
	}
	if err != nil {
		return PromptSubmitPayload{}, err
	}
	var payload PromptSubmitPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return PromptSubmitPayload{}, err
	}
	return payload, nil
}

func PromptText(payload PromptSubmitPayload) string {
	for _, value := range []string{payload.Prompt, payload.Input, payload.UserPrompt, payload.Text} {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

var stopPromptPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)^\s*(?:please\s+)?(?:stop|cancel|abort)\s*(?:now)?\s*[.!]?\s*$`),
	regexp.MustCompile(`(?i)\b(?:stop|cancel|abort)\b`),
	regexp.MustCompile(`(?i)^\s*(?:结束|停止)(?:当前)?(?:工作流|流程|会话)\s*[.!]?\s*$`),
	regexp.MustCompile(`(?i)\b(?:stop|cancel|abort)\s+(?:the\s+)?(?:current|active|running)\s+(?:workflow|task|run|session|mode)\b`),
	regexp.MustCompile(`(?i)\b(?:stop|cancel|abort)\s+(?:this|the)\s+(?:workflow|run|task|session|mode)\b`),
}

func PromptActivatesRalphx(text string) bool {
	return strings.TrimSpace(text) == "$ralphx"
}

func PromptStopsRalphx(text string) bool {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return false
	}
	for _, pattern := range stopPromptPatterns {
		if pattern.MatchString(trimmed) {
			return true
		}
	}
	return false
}
