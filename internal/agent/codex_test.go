package agent

import "testing"

func TestExtractAgentMessageAndSession(t *testing.T) {
	raw := []byte("{\"type\":\"thread.started\",\"thread_id\":\"thread-123\"}\n{\"type\":\"item.completed\",\"item\":{\"type\":\"agent_message\",\"text\":\"{\\\"status\\\":\\\"in_progress\\\",\\\"exit_signal\\\":false,\\\"files_modified\\\":0,\\\"tests_passed\\\":false,\\\"blockers\\\":[],\\\"summary\\\":\\\"ok\\\"}\"}}\n")
	message, sessionID := extractAgentMessageAndSession(raw)
	if sessionID != "thread-123" {
		t.Fatalf("sessionID = %q", sessionID)
	}
	if message == "" {
		t.Fatal("expected message text")
	}
}
