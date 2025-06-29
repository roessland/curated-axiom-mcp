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
		APLQuery: options.APLQuery,
	}

	// Add metadata
	if len(result.Tables) > 0 {
		formatted.Metadata["fields"] = result.Tables[0].Fields
	}
	formatted.Metadata["status"] = result.Status

	// Format as CSV data (primary format for /_apl endpoint)
	csvData := f.formatAsCSV(result, options)
	columnStats := f.generateColumnStats(result, options)
	
	formatted.Data = csvData
	formatted.Summary = f.generateTableSummary(result, options)
	formatted.Metadata["column_stats"] = columnStats

	return formatted, nil
}

// formatAsCSV creates a CSV string optimized for LLMs
func (f *LLMFormatter) formatAsCSV(result *axiom.QueryResult, options FormatOptions) string {
	if len(result.Tables) == 0 || len(result.Tables[0].Columns) == 0 {
		return ""
	}

	table := result.Tables[0]
	headers := extractFieldNames(table.Fields)

	// Start building CSV output
	var csvBuilder strings.Builder
	
	// Write headers
	for i, header := range headers {
		if i > 0 {
			csvBuilder.WriteString(",")
		}
		csvBuilder.WriteString(formatCSVValue(header))
	}
	csvBuilder.WriteString("\n")

	// Write column types as second row
	for i, field := range table.Fields {
		if i > 0 {
			csvBuilder.WriteString(",")
		}
		csvBuilder.WriteString(formatCSVValue(field.Type))
	}
	csvBuilder.WriteString("\n")

	// Write separator row
	csvBuilder.WriteString("---\n")

	// Get total number of rows
	totalRows := 0
	if len(table.Columns) > 0 {
		totalRows = len(table.Columns[0])
	}

	// Limit rows if specified
	maxRows := totalRows
	if options.MaxRows > 0 && totalRows > options.MaxRows {
		maxRows = options.MaxRows
	}

	// Write data rows using the Table.Rows() iterator
	rowIndex := 0
	for row := range table.Rows() {
		if rowIndex >= maxRows {
			break
		}
		
		// Write row values
		for i, value := range row {
			if i > 0 {
				csvBuilder.WriteString(",")
			}
			csvBuilder.WriteString(formatCSVValue(value))
		}
		csvBuilder.WriteString("\n")
		rowIndex++
	}

	return csvBuilder.String()
}

// formatTable creates a CSV format optimized for LLMs (legacy method for compatibility)
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

	// Get total number of rows
	totalRows := 0
	if len(table.Columns) > 0 {
		totalRows = len(table.Columns[0])
	}

	// Limit rows if specified
	maxRows := totalRows
	if options.MaxRows > 0 && totalRows > options.MaxRows {
		maxRows = options.MaxRows
	}

	// Use the Table.Rows() iterator to get properly formatted rows
	rows := make([][]string, 0, maxRows)
	rowIndex := 0
	
	// Iterate through rows using the table iterator
	for row := range table.Rows() {
		if rowIndex >= maxRows {
			break
		}
		
		// Convert row values to CSV-formatted strings
		csvRow := make([]string, len(row))
		for i, value := range row {
			csvRow[i] = formatCSVValue(value)
		}
		
		rows = append(rows, csvRow)
		rowIndex++
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

	if totalRows == 0 {
		// Add query performance info for no results
		if result.Status.ElapsedTime > 0 {
			// Convert nanoseconds to seconds with 1 decimal place
			elapsedSec := float64(result.Status.ElapsedTime) / 1000000000
			return fmt.Sprintf("No results found (query took %.1f s)", elapsedSec)
		}
		return "No results found"
	}

	// Format with results
	var summary strings.Builder
	summary.WriteString(fmt.Sprintf("Found %d records", totalRows))
	
	// Add timing info
	if result.Status.ElapsedTime > 0 {
		// Convert nanoseconds to seconds with 1 decimal place
		elapsedSec := float64(result.Status.ElapsedTime) / 1000000000
		summary.WriteString(fmt.Sprintf(" in %.1f s", elapsedSec))
	}
	
	// Add display info
	if displayRows < totalRows {
		summary.WriteString(fmt.Sprintf(". Showing first %d", displayRows))
	}
	
	// Add field count
	if len(table.Fields) > 0 {
		summary.WriteString(fmt.Sprintf(" with %d fields", len(table.Fields)))
	}
	
	summary.WriteString(".")

	return summary.String()
}

