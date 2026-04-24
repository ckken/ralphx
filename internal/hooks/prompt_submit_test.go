package hooks

import "testing"

func TestPromptActivatesRalphx(t *testing.T) {
	cases := []struct {
		text string
		want bool
	}{
		{"$ralphx", true},
		{"  $ralphx  ", true},
		{"$ralphx 开始", false},
		{"$ralphx continue", false},
		{"please use $ralphx continue", false},
		{"ralphx", false},
		{"ralph continue", false},
		{"doctor only", false},
	}
	for _, tc := range cases {
		if got := PromptActivatesRalphx(tc.text); got != tc.want {
			t.Fatalf("PromptActivatesRalphx(%q) = %v, want %v", tc.text, got, tc.want)
		}
	}
}

func TestPromptStopsRalphx(t *testing.T) {
	cases := []struct {
		text string
		want bool
	}{
		{"stop", true},
		{"测试下是否有 stop", true},
		{"please stop now", true},
		{"cancel", true},
		{"abort", true},
		{"结束当前工作流", true},
		{"停止当前会话", true},
		{"continue", false},
		{"$ralphx 继续", false},
	}
	for _, tc := range cases {
		if got := PromptStopsRalphx(tc.text); got != tc.want {
			t.Fatalf("PromptStopsRalphx(%q) = %v, want %v", tc.text, got, tc.want)
		}
	}
}
