package assets

import (
	"encoding/json"
	"testing"
)

func TestLoopOutputSchemaAvoidsUnsupportedCompositionKeywords(t *testing.T) {
	var schema map[string]any
	if err := json.Unmarshal(defaultSchema, &schema); err != nil {
		t.Fatalf("unmarshal defaultSchema: %v", err)
	}
	for _, keyword := range []string{"allOf", "anyOf", "oneOf", "if", "then", "else"} {
		if _, ok := schema[keyword]; ok {
			t.Fatalf("loop output schema contains unsupported keyword %q", keyword)
		}
	}
	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("properties = %T, want map[string]any", schema["properties"])
	}
	requiredValues, ok := schema["required"].([]any)
	if !ok {
		t.Fatalf("required = %T, want []any", schema["required"])
	}
	required := map[string]bool{}
	for _, value := range requiredValues {
		name, ok := value.(string)
		if !ok {
			t.Fatalf("required item = %T, want string", value)
		}
		required[name] = true
	}
	for name := range properties {
		if !required[name] {
			t.Fatalf("loop output schema property %q is missing from required", name)
		}
	}
}
