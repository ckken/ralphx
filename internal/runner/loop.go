package runner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ckken/ralphx/internal/agent"
	"github.com/ckken/ralphx/internal/assets"
	"github.com/ckken/ralphx/internal/config"
	"github.com/ckken/ralphx/internal/contracts"
	"github.com/ckken/ralphx/internal/parallel"
	"github.com/ckken/ralphx/internal/plan"
	"github.com/ckken/ralphx/internal/prompt"
	"github.com/ckken/ralphx/internal/state"
	"github.com/ckken/ralphx/internal/task"
	"github.com/ckken/ralphx/internal/validate"
	"github.com/ckken/ralphx/internal/vcs"
)

type Loop struct {
	Config config.RunConfig
	Agent  agent.Agent
}

type roundExecution struct {
	Result      contracts.RoundResult
	Forced      bool
	UsedWorkers int
	Summary     string
}

func New(cfg config.RunConfig) Loop {
	return Loop{Config: cfg, Agent: agent.NewCodex(cfg.CodexCmd)}
}

func (l Loop) Run(ctx context.Context) error {
	paths := state.DerivePaths(l.Config.Workdir, l.Config.StateDir)
	if err := paths.Ensure(); err != nil {
		return err
	}
	if err := state.EnsureParallelDirs(paths); err != nil {
		return err
	}
	schemaPath, err := assets.EnsureSchemaFile(paths.Root, l.Config.OutputSchemaFile)
	if err != nil {
		return fmt.Errorf("prepare output schema: %w", err)
	}
	template, err := prompt.LoadTemplate(l.Config.PromptFile)
	if err != nil {
		return fmt.Errorf("load prompt template: %w", err)
	}
	bundle, err := task.Load(l.Config.TaskFile, task.LoadOptions{
		ChecklistPath: l.Config.ChecklistFile,
		SummaryPath:   paths.SummaryFile,
		StatePath:     paths.StateFile,
	})
	if err != nil {
		return err
	}
	start := time.Now()
	iteration := 1
	noProgress := 0
	session, _ := state.LoadSession(paths)

	fmt.Printf("[%s] Starting Go ralphx loop in %s\n", ts(time.Now()), l.Config.Workdir)
	fmt.Printf("[%s] Task: %s\n", ts(time.Now()), l.Config.TaskFile)

	for {
		if l.Config.MaxIterations > 0 && iteration > l.Config.MaxIterations {
			return fmt.Errorf("stopping after reaching MAX_ITERATIONS=%d before task completion", l.Config.MaxIterations)
		}

		if !l.Config.ResumeSession || !state.SessionFresh(session, l.Config.SessionExpiry, time.Now()) {
			if l.Config.ResumeSession && session.ThreadID != "" {
				_ = state.ClearSession(paths)
			}
			session = state.SessionMeta{}
		}

		execResult, nextSession, err := l.runIteration(ctx, iteration, bundle, template, schemaPath, paths, session)
		if err != nil {
			result := contracts.RoundResult{Status: contracts.StatusBlocked, Mode: contracts.ModeBlocked, ExitSignal: false, FilesModified: 0, TestsPassed: false, Blockers: []string{"runner_error"}, Summary: err.Error()}
			_ = state.WriteLastResult(paths, result)
			_ = state.WriteStateWithContext(paths, iteration, result, l.Config.TaskFile, l.Config.ChecklistFile, nil)
			if replanned, guidance, replannedErr := l.tryAutoReplan(ctx, paths, "runner_error"); replannedErr == nil && replanned {
				_ = state.WriteStateWithContext(paths, iteration, result, l.Config.TaskFile, l.Config.ChecklistFile, guidance)
				return fmt.Errorf("runner error triggered auto-replan; review regenerated task/checklist and rerun")
			}
			return err
		}
		session = nextSession
		if l.Config.ResumeSession {
			if session.ThreadID != "" {
				_ = state.WriteSession(paths, session)
			} else {
				_ = state.ClearSession(paths)
			}
		}

		result := execResult.Result
		var guidance *state.Guidance
		if result.Status == contracts.StatusInProgress && result.Mode == contracts.ModeProducePlan {
			var applyErr error
			guidance, applyErr = l.applyProducePlan(bundle, paths, result)
			if applyErr != nil {
				result = contracts.RoundResult{
					Status:        contracts.StatusBlocked,
					Mode:          contracts.ModeBlocked,
					ExitSignal:    false,
					FilesModified: 0,
					TestsPassed:   false,
					Blockers:      []string{"invalid_produce_plan"},
					Summary:       applyErr.Error(),
				}
				_ = state.WriteLastResult(paths, result)
				_ = state.WriteStateWithGuidance(paths, iteration, result, guidance)
				return fmt.Errorf(applyErr.Error())
			}
			result.Summary = guidance.Message
			_ = state.WriteSummary(paths, guidance.Message)
		}
		bundle, err = task.Load(l.Config.TaskFile, task.LoadOptions{
			ChecklistPath: l.Config.ChecklistFile,
			SummaryPath:   paths.SummaryFile,
			StatePath:     paths.StateFile,
		})
		if err != nil {
			return err
		}

		_ = state.WriteSummary(paths, result.Summary)
		_ = state.WriteLastResult(paths, result)
		_ = state.WriteStateWithContext(paths, iteration, result, l.Config.TaskFile, l.Config.ChecklistFile, guidance)
		stats := state.Stats{
			StartedAt:           start.Format("2006-01-02 15:04:05"),
			UpdatedAt:           time.Now().Format("2006-01-02 15:04:05"),
			LoopsCompleted:      iteration,
			TotalElapsedSeconds: int(time.Since(start).Seconds()),
			LastRoundSeconds:    0,
			AverageRoundSeconds: max(1, int(time.Since(start).Seconds())/iteration),
			LastStatus:          string(result.Status),
			LastExitSignal:      result.ExitSignal,
			LastFilesModified:   result.FilesModified,
		}
		_ = state.WriteStats(paths, stats)

		fmt.Printf("[%s] Result: status=%s exit_signal=%t files_modified=%d tests_passed=%t blockers=%d workers=%d\n", ts(time.Now()), result.Status, result.ExitSignal, result.FilesModified, result.TestsPassed, len(result.Blockers), execResult.UsedWorkers)

		if l.Config.TestsCmd != "" && !execResult.Forced && (result.Mode == contracts.ModeExecuteNextStep || result.Mode == contracts.ModeComplete) {
			testLog := filepath.Join(paths.LogDir, fmt.Sprintf("tests-%d.log", iteration))
			if err := validate.Run(ctx, l.Config.Workdir, l.Config.TestsCmd, testLog); err != nil {
				result = contracts.RoundResult{Status: contracts.StatusBlocked, Mode: contracts.ModeBlocked, ExitSignal: false, FilesModified: 0, TestsPassed: false, Blockers: []string{"tests_failed"}, Summary: "Tests failed"}
				_ = state.WriteLastResult(paths, result)
				_ = state.WriteStateWithContext(paths, iteration, result, l.Config.TaskFile, l.Config.ChecklistFile, nil)
				if replanned, guidance, replannedErr := l.tryAutoReplan(ctx, paths, "tests_failed"); replannedErr == nil && replanned {
					_ = state.WriteStateWithContext(paths, iteration, result, l.Config.TaskFile, l.Config.ChecklistFile, guidance)
					return fmt.Errorf("tests failed and auto-replan regenerated task/checklist; see %s", testLog)
				}
				return fmt.Errorf("tests failed; see %s", testLog)
			}
		}

		if result.ExitSignal && result.Status == contracts.StatusComplete {
			fmt.Printf("[%s] Task complete\n", ts(time.Now()))
			return nil
		}
		if result.Status == contracts.StatusBlocked {
			if replanned, guidance, replannedErr := l.tryAutoReplan(ctx, paths, "blocked"); replannedErr == nil && replanned {
				_ = state.WriteStateWithContext(paths, iteration, result, l.Config.TaskFile, l.Config.ChecklistFile, guidance)
				return fmt.Errorf("codex reported blockers and auto-replan regenerated task/checklist")
			}
			return fmt.Errorf("codex reported blockers: %s", strings.Join(result.Blockers, ", "))
		}

		if result.FilesModified > 0 || result.Mode == contracts.ModeProducePlan {
			noProgress = 0
		} else {
			noProgress++
		}
		if l.Config.MaxNoProgress > 0 && noProgress >= l.Config.MaxNoProgress {
			if replanned, guidance, replannedErr := l.tryAutoReplan(ctx, paths, "no_progress"); replannedErr == nil && replanned {
				_ = state.WriteStateWithContext(paths, iteration, result, l.Config.TaskFile, l.Config.ChecklistFile, guidance)
				return fmt.Errorf("stopping after %d no-progress rounds; auto-replan regenerated task/checklist", noProgress)
			}
			return fmt.Errorf("stopping after %d no-progress rounds", noProgress)
		}
		iteration++
	}
}

