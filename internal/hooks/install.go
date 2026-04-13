package hooks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type codexHooksFile struct {
	Hooks map[string][]map[string]any `json:"hooks"`
}

func InstallUserStopHook() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	hooksPath := filepath.Join(home, ".codex", "hooks.json")
	content := codexHooksFile{Hooks: map[string][]map[string]any{}}
	if data, err := os.ReadFile(hooksPath); err == nil {
		_ = json.Unmarshal(data, &content)
	}
	if content.Hooks == nil {
		content.Hooks = map[string][]map[string]any{}
	}

	command := `bash -lc 'command -v ralphx >/dev/null 2>&1 || exit 0; project_root="${CODEX_PROJECT_ROOT:-$PWD}"; ralphx hook stop-guard --workdir "$project_root" --native-json'`
	entry := map[string]any{
		"hooks": []map[string]any{
			{
				"type":          "command",
				"command":       command,
				"timeout":       10,
				"statusMessage": "Running ralphx stop guard",
			},
		},
	}

	content.Hooks["Stop"] = mergeManagedHookEntries(content.Hooks["Stop"], entry)

	promptSubmitCommand := `bash -lc 'command -v ralphx >/dev/null 2>&1 || exit 0; payload="$(mktemp)"; cat >"$payload"; ralphx hook prompt-submit --payload "$payload" --json; rm -f "$payload"'`
	promptSubmitEntry := map[string]any{
		"hooks": []map[string]any{
			{
				"type":          "command",
				"command":       promptSubmitCommand,
				"timeout":       10,
				"statusMessage": "Activating ralphx workflow hooks",
			},
		},
	}
	content.Hooks["UserPromptSubmit"] = mergeManagedHookEntries(content.Hooks["UserPromptSubmit"], promptSubmitEntry)

	if err := os.MkdirAll(filepath.Dir(hooksPath), 0o755); err != nil {
		return "", err
	}
	data, err := json.MarshalIndent(content, "", "  ")
	if err != nil {
		return "", err
	}
	data = append(data, '\n')
	if err := os.WriteFile(hooksPath, data, 0o644); err != nil {
		return "", err
	}
	return hooksPath, nil
}

func UninstallUserStopHook() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	hooksPath := filepath.Join(home, ".codex", "hooks.json")
	data, err := os.ReadFile(hooksPath)
	if err != nil {
		if os.IsNotExist(err) {
			return hooksPath, nil
		}
		return "", err
	}
	var content codexHooksFile
	if err := json.Unmarshal(data, &content); err != nil {
		return "", err
	}
	uninstallManagedEvent(content.Hooks, "Stop")
	uninstallManagedEvent(content.Hooks, "UserPromptSubmit")
	data, err = json.MarshalIndent(content, "", "  ")
	if err != nil {
		return "", err
	}
	data = append(data, '\n')
	if err := os.WriteFile(hooksPath, data, 0o644); err != nil {
		return "", err
	}
	return hooksPath, nil
}

func hasManagedRalphxHook(entry map[string]any) bool {
	hooksValue, ok := entry["hooks"]
	if !ok {
		return false
	}
	hooksList, ok := hooksValue.([]any)
	if !ok {
		return false
	}
	for _, item := range hooksList {
		hook, ok := item.(map[string]any)
		if !ok {
			continue
		}
		command, _ := hook["command"].(string)
		if command != "" && containsManagedHookCommand(command) {
			return true
		}
	}
	return false
}

func containsManagedHookCommand(command string) bool {
	return strings.Contains(command, "ralphx hook stop-guard") || strings.Contains(command, "ralphx hook prompt-submit")
}

func mergeManagedHookEntries(entries []map[string]any, managed map[string]any) []map[string]any {
	filtered := make([]map[string]any, 0, len(entries))
	for _, existing := range entries {
		if hasManagedRalphxHook(existing) {
			continue
		}
		filtered = append(filtered, existing)
	}
	filtered = append(filtered, managed)
	return filtered
}

func uninstallManagedEvent(hooksMap map[string][]map[string]any, event string) {
	entries := hooksMap[event]
	filtered := make([]map[string]any, 0, len(entries))
	for _, existing := range entries {
		if hasManagedRalphxHook(existing) {
			continue
		}
		filtered = append(filtered, existing)
	}
	if len(filtered) == 0 {
		delete(hooksMap, event)
	} else {
		hooksMap[event] = filtered
	}
}
