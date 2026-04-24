package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestAgentsListJsonIncludesCuratedCatalog(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	workdir := t.TempDir()

	stdout, stderr := captureCLIOutput(t, func() int {
		return agentsList([]string{"--workdir", workdir, "--json"})
	})
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	var got struct {
		Catalog []any `json:"catalog"`
	}
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("unmarshal stdout: %v\noutput: %s", err, stdout)
	}
	if len(got.Catalog) != 8 {
		t.Fatalf("len(catalog) = %d, want 8", len(got.Catalog))
	}
}

func TestAgentsInstallProjectWritesFiles(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	workdir := t.TempDir()

	stdout, stderr := captureCLIOutput(t, func() int {
		return agentsInstall([]string{"--workdir", workdir, "--project"})
	})
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	if stdout == "" {
		t.Fatalf("stdout empty, want install summary")
	}
	for _, name := range []string{"workflow-orchestrator", "reviewer", "backend-developer"} {
		if _, err := os.Stat(filepath.Join(workdir, ".codex", "agents", name+".toml")); err != nil {
			t.Fatalf("missing %s: %v", name, err)
		}
	}
}