func (l Loop) runIteration(ctx context.Context, iteration int, bundle task.Bundle, template, schemaPath string, paths state.Paths, session state.SessionMeta) (roundExecution, state.SessionMeta, error) {
	preSnap, err := vcs.CaptureStatusSnapshot(l.Config.Workdir)
	if err != nil {
		return roundExecution{}, state.SessionMeta{}, err
	}
	jobs := l.buildJobs(bundle)
	if len(jobs) > 1 {
		return l.runParallelIteration(ctx, iteration, bundle, template, schemaPath, paths, preSnap, jobs)
	}
	return l.runSingleIteration(ctx, iteration, bundle, template, schemaPath, paths, preSnap, session)
}

func (l Loop) buildJobs(bundle task.Bundle) []parallel.Job {
	if l.Config.Workers <= 1 || strings.TrimSpace(bundle.Checklist.Content) == "" {
		return nil
	}
	items := task.OpenChecklistItems(bundle.Checklist.Content)
	if len(items) <= 1 {
		return nil
	}
	jobs := make([]parallel.Job, 0, len(items))
	for _, item := range items {
		jobs = append(jobs, parallel.Job{
			ID:      fmt.Sprintf("task-%04d", item.Index+1),
			Goal:    item.Text,
			Status:  parallel.JobPending,
			Summary: item.RawLine,
		})
	}
	return jobs
}

