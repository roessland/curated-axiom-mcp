package caxiom

import (
	"bytes"
	"fmt"
	"text/template"
)

// TemplateExecutor handles rendering of APL query templates
type TemplateExecutor struct{}

// NewTemplateExecutor creates a new template executor
func NewTemplateExecutor() *TemplateExecutor {
	return &TemplateExecutor{}
}

// RenderTemplate renders an APL template with the provided parameters
func (te *TemplateExecutor) RenderTemplate(templateAPL string, params map[string]interface{}) (string, error) {
	// Create a new template
	tmpl, err := template.New("apl_query").Parse(templateAPL)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Execute the template with the provided parameters
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, params); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}