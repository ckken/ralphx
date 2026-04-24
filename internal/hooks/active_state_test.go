package hooks

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteReadAndClearActiveState(t *testing.T) {
	dir := t.TempDir()
	if err := WriteActiveState(dir, "$ralphx continue"); err != nil {
		t.Fatalf("WriteActiveState() error = %v", err)
	}
	state, err := ReadActiveState(dir)
	if err != nil {
		t.Fatalf("ReadActiveState() error = %v", err)
	}
	if !state.Active || state.Mode != "ralphx" {
		t.Fatalf("state = %#v", state)
	}
	if state.StopHookActive {
		t.Fatalf("state.StopHookActive = true, want false")
	}
	if err := MarkStopHookActive(dir, "active_workflow"); err != nil {
		t.Fatalf("MarkStopHookActive() error = %v", err)
	}
	marked, err := ReadActiveState(dir)
	if err != nil {
		t.Fatalf("ReadActiveState(marked) error = %v", err)
	}
	if !marked.StopHookActive || marked.StopReason != "active_workflow" {
		t.Fatalf("marked = %#v", marked)
	}
	if err := ClearActiveState(dir); err != nil {
		t.Fatalf("ClearActiveState() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "ralphx-active.json")); !os.IsNotExist(err) {
		t.Fatalf("expected file removed, err=%v", err)
	}
}