func (l Loop) runSingleIteration(ctx context.Context, iteration int, bundle task.Bundle, template, schemaPath string, paths state.Paths, preSnap vcs.Snapshot, session state.SessionMeta) (roundExecution, state.SessionMeta, error) {
	gitStatus := preSnap.Status
	p := prompt.Build(prompt.BuildInput{Iteration: iteration, Workdir: l.Config.Workdir, Bundle: bundle, Template: template, GitStatus: gitStatus})
	rawPath := filepath.Join(paths.Root, fmt.Sprintf("round-%d.txt", iteration))
	logPath := filepath.Join(paths.LogDir, fmt.Sprintf("round-%d.log", iteration))

	roundCtx, cancel := context.WithTimeout(ctx, l.Config.RoundTimeout)
	fmt.Printf("[%s] Round %d: invoking %s\n", ts(time.Now()), iteration, l.Config.CodexCmd)
	resp, runErr := l.Agent.Run(roundCtx, agent.Request{Workdir: l.Config.Workdir, Prompt: p, OutputSchemaPath: schemaPath, RawLogPath: rawPath, ExtraArgs: l.Config.CodexArgs, SessionID: session.ThreadID})
	cancel()
	if shouldResetSession(runErr) && session.ThreadID != "" {
		session = state.SessionMeta{}
		roundCtx, cancel = context.WithTimeout(ctx, l.Config.RoundTimeout)
		fmt.Printf("[%s] Round %d: session expired or missing, retrying with fresh session\n", ts(time.Now()), iteration)
		resp, runErr = l.Agent.Run(roundCtx, agent.Request{Workdir: l.Config.Workdir, Prompt: p, OutputSchemaPath: schemaPath, RawLogPath: rawPath, ExtraArgs: l.Config.CodexArgs})
		cancel()
	}
	_ = os.WriteFile(logPath, resp.RawOutput, 0o644)
	_ = state.WriteLastOutput(paths, string(resp.RawOutput))

	result := resp.Parsed
	if parseOrRunFailed(runErr, result) {
		return roundExecution{}, state.SessionMeta{}, fmt.Errorf(blockedSummary(runErr))
	}

	postBundle, err := task.Load(l.Config.TaskFile, task.LoadOptions{ChecklistPath: l.Config.ChecklistFile, SummaryPath: paths.SummaryFile, StatePath: paths.StateFile})
	if err != nil {
		return roundExecution{}, state.SessionMeta{}, err
	}
	postSnap, err := vcs.CaptureStatusSnapshot(l.Config.Workdir)
	if err != nil {
		return roundExecution{}, state.SessionMeta{}, err
	}
	forced := false
	result, forced = applyGlobalGates(result, preSnap, postSnap, postBundle.Checklist.OpenItems)
	nextSession := session
	if resp.SessionID != "" {
		nextSession = state.SessionMeta{ThreadID: resp.SessionID, UpdatedAt: time.Now().Format("2006-01-02 15:04:05")}
	}
	return roundExecution{Result: result, Forced: forced, UsedWorkers: 1, Summary: result.Summary}, nextSession, nil
}

