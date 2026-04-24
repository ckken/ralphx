package hooks

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type codexHooksFile struct {
	Hooks map[string][]map[string]any `json:"hooks"`
}

type InstallStatus struct {
	HooksPath        string `json:"hooks_path"`
	HooksFileFound   bool   `json:"hooks_file_found"`
	StopInstalled    bool   `json:"stop_installed"`
	PromptInstalled  bool   `json:"prompt_installed"`
	ManagedInstalled bool   `json:"managed_installed"`
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

	stopCommand := `bash -lc 'command -v ralphx >/dev/null 2>&1 || exit 0; project_root="${CODEX_PROJECT_ROOT:-$PWD}"; ralphx hook native --event Stop --workdir "$project_root" --native-json'`
	promptCommand := `bash -lc 'command -v ralphx >/dev/null 2>&1 || exit 0; project_root="${CODEX_PROJECT_ROOT:-$PWD}"; payload="$(mktemp)"; trap '\''rm -f "$payload"'\'' EXIT; cat >"$payload"; ralphx hook native --event UserPromptSubmit --payload "$payload" --workdir "$project_root" --json'`
	stopEntry := map[string]any{
		"hooks": []map[string]any{
			{
				"type":          "command",
				"command":       stopCommand,
				"timeout":       10,
				"statusMessage": "Running ralphx stop guard",
			},
		},
	}
	promptEntry := map[string]any{
		"hooks": []map[string]any{
			{
				"type":          "command",
				"command":       promptCommand,
				"timeout":       10,
				"statusMessage": "Activating ralphx workflow hooks",
			},
		},
	}

	content.Hooks["Stop"] = mergeManagedHookEntries(content.Hooks["Stop"], stopEntry)
	content.Hooks["UserPromptSubmit"] = mergeManagedHookEntries(content.Hooks["UserPromptSubmit"], promptEntry)

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
	fmt.Fprintln(os.Stdout, "note: start a new Codex session before testing \"$ralphx\" so UserPromptSubmit can fire")
	return hooksPath, nil
}

func ReadUserHookInstallStatus() (InstallStatus, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return InstallStatus{}, err
	}
	hooksPath := filepath.Join(home, ".codex", "hooks.json")
	status := InstallStatus{HooksPath: hooksPath}
	data, err := os.ReadFile(hooksPath)
	if err != nil {
		if os.IsNotExist(err) {
			return status, nil
		}
		return InstallStatus{}, err
	}
	status.HooksFileFound = true
	var content codexHooksFile
	if err := json.Unmarshal(data, &content); err != nil {
		return InstallStatus{}, err
	}
	status.StopInstalled = eventHasManagedHook(content.Hooks, "Stop")
	status.PromptInstalled = eventHasManagedHook(content.Hooks, "UserPromptSubmit")
	status.ManagedInstalled = status.StopInstalled && status.PromptInstalled
	return status, nil
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

func eventHasManagedHook(hooksMap map[string][]map[string]any, event string) bool {
	for _, existing := range hooksMap[event] {
		if hasManagedRalphxHook(existing) {
			return true
		}
	}
	return false
}

func hasManagedRalphxHook(entry map[string]any) bool {
	return valueContainsManagedHook(entry)
}

func containsManagedHookCommand(command string) bool {
	return strings.Contains(command, "ralphx hook native") ||
		strings.Contains(command, "ralphx hook stop-guard") ||
		strings.Contains(command, "ralphx hook prompt-submit")
}

func valueContainsManagedHook(value any) bool {
	switch typed := value.(type) {
	case map[string]any:
		if command, _ := typed["command"].(string); command != "" && containsManagedHookCommand(command) {
			return true
		}
		for _, nested := range typed {
			if valueContainsManagedHook(nested) {
				return true
			}
		}
	case []any:
		for _, nested := range typed {
			if valueContainsManagedHook(nested) {
				return true
			}
		}
	case []map[string]any:
		for _, nested := range typed {
			if valueContainsManagedHook(nested) {
				return true
			}
		}
	}
	return false
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
