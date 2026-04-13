package hooks

import "github.com/ckken/ralphx/internal/contracts"

type Event string

const (
	EventSessionStart Event = "session-start"
	EventPromptSubmit Event = "prompt-submit"
	EventPreToolUse   Event = "pre-tool-use"
	EventPostToolUse  Event = "post-tool-use"
	EventTurnComplete Event = "turn-complete"
	EventStop         Event = "stop"
	EventSessionEnd   Event = "session-end"
)

type GuardConfig struct {
	Enabled                   bool
	BlockWhenChecklistOpen    bool
	BlockWhenVerificationMiss bool
	BlockWhenIncomplete       bool
}

type GuardInput struct {
	Event          Event
	Result         contracts.RoundResult
	ChecklistOpen  int
	TestsRequired  bool
	TestsPassedNow bool
}

type Decision struct {
	Allow   bool
	Reason  string
	Message string
}
