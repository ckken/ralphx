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
	"github.com/ckken/ralphx/internal/state"
	"github.com/ckken/ralphx/internal/subagents"
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
		return doctor.Main(rest)
	case "hook":
		return hookMain(rest)
	case "plan":
		return planMain(rest)
	case "replan":
		return replanMain(rest)
	case "skill":
		return skillMain(rest)
	case "agents":
		return agentsMain(rest)
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
	case "run", "doctor", "version", "current", "hook", "plan", "replan", "skill", "agents", "help", "-h", "--help":
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
	fmt.Println("  ralphx hook native --event Stop|UserPromptSubmit [--payload FILE] [--task FILE]")
	fmt.Println("  ralphx hook stop-guard --task FILE [--checklist FILE]")
	fmt.Println("  ralphx plan --goal TEXT --out FILE [--execute]")
	fmt.Println("  ralphx replan --task FILE [--execute]")
	fmt.Println("  ralphx skill install [--project]")
	fmt.Println("  ralphx agents list|discover [--json]")
	fmt.Println("  ralphx agents install [NAME...] [--project] [--json]")
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

func agentsMain(args []string) int {
	if len(args) == 0 {
		return agentsList(nil)
	}

	switch args[0] {
	case "help", "-h", "--help":
		printAgentsUsage()
		return 0
	case "list", "discover":
		return agentsList(args[1:])
	case "install":
		return agentsInstall(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown agents command: %s\n\n", args[0])
		printAgentsUsage()
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
	case "status":
		return hookStatus(args[1:])
	case "uninstall":
		return hookUninstall(args[1:])
	case "native":
		return hookNative(args[1:])
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

func hookNative(args []string) int {
	fs := flag.NewFlagSet("hook native", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	event := fs.String("event", "", "Native hook event name")
	taskPath := fs.String("task", "", "Task markdown file")
	checklistPath := fs.String("checklist", "", "Checklist markdown file")
	workdir := fs.String("workdir", envOr("WORKDIR", mustGetwd()), "Working directory")
	stateDir := fs.String("state-dir", "", "Override state directory (default <workdir>/.ralphx)")
	summaryPath := fs.String("summary", "", "Summary file")
	statePath := fs.String("state-path", "", "Run state file")
	lastResultPath := fs.String("last-result", "", "Last result file")
	payloadPath := fs.String("payload", "", "Path to a JSON payload file")
	testsRequired := fs.Bool("tests-required", false, "Require passing verification before allowing stop")
	testsPassed := fs.Bool("tests-passed", false, "Indicate the latest verification passed")
	jsonOut := fs.Bool("json", true, "Print JSON output")
	nativeJSON := fs.Bool("native-json", false, "Emit native Codex Stop-hook JSON shape")
	help := fs.Bool("help", false, "Show help")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "hook native argument error: %v\n", err)
		return 2
	}
	if *help {
		printHookUsage()
		return 0
	}
	nativeEvent, ok := hooks.NormalizeNativeHookEvent(*event)
	if !ok {
		fmt.Fprintf(os.Stderr, "hook native argument error: unsupported event %q\n", *event)
		return 2
	}
	guardStateDir := resolveHookStateDir(*workdir, *stateDir)
	paths := statePathsForHook(*workdir, guardStateDir)
	resp, err := hooks.DispatchNativeHook(hooks.NativeHookRequest{
		Event:          nativeEvent,
		Workdir:        *workdir,
		StateDir:       guardStateDir,
		PayloadPath:    *payloadPath,
		TaskPath:       *taskPath,
		ChecklistPath:  *checklistPath,
		SummaryPath:    firstNonEmpty(strings.TrimSpace(*summaryPath), paths.summaryPath),
		StatePath:      firstNonEmpty(strings.TrimSpace(*statePath), paths.statePath),
		LastResultPath: firstNonEmpty(strings.TrimSpace(*lastResultPath), paths.lastResultPath),
		TestsRequired:  *testsRequired,
		TestsPassedNow: *testsPassed,
		NativeJSON:     *nativeJSON,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return emitNativeHookResponse(resp, *jsonOut)
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
	nativeJSON := fs.Bool("native-json", false, "Emit native Codex Stop-hook JSON shape")
	help := fs.Bool("help", false, "Show help")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "hook stop-guard argument error: %v\n", err)
		return 2
	}
	if *help {
		printHookUsage()
		return 0
	}
	guardStateDir := resolveHookStateDir(*workdir, *stateDir)
	paths := statePathsForHook(*workdir, guardStateDir)
	resp, err := hooks.DispatchNativeHook(hooks.NativeHookRequest{
		Event:          hooks.NativeHookEventStop,
		Workdir:        *workdir,
		StateDir:       guardStateDir,
		TaskPath:       *taskPath,
		ChecklistPath:  *checklistPath,
		SummaryPath:    paths.summaryPath,
		StatePath:      paths.statePath,
		LastResultPath: paths.lastResultPath,
		TestsRequired:  *testsRequired,
		TestsPassedNow: *testsPassed,
		NativeJSON:     *nativeJSON,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return emitNativeHookResponse(resp, *jsonOut)
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
	resp, err := hooks.DispatchNativeHook(hooks.NativeHookRequest{
		Event:       hooks.NativeHookEventUserPromptSubmit,
		Workdir:     *workdir,
		StateDir:    *stateDir,
		PayloadPath: *payloadPath,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return emitNativeHookResponse(resp, *jsonOut)
}

func emitNativeHookResponse(resp hooks.NativeHookResponse, jsonOut bool) int {
	if resp.Silent {
		return resp.ExitCode
	}
	if resp.Stderr != "" {
		fmt.Fprintln(os.Stderr, resp.Stderr)
	}
	if jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(resp.Stdout)
		return resp.ExitCode
	}
	if resp.Decision.Allow {
		fmt.Fprintln(os.Stdout, "allow")
	} else {
		fmt.Fprintf(os.Stdout, "block: %s - %s\n", resp.Decision.Reason, resp.Decision.Message)
	}
	return resp.ExitCode
}

func hookStatus(args []string) int {
	fs := flag.NewFlagSet("hook status", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	workdir := fs.String("workdir", envOr("WORKDIR", mustGetwd()), "Working directory")
	stateDir := fs.String("state-dir", "", "Override state directory (default <workdir>/.ralphx)")
	help := fs.Bool("help", false, "Show help")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "hook status argument error: %v\n", err)
		return 2
	}
	if *help {
		printHookUsage()
		return 0
	}
	root := *stateDir
	if strings.TrimSpace(root) == "" {
		root = filepath.Join(*workdir, ".ralphx")
	}
	runState, runStateErr := state.LoadRunState(state.DerivePaths(*workdir, root))
	repoLatest, repoErr := hooks.ReadLatest(filepath.Join(root, "last-hook-event.json"))
	activeState, activeErr := hooks.ReadActiveState(root)
	userLatest, userErr := hooks.ReadLatest(filepath.Join(os.Getenv("HOME"), ".codex", "log", "ralphx-last-hook-event.json"))
	installStatus, installErr := hooks.ReadUserHookInstallStatus()
	out := map[string]any{}
	if installErr == nil && installStatus.HooksFileFound {
		out["installed"] = installStatus
	}
	if runStateErr == nil {
		out["state"] = runState
	}
	if repoErr == nil {
		out["repo"] = repoLatest
	}
	if activeErr == nil {
		out["active"] = activeState
	}
	if userErr == nil {
		out["user"] = userLatest
	}
	if len(out) == 0 {
		fmt.Fprintln(os.Stderr, "no hook status found")
		return 1
	}
	fmt.Fprintln(os.Stderr, hookStatusSummary(out))
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(out)
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

	cfg := executionConfig(taskPath, checklistPath, *workdir, planStateDir, firstNonEmpty(strings.TrimSpace(*testsCmd), strings.TrimSpace(outcome.TestsCmd)), *codexBin, splitArgs(*codexArgs))
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

	cfg := executionConfig(taskFile, checklistFile, *workdir, replanStateDir, firstNonEmpty(strings.TrimSpace(*testsCmd), strings.TrimSpace(outcome.TestsCmd)), *codexBin, splitArgs(*codexArgs))
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

func executionConfig(taskFile, checklistFile, workdir, stateDir, testsCmd, codexBin string, codexArgs []string) config.RunConfig {
	return config.RunConfig{
		TaskFile:      taskFile,
		ChecklistFile: checklistFile,
		Workdir:       workdir,
		TestsCmd:      testsCmd,
		CodexCmd:      codexBin,
		CodexArgs:     codexArgs,
		StateDir:      stateDir,
		Workers:       envInt("RALPHX_WORKERS", 1),
		MaxIterations: envInt("MAX_ITERATIONS", 30),
		MaxNoProgress: envInt("MAX_NO_PROGRESS", 3),
		RoundTimeout:  envDurationSeconds("ROUND_TIMEOUT_SECONDS", 1800),
		ResumeSession: envBool("RALPHX_RESUME_SESSION", false),
		SessionExpiry: envDurationHours("SESSION_EXPIRY_HOURS", 24),
		AutoReplan:    envBool("RALPHX_AUTO_REPLAN", true),
	}
}

func agentsList(args []string) int {
	fs := flag.NewFlagSet("agents list", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	workdir := fs.String("workdir", envOr("WORKDIR", mustGetwd()), "Working directory")
	jsonOut := fs.Bool("json", false, "Print JSON output")
	help := fs.Bool("help", false, "Show help")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "agents list argument error: %v\n", err)
		return 2
	}
	if *help {
		printAgentsUsage()
		return 0
	}
	discovery, err := subagents.Discover(*workdir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(discovery)
		return 0
	}
	printAgentsDiscovery(os.Stdout, discovery)
	return 0
}

func agentsInstall(args []string) int {
	fs := flag.NewFlagSet("agents install", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	workdir := fs.String("workdir", envOr("WORKDIR", mustGetwd()), "Working directory")
	project := fs.Bool("project", false, "Install to the current repo ./.codex/agents directory")
	jsonOut := fs.Bool("json", false, "Print JSON output")
	help := fs.Bool("help", false, "Show help")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "agents install argument error: %v\n", err)
		return 2
	}
	if *help {
		printAgentsUsage()
		return 0
	}
	result, err := subagents.Install(*workdir, *project, fs.Args())
	if err != nil {
		if *jsonOut {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			_ = enc.Encode(map[string]any{
				"ok":     false,
				"error":  err.Error(),
				"result": result,
			})
		} else {
			fmt.Fprintln(os.Stderr, err)
		}
		return 1
	}
	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(result)
		return 0
	}
	printAgentsInstall(os.Stdout, result)
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

func printAgentsUsage() {
	fmt.Println("Usage:")
	fmt.Println("  ralphx agents list|discover [--workdir DIR] [--json]")
	fmt.Println("  ralphx agents install [NAME...] [--workdir DIR] [--project] [--json]")
	fmt.Println()
	fmt.Println("Defaults:")
	fmt.Println("  list/discover scans the curated catalog plus the current global and project agent dirs")
	fmt.Println("  install with no names installs the curated set")
	fmt.Println("  --project writes to ./.codex/agents instead of ~/.codex/agents")
}

func printAgentsDiscovery(w *os.File, d subagents.Discovery) {
	fmt.Fprintf(w, "ralphx agents discover\n")
	fmt.Fprintf(w, "workdir: %s\n", d.Workdir)
	fmt.Fprintf(w, "global: %s\n", d.GlobalDir)
	fmt.Fprintf(w, "project: %s\n\n", d.ProjectDir)
	for _, status := range d.Catalog {
		state := "missing"
		if status.Installed {
			state = "installed"
		}
		fmt.Fprintf(w, "- %s [%s] %s\n", status.Spec.Name, state, status.Spec.Description)
		for _, loc := range status.Locations {
			fmt.Fprintf(w, "  %s: %s\n", loc.Scope, loc.Path)
		}
	}
	if len(d.Unknown) > 0 {
		fmt.Fprintln(w, "\nunknown installed agents:")
		for _, item := range d.Unknown {
			fmt.Fprintf(w, "- %s\n", item.Name)
			for _, loc := range item.Locations {
				fmt.Fprintf(w, "  %s: %s\n", loc.Scope, loc.Path)
			}
		}
	}
}

func printAgentsInstall(w *os.File, r subagents.InstallResult) {
	fmt.Fprintf(w, "installed subagents to %s (%s)\n", r.TargetDir, r.Scope)
	for _, loc := range r.Installed {
		fmt.Fprintf(w, "- %s\n", loc.Path)
	}
	if len(r.Missing) > 0 {
		fmt.Fprintf(w, "missing: %s\n", strings.Join(r.Missing, ", "))
	}
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
	fmt.Println("  ralphx hook native --event Stop|UserPromptSubmit [--payload FILE] [--task FILE]")
	fmt.Println("  ralphx hook stop-guard --task FILE [--checklist FILE]")
	fmt.Println("  ralphx hook prompt-submit [--payload FILE]")
	fmt.Println("  ralphx hook install")
	fmt.Println("  ralphx hook status")
	fmt.Println("  ralphx hook uninstall")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  ralphx hook stop-guard --task tasks/demo.md --checklist tasks/demo.checklist.md")
	fmt.Println("  ralphx hook stop-guard --task tasks/demo.md --tests-required --tests-passed")
}

func resolveHookStateDir(workdir, stateDir string) string {
	if strings.TrimSpace(stateDir) != "" {
		return stateDir
	}
	if strings.TrimSpace(workdir) != "" {
		return filepath.Join(workdir, ".ralphx")
	}
	return ".ralphx"
}

func hookStatusSummary(out map[string]any) string {
	parts := make([]string, 0, 3)
	switch active := out["active"].(type) {
	case hooks.ActiveState:
		parts = append(parts, "active="+summaryBool(active.Active))
		if mode := strings.TrimSpace(active.Mode); mode != "" {
			parts = append(parts, "mode="+mode)
		}
	case map[string]any:
		parts = append(parts, "active="+summaryBool(active["active"]))
		if mode, ok := active["mode"].(string); ok && strings.TrimSpace(mode) != "" {
			parts = append(parts, "mode="+mode)
		}
	}
	if installed, ok := out["installed"].(hooks.InstallStatus); ok {
		parts = append(parts, "installed="+summaryBool(installed.ManagedInstalled))
		parts = append(parts, "stopHook="+summaryBool(installed.StopInstalled))
		parts = append(parts, "promptHook="+summaryBool(installed.PromptInstalled))
	}
	if repo, ok := out["repo"].(hooks.LogEntry); ok {
		parts = append(parts, "repo="+string(repo.Event))
		parts = append(parts, "repoReason="+repo.Decision.Reason)
	}
	if user, ok := out["user"].(hooks.LogEntry); ok {
		parts = append(parts, "user="+string(user.Event))
		parts = append(parts, "userReason="+user.Decision.Reason)
	}
	if runState, ok := out["state"].(state.RunState); ok && runState.Hook != nil {
		parts = append(parts, "stateHook="+runState.Hook.Event)
		parts = append(parts, "stateReason="+runState.Hook.Reason)
	}
	if len(parts) == 0 {
		return "[hook status] no details"
	}
	return "[hook status] " + strings.Join(parts, " ")
}

func summaryBool(v any) string {
	if b, ok := v.(bool); ok && b {
		return "true"
	}
	return "false"
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
