package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadQueries loads query definitions from a YAML file
func LoadQueries(filePath string) (*QueryRegistry, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		// If file doesn't exist, return empty registry (queries will come from Axiom starred queries)
		if os.IsNotExist(err) {
			return &QueryRegistry{
				Queries: make(map[string]Query),
			}, nil
		}
		return nil, fmt.Errorf("failed to read queries file %s: %w", filePath, err)
	}

	var registry QueryRegistry
	if err := yaml.Unmarshal(data, &registry); err != nil {
		return nil, fmt.Errorf("failed to parse queries YAML: %w", err)
	}

	// Validate queries
	if err := validateQueries(&registry); err != nil {
		return nil, fmt.Errorf("invalid queries: %w", err)
	}

	return &registry, nil
}

// validateQueries validates the loaded query registry
func validateQueries(registry *QueryRegistry) error {
	// Allow empty queries since they will be fetched from Axiom starred queries
	if len(registry.Queries) == 0 {
		return nil
	}

	for name, query := range registry.Queries {
		if query.Name == "" {
			return fmt.Errorf("query %s: name is required", name)
		}
		if query.Description == "" {
			return fmt.Errorf("query %s: description is required", name)
		}
		if query.APLQuery == "" {
			return fmt.Errorf("query %s: apl_query is required", name)
		}

		// Validate parameters
		for i, param := range query.Parameters {
			if param.Name == "" {
				return fmt.Errorf("query %s: parameter %d name is required", name, i)
			}
			if param.Type == "" {
				return fmt.Errorf("query %s: parameter %s type is required", name, param.Name)
			}
			// Validate parameter type
			if !isValidParameterType(param.Type) {
				return fmt.Errorf("query %s: parameter %s has invalid type %s", name, param.Name, param.Type)
			}
		}
	}

	return nil
}

// isValidParameterType checks if the parameter type is valid
func isValidParameterType(paramType string) bool {
	validTypes := []string{"string", "int", "float", "datetime", "duration", "boolean"}
	for _, validType := range validTypes {
		if paramType == validType {
			return true
		}
	}
	return false
}
