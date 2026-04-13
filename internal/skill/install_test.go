package skill

import (
	"os"
	"path/filepath"
	"strings"
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

func TestInstalledSkillContainsContinuationDiscipline(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	dst, err := Install(t.TempDir(), false)
	if err != nil {
		t.Fatalf("install user scope: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dst, "SKILL.md"))
	if err != nil {
		t.Fatalf("read installed SKILL.md: %v", err)
	}
	text := string(data)
	required := []string{
		"Do not stop after recommending the next step.",
		"You must either produce the next concrete plan or directly execute one bounded next step.",
	}
	for _, needle := range required {
		if !strings.Contains(text, needle) {
			t.Fatalf("installed skill missing %q", needle)
		}
	}
}

func TestRepositorySkillContainsContinuationDiscipline(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "skills", "ralphx", "SKILL.md"))
	if err != nil {
		t.Fatalf("read repository SKILL.md: %v", err)
	}
	text := string(data)
	required := []string{
		"Do not stop after recommending the next step.",
		"You must either produce the next concrete plan or directly execute one bounded next step.",
	}
	for _, needle := range required {
		if !strings.Contains(text, needle) {
			t.Fatalf("repository skill missing %q", needle)
		}
	}
}
