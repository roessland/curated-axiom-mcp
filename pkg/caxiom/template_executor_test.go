package caxiom

import (
	"strings"
	"testing"
)

func TestRenderTemplate(t *testing.T) {
	executor := NewTemplateExecutor()

	// Test template with basic substitution
	templateAPL := `['events']
| where _time > q_start_time:q_start_time = datetime({{.StartTime}})
| where _time < q_end_time:q_end_time = datetime({{.EndTime}})
| where id == q_entity_id:q_entity_id = '{{.EntityId}}'
| limit 200`

	params := map[string]interface{}{
		"StartTime": "2025-06-25T00:00:00Z",
		"EndTime":   "2025-06-26T00:00:00Z",
		"EntityId":  "example-entity-id",
	}

	result, err := executor.RenderTemplate(templateAPL, params)
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	// Check that parameters were substituted
	if !strings.Contains(result, "2025-06-25T00:00:00Z") {
		t.Error("Template should contain StartTime value")
	}
	if !strings.Contains(result, "2025-06-26T00:00:00Z") {
		t.Error("Template should contain EndTime value")
	}
	if !strings.Contains(result, "example-entity-id") {
		t.Error("Template should contain EntityId value")
	}

	// Check that template placeholders were replaced
	if strings.Contains(result, "{{.StartTime}}") {
		t.Error("Template should not contain unreplaced placeholders")
	}
}

func TestRenderTemplateError(t *testing.T) {
	executor := NewTemplateExecutor()

	// Test with invalid template syntax
	invalidTemplate := `['events'] | where id == {{.InvalidSyntax`

	params := map[string]interface{}{
		"InvalidSyntax": "test",
	}

	_, err := executor.RenderTemplate(invalidTemplate, params)
	if err == nil {
		t.Error("Expected error for invalid template syntax")
	}
}