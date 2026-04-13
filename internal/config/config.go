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
	ResumeSession    bool
	SessionExpiry    time.Duration
	AutoReplan       bool
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
		ResumeSession:    envBool("RALPHX_RESUME_SESSION", false),
		SessionExpiry:    envDurationHours("SESSION_EXPIRY_HOURS", 24),
		AutoReplan:       envBool("RALPHX_AUTO_REPLAN", true),
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
	sessionExpiry := fs.Duration("session-expiry", cfg.SessionExpiry, "Session expiration window before starting a fresh session")
	codexArgs := fs.String("codex-args", strings.Join(cfg.CodexArgs, " "), "Extra args passed to codex")
	stateDir := fs.String("state-dir", "", "Override state directory (default <workdir>/.ralphx)")
	fs.BoolVar(&cfg.ResumeSession, "resume", cfg.ResumeSession, "Resume the previous Codex session when available")
	fs.BoolVar(&cfg.AutoReplan, "auto-replan", cfg.AutoReplan, "Automatically regenerate task/checklist when blocked or stale")
	help := fs.Bool("help", false, "Show help")

	if err := fs.Parse(args); err != nil {
		return RunConfig{}, false, err
	}
	if *help {
		printRunUsage()
		return RunConfig{}, true, nil
	}

	cfg.RoundTimeout = *timeout
	cfg.SessionExpiry = *sessionExpiry
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
	fmt.Println("  ralphx run --task FILE [--checklist FILE] [--workdir DIR] [--resume] [--session-expiry DURATION] [--auto-replan]")
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

func envBool(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	switch strings.ToLower(value) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
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

func envDurationHours(key string, fallbackHours int) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return time.Duration(fallbackHours) * time.Hour
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return time.Duration(fallbackHours) * time.Hour
	}
	return time.Duration(parsed) * time.Hour
}

func splitArgs(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return strings.Fields(value)
}
