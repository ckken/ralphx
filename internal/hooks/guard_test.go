package hooks

import (
	"testing"

	"github.com/ckken/ralphx/internal/contracts"
)

func TestEvaluateStopGuardBlocksIncompleteWork(t *testing.T) {
	decision := EvaluateStopGuard(GuardConfig{
		Enabled:                   true,
		BlockWhenChecklistOpen:    true,
		BlockWhenVerificationMiss: true,
		BlockWhenIncomplete:       true,
	}, GuardInput{
		Event:         EventStop,
		Result:        contracts.RoundResult{Status: contracts.StatusInProgress, Mode: contracts.ModeProducePlan},
		ChecklistOpen: 2,
		TestsRequired: true,
	})
	if decision.Allow {
		t.Fatal("expected stop guard to block incomplete work")
	}
	if decision.Reason != "task_incomplete" {
		t.Fatalf("reason = %q", decision.Reason)
	}
}

func TestEvaluateStopGuardBlocksMissingVerification(t *testing.T) {
	decision := EvaluateStopGuard(GuardConfig{
		Enabled:                   true,
		BlockWhenChecklistOpen:    true,
		BlockWhenVerificationMiss: true,
		BlockWhenIncomplete:       true,
	}, GuardInput{
		Event:          EventStop,
		Result:         contracts.RoundResult{Status: contracts.StatusComplete, Mode: contracts.ModeComplete, ExitSignal: true},
		ChecklistOpen:  0,
		TestsRequired:  true,
		TestsPassedNow: false,
	})
	if decision.Allow {
		t.Fatal("expected stop guard to block missing verification")
	}
	if decision.Reason != "verification_missing" {
		t.Fatalf("reason = %q", decision.Reason)
	}
}

func TestEvaluateStopGuardAllowsCleanCompletion(t *testing.T) {
	decision := EvaluateStopGuard(GuardConfig{
		Enabled:                   true,
		BlockWhenChecklistOpen:    true,
		BlockWhenVerificationMiss: true,
		BlockWhenIncomplete:       true,
	}, GuardInput{
		Event:          EventStop,
		Result:         contracts.RoundResult{Status: contracts.StatusComplete, Mode: contracts.ModeComplete, ExitSignal: true},
		ChecklistOpen:  0,
		TestsRequired:  true,
		TestsPassedNow: true,
	})
	if !decision.Allow {
		t.Fatalf("expected allow, got reason %q", decision.Reason)
	}
}
