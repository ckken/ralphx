package hooks

import (
	"fmt"

	"github.com/ckken/ralphx/internal/contracts"
)

func EvaluateStopGuard(cfg GuardConfig, in GuardInput) Decision {
	if !cfg.Enabled {
		return Decision{Allow: true}
	}

	if cfg.BlockWhenIncomplete {
		if in.Result.Status == contracts.StatusInProgress || in.Result.Status == contracts.StatusBlocked {
			return Decision{
				Allow:   false,
				Reason:  "task_incomplete",
				Message: "Do not stop yet. Continue by executing the next bounded step or producing a concrete next-step plan.",
			}
		}
	}

	if cfg.BlockWhenChecklistOpen && in.ChecklistOpen > 0 {
		return Decision{
			Allow:   false,
			Reason:  "checklist_open",
			Message: fmt.Sprintf("Checklist still has %d open items. Continue the current branch of work.", in.ChecklistOpen),
		}
	}

	if cfg.BlockWhenVerificationMiss && in.TestsRequired && !in.TestsPassedNow {
		return Decision{
			Allow:   false,
			Reason:  "verification_missing",
			Message: "Verification evidence is missing or stale. Run the required checks before stopping.",
		}
	}

	return Decision{Allow: true}
}
