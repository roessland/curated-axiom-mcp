package axiom

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
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

// QueryResult represents the response from Axiom
type QueryResult struct {
	Status   string   `json:"status"`
	Matches  [][]any  `json:"matches"`
	Buckets  Buckets  `json:"buckets"`
	Fields   []Field  `json:"fields"`
	Warnings []string `json:"warnings,omitempty"`
}

type Buckets struct {
	Series []BucketSeries `json:"series"`
	Totals []Total        `json:"totals"`
}

type BucketSeries struct {
	Interval string   `json:"interval"`
	Targets  []Target `json:"targets"`
}

type Target struct {
	Group string      `json:"group"`
	Data  []DataPoint `json:"data"`
}

type DataPoint struct {
	Time  string  `json:"time"`
	Value float64 `json:"value"`
}

type Total struct {
	Field string  `json:"field"`
	Value float64 `json:"value"`
}

type Field struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// NewClient creates a new Axiom client
func NewClient(cfg *config.AxiomConfig) *Client {
	// Convert app URL to API URL
	apiURL := convertAppURLToAPIURL(cfg.URL)

	return &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: apiURL,
	}
}

// convertAppURLToAPIURL converts the app URL format to API URL format
// e.g., "https://app.axiom.co" -> "https://api.axiom.co/v1"
//
//	"https://app.eu.axiom.co" -> "https://api.eu.axiom.co/v1"
func convertAppURLToAPIURL(appURL string) string {
	if appURL == "" {
		return "https://api.axiom.co/v1"
	}

	// Replace "app." with "api." and add "/v1" suffix
	apiURL := strings.Replace(appURL, "app.", "api.", 1)
	if !strings.HasSuffix(apiURL, "/v1") {
		apiURL = apiURL + "/v1"
	}

	return apiURL
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

	// Create the HTTP request - use the _apl endpoint
	url := fmt.Sprintf("%s/datasets/_apl?format=tabular", c.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.Token)
	if c.config.OrgID != "" {
		req.Header.Set("X-Axiom-Org-Id", c.config.OrgID)
	}

	// Comprehensive debug logging for the request
	log.Printf("[DEBUG] ==================== HTTP REQUEST ====================")
	log.Printf("[DEBUG] Method: %s", req.Method)
	log.Printf("[DEBUG] URL: %s", req.URL.String())
	log.Printf("[DEBUG] Host: %s", req.URL.Host)
	log.Printf("[DEBUG] Path: %s", req.URL.Path)
	log.Printf("[DEBUG] RawQuery: %s", req.URL.RawQuery)
	log.Printf("[DEBUG] Content-Length: %d", req.ContentLength)

	// Log all request headers (with token masking)
	log.Printf("[DEBUG] Request Headers:")
	for name, values := range req.Header {
		for _, value := range values {
			if name == "Authorization" {
				log.Printf("[DEBUG]   %s: Bearer ***masked***", name)
			} else {
				log.Printf("[DEBUG]   %s: %s", name, value)
			}
		}
	}

	log.Printf("[DEBUG] Request Body: %s", string(jsonData))
	if c.config.OrgID != "" {
		log.Printf("[DEBUG] Org ID: %s", c.config.OrgID)
	}
	log.Printf("[DEBUG] ====================================================")

	// Execute the request
	log.Printf("[DEBUG] Executing HTTP request...")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("[ERROR] HTTP request execution failed: %v", err)
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() { iferr.Log(resp.Body.Close()) }()

	log.Printf("[DEBUG] ==================== HTTP RESPONSE ===================")
	log.Printf("[DEBUG] Status Code: %d", resp.StatusCode)
	log.Printf("[DEBUG] Status: %s", resp.Status)
	log.Printf("[DEBUG] Proto: %s", resp.Proto)
	log.Printf("[DEBUG] Content-Length: %d", resp.ContentLength)
	log.Printf("[DEBUG] Transfer-Encoding: %v", resp.TransferEncoding)
	log.Printf("[DEBUG] Connection Close: %t", resp.Close)

	// Log all response headers
	log.Printf("[DEBUG] Response Headers:")
	for name, values := range resp.Header {
		for _, value := range values {
			log.Printf("[DEBUG]   %s: %s", name, value)
		}
	}

	// Read the response with detailed logging
	log.Printf("[DEBUG] Attempting to read response body...")
	log.Printf("[DEBUG] Body reader type: %T", resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[ERROR] Failed to read response body: %v", err)
		log.Printf("[ERROR] Error type: %T", err)
		log.Printf("[ERROR] Response Proto: %s", resp.Proto)
		log.Printf("[ERROR] Response Content-Length: %d", resp.ContentLength)
		log.Printf("[ERROR] Response Transfer-Encoding: %v", resp.TransferEncoding)
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	log.Printf("[DEBUG] Successfully read %d bytes from response body", len(body))
	if len(body) > 0 {
		if len(body) < 2000 {
			log.Printf("[DEBUG] Response Body: %s", string(body))
		} else {
			log.Printf("[DEBUG] Response Body (first 2000 chars): %s...", string(body[:2000]))
		}
	} else {
		log.Printf("[DEBUG] Response Body: <empty>")
	}
	log.Printf("[DEBUG] =======================================================")

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return nil, utils.NewAxiomError(resp.StatusCode, string(body))
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
	url := fmt.Sprintf("%s/user", c.baseURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create test request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.config.Token)
	if c.config.OrgID != "" {
		req.Header.Set("X-Axiom-Org-Id", c.config.OrgID)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute test request: %w", err)
	}
	defer func() { iferr.Log(resp.Body.Close()) }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return utils.NewAxiomError(resp.StatusCode, string(body))
	}

	return nil
}
