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
	if !strings.Contains(text, "ralphx hook stop-guard") {
		t.Fatalf("hooks.json missing managed command: %q", text)
	}
	if !strings.Contains(text, "ralphx hook prompt-submit") {
		t.Fatalf("hooks.json missing prompt-submit command: %q", text)
	}
	if filepath.Base(path) != "hooks.json" {
		t.Fatalf("unexpected path: %s", path)
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
	if strings.Contains(string(data), "ralphx hook stop-guard") {
		t.Fatalf("managed hook still present: %q", string(data))
	}
	if strings.Contains(string(data), "ralphx hook prompt-submit") {
		t.Fatalf("managed prompt hook still present: %q", string(data))
	}
}
