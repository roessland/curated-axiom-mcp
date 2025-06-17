package server

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/roessland/curated-axiom-mcp/pkg/axiom"
	"github.com/roessland/curated-axiom-mcp/pkg/config"
	"github.com/roessland/curated-axiom-mcp/pkg/formatter"
	"github.com/roessland/curated-axiom-mcp/pkg/utils"
)

// StartStdioServer starts the MCP server in stdio mode
func StartStdioServer(appConfig *config.AppConfig, registry *config.Registry) error {
	// Create MCP server
	s := server.NewDefaultServer("curated-axiom-mcp", "1.0.0")

	// Register handlers
	s.HandleListTools(func(ctx context.Context, cursor *string) (*mcp.ListToolsResult, error) {
		queries, err := registry.GetAllQueries()
		if err != nil {
			return nil, err
		}

		var tools []mcp.Tool
		for name, query := range queries {
			tool := mcp.Tool{
				Name:        name,
				Description: query.Description,
				InputSchema: mcp.ToolInputSchema{
					Type:       "object",
					Properties: make(map[string]interface{}),
				},
			}

			// Add parameters to schema
			for _, param := range query.Parameters {
				paramSchema := map[string]interface{}{
					"type":        convertParamType(param.Type),
					"description": param.Description,
				}

				if len(param.Enum) > 0 {
					paramSchema["enum"] = param.Enum
				}

				tool.InputSchema.Properties[param.Name] = paramSchema
			}

			tools = append(tools, tool)
		}

		return &mcp.ListToolsResult{
			Tools: tools,
		}, nil
	})

	// Register call tool handler
	s.HandleCallTool(func(ctx context.Context, name string, arguments map[string]interface{}) (*mcp.CallToolResult, error) {
		// Get the query definition
		query, err := registry.GetQuery(name)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{Type: "text", Text: fmt.Sprintf("Query not found: %v", err)},
				},
			}, nil
		}

		// Convert arguments to string map for validation
		params := make(map[string]string)
		for key, value := range arguments {
			params[key] = fmt.Sprintf("%v", value)
		}

		// Validate and convert parameters
		convertedParams, err := utils.ValidateAndConvertParams(params, query)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{Type: "text", Text: fmt.Sprintf("Parameter validation failed: %v", err)},
				},
			}, nil
		}

		// Create Axiom client
		client := axiom.NewClient(&appConfig.Axiom)

		// Execute the query
		result, err := client.ExecuteQueryByName(name, convertedParams, registry)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{Type: "text", Text: fmt.Sprintf("Query execution failed: %v", err)},
				},
			}, nil
		}

		// Format result for LLM
		llmFormatter := formatter.NewLLMFormatter()
		formatOptions := formatter.FormatOptions{
			Format:      "table",
			LLMFriendly: query.LLMFriendly,
			MaxRows:     100,
			IncludeRaw:  false,
		}

		formatted, err := llmFormatter.Format(result, formatOptions)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{Type: "text", Text: fmt.Sprintf("Failed to format results: %v", err)},
				},
			}, nil
		}

		// Convert to JSON for MCP response
		jsonData, err := json.MarshalIndent(formatted, "", "  ")
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{Type: "text", Text: fmt.Sprintf("Failed to serialize results: %v", err)},
				},
			}, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{Type: "text", Text: string(jsonData)},
			},
		}, nil
	})

	// Start stdio server
	fmt.Fprintf(os.Stderr, "Starting MCP server (stdio mode)...\n")
	if err := server.ServeStdio(s); err != nil {
		return fmt.Errorf("stdio server error: %w", err)
	}

	return nil
}

// convertParamType converts our parameter types to JSON Schema types
func convertParamType(paramType string) string {
	switch paramType {
	case "int":
		return "integer"
	case "float":
		return "number"
	case "boolean":
		return "boolean"
	case "datetime", "duration", "string":
		return "string"
	default:
		return "string"
	}
}
