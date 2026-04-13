package cli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ckken/ralphx/internal/assets"
	"github.com/ckken/ralphx/internal/config"
	"github.com/ckken/ralphx/internal/current"
	"github.com/ckken/ralphx/internal/doctor"
	"github.com/ckken/ralphx/internal/hooks"
	"github.com/ckken/ralphx/internal/plan"
	"github.com/ckken/ralphx/internal/runner"
	"github.com/ckken/ralphx/internal/skill"
	"github.com/ckken/ralphx/internal/version"
)

func Main(args []string) int {
	if len(args) == 0 {
		printUsage()
		return 0
	}

	command, rest := normalizeCommand(args)
	switch command {
	case "help", "-h", "--help":
		printUsage()
		return 0
	case "version":
		fmt.Println(version.String())
		return 0
	case "current":
		return current.Main(os.Stdout)
	case "doctor":
		return doctor.Run(os.Stdout)
	case "hook":
		return hookMain(rest)
	case "plan":
		return planMain(rest)
	case "replan":
		return replanMain(rest)
	case "skill":
		return skillMain(rest)
	case "run":
		return run(rest)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", command)
		printUsage()
		return 1
	}
}

func run(args []string) int {
	cfg, helpShown, err := config.ParseRunArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "run argument error: %v\n", err)
		return 2
	}
	if helpShown {
		return 0
	}
	if strings.TrimSpace(cfg.TaskFile) == "" {
		fmt.Fprintln(os.Stderr, "missing required --task")
		return 2
	}

	loop := runner.New(cfg)
	if err := loop.Run(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func normalizeCommand(args []string) (string, []string) {
	if len(args) == 0 {
		return "help", nil
	}
	first := args[0]
	switch first {
	case "run", "doctor", "version", "current", "hook", "plan", "replan", "skill", "help", "-h", "--help":
		return first, args[1:]
	default:
		if strings.HasPrefix(first, "-") {
			return "run", args
		}
		return first, args[1:]
	}
}

func printUsage() {
	fmt.Println("ralphx - lean outer-loop runner for Codex and coding agents")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  ralphx run --task FILE [--checklist FILE] [--workdir DIR] [--resume] [--session-expiry DURATION]")
	fmt.Println("  ralphx doctor")
	fmt.Println("  ralphx hook stop-guard --task FILE [--checklist FILE]")
	fmt.Println("  ralphx plan --goal TEXT --out FILE [--execute]")
	fmt.Println("  ralphx replan --task FILE [--execute]")
	fmt.Println("  ralphx skill install [--project]")
	fmt.Println("  ralphx version")
	fmt.Println("  ralphx current")
	fmt.Println()
	fmt.Println("Compatibility:")
	fmt.Println("  ralphx --task FILE ...      same as: ralphx run --task FILE ...")
}

func skillMain(args []string) int {
	if len(args) == 0 {
		printSkillUsage()
		return 0
	}

	switch args[0] {
	case "help", "-h", "--help":
		printSkillUsage()
		return 0
	case "install":
		return skillInstall(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown skill command: %s\n\n", args[0])
		printSkillUsage()
		return 1
	}
}

func hookMain(args []string) int {
	if len(args) == 0 {
		printHookUsage()
		return 0
	}
	switch args[0] {
	case "help", "-h", "--help":
		printHookUsage()
		return 0
	case "install":
		return hookInstall(args[1:])
	case "uninstall":
		return hookUninstall(args[1:])
	case "prompt-submit":
		return hookPromptSubmit(args[1:])
	case "stop-guard":
		return hookStopGuard(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown hook command: %s\n\n", args[0])
		printHookUsage()
		return 1
	}
}

func hookStopGuard(args []string) int {
	fs := flag.NewFlagSet("hook stop-guard", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	taskPath := fs.String("task", "", "Task markdown file")
	checklistPath := fs.String("checklist", "", "Checklist markdown file")
	workdir := fs.String("workdir", envOr("WORKDIR", mustGetwd()), "Working directory")
	stateDir := fs.String("state-dir", "", "Override state directory (default <workdir>/.ralphx)")
	testsRequired := fs.Bool("tests-required", false, "Require passing verification before allowing stop")
	testsPassed := fs.Bool("tests-passed", false, "Indicate the latest verification passed")
	jsonOut := fs.Bool("json", true, "Print JSON output")
	help := fs.Bool("help", false, "Show help")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "hook stop-guard argument error: %v\n", err)
		return 2
	}
	if *help {
		printHookUsage()
		return 0
	}
	guardStateDir := *stateDir
	if strings.TrimSpace(guardStateDir) == "" {
		guardStateDir = filepath.Join(*workdir, ".ralphx")
	}
	paths := statePathsForHook(*workdir, guardStateDir)
	input, err := hooks.LoadStopGuardInput(*taskPath, *checklistPath, paths.summaryPath, paths.statePath, paths.lastResultPath, *testsRequired, *testsPassed)
	if err != nil {
		if err == hooks.ErrNoTaskContext {
			decision := hooks.Decision{
				Allow:   true,
				Reason:  "no_task_context",
				Message: "No ralphx task context found in the current workspace; skipping stop guard.",
			}
			hookLogDir := filepath.Join(guardStateDir, "logs")
			_ = hooks.AppendLog(hookLogDir, hooks.LogEntry{
				Event:    hooks.EventStop,
				Decision: decision,
				Result:   map[string]any{"workdir": *workdir},
			})
			_ = hooks.AppendUserLog(hooks.LogEntry{
				Event:    hooks.EventStop,
				Decision: decision,
				Result:   map[string]any{"workdir": *workdir},
			})
			fmt.Fprintf(os.Stderr, "[hook stop-guard] allow: %s\n", decision.Message)
			if *jsonOut {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				_ = enc.Encode(decision)
			} else {
				fmt.Fprintln(os.Stdout, "allow")
			}
			return 0
		}
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	decision := hooks.EvaluateStopGuard(hooks.GuardConfig{
		Enabled:                   true,
		BlockWhenChecklistOpen:    true,
		BlockWhenVerificationMiss: true,
		BlockWhenIncomplete:       true,
	}, input)
	hookLogDir := filepath.Join(guardStateDir, "logs")
	_ = hooks.AppendLog(hookLogDir, hooks.LogEntry{
		Event:         hooks.EventStop,
		TaskPath:      *taskPath,
		ChecklistPath: *checklistPath,
		Decision:      decision,
		Result:        input.Result,
	})
	_ = hooks.AppendUserLog(hooks.LogEntry{
		Event:         hooks.EventStop,
		TaskPath:      *taskPath,
		ChecklistPath: *checklistPath,
		Decision:      decision,
		Result:        input.Result,
	})
	if decision.Allow {
		fmt.Fprintf(os.Stderr, "[hook stop-guard] allow: %s\n", firstNonEmpty(decision.Message, "stop allowed"))
	} else {
		fmt.Fprintf(os.Stderr, "[hook stop-guard] block (%s): %s\n", decision.Reason, decision.Message)
	}
	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(decision)
	} else if decision.Allow {
		fmt.Fprintln(os.Stdout, "allow")
	} else {
		fmt.Fprintf(os.Stdout, "block: %s - %s\n", decision.Reason, decision.Message)
	}
	if decision.Allow {
		return 0
	}
	return 3
}

func hookPromptSubmit(args []string) int {
	fs := flag.NewFlagSet("hook prompt-submit", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	workdir := fs.String("workdir", "", "Working directory")
	stateDir := fs.String("state-dir", "", "Override state directory (default <workdir>/.ralphx)")
	payloadPath := fs.String("payload", "", "Path to a JSON payload file")
	jsonOut := fs.Bool("json", true, "Print JSON output")
	help := fs.Bool("help", false, "Show help")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "hook prompt-submit argument error: %v\n", err)
		return 2
	}
	if *help {
		printHookUsage()
		return 0
	}
	payload, err := hooks.LoadPromptSubmitPayload(*payloadPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	text := hooks.PromptText(payload)
	active := hooks.PromptActivatesRalphx(text)
	effectiveWorkdir := strings.TrimSpace(*workdir)
	if effectiveWorkdir == "" {
		effectiveWorkdir = strings.TrimSpace(payload.Cwd)
	}
	if effectiveWorkdir == "" {
		effectiveWorkdir = mustGetwd()
	}
	guardStateDir := *stateDir
	if strings.TrimSpace(guardStateDir) == "" {
		guardStateDir = filepath.Join(effectiveWorkdir, ".ralphx")
	}
	hookLogDir := filepath.Join(guardStateDir, "logs")
	decision := hooks.Decision{
		Allow:   true,
		Reason:  "prompt_submit",
		Message: "ralphx inactive",
	}
	if active {
		decision.Message = "ralphx mode active"
	}
	_ = hooks.AppendLog(hookLogDir, hooks.LogEntry{
		Event:    hooks.EventPromptSubmit,
		TaskPath: "",
		Decision: decision,
		Result: map[string]any{
			"active":  active,
			"text":    text,
			"workdir": effectiveWorkdir,
		},
	})
	_ = hooks.AppendUserLog(hooks.LogEntry{
		Event:    hooks.EventPromptSubmit,
		TaskPath: "",
		Decision: decision,
		Result: map[string]any{
			"active":  active,
			"text":    text,
			"workdir": effectiveWorkdir,
		},
	})
	if active {
		fmt.Fprintln(os.Stderr, "[hook prompt-submit] ralphx mode active")
	}
	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		payload := map[string]any{
			"active":  active,
			"message": decision.Message,
		}
		if active {
			payload = map[string]any{
				"hookSpecificOutput": map[string]any{
					"hookEventName":     "UserPromptSubmit",
					"additionalContext": "ralphx workflow active. Continue the current branch of work; do not stop at advice alone. Execute one bounded next step or emit a concrete next-step plan.",
				},
			}
		}
		_ = enc.Encode(payload)
	}
	return 0
}

func hookInstall(args []string) int {
	fs := flag.NewFlagSet("hook install", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	help := fs.Bool("help", false, "Show help")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "hook install argument error: %v\n", err)
		return 2
	}
	if *help {
		printHookUsage()
		return 0
	}
	path, err := hooks.InstallUserStopHook()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Fprintf(os.Stdout, "installed hooks: %s\n", path)
	return 0
}

