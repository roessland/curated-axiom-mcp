package formatter

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/roessland/curated-axiom-mcp/pkg/axiom"
)

// CLIFormatter formats results for command-line display
type CLIFormatter struct{}

// NewCLIFormatter creates a new CLI formatter
func NewCLIFormatter() *CLIFormatter {
	return &CLIFormatter{}
}

// FormatForCLI formats the result for CLI output
func (f *CLIFormatter) FormatForCLI(result *axiom.QueryResult, options FormatOptions) (string, error) {
	formatted, err := f.Format(result, options)
	if err != nil {
		return "", err
	}

	var output strings.Builder

	// Add summary
	output.WriteString(fmt.Sprintf("📊 %s\n\n", formatted.Summary))

	// Add warnings if any
	if len(formatted.Warnings) > 0 {
		output.WriteString("⚠️  Warnings:\n")
		for _, warning := range formatted.Warnings {
			output.WriteString(fmt.Sprintf("   • %s\n", warning))
		}
		output.WriteString("\n")
	}

	// Format data based on type
	switch data := formatted.Data.(type) {
	case *TableResult:
		output.WriteString(f.formatTableForCLI(data))
	case *TimeSeriesResult:
		output.WriteString(f.formatTimeSeriesForCLI(data))
	case *SummaryResult:
		output.WriteString(f.formatSummaryForCLI(data))
	default:
		// Fallback to JSON
		jsonBytes, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to format data: %w", err)
		}
		output.WriteString(string(jsonBytes))
	}

	return output.String(), nil
}

// Format implements the Formatter interface
func (f *CLIFormatter) Format(result *axiom.QueryResult, options FormatOptions) (*FormattedResult, error) {
	// Use LLM formatter as base and then format for CLI
	llmFormatter := NewLLMFormatter()
	return llmFormatter.Format(result, options)
}

// formatTableForCLI creates a nice table output for the CLI
func (f *CLIFormatter) formatTableForCLI(table *TableResult) string {
	if len(table.Rows) == 0 {
		return "No data to display.\n"
	}

	var output strings.Builder

	// Calculate column widths
	colWidths := make([]int, len(table.Headers))

	// Start with header widths
	for i, header := range table.Headers {
		colWidths[i] = len(header)
	}

	// Check row widths
	for _, row := range table.Rows {
		for i, cell := range row {
			if i < len(colWidths) && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	// Add some padding
	for i := range colWidths {
		colWidths[i] += 2
	}

	// Create header separator
	separator := strings.Builder{}
	for i, width := range colWidths {
		if i > 0 {
			separator.WriteString("+")
		}
		separator.WriteString(strings.Repeat("-", width))
	}
	separatorLine := separator.String()

	// Print header
	output.WriteString(separatorLine + "\n")
	for i, header := range table.Headers {
		if i > 0 {
			output.WriteString("|")
		}
		output.WriteString(fmt.Sprintf(" %-*s", colWidths[i]-1, header))
	}
	output.WriteString("\n" + separatorLine + "\n")

	// Print rows
	for _, row := range table.Rows {
		for i, cell := range row {
			if i > 0 {
				output.WriteString("|")
			}
			output.WriteString(fmt.Sprintf(" %-*s", colWidths[i]-1, cell))
		}
		output.WriteString("\n")
	}
	output.WriteString(separatorLine + "\n")

	// Add row count info
	if table.Total > len(table.Rows) {
		output.WriteString(fmt.Sprintf("\nShowing %d of %d total rows\n", len(table.Rows), table.Total))
	}

	return output.String()
}

// formatTimeSeriesForCLI formats time series data for CLI
func (f *CLIFormatter) formatTimeSeriesForCLI(ts *TimeSeriesResult) string {
	var output strings.Builder

	for i, series := range ts.Series {
		output.WriteString(fmt.Sprintf("📈 %s\n", series.Name))

		if len(series.Points) == 0 {
			output.WriteString("   No data points\n\n")
			continue
		}

		// Show first few points
		maxPoints := 10
		pointsToShow := len(series.Points)
		if pointsToShow > maxPoints {
			pointsToShow = maxPoints
		}

		for j := 0; j < pointsToShow; j++ {
			point := series.Points[j]
			output.WriteString(fmt.Sprintf("   %s: %v\n", point.Time, point.Value))
		}

		if len(series.Points) > maxPoints {
			output.WriteString(fmt.Sprintf("   ... and %d more points\n", len(series.Points)-maxPoints))
		}

		if i < len(ts.Series)-1 {
			output.WriteString("\n")
		}
	}

	return output.String()
}

// formatSummaryForCLI formats summary data for CLI
func (f *CLIFormatter) formatSummaryForCLI(summary *SummaryResult) string {
	var output strings.Builder

	if len(summary.Totals) > 0 {
		output.WriteString("📊 Totals:\n")
		for field, value := range summary.Totals {
			output.WriteString(fmt.Sprintf("   %-20s: %v\n", field, value))
		}
	}

	if len(summary.Counts) > 0 {
		if len(summary.Totals) > 0 {
			output.WriteString("\n")
		}
		output.WriteString("🔢 Counts:\n")
		for field, count := range summary.Counts {
			output.WriteString(fmt.Sprintf("   %-20s: %d\n", field, count))
		}
	}

	if len(summary.Averages) > 0 {
		if len(summary.Totals) > 0 || len(summary.Counts) > 0 {
			output.WriteString("\n")
		}
		output.WriteString("📊 Averages:\n")
		for field, avg := range summary.Averages {
			output.WriteString(fmt.Sprintf("   %-20s: %.2f\n", field, avg))
		}
	}

	return output.String()
}
