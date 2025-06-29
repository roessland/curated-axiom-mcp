package caxiom

import (
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParsedQuery represents a starred query that has been parsed for MCP use
type ParsedQuery struct {
	Name         string
	OriginalAPL  string
	TemplateAPL  string
	Metadata     *QueryMetadata
}

// QueryMetadata represents the YAML metadata extracted from query comments
type QueryMetadata struct {
	CuratedAxiomMCP CuratedAxiomMCPConfig `yaml:"CuratedAxiomMCP"`
}

// CuratedAxiomMCPConfig represents the configuration for curated axiom MCP
type CuratedAxiomMCPConfig struct {
	ToolName    string                `yaml:"ToolName,omitempty"`
	Params      []ParameterDefinition `yaml:"Params,omitempty"`
	Constraints []string              `yaml:"Constraints,omitempty"`
	Description string                `yaml:"Description,omitempty"`
}

// ParameterDefinition represents a parameter definition from the YAML metadata
type ParameterDefinition struct {
	Name        string `yaml:"Name"`
	Type        string `yaml:"Type"`
	Example     string `yaml:"Example,omitempty"`
	Description string `yaml:"Description,omitempty"`
}

// ParseStarredQuery parses a starred query's APL for MCP usage
func ParseStarredQuery(queryName, apl string) (*ParsedQuery, error) {
	// Check if query contains any CuratedAxiomMCP marker
	hasCuratedMarker := strings.Contains(apl, "CuratedAxiomMCP")
	hasProperMarker := strings.Contains(apl, "CuratedAxiomMCP:")
	
	if !hasCuratedMarker {
		return nil, fmt.Errorf("query does not contain CuratedAxiomMCP marker")
	}
	
	if hasCuratedMarker && !hasProperMarker {
		return nil, fmt.Errorf("query contains 'CuratedAxiomMCP' but not the proper 'CuratedAxiomMCP:' marker format")
	}

	// Extract YAML metadata from comments
	metadata, err := extractYAMLMetadata(apl)
	if err != nil {
		return nil, fmt.Errorf("failed to extract YAML metadata: %w", err)
	}

	// Convert parameter declarations to template format
	templateAPL, err := convertToTemplate(apl)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to template: %w", err)
	}

	return &ParsedQuery{
		Name:         queryName,
		OriginalAPL:  apl,
		TemplateAPL:  templateAPL,
		Metadata:     metadata,
	}, nil
}

// extractYAMLMetadata extracts and parses YAML metadata from APL comments
func extractYAMLMetadata(apl string) (*QueryMetadata, error) {
	lines := strings.Split(apl, "\n")
	var yamlLines []string
	inYAMLSection := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Look for the start of YAML section
		if strings.Contains(trimmed, "// CuratedAxiomMCP:") {
			inYAMLSection = true
			// Include this line as it contains the root YAML key
			if strings.HasPrefix(line, "// ") {
				yamlContent := strings.TrimPrefix(line, "// ")
				yamlLines = append(yamlLines, yamlContent)
			} else if strings.HasPrefix(line, "//") {
				yamlContent := strings.TrimPrefix(line, "//")
				yamlLines = append(yamlLines, yamlContent)
			}
			continue
		}

		// If we're in the YAML section
		if inYAMLSection {
			// Stop if we hit an empty line or non-comment line
			if trimmed == "" || !strings.HasPrefix(trimmed, "//") {
				break
			}

			// Remove the comment prefix but preserve the original indentation structure
			if strings.HasPrefix(line, "// ") {
				yamlContent := strings.TrimPrefix(line, "// ")
				yamlLines = append(yamlLines, yamlContent)
			} else if strings.HasPrefix(line, "//") {
				yamlContent := strings.TrimPrefix(line, "//")
				yamlLines = append(yamlLines, yamlContent)
			}
		}
	}

	if len(yamlLines) == 0 {
		return &QueryMetadata{}, nil
	}

	// Join the YAML lines and parse
	yamlContent := strings.Join(yamlLines, "\n")
	var metadata QueryMetadata
	if err := yaml.Unmarshal([]byte(yamlContent), &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse YAML metadata (content: %q): %w", yamlContent, err)
	}

	return &metadata, nil
}

// convertToTemplate converts APL with ///param= annotations to Go template format
func convertToTemplate(apl string) (string, error) {
	// Regular expression to match parameter declarations with ///param= annotations
	// Matches: param_name:type = value [,] ///param=template_replacement
	// Handle both comma and no-comma cases
	paramRegex := regexp.MustCompile(`(\w+):(\w+)\s*=\s*([^,\n]+?)(?:,)?\s*///param=([^,\n]+)`)

	// Find all parameter replacements
	templateAPL := apl
	matches := paramRegex.FindAllStringSubmatch(apl, -1)

	for _, match := range matches {
		if len(match) < 5 {
			continue
		}

		paramName := match[1]
		paramType := match[2]
		templateReplacement := strings.TrimSpace(match[4])

		// Replace the entire parameter declaration with the template version
		originalDeclaration := match[0]
		newDeclaration := paramName + ":" + paramType + " = " + templateReplacement
		templateAPL = strings.ReplaceAll(templateAPL, originalDeclaration, newDeclaration)
	}

	return templateAPL, nil
}