// ColumnStats represents statistics for a single column
type ColumnStats struct {
	First            string            `json:"first"`
	Last             string            `json:"last"`
	NullCount        int               `json:"null_count"`
	UniqueValues     map[string]int    `json:"unique_values"`
	TotalUniqueCount int               `json:"total_unique_count"` // Actual count of unique values seen
	Examples         []string          `json:"examples"`
	Type             string            `json:"type"`
}

// generateColumnStats creates statistics for each column
func (f *LLMFormatter) generateColumnStats(result *axiom.QueryResult, options FormatOptions) map[string]ColumnStats {
	if len(result.Tables) == 0 || len(result.Tables[0].Columns) == 0 {
		return make(map[string]ColumnStats)
	}

	table := result.Tables[0]
	stats := make(map[string]ColumnStats)
	
	// Get total number of rows
	totalRows := 0
	if len(table.Columns) > 0 {
		totalRows = len(table.Columns[0])
	}
	
	if totalRows == 0 {
		return stats
	}

	// Determine sampling strategy - sample if > 1000 rows
	sampleRate := 1
	if totalRows > 1000 {
		sampleRate = totalRows / 500 // Take ~500 samples
		if sampleRate < 1 {
			sampleRate = 1
		}
	}

	// Process each column
	for colIndex, field := range table.Fields {
		if colIndex >= len(table.Columns) {
			continue
		}
		
		column := table.Columns[colIndex]
		colStats := ColumnStats{
			Type:         field.Type,
			UniqueValues: make(map[string]int),
			Examples:     make([]string, 0),
		}
		
		valueCounts := make(map[string]int)
		uniqueValuesSet := make(map[string]bool) // Track all unique values
		examplesSet := make(map[string]bool)     // Track examples to avoid duplicates
		var firstValue, lastValue string
		var examples []string
		nullCount := 0
		
		// Sample through the data
		for i := 0; i < len(column); i += sampleRate {
			if i >= len(column) {
				break
			}
			
			value := column[i]
			valueStr := formatCellValue(value)
			
			// Track first and last (including nulls)
			if i == 0 {
				firstValue = valueStr
			}
			if i == len(column)-sampleRate || i+sampleRate >= len(column) {
				lastValue = valueStr
			}
			
			// Count nulls/empty
			if value == nil || valueStr == "" {
				nullCount++
				continue
			}
			
			// Track all unique values for total count
			uniqueValuesSet[valueStr] = true
			
			// Count frequencies for top values (limit tracking to top 100 to avoid huge maps)
			if len(valueCounts) < 100 {
				valueCounts[valueStr]++
			} else {
				// Still count if we've seen this value before
				if _, exists := valueCounts[valueStr]; exists {
					valueCounts[valueStr]++
				}
			}
			
			// Collect unique examples (limit to 5, prefer longer/complex values)
			if !examplesSet[valueStr] && len(examples) < 5 {
				examples = append(examples, valueStr)
				examplesSet[valueStr] = true
			} else if !examplesSet[valueStr] && len(valueStr) > 50 {
				// Replace shortest example with longer one if we have a complex value
				shortestIdx := 0
				shortestLen := len(examples[0])
				for idx, ex := range examples {
					if len(ex) < shortestLen {
						shortestIdx = idx
						shortestLen = len(ex)
					}
				}
				if len(valueStr) > shortestLen {
					delete(examplesSet, examples[shortestIdx])
					examples[shortestIdx] = valueStr
					examplesSet[valueStr] = true
				}
			}
		}
		
		colStats.First = firstValue
		colStats.Last = lastValue
		colStats.NullCount = nullCount
		colStats.UniqueValues = valueCounts
		colStats.TotalUniqueCount = len(uniqueValuesSet)
		colStats.Examples = examples
		
		stats[field.Name] = colStats
	}
	
	return stats
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

// formatCSVValue formats a value for CSV output using single quotes when needed
func formatCSVValue(value interface{}) string {
	if value == nil {
		return ""
	}
	
	str := fmt.Sprintf("%v", value)
	
	// Check if the value needs quoting (contains comma, newline, or single quote)
	needsQuoting := strings.ContainsAny(str, ",\n\r'")
	
	if needsQuoting {
		// Escape single quotes by doubling them
		escaped := strings.ReplaceAll(str, "'", "''")
		return "'" + escaped + "'"
	}
	
	return str
}
