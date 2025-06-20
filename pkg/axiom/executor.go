package axiom

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/roessland/curated-axiom-mcp/pkg/config"
)

// ExecuteQueryByName executes a named query with parameters
func (c *Client) ExecuteQueryByName(queryName string, params map[string]interface{}, registry *config.Registry) (*QueryResult, error) {
	// Get the query definition
	query, err := registry.GetQuery(queryName)
	if err != nil {
		return nil, err
	}

	// Substitute parameters in the APL query
	apl, err := substituteParameters(query.APLQuery, params)
	if err != nil {
		return nil, fmt.Errorf("failed to substitute parameters: %w", err)
	}

	// Extract dataset from APL query or use configured default
	dataset := query.Dataset
	if dataset == "" {
		dataset = extractDatasetFromAPL(apl)
	}
	if dataset == "" {
		dataset = c.config.Dataset
	}

	// Execute the query
	result, err := c.ExecuteQuery(apl, dataset)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// extractDatasetFromAPL extracts the dataset name from an APL query
// Looks for patterns like ['dataset_name'] or ["dataset_name"] at the beginning of the query
func extractDatasetFromAPL(apl string) string {
	// Match patterns like ['dataset'] or ["dataset"] at the start of lines
	re := regexp.MustCompile(`(?m)^\s*\[['"]([^'"]+)['"]\]`)
	matches := re.FindStringSubmatch(apl)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// substituteParameters replaces parameter placeholders in the APL query
func substituteParameters(apl string, params map[string]interface{}) (string, error) {
	result := apl

	for paramName, paramValue := range params {
		placeholder := fmt.Sprintf("{%s}", paramName)
		valueStr := formatParameterValue(paramValue)
		result = strings.ReplaceAll(result, placeholder, valueStr)
	}

	// Check for unsubstituted placeholders
	if strings.Contains(result, "{") && strings.Contains(result, "}") {
		// Find unsubstituted placeholders
		var unsubstituted []string
		start := 0
		for {
			openBrace := strings.Index(result[start:], "{")
			if openBrace == -1 {
				break
			}
			openBrace += start

			closeBrace := strings.Index(result[openBrace:], "}")
			if closeBrace == -1 {
				break
			}
			closeBrace += openBrace

			placeholder := result[openBrace+1 : closeBrace]
			unsubstituted = append(unsubstituted, placeholder)
			start = closeBrace + 1
		}

		if len(unsubstituted) > 0 {
			return "", fmt.Errorf("unsubstituted parameters: %v", unsubstituted)
		}
	}

	return result, nil
}

// formatParameterValue converts a parameter value to its string representation for APL
func formatParameterValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		// Escape quotes in string values
		escaped := strings.ReplaceAll(v, `"`, `\"`)
		return fmt.Sprintf(`"%s"`, escaped)
	case int:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	case float64:
		return fmt.Sprintf("%g", v)
	case bool:
		if v {
			return "true"
		}
		return "false"
	default:
		// For other types (like datetime strings), use string representation
		return fmt.Sprintf(`"%v"`, v)
	}
}
