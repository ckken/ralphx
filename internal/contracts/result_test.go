package contracts

import "testing"

func TestRoundResultValidateRejectsAdviceOnlyExecuteNextStep(t *testing.T) {
	result := RoundResult{
		Status:        StatusInProgress,
		Mode:          ModeExecuteNextStep,
		ExitSignal:    false,
		FilesModified: 0,
		TestsPassed:   false,
		Summary:       "I recommend the next step",
	}
	if err := result.Validate(); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestRoundResultValidateAcceptsProducePlan(t *testing.T) {
	result := RoundResult{
		Status:      StatusInProgress,
		Mode:        ModeProducePlan,
		ExitSignal:  false,
		Summary:     "planned the next bounded slice",
		NextStep:    "Add RegistryState and attach it to EmitSession",
		TestsPassed: false,
	}
	if err := result.Validate(); err != nil {
		t.Fatalf("expected valid result, got %v", err)
	}
}
