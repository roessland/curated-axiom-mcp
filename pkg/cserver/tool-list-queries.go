package cserver

import (
	"context"
	"fmt"
	"sort"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/roessland/curated-axiom-mcp/pkg/config"
)

var listQueriesTool = mcp.NewTool("list_queries",
	mcp.WithDescription("List all available queries"),
	mcp.WithDestructiveHintAnnotation(false),
	mcp.WithOpenWorldHintAnnotation(false),
	mcp.WithIdempotentHintAnnotation(true),
	mcp.WithReadOnlyHintAnnotation(true),
)

func ListQueriesHandler(registry *config.Registry) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		queries, err := registry.ListQueries()
		if err != nil {
			return failedResult("failed to list queries"), nil
		}

		if len(queries) == 0 {
			return successResult("No queries available."), nil
		}

		// Sort queries by name
		names := make([]string, 0, len(queries))
		for name := range queries {
			names = append(names, name)
		}
		sort.Strings(names)

		// Display queries
		content := fmt.Sprintf("Available Queries (%d total):\n\n", len(queries))

		// Calculate max name length for alignment
		maxNameLen := 0
		for _, name := range names {
			if len(name) > maxNameLen {
				maxNameLen = len(name)
			}
		}

		for _, name := range names {
			description := queries[name]
			// Truncate long descriptions
			if len(description) > 80 {
				description = description[:77] + "..."
			}
			content += fmt.Sprintf("  %-*s  %s\n", maxNameLen, name, description)
		}

		return successResult(content), nil
	}
}
