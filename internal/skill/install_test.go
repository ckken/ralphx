package skill

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstallUserScopeWritesFiles(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	dst, err := Install(t.TempDir(), false)
	if err != nil {
		t.Fatalf("install user scope: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "SKILL.md")); err != nil {
		t.Fatalf("missing SKILL.md: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "agents", "openai.yaml")); err != nil {
		t.Fatalf("missing openai.yaml: %v", err)
	}
}
