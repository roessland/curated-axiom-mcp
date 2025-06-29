package cserver

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/roessland/curated-axiom-mcp/pkg/axiom"
	"github.com/roessland/curated-axiom-mcp/pkg/config"
)

var starredQueriesTool = mcp.NewTool("debug_starred_queries")

// DebugStarredQueriesHandler lists all starred queries in Axiom
func DebugStarredQueriesHandler(appConfig *config.AppConfig) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Log the tool call request
		logToolCall("debug_starred_queries", request)
		
		clientConfig := &axiom.AxiomConfig{
			Token:   appConfig.Axiom.Token,
			OrgID:   appConfig.Axiom.OrgID,
			Dataset: appConfig.Axiom.Dataset,
			URL:     appConfig.Axiom.URL,
		}
		client := axiom.NewClient(clientConfig)
		queries, err := client.StarredQueries()
		if err != nil {
			return failedResult("failed to fetch starred queries"), nil
		}
		if len(queries) == 0 {
			return successResult("No starred queries found."), nil
		}
		content := fmt.Sprintf("Starred Queries (%d):\n\n", len(queries))
		for _, q := range queries {
			content += fmt.Sprintf("- %s (Dataset: %s)\n", q.Name, q.Dataset)
			content += fmt.Sprintf("  %s\n", q.Query.APL)
		}

		return successResult(content), nil
	}
}