func (l Loop) runParallelIteration(ctx context.Context, iteration int, bundle task.Bundle, template, schemaPath string, paths state.Paths, preSnap vcs.Snapshot, jobs []parallel.Job) (roundExecution, state.SessionMeta, error) {
	fmt.Printf("[%s] Round %d: scheduling %d jobs across %d workers\n", ts(time.Now()), iteration, len(jobs), min(l.Config.Workers, len(jobs)))
	workerFn := parallel.FuncWorker(func(workerCtx context.Context, job parallel.Job) (parallel.WorkerResult, error) {
		return l.executeJob(workerCtx, iteration, bundle, template, schemaPath, paths, preSnap.Status, job)
	})
	scheduler := parallel.LocalScheduler{Workers: l.Config.Workers, Worker: workerFn}
	results, err := scheduler.RunRound(ctx, jobs)
	if err != nil {
		return roundExecution{}, state.SessionMeta{}, err
	}
	for _, res := range results {
		resultPath := filepath.Join(state.ResultsDir(paths), res.JobID+".result.json")
		_ = state.WriteJSON(resultPath, res)
	}
	completedIndexes := []int{}
	blocked := []string{}
	filesModified := 0
	allComplete := len(results) > 0
	for _, res := range results {
		filesModified += res.FilesModified
		if res.Status == string(contracts.StatusComplete) && res.ExitSignal {
			idx := parseJobIndex(res.JobID)
			if idx >= 0 {
				completedIndexes = append(completedIndexes, idx)
			}
			continue
		}
		allComplete = false
		if len(res.Blockers) > 0 {
			blocked = append(blocked, res.Blockers...)
		} else {
			blocked = append(blocked, fmt.Sprintf("job %s returned %s", res.JobID, res.Status))
		}
	}
	if len(completedIndexes) > 0 {
		_ = task.MarkChecklistItemsDone(bundle.Checklist.Path, completedIndexes)
	}
	postBundle, err := task.Load(l.Config.TaskFile, task.LoadOptions{ChecklistPath: l.Config.ChecklistFile, SummaryPath: paths.SummaryFile, StatePath: paths.StateFile})
	if err != nil {
		return roundExecution{}, state.SessionMeta{}, err
	}
	postSnap, err := vcs.CaptureStatusSnapshot(l.Config.Workdir)
	if err != nil {
		return roundExecution{}, state.SessionMeta{}, err
	}
	if allComplete && postBundle.Checklist.OpenItems == 0 {
		result := contracts.RoundResult{Status: contracts.StatusComplete, Mode: contracts.ModeComplete, ExitSignal: true, FilesModified: max(1, filesModified), TestsPassed: true, Blockers: nil, Summary: fmt.Sprintf("completed %d checklist jobs", len(results))}
		result, forced := applyGlobalGates(result, preSnap, postSnap, postBundle.Checklist.OpenItems)
		return roundExecution{Result: result, Forced: forced, UsedWorkers: min(l.Config.Workers, len(jobs)), Summary: result.Summary}, state.SessionMeta{}, nil
	}
	if len(blocked) > 0 && len(completedIndexes) == 0 {
		return roundExecution{Result: contracts.RoundResult{Status: contracts.StatusBlocked, Mode: contracts.ModeBlocked, ExitSignal: false, FilesModified: filesModified, TestsPassed: false, Blockers: contracts.NormalizeBlockers(blocked), Summary: "parallel jobs blocked"}, UsedWorkers: min(l.Config.Workers, len(jobs))}, state.SessionMeta{}, nil
	}
	return roundExecution{Result: contracts.RoundResult{Status: contracts.StatusInProgress, Mode: contracts.ModeExecuteNextStep, ExitSignal: false, FilesModified: max(1, filesModified), TestsPassed: false, Blockers: nil, Summary: fmt.Sprintf("completed %d/%d checklist jobs", len(completedIndexes), len(jobs))}, UsedWorkers: min(l.Config.Workers, len(jobs))}, state.SessionMeta{}, nil
}

