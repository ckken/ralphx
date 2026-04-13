package contracts

import "strings"

type RoundStatus string
type ActionMode string

const (
	StatusInProgress RoundStatus = "in_progress"
	StatusBlocked    RoundStatus = "blocked"
	StatusComplete   RoundStatus = "complete"

	ModeExecuteNextStep ActionMode = "execute_next_step"
	ModeProducePlan     ActionMode = "produce_plan"
	ModeBlocked         ActionMode = "blocked"
	ModeComplete        ActionMode = "complete"
)

type RoundResult struct {
	Status          RoundStatus `json:"status"`
	Mode            ActionMode  `json:"mode"`
	ExitSignal      bool        `json:"exit_signal"`
	FilesModified   int         `json:"files_modified"`
	TestsPassed     bool        `json:"tests_passed"`
	Blockers        []string    `json:"blockers"`
	Summary         string      `json:"summary"`
	NextStep        string      `json:"next_step,omitempty"`
	ChecklistUpdate []string    `json:"checklist_update,omitempty"`
}

func (r RoundResult) Validate() error {
	switch r.Status {
	case StatusInProgress, StatusBlocked, StatusComplete:
	default:
		return &ValidationError{Message: "invalid status: " + string(r.Status)}
	}
	if r.FilesModified < 0 {
		return &ValidationError{Message: "files_modified must be >= 0"}
	}
	switch r.Mode {
	case ModeExecuteNextStep, ModeProducePlan, ModeBlocked, ModeComplete:
	default:
		return &ValidationError{Message: "invalid mode: " + string(r.Mode)}
	}
	switch r.Status {
	case StatusInProgress:
		if r.Mode != ModeExecuteNextStep && r.Mode != ModeProducePlan {
			return &ValidationError{Message: "in_progress requires mode execute_next_step or produce_plan"}
		}
		if r.Mode == ModeExecuteNextStep && r.FilesModified <= 0 {
			return &ValidationError{Message: "execute_next_step requires files_modified > 0"}
		}
		if r.Mode == ModeProducePlan && strings.TrimSpace(r.NextStep) == "" && len(NormalizeBlockers(r.ChecklistUpdate)) == 0 {
			return &ValidationError{Message: "produce_plan requires next_step or checklist_update"}
		}
	case StatusBlocked:
		if r.Mode != ModeBlocked {
			return &ValidationError{Message: "blocked requires mode blocked"}
		}
	case StatusComplete:
		if r.Mode != ModeComplete {
			return &ValidationError{Message: "complete requires mode complete"}
		}
	}
	return nil
}

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

func NormalizeBlockers(in []string) []string {
	out := make([]string, 0, len(in))
	for _, item := range in {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		out = append(out, item)
	}
	return out
}
