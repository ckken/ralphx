package hooks

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ckken/ralphx/internal/state"
)

type NativeHookEvent string

const (
	NativeHookEventUserPromptSubmit NativeHookEvent = "UserPromptSubmit"
	NativeHookEventStop             NativeHookEvent = "Stop"
)

type NativeHookRequest struct {
	Event          NativeHookEvent
	Workdir        string
	StateDir       string
	PayloadPath    string
	TaskPath       string
	ChecklistPath  string
	SummaryPath    string
	StatePath      string
	LastResultPath string
	TestsRequired  bool
	TestsPassedNow bool
	NativeJSON     bool
}

type NativeHookResponse struct {
	ExitCode int
	Stdout   any
	Stderr   string
	Decision Decision
	Result   any
	Silent   bool
}

func NormalizeNativeHookEvent(raw string) (NativeHookEvent, bool) {
	switch strings.TrimSpace(raw) {
	case string(NativeHookEventUserPromptSubmit), "prompt-submit":
		return NativeHookEventUserPromptSubmit, true
	case string(NativeHookEventStop), "stop-guard":
		return NativeHookEventStop, true
	default:
		return "", false
	}
}

func DispatchNativeHook(req NativeHookRequest) (NativeHookResponse, error) {
	switch req.Event {
	case NativeHookEventUserPromptSubmit:
		return dispatchNativePromptSubmit(req)
	case NativeHookEventStop:
		return dispatchNativeStopGuard(req)
	default:
		return NativeHookResponse{}, fmt.Errorf("unsupported native hook event: %q", req.Event)
	}
}

func dispatchNativePromptSubmit(req NativeHookRequest) (NativeHookResponse, error) {
	payload, err := LoadPromptSubmitPayload(req.PayloadPath)
	if err != nil {
		return NativeHookResponse{}, err
	}

	text := PromptText(payload)
	effectiveWorkdir := strings.TrimSpace(req.Workdir)
	if effectiveWorkdir == "" {
		effectiveWorkdir = strings.TrimSpace(payload.Cwd)
	}
	stateDir := resolveNativeHookStateDir(effectiveWorkdir, req.StateDir)
	existingState, existingErr := ReadActiveState(stateDir)
	stateActive := existingErr == nil && existingState.Active
	promptActive := PromptActivatesRalphx(text)
	decision := Decision{
		Allow:   true,
		Reason:  "prompt_submit",
		Message: "ralphx inactive",
	}
	stopping := PromptStopsRalphx(text)
	if stopping {
		decision.Reason = "prompt_stop"
		decision.Message = "ralphx workflow stopped"
	} else if promptActive {
		decision.Message = "ralphx mode active"
	}

	if stopping {
		entry := LogEntry{
			Event:    EventPromptSubmit,
			Decision: decision,
			Result: map[string]any{
				"prompt_active": promptActive,
				"state_active":  stateActive,
				"stopping":      stopping,
				"text":          text,
				"workdir":       effectiveWorkdir,
			},
		}
		if err := writeNativeHookEntry(stateDir, entry); err != nil {
			return NativeHookResponse{}, err
		}
		if err := ClearActiveState(stateDir); err != nil {
			return NativeHookResponse{}, err
		}
		resp := NativeHookResponse{
			ExitCode: 0,
			Decision: decision,
			Result:   entry.Result,
			Stdout:   map[string]any{},
		}
		resp.Stderr = "[hook prompt-submit] ralphx workflow stopped"
		return resp, nil
	}

	if promptActive {
		entry := LogEntry{
			Event:    EventPromptSubmit,
			Decision: decision,
			Result: map[string]any{
				"prompt_active": promptActive,
				"state_active":  stateActive,
				"stopping":      stopping,
				"text":          text,
				"workdir":       effectiveWorkdir,
			},
		}
		if err := writeNativeHookEntry(stateDir, entry); err != nil {
			return NativeHookResponse{}, err
		}
		if err := WriteActiveState(stateDir, text); err != nil {
			return NativeHookResponse{}, err
		}
		resp := NativeHookResponse{
			ExitCode: 0,
			Decision: decision,
			Result:   entry.Result,
			Stdout: map[string]any{
				"hookSpecificOutput": map[string]any{
					"hookEventName":     string(NativeHookEventUserPromptSubmit),
					"additionalContext": "ralphx active.",
				},
			},
			Stderr: "[hook prompt-submit] mode=ralphx active (ralphx active)",
		}
		return resp, nil
	}

	return NativeHookResponse{
		ExitCode: 0,
		Decision: decision,
		Silent:   true,
	}, nil
}

