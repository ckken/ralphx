package contracts

import "strings"

type RoundStatus string

const (
	StatusInProgress RoundStatus = "in_progress"
	StatusBlocked    RoundStatus = "blocked"
	StatusComplete   RoundStatus = "complete"
)

type RoundResult struct {
	Status        RoundStatus `json:"status"`
	ExitSignal    bool        `json:"exit_signal"`
	FilesModified int         `json:"files_modified"`
	TestsPassed   bool        `json:"tests_passed"`
	Blockers      []string    `json:"blockers"`
	Summary       string      `json:"summary"`
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
