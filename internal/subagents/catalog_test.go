package subagents

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstallUserScopeWritesCuratedAgents(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	workdir := t.TempDir()

	result, err := Install(workdir, false, []string{"workflow-orchestrator", "reviewer"})
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if result.Scope != "global" {
		t.Fatalf("Scope = %q, want global", result.Scope)
	}
	if len(result.Installed) != 2 {
		t.Fatalf("len(Installed) = %d, want 2", len(result.Installed))
	}
	for _, name := range []string{"workflow-orchestrator", "reviewer"} {
		if _, err := os.Stat(filepath.Join(home, ".codex", "agents", name+".toml")); err != nil {
			t.Fatalf("missing %s: %v", name, err)
		}
	}
}

func TestDiscoverFindsProjectAndGlobalAgents(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	workdir := t.TempDir()

	if _, err := Install(workdir, false, []string{"reviewer"}); err != nil {
		t.Fatalf("global install: %v", err)
	}
	if _, err := Install(workdir, true, []string{"backend-developer"}); err != nil {
		t.Fatalf("project install: %v", err)
	}

	got, err := Discover(workdir)
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}
	if got.ProjectDir != filepath.Join(workdir, ".codex", "agents") {
		t.Fatalf("ProjectDir = %q, want %q", got.ProjectDir, filepath.Join(workdir, ".codex", "agents"))
	}
	assertStatus := func(name string, wantInstalled bool, wantScope string) {
		t.Helper()
		for _, status := range got.Catalog {
			if status.Spec.Name != name {
				continue
			}
			if status.Installed != wantInstalled {
				t.Fatalf("%s Installed = %v, want %v", name, status.Installed, wantInstalled)
			}
			if status.Effective == nil {
				t.Fatalf("%s Effective = nil", name)
			}
			if status.Effective.Scope != wantScope {
				t.Fatalf("%s scope = %q, want %q", name, status.Effective.Scope, wantScope)
			}
			return
		}
		t.Fatalf("missing catalog status for %s", name)
	}
	assertStatus("reviewer", true, "global")
	assertStatus("backend-developer", true, "project")
}
