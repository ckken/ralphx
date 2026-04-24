package hooks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallUserStopHookWritesHooksJson(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	path, err := InstallUserStopHook()
	if err != nil {
		t.Fatalf("InstallUserStopHook() error = %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read hooks.json: %v", err)
	}
	text := string(data)
	if !strings.Contains(text, "ralphx hook native --event Stop") {
		t.Fatalf("hooks.json missing managed stop command: %q", text)
	}
	if !strings.Contains(text, "ralphx hook native --event UserPromptSubmit") {
		t.Fatalf("hooks.json missing prompt-submit command: %q", text)
	}
	if filepath.Base(path) != "hooks.json" {
		t.Fatalf("unexpected path: %s", path)
	}
	status, err := ReadUserHookInstallStatus()
	if err != nil {
		t.Fatalf("ReadUserHookInstallStatus() error = %v", err)
	}
	if !status.HooksFileFound || !status.StopInstalled || !status.PromptInstalled || !status.ManagedInstalled {
		t.Fatalf("install status = %#v, want fully installed", status)
	}
}

func TestUninstallUserStopHookRemovesManagedEntry(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	path, err := InstallUserStopHook()
	if err != nil {
		t.Fatalf("InstallUserStopHook() error = %v", err)
	}
	if _, err := UninstallUserStopHook(); err != nil {
		t.Fatalf("UninstallUserStopHook() error = %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read hooks.json: %v", err)
	}
	if strings.Contains(string(data), "ralphx hook native --event Stop") {
		t.Fatalf("managed hook still present: %q", string(data))
	}
	if strings.Contains(string(data), "ralphx hook native --event UserPromptSubmit") {
		t.Fatalf("managed prompt hook still present: %q", string(data))
	}
}

func TestInstallUserStopHookRemovesLegacyPromptSubmitEntry(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	hooksPath := filepath.Join(home, ".codex", "hooks.json")
	if err := os.MkdirAll(filepath.Dir(hooksPath), 0o755); err != nil {
		t.Fatalf("mkdir hooks dir: %v", err)
	}

	legacy := `{
  "hooks": {
    "UserPromptSubmit": [
      {
        "hooks": [
          {
            "command": "bash -lc 'command -v ralphx >/dev/null 2>&1 || exit 0; payload=\"$(mktemp)\"; cat >\"$payload\"; ralphx hook native --event UserPromptSubmit --payload \"$payload\" --json; rm -f \"$payload\"'",
            "statusMessage": "Activating ralphx workflow hooks",
            "timeout": 10,
            "type": "command"
          }
        ]
      }
    ]
  }
}`
	if err := os.WriteFile(hooksPath, []byte(legacy), 0o644); err != nil {
		t.Fatalf("write hooks.json: %v", err)
	}

	path, err := InstallUserStopHook()
	if err != nil {
		t.Fatalf("InstallUserStopHook() error = %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read hooks.json: %v", err)
	}
	text := string(data)
	if strings.Contains(text, `command": "bash -lc 'command -v ralphx >/dev/null 2>&1 || exit 0; payload="$(mktemp)"; cat >"$payload"; ralphx hook native --event UserPromptSubmit --payload "$payload" --json; rm -f "$payload"'`) {
		t.Fatalf("legacy prompt hook still present: %q", text)
	}
	if !strings.Contains(text, "ralphx hook native --event Stop") {
		t.Fatalf("hooks.json missing managed stop command: %q", text)
	}
	if !strings.Contains(text, "ralphx hook native --event UserPromptSubmit") {
		t.Fatalf("hooks.json missing managed prompt command: %q", text)
	}
}
