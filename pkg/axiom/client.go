package axiom

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/roessland/curated-axiom-mcp/pkg/config"
	"github.com/roessland/curated-axiom-mcp/pkg/utils"
	"github.com/roessland/curated-axiom-mcp/pkg/utils/iferr"
)

// Client wraps the Axiom HTTP API
type Client struct {
	config     *config.AxiomConfig
	httpClient *http.Client
	baseURL    string
}

// QueryRequest represents an Axiom query request
type QueryRequest struct {
	APL       string `json:"apl"`
	StartTime string `json:"startTime,omitempty"`
	EndTime   string `json:"endTime,omitempty"`
}

// QueryResult represents the response from Axiom's tabular format
type QueryResult struct {
	Format        string                 `json:"format"`
	Status        QueryStatus            `json:"status"`
	Tables        []Table                `json:"tables"`
	DatasetNames  []string               `json:"datasetNames"`
	FieldsMetaMap map[string][]FieldMeta `json:"fieldsMetaMap"`
}

type QueryStatus struct {
	ElapsedTime    int64  `json:"elapsedTime"`
	MinCursor      string `json:"minCursor"`
	MaxCursor      string `json:"maxCursor"`
	BlocksExamined int64  `json:"blocksExamined"`
	BlocksCached   int64  `json:"blocksCached"`
	BlocksMatched  int64  `json:"blocksMatched"`
	BlocksSkipped  int64  `json:"blocksSkipped"`
	RowsExamined   int64  `json:"rowsExamined"`
	RowsMatched    int64  `json:"rowsMatched"`
	NumGroups      int64  `json:"numGroups"`
	IsPartial      bool   `json:"isPartial"`
	CacheStatus    int    `json:"cacheStatus"`
	MinBlockTime   string `json:"minBlockTime"`
	MaxBlockTime   string `json:"maxBlockTime"`
}

type Table struct {
	Name    string    `json:"name"`
	Sources []Source  `json:"sources"`
	Fields  []Field   `json:"fields"`
	Order   []OrderBy `json:"order"`
	Groups  []any     `json:"groups"`
	Range   TimeRange `json:"range"`
	Columns [][]any   `json:"columns"`
}

type Source struct {
	Name string `json:"name"`
}

type Field struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type OrderBy struct {
	Field string `json:"field"`
	Desc  bool   `json:"desc"`
}

type TimeRange struct {
	Field string `json:"field"`
	Start string `json:"start"`
	End   string `json:"end"`
}

type FieldMeta struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Unit        string `json:"unit"`
	Hidden      bool   `json:"hidden"`
	Description string `json:"description"`
}

// StarredQuery represents a starred query object from the Axiom API
// You may want to expand this struct based on the actual API response fields
// See: https://axiom.co/docs/restapi/query (Saved queries endpoints)
type StarredQuery struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	APL         string `json:"apl"`
	// Add more fields as needed based on the API response
}

// NewClient creates a new Axiom client
func NewClient(cfg *config.AxiomConfig) *Client {
	baseURL := cfg.URL
	if baseURL == "" {
		baseURL = "https://api.axiom.co"
	}

	return &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: baseURL,
	}
}

// doRequest performs an HTTP request with comprehensive logging and error handling
func (c *Client) doRequest(method, url string, body []byte, description string) ([]byte, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set common headers
	req.Header.Set("Authorization", "Bearer "+c.config.Token)
	if c.config.OrgID != "" {
		req.Header.Set("X-Axiom-Org-Id", c.config.OrgID)
	}
	if method == "POST" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Debug logging for the request
	slog.Debug(fmt.Sprintf("HTTP request for %s", description),
		"method", req.Method,
		"url", req.URL.String(),
		"org_id", c.config.OrgID,
		"body_size", len(body),
	)

	if len(body) > 0 {
		slog.Debug(fmt.Sprintf("%s request body", description), "body", string(body))
	}

	// Log headers (with token masking)
	if slog.Default().Enabled(context.TODO(), slog.LevelDebug) {
		headers := make(map[string]string)
		for name, values := range req.Header {
			for _, value := range values {
				if name == "Authorization" {
					headers[name] = "Bearer ***masked***"
				} else {
					headers[name] = value
				}
			}
		}
		slog.Debug(fmt.Sprintf("%s request headers", description), "headers", headers)
	}

	// Execute the request
	slog.Debug(fmt.Sprintf("Executing %s request", description))
	resp, err := c.httpClient.Do(req)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to execute %s request", description), "error", err)
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() { iferr.Log(resp.Body.Close()) }()

	// Debug logging for the response
	slog.Debug(fmt.Sprintf("%s response details", description),
		"status_code", resp.StatusCode,
		"status", resp.Status,
		"content_length", resp.ContentLength,
	)

	// Log response headers
	if slog.Default().Enabled(context.TODO(), slog.LevelDebug) {
		headers := make(map[string]string)
		for name, values := range resp.Header {
			for _, value := range values {
				headers[name] = value
			}
		}
		slog.Debug(fmt.Sprintf("%s response headers", description), "headers", headers)
	}

	// Read the response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to read %s response body", description),
			"error", err,
			"status_code", resp.StatusCode,
		)
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	slog.Debug(fmt.Sprintf("Read %s response body", description), "body_size", len(responseBody))
	if slog.Default().Enabled(context.TODO(), slog.LevelDebug) && len(responseBody) > 0 {
		if len(responseBody) < 2000 {
			slog.Debug(fmt.Sprintf("%s response body", description), "body", string(responseBody))
		} else {
			slog.Debug(fmt.Sprintf("%s response body (truncated)", description), "body", string(responseBody[:2000])+"...")
		}
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		slog.Error(fmt.Sprintf("%s request failed", description),
			"status_code", resp.StatusCode,
			"response_body", string(responseBody))
		return nil, utils.NewAxiomError(resp.StatusCode, string(responseBody))
	}

	slog.Debug(fmt.Sprintf("Successfully completed %s request", description))
	return responseBody, nil
}

// ExecuteQuery executes an APL query against Axiom
func (c *Client) ExecuteQuery(apl string, dataset string) (*QueryResult, error) {
	if dataset == "" {
		dataset = c.config.Dataset
	}

	if dataset == "" {
		return nil, fmt.Errorf("no dataset specified")
	}

	// Prepare the query request
	queryReq := QueryRequest{
		APL: apl,
	}

	jsonData, err := json.Marshal(queryReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query request: %w", err)
	}

	// Execute the request
	url := fmt.Sprintf("%s/v1/datasets/_apl?format=tabular", c.baseURL)
	body, err := c.doRequest("POST", url, jsonData, "query execution")
	if err != nil {
		return nil, err
	}

	// Parse the response
	var result QueryResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// TestConnection tests the connection to Axiom
func (c *Client) TestConnection() error {
	url := fmt.Sprintf("%s/v1/user", c.baseURL)
	_, err := c.doRequest("GET", url, nil, "connection test")
	return err
}

// StarredQueries fetches all starred queries for the authenticated user
func (c *Client) StarredQueries() ([]StarredQuery, error) {
	url := fmt.Sprintf("%s/v2/apl-starred-queries?who=all", c.baseURL)
	body, err := c.doRequest("GET", url, nil, "starred queries")
	if err != nil {
		return nil, err
	}

	var queries []StarredQuery
	if err := json.Unmarshal(body, &queries); err != nil {
		slog.Error("Failed to parse starred queries response", "error", err)
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	slog.Debug("Successfully fetched starred queries", "count", len(queries))
	return queries, nil
}