func (l Loop) executeJob(ctx context.Context, iteration int, bundle task.Bundle, template, schemaPath string, paths state.Paths, gitStatus string, job parallel.Job) (parallel.WorkerResult, error) {
	idx := parseJobIndex(job.ID)
	items := task.OpenChecklistItems(bundle.Checklist.Content)
	if idx < 0 || idx >= len(items) {
		return parallel.WorkerResult{JobID: job.ID, WorkerID: fmt.Sprintf("worker-%02d", 1), Status: string(contracts.StatusBlocked), ExitSignal: false, FilesModified: 0, TestsPassed: false, Blockers: []string{"invalid_job_index"}, Summary: job.Goal}, nil
	}
	item := items[idx]
	workerID := fmt.Sprintf("worker-%02d", idx+1)
	workerStatePath := filepath.Join(state.WorkersDir(paths), workerID+".json")
	workerLogPath := filepath.Join(paths.LogDir, workerID+".stdout.log")
	_ = state.WriteJSON(workerStatePath, parallel.WorkerState{ID: workerID, Lifecycle: parallel.WorkerRunning, JobID: job.ID, StartedAt: time.Now().Format(time.RFC3339), UpdatedAt: time.Now().Format(time.RFC3339), LogPath: workerLogPath, ResultPath: filepath.Join(state.ResultsDir(paths), job.ID+".result.json")})

	jobBundle := bundle
	jobBundle.Checklist.Content = "- [ ] " + item.Text
	jobBundle.Checklist.OpenItems = 1
	jobBundle.Checklist.Path = bundle.Checklist.Path
	jobPrompt := prompt.Build(prompt.BuildInput{Iteration: iteration, Workdir: l.Config.Workdir, Bundle: jobBundle, Template: template, GitStatus: gitStatus}) + "\nFocus only on this assigned checklist item and do not claim overall task completion unless your assigned slice is done.\n"
	rawPath := filepath.Join(paths.Root, fmt.Sprintf("round-%d-%s.txt", iteration, job.ID))
	jobCtx, cancel := context.WithTimeout(ctx, l.Config.RoundTimeout)
	resp, err := l.Agent.Run(jobCtx, agent.Request{Workdir: l.Config.Workdir, Prompt: jobPrompt, OutputSchemaPath: schemaPath, RawLogPath: rawPath, ExtraArgs: l.Config.CodexArgs})
	cancel()
	_ = os.WriteFile(workerLogPath, resp.RawOutput, 0o644)
	result := resp.Parsed
	if parseOrRunFailed(err, result) {
		return parallel.WorkerResult{JobID: job.ID, WorkerID: workerID, Status: string(contracts.StatusBlocked), ExitSignal: false, FilesModified: 0, TestsPassed: false, Blockers: []string{"invalid_json"}, Summary: blockedSummary(err)}, nil
	}
	return parallel.WorkerResult{JobID: job.ID, WorkerID: workerID, Status: string(result.Status), ExitSignal: result.ExitSignal, FilesModified: result.FilesModified, TestsPassed: result.TestsPassed, Blockers: result.Blockers, Summary: result.Summary}, nil
}

