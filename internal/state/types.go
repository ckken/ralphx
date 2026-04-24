package state

import "github.com/ckken/ralphx/internal/contracts"

type RunState struct {
	Iteration     int                   `json:"iteration"`
	UpdatedAt     string                `json:"updated_at"`
	TaskFile      string                `json:"task_file,omitempty"`
	ChecklistFile string                `json:"checklist_file,omitempty"`
	Result        contracts.RoundResult `json:"result"`
	Guidance      *Guidance             `json:"guidance,omitempty"`
	Hook          *HookState            `json:"hook,omitempty"`
}

type HookState struct {
	Event     string `json:"event"`
	Allow     bool   `json:"allow"`
	Reason    string `json:"reason"`
	Message   string `json:"message"`
	Result    any    `json:"result,omitempty"`
	UpdatedAt string `json:"updated_at"`
}

type Guidance struct {
	Reason        string   `json:"reason"`
	Message       string   `json:"message"`
	TaskFile      string   `json:"task_file,omitempty"`
	ChecklistFile string   `json:"checklist_file,omitempty"`
	NextStep      string   `json:"next_step,omitempty"`
	ChecklistNote []string `json:"checklist_update,omitempty"`
	GeneratedAt   string   `json:"generated_at"`
}

type Stats struct {
	StartedAt           string `json:"started_at"`
	UpdatedAt           string `json:"updated_at"`
	LoopsCompleted      int    `json:"loops_completed"`
	TotalElapsedSeconds int    `json:"total_elapsed_seconds"`
	LastRoundSeconds    int    `json:"last_round_seconds"`
	AverageRoundSeconds int    `json:"average_round_seconds"`
	LastStatus          string `json:"last_status"`
	LastExitSignal      bool   `json:"last_exit_signal"`
	LastFilesModified   int    `json:"last_files_modified"`
}

type SessionMeta struct {
	ThreadID  string `json:"thread_id"`
	UpdatedAt string `json:"updated_at"`
}

type Paths struct {
	Root           string
	LogDir         string
	StateFile      string
	LastOutputFile string
	LastJSONFile   string
	SummaryFile    string
	StatsFile      string
	SessionFile    string
}
