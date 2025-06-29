package cserver

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/roessland/curated-axiom-mcp/pkg/formatter"
)

// writeDebugLog writes debug information to stdout.log file
func writeDebugLog(content string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return // Silently fail
	}
	
	logDir := filepath.Join(homeDir, ".config", "curated-axiom-mcp")
	logFile := filepath.Join(logDir, "stdout.log")
	
	// Ensure directory exists
	os.MkdirAll(logDir, 0755)
	
	// Open file in append mode
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return // Silently fail
	}
	defer f.Close()
	
	// Write with timestamp
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	f.WriteString(fmt.Sprintf("[%s] %s\n\n", timestamp, content))
}

// logToolCall logs the tool call request for debugging
func logToolCall(toolName string, request mcp.CallToolRequest) {
	// Convert request arguments to JSON for readable output
	argsJson, _ := json.MarshalIndent(request.Params.Arguments, "", "  ")
	
	content := fmt.Sprintf("==== TOOL CALL REQUEST ====\nTool: %s\nArguments:\n%s", 
		toolName, string(argsJson))
	writeDebugLog(content)
}

// logToolResponse logs the tool call response for debugging
func logToolResponse(response *mcp.CallToolResult) {
	// Extract text content from response
	var responseText string
	if len(response.Content) > 0 {
		if textContent, ok := response.Content[0].(mcp.TextContent); ok {
			responseText = textContent.Text
		}
	}
	
	// Truncate very long responses for readability
	if len(responseText) > 5000 {
		responseText = responseText[:5000] + "\n... (truncated, " + 
			fmt.Sprintf("%d", len(responseText)) + " total chars)"
	}
	
	content := fmt.Sprintf("==== TOOL CALL RESPONSE ====\n%s", responseText)
	writeDebugLog(content)
}

