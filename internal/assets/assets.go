package assets

import (
	"embed"
	_ "embed"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed prompts/loop-system-prompt.md
var defaultPrompt string

//go:embed schemas/loop-output.schema.json
var defaultSchema []byte

//go:embed schemas/plan-output.schema.json
var defaultPlanSchema []byte

//go:embed subagents/*.toml
var subagentFS embed.FS

func DefaultPrompt() string {
	return defaultPrompt
}

func EnsureSchemaFile(rootDir, overridePath string) (string, error) {
	if overridePath != "" {
		return overridePath, nil
	}
	runtimeDir := filepath.Join(rootDir, "runtime")
	if err := os.MkdirAll(runtimeDir, 0o755); err != nil {
		return "", err
	}
	target := filepath.Join(runtimeDir, "loop-output.schema.json")
	if err := os.WriteFile(target, defaultSchema, 0o644); err != nil {
		return "", err
	}
	return target, nil
}

func EnsurePlanSchemaFile(rootDir, overridePath string) (string, error) {
	if overridePath != "" {
		return overridePath, nil
	}
	runtimeDir := filepath.Join(rootDir, "runtime")
	if err := os.MkdirAll(runtimeDir, 0o755); err != nil {
		return "", err
	}
	target := filepath.Join(runtimeDir, "plan-output.schema.json")
	if err := os.WriteFile(target, defaultPlanSchema, 0o644); err != nil {
		return "", err
	}
	return target, nil
}

func SubagentBytes(name string) ([]byte, bool) {
	data, err := subagentFS.ReadFile("subagents/" + name + ".toml")
	if err != nil {
		return nil, false
	}
	return data, true
}

func SubagentNames() ([]string, error) {
	entries, err := fs.ReadDir(subagentFS, "subagents")
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if filepath.Ext(name) != ".toml" {
			continue
		}
		out = append(out, strings.TrimSuffix(name, ".toml"))
	}
	return out, nil
}
