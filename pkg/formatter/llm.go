package formatter

import (
	"fmt"
	"strings"

	"github.com/axiomhq/axiom-go/axiom/query"
	"github.com/roessland/curated-axiom-mcp/pkg/axiom"
)

// LLMFormatter formats results for optimal LLM consumption
type LLMFormatter struct{}

// NewLLMFormatter creates a new LLM formatter
func NewLLMFormatter() *LLMFormatter {
	return &LLMFormatter{}
}

// Format formats the Axiom result for LLM consumption
func (f *LLMFormatter) Format(result *axiom.QueryResult, options FormatOptions) (*FormattedResult, error) {
	// Count total rows from first table
	totalRows := 0
	if len(result.Tables) > 0 && len(result.Tables[0].Columns) > 0 {
		totalRows = len(result.Tables[0].Columns[0])
	}

	formatted := &FormattedResult{
		Count:    totalRows,
		Warnings: []string{}, // No warnings in tabular format
		Metadata: make(map[string]interface{}),
	}

	// Add metadata
	if len(result.Tables) > 0 {
		formatted.Metadata["fields"] = result.Tables[0].Fields
	}
	formatted.Metadata["status"] = result.Status

	// Format as tabular data (primary format for /_apl endpoint)
	formatted.Data = f.formatTable(result, options)
	formatted.Summary = f.generateTableSummary(result, options)

	return formatted, nil
}

// formatTable creates a clean table format optimized for LLMs
func (f *LLMFormatter) formatTable(result *axiom.QueryResult, options FormatOptions) *TableResult {
	if len(result.Tables) == 0 || len(result.Tables[0].Columns) == 0 {
		headers := []string{}
		if len(result.Tables) > 0 {
			headers = extractFieldNames(result.Tables[0].Fields)
		}
		return &TableResult{
			Headers: headers,
			Rows:    [][]string{},
			Total:   0,
		}
	}

	table := result.Tables[0]
	headers := extractFieldNames(table.Fields)

	// Get number of rows from first column
	totalRows := len(table.Columns[0])

	// Limit rows if specified
	maxRows := totalRows
	if options.MaxRows > 0 && totalRows > options.MaxRows {
		maxRows = options.MaxRows
	}

	// Convert column-based data to row-based format
	rows := make([][]string, maxRows)
	for i := 0; i < maxRows; i++ {
		row := make([]string, len(table.Columns))
		for j, column := range table.Columns {
			if i < len(column) {
				row[j] = formatCellValue(column[i])
			} else {
				row[j] = ""
			}
		}
		rows[i] = row
	}

	return &TableResult{
		Headers: headers,
		Rows:    rows,
		Total:   totalRows,
	}
}

// generateTableSummary creates a human-readable summary of table data
func (f *LLMFormatter) generateTableSummary(result *axiom.QueryResult, options FormatOptions) string {
	if len(result.Tables) == 0 {
		return "No data tables found"
	}

	table := result.Tables[0]
	totalRows := 0
	if len(table.Columns) > 0 {
		totalRows = len(table.Columns[0])
	}

	displayRows := totalRows
	if options.MaxRows > 0 && totalRows > options.MaxRows {
		displayRows = options.MaxRows
	}

	var parts []string

	if totalRows == 0 {
		parts = append(parts, "No results found")
	} else {
		parts = append(parts, fmt.Sprintf("Found %d records", totalRows))

		if displayRows < totalRows {
			parts = append(parts, fmt.Sprintf("showing first %d", displayRows))
		}

		if len(table.Fields) > 0 {
			fieldNames := extractFieldNames(table.Fields)
			parts = append(parts, fmt.Sprintf("with %d fields: %s",
				len(fieldNames), strings.Join(fieldNames, ", ")))
		}
	}

	// Add query performance info
	if result.Status.ElapsedTime > 0 {
		parts = append(parts, fmt.Sprintf("(query took %dms)", result.Status.ElapsedTime))
	}

	return strings.Join(parts, " ")
}

// formatTimeSeries - removed as tabular format doesn't support time series
// formatSummary - removed as tabular format doesn't support summary buckets
// generateTimeSeriesSummary - removed as tabular format doesn't support time series
// generateSummary - removed as tabular format doesn't support summary buckets

// Helper functions

func extractFieldNames(fields []query.Field) []string {
	names := make([]string, len(fields))
	for i, field := range fields {
		names[i] = field.Name
	}
	return names
}

func formatCellValue(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%v", value)
}
