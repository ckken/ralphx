package plan

import "testing"

func TestMergeChecklistPreservesCompletedAndAppendsNew(t *testing.T) {
	existing := "# task checklist\n\n- [x] done one\n- [ ] open one\n"
	next := []string{"open one", "next two"}
	got := mergeChecklist(existing, next)
	if len(got) != 3 {
		t.Fatalf("len = %d, want 3", len(got))
	}
	if got[0] != "done one" {
		t.Fatalf("got[0] = %q", got[0])
	}
	if got[1] != "open one" {
		t.Fatalf("got[1] = %q", got[1])
	}
	if got[2] != "next two" {
		t.Fatalf("got[2] = %q", got[2])
	}
}
