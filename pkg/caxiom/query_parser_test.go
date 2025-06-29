package caxiom

import (
	"strings"
	"testing"
)

func TestParseStarredQuery(t *testing.T) {
	apl := `declare query_parameters (
    q_start_time:datetime = datetime(2025-06-25T00:00:00Z), ///param=datetime({{.StartTime}}),
    q_end_time:datetime = datetime(2025-06-26T00:00:00Z), ///param=datetime({{.EndTime}}),
    q_entity_id:string = 'example-entity-id' /// param='{{.EntityId}}'
);
['events']
| where _time > q_start_time and _time < q_end_time
| where id == q_entity_id
| limit 200

// CuratedAxiomMCP:
//   ToolName: entity_data
//   Params:
//     - Name: EntityId
//       Type: string
//       Example: example-entity-id
//     - Name: StartTime
//       Type: date-time
//       Example: 2025-06-25T00:00:00Z
//       Description: Start of interval to query
//     - Name: EndTime
//       Type: date-time
//       Example: 2025-06-26T00:00:00Z
//       Description: End of interval to query
//   Constraints:
//     - time between StartTime and EndTime must be 24 hours or less, to avoid query timing out`

	parsed, err := ParseStarredQuery("test-query", apl)
	if err != nil {
		t.Fatalf("Failed to parse query: %v", err)
	}

	// Test metadata parsing
	if parsed.Metadata.CuratedAxiomMCP.ToolName != "entity_data" {
		t.Errorf("Expected ToolName 'entity_data', got '%s'", parsed.Metadata.CuratedAxiomMCP.ToolName)
	}

	if len(parsed.Metadata.CuratedAxiomMCP.Params) != 3 {
		t.Errorf("Expected 3 parameters, got %d", len(parsed.Metadata.CuratedAxiomMCP.Params))
	}

	// Check parameter names
	expectedParams := []string{"EntityId", "StartTime", "EndTime"}
	for i, param := range parsed.Metadata.CuratedAxiomMCP.Params {
		if i < len(expectedParams) && param.Name != expectedParams[i] {
			t.Errorf("Expected parameter %d to be '%s', got '%s'", i, expectedParams[i], param.Name)
		}
	}

	// Test template conversion
	if !strings.Contains(parsed.TemplateAPL, "{{.StartTime}}") {
		t.Errorf("Template should contain {{.StartTime}}")
	}
	if !strings.Contains(parsed.TemplateAPL, "{{.EndTime}}") {
		t.Errorf("Template should contain {{.EndTime}}")
	}
	if !strings.Contains(parsed.TemplateAPL, "{{.EntityId}}") {
		t.Errorf("Template should contain {{.EntityId}}")
	}
}

func TestParseStarredQueryWithoutMarker(t *testing.T) {
	apl := `['events'] | limit 10`
	
	_, err := ParseStarredQuery("test-query", apl)
	if err == nil {
		t.Error("Expected error for query without CuratedAxiomMCP marker")
	}
}