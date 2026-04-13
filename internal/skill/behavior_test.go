package skill

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

type behaviorResponse struct {
	AssistantResponse string `json:"assistant_response"`
}

func TestSkillContinuesAfterIdentifyingNextEdge(t *testing.T) {
	if os.Getenv("RALPHX_RUN_BEHAVIOR_TESTS") != "1" {
		t.Skip("set RALPHX_RUN_BEHAVIOR_TESTS=1 to run forward behavior tests")
	}
	if _, err := exec.LookPath("codex"); err != nil {
		t.Skip("codex not available")
	}

	skillPath := filepath.Join("..", "..", "skills", "ralphx", "SKILL.md")
	skillData, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("read skill: %v", err)
	}

	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "schema.json")
	outputPath := filepath.Join(tmpDir, "output.json")
	schema := `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "additionalProperties": false,
  "required": ["assistant_response"],
  "properties": {
    "assistant_response": { "type": "string", "minLength": 1 }
  }
}`
	if err := os.WriteFile(schemaPath, []byte(schema), 0o644); err != nil {
		t.Fatalf("write schema: %v", err)
	}

	prompt := strings.TrimSpace(`
You are testing whether a skill enforces continuation instead of stopping at advice.

Skill:
` + string(skillData) + `

User message:
$ralphx 开始

Conversation context:
现在剩下最明显的 emit 架构边是 registry 单例本身：component/name/css/extra-file 这些还是 process-wide mutable singleton，平台代码还在直接读它们。下一轮就该开始把 registry state 从“全局实例”往“session-local handle”收。

Task:
Reply as the assistant should reply to the user. Do not explain the policy. Just produce the assistant reply.
`)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(
		ctx,
		"codex",
		"exec",
		"--skip-git-repo-check",
		"--dangerously-bypass-approvals-and-sandbox",
		"-C", filepath.Join("..", ".."),
		"--output-schema", schemaPath,
		"-o", outputPath,
		"-",
	)
	cmd.Stdin = strings.NewReader(prompt)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("codex exec failed: %v\n%s", err, string(out))
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	var resp behaviorResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		t.Fatalf("parse output: %v\n%s", err, string(data))
	}

	reply := strings.TrimSpace(resp.AssistantResponse)
	if reply == "" {
		t.Fatal("assistant response is empty")
	}
	if looksAdviceOnly(reply) {
		t.Fatalf("assistant stopped at advice instead of continuing:\n%s", reply)
	}
	if !looksActionable(reply) {
		t.Fatalf("assistant response does not look like a concrete next step or plan:\n%s", reply)
	}
}

func looksAdviceOnly(text string) bool {
	lower := strings.ToLower(text)
	if strings.Contains(lower, "下一轮") || strings.Contains(lower, "下一步我建议") || strings.Contains(lower, "建议下一步") {
		return !looksActionable(text)
	}
	return false
}

func looksActionable(text string) bool {
	actionMarkers := []string{
		"- [ ]",
		"下一步计划",
		"先做",
		"直接做",
		"bounded slice",
		"RegistryState",
		"EmitSession",
	}
	for _, marker := range actionMarkers {
		if strings.Contains(text, marker) {
			return true
		}
	}
	numbered := regexp.MustCompile(`(?m)^[0-9]+\.\s+`)
	return numbered.MatchString(text)
}