// formatAsMarkdown formats a FormattedResult as markdown text
func formatAsMarkdown(result *formatter.FormattedResult) string {
	var builder strings.Builder
	
	// Write summary
	builder.WriteString(result.Summary)
	builder.WriteString("\n\n")
	
	// Add APL section if query is available
	if result.APLQuery != "" {
		builder.WriteString("## APL\n\n")
		builder.WriteString("```apl\n")
		builder.WriteString(result.APLQuery)
		builder.WriteString("\n```\n\n")
	}
	
	// Add column statistics if available
	if columnStats, ok := result.Metadata["column_stats"].(map[string]formatter.ColumnStats); ok && len(columnStats) > 0 {
		builder.WriteString("## Column Stats\n\n")
		
		for colName, stats := range columnStats {
			builder.WriteString(fmt.Sprintf("### %s\n\n", colName))
			builder.WriteString(fmt.Sprintf("- **Type**: %s\n", stats.Type))
			
			// Always show first/last, even if empty
			if stats.First == "" {
				builder.WriteString("- **First**: (empty)\n")
			} else {
				builder.WriteString(fmt.Sprintf("- **First**: %s\n", stats.First))
			}
			if stats.Last == "" {
				builder.WriteString("- **Last**: (empty)\n")
			} else {
				builder.WriteString(fmt.Sprintf("- **Last**: %s\n", stats.Last))
			}
			
			if len(stats.UniqueValues) > 0 || stats.TotalUniqueCount > 0 {
				// Show total unique count first
				if stats.TotalUniqueCount > len(stats.UniqueValues) {
					builder.WriteString(fmt.Sprintf("- **Unique values**: %d total", stats.TotalUniqueCount))
					if len(stats.UniqueValues) > 0 {
						builder.WriteString(", top values: ")
					}
				} else {
					builder.WriteString("- **Unique values**: ")
				}
				
				// Show top values (limit to 8 for readability)
				if len(stats.UniqueValues) > 0 {
					// Sort by frequency (descending)
					type valueCount struct {
						value string
						count int
					}
					var sortedValues []valueCount
					for value, count := range stats.UniqueValues {
						sortedValues = append(sortedValues, valueCount{value, count})
					}
					
					// Simple bubble sort by count (descending)
					for i := 0; i < len(sortedValues)-1; i++ {
						for j := 0; j < len(sortedValues)-i-1; j++ {
							if sortedValues[j].count < sortedValues[j+1].count {
								sortedValues[j], sortedValues[j+1] = sortedValues[j+1], sortedValues[j]
							}
						}
					}
					
					// Show top 8 values
					maxShow := 8
					if len(sortedValues) < maxShow {
						maxShow = len(sortedValues)
					}
					
					for i := 0; i < maxShow; i++ {
						if i > 0 {
							builder.WriteString(", ")
						}
						
						vc := sortedValues[i]
						if vc.value == "" {
							builder.WriteString(fmt.Sprintf("(empty) (%d)", vc.count))
						} else {
							// Truncate long values in the summary
							displayValue := vc.value
							if len(displayValue) > 25 {
								displayValue = displayValue[:25] + "..."
							}
							builder.WriteString(fmt.Sprintf("%s (%d)", displayValue, vc.count))
						}
					}
					
					// Add "and X more" if there are more values
					if len(stats.UniqueValues) > maxShow {
						remaining := len(stats.UniqueValues) - maxShow
						builder.WriteString(fmt.Sprintf(", and %d more", remaining))
					}
				}
				builder.WriteString("\n")
			}
			
			if stats.NullCount > 0 {
				builder.WriteString(fmt.Sprintf("- **Null/empty count**: %d\n", stats.NullCount))
			}
			
			if len(stats.Examples) > 0 {
				builder.WriteString("- **Examples**: ")
				for i, example := range stats.Examples {
					if i > 0 {
						builder.WriteString(", ")
					}
					// Show full examples for complex data like JSON
					if len(example) > 100 {
						builder.WriteString(fmt.Sprintf("`%s`", example[:100]+"..."))
					} else {
						builder.WriteString(fmt.Sprintf("`%s`", example))
					}
				}
				builder.WriteString("\n")
			}
			
			builder.WriteString("\n")
		}
	}

	// If there's CSV data, format it nicely
	if csvData, ok := result.Data.(string); ok && csvData != "" {
		builder.WriteString("## Results\n\n")
		
		// Split CSV into lines
		lines := strings.Split(strings.TrimSpace(csvData), "\n")
		if len(lines) > 0 {
			// Write CSV with proper markdown formatting
			for i, line := range lines {
				if i == 0 {
					// Header line - add separator after it
					builder.WriteString(line)
					builder.WriteString("\n")
					// Add separator line
					headerCount := strings.Count(line, ",") + 1
					for j := 0; j < headerCount; j++ {
						if j > 0 {
							builder.WriteString(",")
						}
						builder.WriteString("---")
					}
					builder.WriteString("\n")
				} else {
					builder.WriteString(line)
					builder.WriteString("\n")
				}
			}
		}
	}
	
	return builder.String()
}

// Respond to LLM that tool call was successful
func successResult(content string) *mcp.CallToolResult {
	// Check response size and log warning if over 20KB
	responseSize := len(content)
	const maxRecommendedSize = 20 * 1024 // 20KB
	
	if responseSize > maxRecommendedSize {
		slog.Warn("MCP tool response exceeds recommended size - consider summarizing large responses to avoid wasting context",
			"size_bytes", responseSize,
			"size_kb", responseSize/1024, 
			"max_recommended_kb", maxRecommendedSize/1024)
	}
	
	result := &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{Type: "text", Text: content},
		},
	}
	
	// Log response for debugging
	logToolResponse(result)
	
	return result
}

// Respond to LLM that tool call failed
func failedResult(content string) *mcp.CallToolResult {
	result := &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{Type: "text", Text: content},
		},
	}
	
	// Log response for debugging
	logToolResponse(result)
	
	return result
}

// Respond to LLM that tool call failed
func errorResult(err error) *mcp.CallToolResult {
	result := &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{Type: "text", Text: fmt.Sprintf("%v", err)},
		},
	}
	
	// Log response for debugging
	logToolResponse(result)
	
	return result
}
