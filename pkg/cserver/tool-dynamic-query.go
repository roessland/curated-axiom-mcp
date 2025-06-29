package cserver

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/roessland/curated-axiom-mcp/pkg/axiom"
	"github.com/roessland/curated-axiom-mcp/pkg/caxiom"
	"github.com/roessland/curated-axiom-mcp/pkg/config"
	"github.com/roessland/curated-axiom-mcp/pkg/formatter"
)

// CreateDynamicQueryHandler creates a handler for a dynamic query tool
func CreateDynamicQueryHandler(toolName string, registry *config.Registry, appConfig *config.AppConfig) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Log the tool call request
		logToolCall(toolName, request)
		
		// Get the dynamic query definition
		query, err := registry.GetDynamicQuery(toolName)
		if err != nil {
			return errorResult(fmt.Errorf("query not found: %w", err)), nil
		}

		// Extract parameters from the request
		params := make(map[string]interface{})
		for _, param := range query.Parameters {
			value, err := request.RequireString(param.Name)
			if err != nil {
				if param.Required {
					return errorResult(fmt.Errorf("missing required parameter %s: %w", param.Name, err)), nil
				}
				// Use empty string for optional missing parameters
				params[param.Name] = ""
			} else {
				params[param.Name] = value
			}
		}

		// Render the template with provided parameters
		templateExecutor := caxiom.NewTemplateExecutor()
		renderedAPL, err := templateExecutor.RenderTemplate(query.TemplateAPL, params)
		if err != nil {
			return errorResult(fmt.Errorf("failed to render query template: %w", err)), nil
		}
		
		// Debug: log the rendered APL
		slog.Debug("Rendered APL query", "tool_name", toolName, "rendered_apl", renderedAPL)

		// Create Axiom client and execute the query
		clientConfig := &axiom.AxiomConfig{
			Token:   appConfig.Axiom.Token,
			OrgID:   appConfig.Axiom.OrgID,
			Dataset: appConfig.Axiom.Dataset,
			URL:     appConfig.Axiom.URL,
		}
		client := axiom.NewClient(clientConfig)
		result, err := client.ExecuteQuery(renderedAPL)
		if err != nil {
			return failedResult(fmt.Sprintf("query execution failed: %v", err)), nil
		}

		// Format result for LLM
		llmFormatter := formatter.NewLLMFormatter()
		formatOptions := formatter.FormatOptions{
			Format:      "table",
			LLMFriendly: true,
			MaxRows:     100,
			APLQuery:    renderedAPL, // Pass the rendered APL query
		}

		formatted, err := llmFormatter.Format(result, formatOptions)
		if err != nil {
			return failedResult("failed to format results"), nil
		}

		// Format as markdown/plaintext response
		textResponse := formatAsMarkdown(formatted)
		return successResult(textResponse), nil
	}
}