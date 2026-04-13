package hooks

import (
	"encoding/json"
	"os"
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

func PromptActivatesRalphx(text string) bool {
	lower := strings.ToLower(text)
	return strings.Contains(lower, "$ralphx") || strings.Contains(lower, " ralphx") || strings.HasPrefix(lower, "ralphx ")
}