func dispatchNativeStopGuard(req NativeHookRequest) (NativeHookResponse, error) {
	guardStateDir := resolveNativeHookStateDir(req.Workdir, req.StateDir)
	activeState, activeErr := ReadActiveState(guardStateDir)
	if activeErr == nil && activeState.Active {
		if activeState.StopHookActive {
			return NativeHookResponse{
				ExitCode: 0,
				Decision: Decision{
					Allow:   true,
					Reason:  "stop_hook_active",
					Message: "stop hook already handled",
				},
				Silent: true,
			}, nil
		}

		decision := Decision{
			Allow:   false,
			Reason:  "active_workflow",
			Message: "Ralph is still active in this workspace. Continue the current branch of work instead of stopping.",
		}
		entry := LogEntry{
			Event:    EventStop,
			Decision: decision,
			Result: map[string]any{
				"workdir":          req.Workdir,
				"mode":             activeState.Mode,
				"prompt":           activeState.Prompt,
				"active":           activeState.Active,
				"stop_hook_active": activeState.StopHookActive,
			},
		}
		if err := writeNativeHookEntry(guardStateDir, entry); err != nil {
			return NativeHookResponse{}, err
		}
		if err := MarkStopHookActive(guardStateDir, decision.Reason); err != nil {
			return NativeHookResponse{}, err
		}
		resp := NativeHookResponse{
			ExitCode: 3,
			Decision: decision,
			Result:   entry.Result,
		}
		if req.NativeJSON {
			resp.ExitCode = 0
			resp.Stdout = nativeStopOutput(decision)
		} else {
			resp.Stdout = decision
		}
		resp.Stderr = fmt.Sprintf("[hook stop-guard] block (%s) $ralphx", decision.Reason)
		return resp, nil
	}

	input, err := LoadStopGuardInput(req.TaskPath, req.ChecklistPath, req.SummaryPath, req.StatePath, req.LastResultPath, req.TestsRequired, req.TestsPassedNow)
	if err != nil {
		if err == ErrNoTaskContext {
			decision := Decision{
				Allow:   true,
				Reason:  "no_task_context",
				Message: "No ralphx task context found in the current workspace; skipping stop guard.",
			}
			entry := LogEntry{
				Event:    EventStop,
				Decision: decision,
				Result: map[string]any{
					"workdir": req.Workdir,
				},
			}
			if err := writeNativeHookEntry(guardStateDir, entry); err != nil {
				return NativeHookResponse{}, err
			}
			resp := NativeHookResponse{
				ExitCode: 0,
				Decision: decision,
				Result:   entry.Result,
			}
			if req.NativeJSON {
				resp.Stdout = map[string]any{}
			} else {
				resp.Stdout = decision
			}
			resp.Stderr = fmt.Sprintf("[hook stop-guard] allow: %s $ralphx", decision.Message)
			return resp, nil
		}
		return NativeHookResponse{}, err
	}

	decision := EvaluateStopGuard(GuardConfig{
		Enabled:                   true,
		BlockWhenChecklistOpen:    true,
		BlockWhenVerificationMiss: true,
		BlockWhenIncomplete:       true,
	}, input)
	entry := LogEntry{
		Event:         EventStop,
		TaskPath:      req.TaskPath,
		ChecklistPath: req.ChecklistPath,
		Decision:      decision,
		Result:        input.Result,
	}
	if err := writeNativeHookEntry(guardStateDir, entry); err != nil {
		return NativeHookResponse{}, err
	}

	resp := NativeHookResponse{
		ExitCode: 0,
		Decision: decision,
		Result:   input.Result,
	}
	if decision.Allow {
		message := strings.TrimSpace(decision.Message)
		if message == "" {
			message = "stop allowed"
		}
		if err := ClearActiveState(guardStateDir); err != nil {
			return NativeHookResponse{}, err
		}
		resp.Stderr = fmt.Sprintf("[hook stop-guard] allow: %s $ralphx", message)
		if req.NativeJSON {
			resp.Stdout = map[string]any{}
		} else {
			resp.Stdout = decision
		}
		return resp, nil
	}

	resp.ExitCode = 3
	resp.Stderr = fmt.Sprintf("[hook stop-guard] block (%s) $ralphx", decision.Reason)
	if req.NativeJSON {
		resp.ExitCode = 0
		resp.Stdout = nativeStopOutput(decision)
	} else {
		resp.Stdout = decision
	}
	return resp, nil
}

func resolveNativeHookStateDir(workdir, stateDir string) string {
	if strings.TrimSpace(stateDir) != "" {
		return stateDir
	}
	if strings.TrimSpace(workdir) != "" {
		return filepath.Join(workdir, ".ralphx")
	}
	return ".ralphx"
}

func writeNativeHookEntry(stateDir string, entry LogEntry) error {
	if entry.Timestamp == "" {
		entry.Timestamp = time.Now().Format(time.RFC3339)
	}
	hookLogDir := filepath.Join(stateDir, "logs")
	if err := AppendLog(hookLogDir, entry); err != nil {
		return err
	}
	if err := WriteLatest(filepath.Join(stateDir, "last-hook-event.json"), entry); err != nil {
		return err
	}
	if err := AppendUserLog(entry); err != nil {
		return err
	}
	if err := WriteUserLatest(entry); err != nil {
		return err
	}
	return persistNativeHookState(stateDir, entry)
}

func nativeStopOutput(decision Decision) map[string]any {
	out := map[string]any{
		"decision": map[bool]string{true: "allow", false: "block"}[decision.Allow],
		"reason":   decision.Reason,
	}
	return out
}

func persistNativeHookState(stateDir string, entry LogEntry) error {
	paths := state.DerivePaths("", stateDir)
	runState, err := state.LoadRunState(paths)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err != nil {
		runState = state.RunState{}
	}
	runState.Hook = &state.HookState{
		Event:     string(entry.Event),
		Allow:     entry.Decision.Allow,
		Reason:    entry.Decision.Reason,
		Message:   entry.Decision.Message,
		Result:    entry.Result,
		UpdatedAt: entry.Timestamp,
	}
	return state.WriteRunState(paths, runState)
}
