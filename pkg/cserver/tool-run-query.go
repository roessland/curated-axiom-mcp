package cserver

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/roessland/curated-axiom-mcp/pkg/axiom"
	"github.com/roessland/curated-axiom-mcp/pkg/config"
	"github.com/roessland/curated-axiom-mcp/pkg/formatter"
)

var runQueryTool = mcp.NewTool("run_query",
	mcp.WithDescription("Run an APL query"),
	mcp.WithDestructiveHintAnnotation(false),
	mcp.WithOpenWorldHintAnnotation(false),
	mcp.WithIdempotentHintAnnotation(true),
	mcp.WithReadOnlyHintAnnotation(true),
	mcp.WithString("apl", mcp.Required()),
)

func RunQueryHandler(registry *config.Registry, appConfig *config.AppConfig) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		apl, err := request.RequireString("apl")
		if err != nil {
			return errorResult(err), nil
		}

		// Create Axiom client
		clientConfig := &axiom.AxiomConfig{
			Token:   appConfig.Axiom.Token,
			OrgID:   appConfig.Axiom.OrgID,
			Dataset: appConfig.Axiom.Dataset,
			URL:     appConfig.Axiom.URL,
		}
		client := axiom.NewClient(clientConfig)

		// Execute the query
		result, err := client.ExecuteQuery(apl)
		if err != nil {
			return failedResult("query execution failed"), nil
		}

		// Format result for LLM
		llmFormatter := formatter.NewLLMFormatter()
		formatOptions := formatter.FormatOptions{
			Format:      "table",
			LLMFriendly: true,
			MaxRows:     100,
		}

		formatted, err := llmFormatter.Format(result, formatOptions)
		if err != nil {
			return failedResult("failed to format results"), nil
		}

		// Convert to JSON for MCP response
		jsonData, err := json.MarshalIndent(formatted, "", "  ")
		if err != nil {
			return failedResult("failed to serialize results"), nil
		}

		return successResult(string(jsonData)), nil
	}
}
