package parallel

import "context"

type JobStatus string

type WorkerLifecycle string

const (
	JobPending   JobStatus = "pending"
	JobRunning   JobStatus = "running"
	JobBlocked   JobStatus = "blocked"
	JobComplete  JobStatus = "complete"
	JobCancelled JobStatus = "cancelled"

	WorkerStarting WorkerLifecycle = "starting"
	WorkerRunning  WorkerLifecycle = "running"
	WorkerStopping WorkerLifecycle = "stopping"
	WorkerExited   WorkerLifecycle = "exited"
	WorkerLost     WorkerLifecycle = "lost"
)

type Job struct {
	ID        string    `json:"id"`
	Goal      string    `json:"goal"`
	Scope     []string  `json:"scope,omitempty"`
	Verify    string    `json:"verify,omitempty"`
	WorkerID  string    `json:"worker_id,omitempty"`
	Status    JobStatus `json:"status"`
	Summary   string    `json:"summary,omitempty"`
	DependsOn []string  `json:"depends_on,omitempty"`
}

type WorkerState struct {
	ID         string          `json:"id"`
	Lifecycle  WorkerLifecycle `json:"lifecycle"`
	JobID      string          `json:"job_id,omitempty"`
	StartedAt  string          `json:"started_at,omitempty"`
	UpdatedAt  string          `json:"updated_at,omitempty"`
	LogPath    string          `json:"log_path,omitempty"`
	ResultPath string          `json:"result_path,omitempty"`
}

type WorkerResult struct {
	JobID         string   `json:"job_id"`
	WorkerID      string   `json:"worker_id"`
	Status        string   `json:"status"`
	ExitSignal    bool     `json:"exit_signal"`
	FilesModified int      `json:"files_modified"`
	TestsPassed   bool     `json:"tests_passed"`
	Blockers      []string `json:"blockers,omitempty"`
	Summary       string   `json:"summary,omitempty"`
}

type Worker interface {
	Execute(ctx context.Context, job Job) (WorkerResult, error)
}

type Scheduler interface {
	RunRound(ctx context.Context, jobs []Job) ([]WorkerResult, error)
}
