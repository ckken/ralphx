package doctor

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunTextReportsChecks(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	binDir := t.TempDir()
	makeExecutables(t, binDir, "bash", "python3", "git", "gh", "codex", "jq")
	t.Setenv("PATH", binDir)

	var stdout, stderr bytes.Buffer
	exitCode := run(nil, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0", exitCode)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	text := stdout.String()
	required := []string{
		"ralphx doctor",
		"BIN_DIR=" + filepath.Join(home, ".local", "bin"),
		"[ok] bash -> " + filepath.Join(binDir, "bash"),
		"[missing] PATH does not contain " + filepath.Join(home, ".local", "bin"),
	}
	for _, needle := range required {
		if !strings.Contains(text, needle) {
			t.Fatalf("output missing %q:\n%s", needle, text)
		}
	}
}

func TestRunJSONReportsStableShape(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	binDir := t.TempDir()
	makeExecutables(t, binDir, "bash", "python3", "git", "gh", "codex", "jq")
	t.Setenv("PATH", binDir)

	var stdout, stderr bytes.Buffer
	exitCode := run([]string{"--json"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0", exitCode)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	var got Result
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal json: %v\noutput: %s", err, stdout.String())
	}
	if got.Tool != "doctor" {
		t.Fatalf("Tool = %q, want doctor", got.Tool)
	}
	if got.ExitCode != 0 {
		t.Fatalf("ExitCode = %d, want 0", got.ExitCode)
	}
	if got.Ok != true {
		t.Fatalf("Ok = %v, want true", got.Ok)
	}
	if got.PathContainsBinDir {
		t.Fatalf("PathContainsBinDir = true, want false")
	}
	if len(got.Checks) != 6 {
		t.Fatalf("len(Checks) = %d, want 6", len(got.Checks))
	}
}

func makeExecutables(t *testing.T, dir string, names ...string) {
	t.Helper()
	for _, name := range names {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
			t.Fatalf("write executable %s: %v", name, err)
		}
	}
}