func applyGlobalGates(result contracts.RoundResult, preSnap, postSnap vcs.Snapshot, checklistOpenItems int) (contracts.RoundResult, bool) {
	forced := false
	if result.Status == contracts.StatusComplete && result.ExitSignal && result.FilesModified <= 0 && preSnap.Status == postSnap.Status {
		forced = true
		result = contracts.RoundResult{Status: contracts.StatusInProgress, Mode: contracts.ModeProducePlan, ExitSignal: false, FilesModified: 0, TestsPassed: false, Blockers: nil, Summary: "Ignored premature completion because no new changes were detected. " + result.Summary, NextStep: "Continue with the next remaining bounded step instead of declaring completion."}
	}
	if result.Status == contracts.StatusComplete && result.ExitSignal && checklistOpenItems > 0 {
		forced = true
		result = contracts.RoundResult{Status: contracts.StatusInProgress, Mode: contracts.ModeProducePlan, ExitSignal: false, FilesModified: 0, TestsPassed: false, Blockers: nil, Summary: fmt.Sprintf("Ignored premature completion because checklist still has %d open items. %s", checklistOpenItems, result.Summary), NextStep: "Continue with the remaining checklist items."}
	}
	return result, forced
}

func shouldResetSession(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "thread/resume failed") || strings.Contains(msg, "no rollout found for thread id")
}

func (l Loop) tryAutoReplan(ctx context.Context, paths state.Paths, reason string) (bool, *state.Guidance, error) {
	if !l.Config.AutoReplan {
		return false, nil, nil
	}
	schemaPath, err := assets.EnsurePlanSchemaFile(paths.Root, "")
	if err != nil {
		return false, nil, err
	}
	rawLogPath := filepath.Join(paths.LogDir, "auto-replan.log")
	if err := plan.EnsureLogDir(rawLogPath); err != nil {
		return false, nil, err
	}
	replanTimeout := l.Config.RoundTimeout
	if replanTimeout <= 0 {
		replanTimeout = 30 * time.Minute
	}
	replanCtx, cancel := context.WithTimeout(ctx, replanTimeout)
	defer cancel()
	outcome, _, _, err := plan.Replan(replanCtx, plan.ReplanRequest{
		TaskPath:          l.Config.TaskFile,
		ChecklistPath:     l.Config.ChecklistFile,
		SummaryPath:       paths.SummaryFile,
		StatePath:         paths.StateFile,
		Workdir:           l.Config.Workdir,
		CodexCmd:          l.Config.CodexCmd,
		CodexArgs:         l.Config.CodexArgs,
		OutputSchemaPath:  schemaPath,
		RawLogPath:        rawLogPath,
		PreserveCompleted: true,
	})
	if err != nil {
		return false, nil, err
	}
	taskFile, checklistFile, err := plan.WriteFiles(l.Config.TaskFile, outcome)
	if err != nil {
		return false, nil, err
	}
	message := fmt.Sprintf("Auto-replanned due to %s. Review regenerated task/checklist and rerun.", reason)
	_ = state.WriteSummary(paths, message)
	guidance := &state.Guidance{
		Reason:        reason,
		Message:       message,
		TaskFile:      taskFile,
		ChecklistFile: checklistFile,
		NextStep:      outcome.TaskMarkdown,
		ChecklistNote: outcome.Checklist,
		GeneratedAt:   time.Now().Format("2006-01-02 15:04:05"),
	}
	return true, guidance, nil
}

