package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type RunConfig struct {
	TaskFile         string
	ChecklistFile    string
	Workdir          string
	PromptFile       string
	OutputSchemaFile string
	TestsCmd         string
	CodexCmd         string
	CodexArgs        []string
	StateDir         string
	Workers          int
	MaxIterations    int
	MaxNoProgress    int
	RoundTimeout     time.Duration
}

func ParseRunArgs(args []string) (RunConfig, bool, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return RunConfig{}, false, err
	}

	cfg := RunConfig{
		Workdir:          envOr("WORKDIR", cwd),
		ChecklistFile:    os.Getenv("CHECKLIST_FILE"),
		PromptFile:       os.Getenv("PROMPT_FILE"),
		OutputSchemaFile: os.Getenv("OUTPUT_SCHEMA_FILE"),
		TestsCmd:         os.Getenv("TESTS_CMD"),
		CodexCmd:         envOr("CODEX_CMD", "codex"),
		CodexArgs:        splitArgs(os.Getenv("CODEX_ARGS")),
		Workers:          envInt("RALPHX_WORKERS", 1),
		MaxIterations:    envInt("MAX_ITERATIONS", 30),
		MaxNoProgress:    envInt("MAX_NO_PROGRESS", 3),
		RoundTimeout:     envDurationSeconds("ROUND_TIMEOUT_SECONDS", 1800),
	}

	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.StringVar(&cfg.TaskFile, "task", "", "Task markdown file")
	fs.StringVar(&cfg.ChecklistFile, "checklist", cfg.ChecklistFile, "Checklist markdown file")
	fs.StringVar(&cfg.Workdir, "workdir", cfg.Workdir, "Working directory")
	fs.StringVar(&cfg.PromptFile, "prompt", cfg.PromptFile, "Prompt template path")
	fs.StringVar(&cfg.OutputSchemaFile, "schema", cfg.OutputSchemaFile, "Output schema path")
	fs.StringVar(&cfg.TestsCmd, "tests-cmd", cfg.TestsCmd, "Validation command to run after successful rounds")
	fs.StringVar(&cfg.CodexCmd, "codex-bin", cfg.CodexCmd, "Codex executable name/path")
	fs.IntVar(&cfg.Workers, "workers", cfg.Workers, "Worker count for future parallel mode; 1 keeps current single-runner behavior")
	fs.IntVar(&cfg.MaxIterations, "max-iterations", cfg.MaxIterations, "Maximum iterations; 0 means unlimited")
	fs.IntVar(&cfg.MaxNoProgress, "max-no-progress", cfg.MaxNoProgress, "Maximum no-progress rounds; 0 means unlimited")
	timeout := fs.Duration("round-timeout", cfg.RoundTimeout, "Per-round timeout")
	codexArgs := fs.String("codex-args", strings.Join(cfg.CodexArgs, " "), "Extra args passed to codex")
	stateDir := fs.String("state-dir", "", "Override state directory (default <workdir>/.ralphx)")
	help := fs.Bool("help", false, "Show help")

	if err := fs.Parse(args); err != nil {
		return RunConfig{}, false, err
	}
	if *help {
		printRunUsage()
		return RunConfig{}, true, nil
	}

	cfg.RoundTimeout = *timeout
	cfg.CodexArgs = splitArgs(*codexArgs)
	if *stateDir != "" {
		cfg.StateDir = *stateDir
	} else if cfg.StateDir == "" {
		cfg.StateDir = filepath.Join(cfg.Workdir, ".ralphx")
	}

	return cfg, false, nil
}

func printRunUsage() {
	fmt.Println("Usage:")
	fmt.Println("  ralphx run --task FILE [--checklist FILE] [--workdir DIR]")
	fmt.Println()
	fmt.Println("Compatibility:")
	fmt.Println("  ralphx --task FILE ... maps to `ralphx run --task FILE ...`")
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envDurationSeconds(key string, fallbackSeconds int) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return time.Duration(fallbackSeconds) * time.Second
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return time.Duration(fallbackSeconds) * time.Second
	}
	return time.Duration(parsed) * time.Second
}

func splitArgs(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return strings.Fields(value)
}
