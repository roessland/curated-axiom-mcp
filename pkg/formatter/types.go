package formatter

import "github.com/roessland/curated-axiom-mcp/pkg/axiom"

// FormattedResult represents a formatted query result
type FormattedResult struct {
	Summary  string                 `json:"summary"`
	Data     interface{}            `json:"data"`
	Count    int                    `json:"count"`
	Warnings []string               `json:"warnings,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// TableResult represents data in table format
type TableResult struct {
	Headers []string   `json:"headers"`
	Rows    [][]string `json:"rows"`
	Total   int        `json:"total"`
}

// TimeSeriesResult represents time series data
type TimeSeriesResult struct {
	Series []Series `json:"series"`
}

type Series struct {
	Name   string      `json:"name"`
	Points []DataPoint `json:"points"`
}

type DataPoint struct {
	Time  string      `json:"time"`
	Value interface{} `json:"value"`
}

// SummaryResult represents aggregated/summary data
type SummaryResult struct {
	Totals   map[string]interface{} `json:"totals"`
	Counts   map[string]int         `json:"counts"`
	Averages map[string]float64     `json:"averages,omitempty"`
}

// FormatOptions controls how results are formatted
type FormatOptions struct {
	Format      string // "table", "json", "summary", "timeseries"
	LLMFriendly bool   // Whether to optimize for LLM consumption
	MaxRows     int    // Maximum number of rows to include
	IncludeRaw  bool   // Whether to include raw data
}

// DefaultFormatOptions returns sensible defaults
func DefaultFormatOptions() FormatOptions {
	return FormatOptions{
		Format:      "table",
		LLMFriendly: true,
		MaxRows:     100,
		IncludeRaw:  false,
	}
}

// Formatter interface for different output formats
type Formatter interface {
	Format(result *axiom.QueryResult, options FormatOptions) (*FormattedResult, error)
}