func (l Loop) applyProducePlan(bundle task.Bundle, paths state.Paths, result contracts.RoundResult) (*state.Guidance, error) {
	if strings.TrimSpace(result.NextStep) == "" && len(result.ChecklistUpdate) == 0 {
		return nil, fmt.Errorf("produce_plan did not include next_step or checklist_update")
	}

	taskFile := bundle.Task.Path
	checklistFile := bundle.Checklist.Path
	if strings.TrimSpace(taskFile) == "" {
		taskFile = l.Config.TaskFile
	}
	if strings.TrimSpace(taskFile) == "" {
		return nil, fmt.Errorf("produce_plan requires a task file path")
	}
	if strings.TrimSpace(result.NextStep) != "" {
		updatedTask := appendPlannedNextStep(bundle.Task.Content, result.NextStep)
		if err := os.WriteFile(taskFile, []byte(updatedTask), 0o644); err != nil {
			return nil, err
		}
	}
	if len(result.ChecklistUpdate) > 0 {
		items := result.ChecklistUpdate
		if strings.TrimSpace(bundle.Checklist.Content) != "" {
			items = planMergeChecklist(bundle.Checklist.Content, result.ChecklistUpdate)
		}
		if strings.TrimSpace(checklistFile) == "" {
			checklistFile = plan.ChecklistPath(taskFile)
		}
		checklistBody := renderChecklistForPlan(checklistTitle(bundle, taskFile), items)
		if err := os.WriteFile(checklistFile, []byte(checklistBody), 0o644); err != nil {
			return nil, err
		}
	}
	message := "Applied produce_plan to task/state"
	if strings.TrimSpace(result.NextStep) != "" {
		message = "Applied produce_plan and updated the next step in the task file"
	}
	if len(result.ChecklistUpdate) > 0 {
		message = "Applied produce_plan and updated the checklist"
	}
	return &state.Guidance{
		Reason:        "produce_plan",
		Message:       message,
		TaskFile:      taskFile,
		ChecklistFile: checklistFile,
		NextStep:      result.NextStep,
		ChecklistNote: result.ChecklistUpdate,
		GeneratedAt:   time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

func appendPlannedNextStep(taskContent, nextStep string) string {
	base := strings.TrimSpace(taskContent)
	section := "## Planned Next Step\n\n" + strings.TrimSpace(nextStep)
	if base == "" {
		return section + "\n"
	}
	marker := "\n## Planned Next Step"
	if idx := strings.Index(base, marker); idx >= 0 {
		base = strings.TrimSpace(base[:idx])
	}
	return base + "\n\n" + section + "\n"
}

func checklistTitle(bundle task.Bundle, taskFile string) string {
	taskContent := strings.TrimSpace(bundle.Task.Content)
	if strings.HasPrefix(taskContent, "# ") {
		if lineEnd := strings.Index(taskContent, "\n"); lineEnd >= 0 {
			return strings.TrimSpace(taskContent[2:lineEnd])
		}
		return strings.TrimSpace(taskContent[2:])
	}
	base := strings.TrimSuffix(filepath.Base(taskFile), filepath.Ext(taskFile))
	if strings.TrimSpace(base) == "" {
		return "task"
	}
	return base
}

func renderChecklistForPlan(title string, items []string) string {
	lines := []string{fmt.Sprintf("# %s checklist", strings.TrimSpace(title)), ""}
	for _, item := range items {
		if strings.TrimSpace(item) == "" {
			continue
		}
		lines = append(lines, "- [ ] "+strings.TrimSpace(item))
	}
	return strings.Join(lines, "\n") + "\n"
}

func planMergeChecklist(existing string, next []string) []string {
	completed := completedChecklistTextsForPlan(existing)
	seen := map[string]bool{}
	out := make([]string, 0, len(completed)+len(next))
	for _, item := range completed {
		key := normalizeChecklistTextForPlan(item)
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, item)
	}
	for _, item := range next {
		key := normalizeChecklistTextForPlan(item)
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, item)
	}
	return out
}

func completedChecklistTextsForPlan(content string) []string {
	lines := strings.Split(content, "\n")
	out := make([]string, 0)
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "- [x] "):
			out = append(out, strings.TrimSpace(trimmed[6:]))
		case strings.HasPrefix(trimmed, "* [x] "):
			out = append(out, strings.TrimSpace(trimmed[6:]))
		case strings.HasPrefix(trimmed, "- [X] "):
			out = append(out, strings.TrimSpace(trimmed[6:]))
		case strings.HasPrefix(trimmed, "* [X] "):
			out = append(out, strings.TrimSpace(trimmed[6:]))
		}
	}
	return out
}

func normalizeChecklistTextForPlan(text string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(text)), " "))
}

func parseOrRunFailed(err error, result contracts.RoundResult) bool {
	if result.Status == "" {
		return true
	}
	return false && err != nil
}

func blockedSummary(err error) string {
	if err == nil {
		return "Codex did not return a JSON object"
	}
	return "Codex did not return a JSON object: " + err.Error()
}

func parseJobIndex(jobID string) int {
	var n int
	if _, err := fmt.Sscanf(jobID, "task-%04d", &n); err != nil {
		return -1
	}
	return n - 1
}

func ts(t time.Time) string { return t.Format("2006-01-02 15:04:05") }
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
