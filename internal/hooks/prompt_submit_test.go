package hooks

import "testing"

func TestPromptActivatesRalphx(t *testing.T) {
	cases := []struct {
		text string
		want bool
	}{
		{"$ralphx 开始", true},
		{"please use $ralphx continue", true},
		{"ralphx continue", true},
		{"doctor only", false},
	}
	for _, tc := range cases {
		if got := PromptActivatesRalphx(tc.text); got != tc.want {
			t.Fatalf("PromptActivatesRalphx(%q) = %v, want %v", tc.text, got, tc.want)
		}
	}
}