func hookUninstall(args []string) int {
	fs := flag.NewFlagSet("hook uninstall", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	help := fs.Bool("help", false, "Show help")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "hook uninstall argument error: %v\n", err)
		return 2
	}
	if *help {
		printHookUsage()
		return 0
	}
	path, err := hooks.UninstallUserStopHook()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Fprintf(os.Stdout, "removed managed hooks from: %s\n", path)
	return 0
}

func planMain(args []string) int {
	fs := flag.NewFlagSet("plan", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	goal := fs.String("goal", "", "Goal statement to convert into a task and checklist")
	out := fs.String("out", "", "Task markdown file to write")
	workdir := fs.String("workdir", envOr("WORKDIR", mustGetwd()), "Working directory")
	execute := fs.Bool("execute", false, "Run the generated task immediately after writing files")
	testsCmd := fs.String("tests-cmd", os.Getenv("TESTS_CMD"), "Validation command to use when --execute is set")
	codexBin := fs.String("codex-bin", envOr("CODEX_CMD", "codex"), "Codex executable name/path")
	codexArgs := fs.String("codex-args", os.Getenv("CODEX_ARGS"), "Extra args passed to codex for planning")
	stateDir := fs.String("state-dir", "", "Override state directory for --execute")
	help := fs.Bool("help", false, "Show help")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "plan argument error: %v\n", err)
		return 2
	}
	if *help {
		printPlanUsage()
		return 0
	}
	if strings.TrimSpace(*goal) == "" {
		fmt.Fprintln(os.Stderr, "missing required --goal")
		return 2
	}
	if strings.TrimSpace(*out) == "" {
		fmt.Fprintln(os.Stderr, "missing required --out")
		return 2
	}
	outPath := *out
	if !filepath.IsAbs(outPath) {
		outPath = filepath.Join(*workdir, outPath)
	}

	planStateDir := *stateDir
	if strings.TrimSpace(planStateDir) == "" {
		planStateDir = filepath.Join(*workdir, ".ralphx")
	}
	schemaPath, err := assets.EnsurePlanSchemaFile(planStateDir, "")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	rawLogPath := filepath.Join(planStateDir, "logs", "plan.log")
	if err := os.MkdirAll(filepath.Dir(rawLogPath), 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	outcome, _, err := plan.Run(context.Background(), plan.Request{
		Goal:             *goal,
		Workdir:          *workdir,
		CodexCmd:         *codexBin,
		CodexArgs:        splitArgs(*codexArgs),
		OutputSchemaPath: schemaPath,
		RawLogPath:       rawLogPath,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	taskPath, checklistPath, err := plan.WriteFiles(outPath, outcome)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	fmt.Fprintf(os.Stdout, "planned task: %s\n", taskPath)
	fmt.Fprintf(os.Stdout, "planned checklist: %s\n", checklistPath)

	if !*execute {
		return 0
	}

	cfg := config.RunConfig{
		TaskFile:      taskPath,
		ChecklistFile: checklistPath,
		Workdir:       *workdir,
		TestsCmd:      firstNonEmpty(strings.TrimSpace(*testsCmd), strings.TrimSpace(outcome.TestsCmd)),
		CodexCmd:      *codexBin,
		CodexArgs:     splitArgs(*codexArgs),
		StateDir:      planStateDir,
		Workers:       envInt("RALPHX_WORKERS", 1),
		MaxIterations: envInt("MAX_ITERATIONS", 30),
		MaxNoProgress: envInt("MAX_NO_PROGRESS", 3),
		RoundTimeout:  envDurationSeconds("ROUND_TIMEOUT_SECONDS", 1800),
		ResumeSession: envBool("RALPHX_RESUME_SESSION", false),
		SessionExpiry: envDurationHours("SESSION_EXPIRY_HOURS", 24),
	}
	if err := runner.New(cfg).Run(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func replanMain(args []string) int {
	fs := flag.NewFlagSet("replan", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	taskPath := fs.String("task", "", "Existing task markdown file to replan")
	checklistPath := fs.String("checklist", "", "Checklist markdown file")
	workdir := fs.String("workdir", envOr("WORKDIR", mustGetwd()), "Working directory")
	execute := fs.Bool("execute", false, "Run the replanned task immediately after writing files")
	testsCmd := fs.String("tests-cmd", os.Getenv("TESTS_CMD"), "Validation command to use when --execute is set")
	codexBin := fs.String("codex-bin", envOr("CODEX_CMD", "codex"), "Codex executable name/path")
	codexArgs := fs.String("codex-args", os.Getenv("CODEX_ARGS"), "Extra args passed to codex for replanning")
	stateDir := fs.String("state-dir", "", "Override state directory")
	preserveCompleted := fs.Bool("preserve-completed", true, "Preserve completed checklist items in the regenerated checklist")
	help := fs.Bool("help", false, "Show help")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "replan argument error: %v\n", err)
		return 2
	}
	if *help {
		printReplanUsage()
		return 0
	}
	if strings.TrimSpace(*taskPath) == "" {
		fmt.Fprintln(os.Stderr, "missing required --task")
		return 2
	}

	replanStateDir := *stateDir
	if strings.TrimSpace(replanStateDir) == "" {
		replanStateDir = filepath.Join(*workdir, ".ralphx")
	}
	summaryPath, statePath := plan.DefaultReplanPaths(*workdir, replanStateDir)
	schemaPath, err := assets.EnsurePlanSchemaFile(replanStateDir, "")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	rawLogPath := filepath.Join(replanStateDir, "logs", "replan.log")
	if err := plan.EnsureLogDir(rawLogPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	outcome, _, _, err := plan.Replan(context.Background(), plan.ReplanRequest{
		TaskPath:          *taskPath,
		ChecklistPath:     *checklistPath,
		SummaryPath:       summaryPath,
		StatePath:         statePath,
		Workdir:           *workdir,
		CodexCmd:          *codexBin,
		CodexArgs:         splitArgs(*codexArgs),
		OutputSchemaPath:  schemaPath,
		RawLogPath:        rawLogPath,
		PreserveCompleted: *preserveCompleted,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	taskAbs := *taskPath
	if !filepath.IsAbs(taskAbs) {
		taskAbs = filepath.Join(*workdir, taskAbs)
	}
	taskFile, checklistFile, err := plan.WriteFiles(taskAbs, outcome)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	fmt.Fprintf(os.Stdout, "replanned task: %s\n", taskFile)
	fmt.Fprintf(os.Stdout, "replanned checklist: %s\n", checklistFile)

	if !*execute {
		return 0
	}

	cfg := config.RunConfig{
		TaskFile:      taskFile,
		ChecklistFile: checklistFile,
		Workdir:       *workdir,
		TestsCmd:      firstNonEmpty(strings.TrimSpace(*testsCmd), strings.TrimSpace(outcome.TestsCmd)),
		CodexCmd:      *codexBin,
		CodexArgs:     splitArgs(*codexArgs),
		StateDir:      replanStateDir,
		Workers:       envInt("RALPHX_WORKERS", 1),
		MaxIterations: envInt("MAX_ITERATIONS", 30),
		MaxNoProgress: envInt("MAX_NO_PROGRESS", 3),
		RoundTimeout:  envDurationSeconds("ROUND_TIMEOUT_SECONDS", 1800),
		ResumeSession: envBool("RALPHX_RESUME_SESSION", false),
		SessionExpiry: envDurationHours("SESSION_EXPIRY_HOURS", 24),
	}
	if err := runner.New(cfg).Run(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func skillInstall(args []string) int {
	fs := flag.NewFlagSet("skill install", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	project := fs.Bool("project", false, "Install to the current repo ./.codex/skills directory")
	help := fs.Bool("help", false, "Show help")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "skill install argument error: %v\n", err)
		return 2
	}
	if *help {
		printSkillUsage()
		return 0
	}

	workdir, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	path, err := skill.Install(workdir, *project)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Fprintf(os.Stdout, "installed skill: %s\n", path)
	return 0
}

func printSkillUsage() {
	fmt.Println("Usage:")
	fmt.Println("  ralphx skill install [--project]")
	fmt.Println()
	fmt.Println("Defaults:")
	fmt.Println("  Without --project, installs to ~/.codex/skills/ralphx")
	fmt.Println("  With --project, installs to ./.codex/skills/ralphx")
}

func printPlanUsage() {
	fmt.Println("Usage:")
	fmt.Println("  ralphx plan --goal TEXT --out FILE [--execute]")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  ralphx plan --goal \"add health endpoint\" --out tasks/health.md")
	fmt.Println("  ralphx plan --goal \"finish migration\" --out tasks/migration.md --execute")
}

func printReplanUsage() {
	fmt.Println("Usage:")
	fmt.Println("  ralphx replan --task FILE [--execute]")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  ralphx replan --task tasks/migration.md")
	fmt.Println("  ralphx replan --task tasks/migration.md --execute")
}

func printHookUsage() {
	fmt.Println("Usage:")
	fmt.Println("  ralphx hook stop-guard --task FILE [--checklist FILE]")
	fmt.Println("  ralphx hook prompt-submit [--payload FILE]")
	fmt.Println("  ralphx hook install")
	fmt.Println("  ralphx hook uninstall")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  ralphx hook stop-guard --task tasks/demo.md --checklist tasks/demo.checklist.md")
	fmt.Println("  ralphx hook stop-guard --task tasks/demo.md --tests-required --tests-passed")
}

type hookStatePaths struct {
	summaryPath    string
	statePath      string
	lastResultPath string
}

func statePathsForHook(workdir, stateDir string) hookStatePaths {
	root := stateDir
	if strings.TrimSpace(root) == "" {
		root = filepath.Join(workdir, ".ralphx")
	}
	return hookStatePaths{
		summaryPath:    filepath.Join(root, "summary.txt"),
		statePath:      filepath.Join(root, "state.json"),
		lastResultPath: filepath.Join(root, "last-result.json"),
	}
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
	var parsed int
	if _, err := fmt.Sscanf(value, "%d", &parsed); err != nil {
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
	var parsed int
	if _, err := fmt.Sscanf(value, "%d", &parsed); err != nil {
		return time.Duration(fallbackSeconds) * time.Second
	}
	return time.Duration(parsed) * time.Second
}

func envDurationHours(key string, fallbackHours int) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return time.Duration(fallbackHours) * time.Hour
	}
	var parsed int
	if _, err := fmt.Sscanf(value, "%d", &parsed); err != nil {
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

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func mustGetwd() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}
