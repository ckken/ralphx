package assets

import (
	_ "embed"
	"os"
	"path/filepath"
)

//go:embed prompts/loop-system-prompt.md
var defaultPrompt string

//go:embed schemas/loop-output.schema.json
var defaultSchema []byte

//go:embed schemas/plan-output.schema.json
var defaultPlanSchema []byte

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
