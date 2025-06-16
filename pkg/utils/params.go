package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/roessland/curated-axiom-mcp/pkg/config"
)

// ParseParams parses command line parameters in the format key=value
func ParseParams(args []string) (map[string]string, error) {
	params := make(map[string]string)

	for _, arg := range args {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid parameter format: %s (expected key=value)", arg)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if key == "" {
			return nil, fmt.Errorf("empty parameter key in: %s", arg)
		}

		params[key] = value
	}

	return params, nil
}

// ValidateAndConvertParams validates parameters against query definition and converts types
func ValidateAndConvertParams(params map[string]string, query *config.Query) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Check for required parameters
	for _, param := range query.Parameters {
		value, exists := params[param.Name]

		if !exists {
			if param.Required {
				return nil, fmt.Errorf("required parameter %s is missing", param.Name)
			}
			// Use default value if available
			if param.Default != nil {
				result[param.Name] = param.Default
			}
			continue
		}

		// Convert and validate the parameter
		convertedValue, err := convertParameter(value, param)
		if err != nil {
			return nil, fmt.Errorf("parameter %s: %w", param.Name, err)
		}

		result[param.Name] = convertedValue
	}

	// Check for unknown parameters
	for key := range params {
		found := false
		for _, param := range query.Parameters {
			if param.Name == key {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("unknown parameter: %s", key)
		}
	}

	return result, nil
}

// convertParameter converts a string value to the appropriate type
func convertParameter(value string, param config.Parameter) (interface{}, error) {
	// Check enum values first
	if len(param.Enum) > 0 {
		found := false
		for _, enumValue := range param.Enum {
			if fmt.Sprintf("%v", enumValue) == value {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("value %s is not in allowed values %v", value, param.Enum)
		}
	}

	// Convert based on type
	switch param.Type {
	case "string":
		return value, nil

	case "int":
		intVal, err := strconv.Atoi(value)
		if err != nil {
			return nil, fmt.Errorf("invalid integer value: %s", value)
		}
		return intVal, nil

	case "float":
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid float value: %s", value)
		}
		return floatVal, nil

	case "boolean":
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return nil, fmt.Errorf("invalid boolean value: %s", value)
		}
		return boolVal, nil

	case "datetime":
		// Try multiple datetime formats
		formats := []string{
			time.RFC3339,
			"2006-01-02T15:04:05",
			"2006-01-02 15:04:05",
			"2006-01-02",
		}

		var parsedTime time.Time
		var err error
		for _, format := range formats {
			parsedTime, err = time.Parse(format, value)
			if err == nil {
				break
			}
		}
		if err != nil {
			return nil, fmt.Errorf("invalid datetime value %s (expected RFC3339 or similar format)", value)
		}
		return parsedTime.Format(time.RFC3339), nil

	case "duration":
		duration, err := time.ParseDuration(value)
		if err != nil {
			return nil, fmt.Errorf("invalid duration value: %s", value)
		}
		return duration.String(), nil

	default:
		return nil, fmt.Errorf("unsupported parameter type: %s", param.Type)
	}
}
