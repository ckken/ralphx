package hooks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type ActiveState struct {
	Mode           string `json:"mode"`
	Active         bool   `json:"active"`
	Prompt         string `json:"prompt,omitempty"`
	StopHookActive bool   `json:"stop_hook_active,omitempty"`
	StopReason     string `json:"stop_reason,omitempty"`
	UpdatedAt      string `json:"updated_at"`
}

func WriteActiveState(stateDir, prompt string) error {
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		return err
	}
	state := ActiveState{
		Mode:           "ralphx",
		Active:         true,
		Prompt:         prompt,
		StopHookActive: false,
		UpdatedAt:      time.Now().Format(time.RFC3339),
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(filepath.Join(stateDir, "ralphx-active.json"), data, 0o644)
}

func MarkStopHookActive(stateDir, reason string) error {
	state, err := ReadActiveState(stateDir)
	if err != nil {
		return err
	}
	state.StopHookActive = true
	state.StopReason = reason
	state.UpdatedAt = time.Now().Format(time.RFC3339)
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(filepath.Join(stateDir, "ralphx-active.json"), data, 0o644)
}

func ReadActiveState(stateDir string) (ActiveState, error) {
	data, err := os.ReadFile(filepath.Join(stateDir, "ralphx-active.json"))
	if err != nil {
		return ActiveState{}, err
	}
	var state ActiveState
	if err := json.Unmarshal(data, &state); err != nil {
		return ActiveState{}, err
	}
	return state, nil
}

func ClearActiveState(stateDir string) error {
	if err := os.Remove(filepath.Join(stateDir, "ralphx-active.json")); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
