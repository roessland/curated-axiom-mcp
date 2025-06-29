package cserver

import (
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// Respond to LLM that tool call was successful
func successResult(content string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{Type: "text", Text: content},
		},
	}
}

// Respond to LLM that tool call failed
func failedResult(content string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{Type: "text", Text: content},
		},
	}
}

// Respond to LLM that tool call failed
func errorResult(err error) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{Type: "text", Text: fmt.Sprintf("%v", err)},
		},
	}
}
