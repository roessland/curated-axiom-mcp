package formatter

import (
	"fmt"
	"strings"

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
	formatted := &FormattedResult{
		Count:    len(result.Matches),
		Warnings: result.Warnings,
		Metadata: make(map[string]interface{}),
	}

	// Add metadata
	formatted.Metadata["fields"] = result.Fields
	formatted.Metadata["status"] = result.Status

	// Determine the best format based on data structure
	if len(result.Buckets.Series) > 0 {
		// Time series data
		formatted.Data = f.formatTimeSeries(result, options)
		formatted.Summary = f.generateTimeSeriesSummary(result)
	} else if len(result.Buckets.Totals) > 0 {
		// Aggregated data
		formatted.Data = f.formatSummary(result, options)
		formatted.Summary = f.generateSummary(result)
	} else {
		// Tabular data
		formatted.Data = f.formatTable(result, options)
		formatted.Summary = f.generateTableSummary(result, options)
	}

	return formatted, nil
}

// formatTable creates a clean table format optimized for LLMs
func (f *LLMFormatter) formatTable(result *axiom.QueryResult, options FormatOptions) *TableResult {
	if len(result.Matches) == 0 {
		return &TableResult{
			Headers: extractFieldNames(result.Fields),
			Rows:    [][]string{},
			Total:   0,
		}
	}

	headers := extractFieldNames(result.Fields)

	// Limit rows if specified
	maxRows := len(result.Matches)
	if options.MaxRows > 0 && maxRows > options.MaxRows {
		maxRows = options.MaxRows
	}

	rows := make([][]string, maxRows)
	for i := 0; i < maxRows; i++ {
		row := make([]string, len(result.Matches[i]))
		for j, cell := range result.Matches[i] {
			row[j] = formatCellValue(cell)
		}
		rows[i] = row
	}

	return &TableResult{
		Headers: headers,
		Rows:    rows,
		Total:   len(result.Matches),
	}
}

// formatTimeSeries creates time series format
func (f *LLMFormatter) formatTimeSeries(result *axiom.QueryResult, options FormatOptions) *TimeSeriesResult {
	series := make([]Series, len(result.Buckets.Series))

	for i, bucketSeries := range result.Buckets.Series {
		seriesData := Series{
			Name:   fmt.Sprintf("Series %d", i+1),
			Points: []DataPoint{},
		}

		for _, target := range bucketSeries.Targets {
			for _, dataPoint := range target.Data {
				seriesData.Points = append(seriesData.Points, DataPoint{
					Time:  dataPoint.Time,
					Value: dataPoint.Value,
				})
			}
		}

		series[i] = seriesData
	}

	return &TimeSeriesResult{
		Series: series,
	}
}

// formatSummary creates summary format
func (f *LLMFormatter) formatSummary(result *axiom.QueryResult, options FormatOptions) *SummaryResult {
	summary := &SummaryResult{
		Totals: make(map[string]interface{}),
		Counts: make(map[string]int),
	}

	for _, total := range result.Buckets.Totals {
		summary.Totals[total.Field] = total.Value
	}

	return summary
}

// generateTableSummary creates a human-readable summary of table data
func (f *LLMFormatter) generateTableSummary(result *axiom.QueryResult, options FormatOptions) string {
	totalRows := len(result.Matches)
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

		if len(result.Fields) > 0 {
			fieldNames := extractFieldNames(result.Fields)
			parts = append(parts, fmt.Sprintf("with %d fields: %s",
				len(fieldNames), strings.Join(fieldNames, ", ")))
		}
	}

	if len(result.Warnings) > 0 {
		parts = append(parts, fmt.Sprintf("with %d warnings", len(result.Warnings)))
	}

	return strings.Join(parts, " ")
}

// generateTimeSeriesSummary creates a summary for time series data
func (f *LLMFormatter) generateTimeSeriesSummary(result *axiom.QueryResult) string {
	seriesCount := len(result.Buckets.Series)
	totalPoints := 0

	for _, series := range result.Buckets.Series {
		for _, target := range series.Targets {
			totalPoints += len(target.Data)
		}
	}

	return fmt.Sprintf("Time series data with %d series and %d total data points",
		seriesCount, totalPoints)
}

// generateSummary creates a summary for aggregated data
func (f *LLMFormatter) generateSummary(result *axiom.QueryResult) string {
	totalCount := len(result.Buckets.Totals)
	if totalCount == 0 {
		return "Summary data with no aggregations"
	}

	fieldNames := make([]string, len(result.Buckets.Totals))
	for i, total := range result.Buckets.Totals {
		fieldNames[i] = total.Field
	}

	return fmt.Sprintf("Aggregated data for %d fields: %s",
		totalCount, strings.Join(fieldNames, ", "))
}

// Helper functions

func extractFieldNames(fields []axiom.Field) []string {
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
