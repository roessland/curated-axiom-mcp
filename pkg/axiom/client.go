package axiom

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/axiom-go/axiom/query"
)

// Client wraps the Axiom client for our specific use cases
type Client struct {
	client *axiom.Client
	config *AxiomConfig
}

// AxiomConfig represents Axiom configuration (avoiding import cycle)
type AxiomConfig struct {
	Token   string
	OrgID   string
	Dataset string
	URL     string
}

// NewClient creates a new Axiom client from config
func NewClient(config *AxiomConfig) *Client {
	opts := []axiom.Option{
		axiom.SetToken(config.Token),
	}
	if config.URL != "" {
		opts = append(opts, axiom.SetURL(config.URL))
	}
	if config.OrgID != "" {
		opts = append(opts, axiom.SetOrganizationID(config.OrgID))
	}

	client, err := axiom.NewClient(opts...)
	if err != nil {
		// This shouldn't happen with valid config, but just in case
		panic(fmt.Sprintf("failed to create axiom client: %v", err))
	}

	return &Client{
		client: client,
		config: config,
	}
}

// QueryResult is an alias for the axiom query result
type QueryResult = query.Result

// StarredQuery represents a starred query from Axiom
type StarredQuery struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Dataset  string                 `json:"dataset"`
	Query    StarredQueryContent    `json:"query"`
	Kind     string                 `json:"kind"`
	Metadata map[string]interface{} `json:"metadata"`
	Who      string                 `json:"who"`
}

// StarredQueryContent represents the content of a starred query
type StarredQueryContent struct {
	APL string `json:"apl"`
}

// ExecuteQuery executes an APL query and returns the result
func (c *Client) ExecuteQuery(apl string) (*QueryResult, error) {
	ctx := context.Background()

	result, err := c.client.Query(ctx, apl)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}

	return result, nil
}

// StarredQueries fetches all starred queries from Axiom
func (c *Client) StarredQueries() ([]StarredQuery, error) {
	ctx := context.Background()

	// Construct the full URL with query parameters
	baseURL := c.config.URL
	if baseURL == "" {
		baseURL = "https://api.axiom.co"
	}

	fullURL, err := url.JoinPath(baseURL, "/v2/apl-starred-queries")
	if err != nil {
		return nil, fmt.Errorf("failed to construct URL: %w", err)
	}

	// Add query parameters
	parsedURL, err := url.Parse(fullURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	query := parsedURL.Query()
	query.Set("who", "all") // Get all starred queries
	parsedURL.RawQuery = query.Encode()
	fullURL = parsedURL.String()

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add required headers
	req.Header.Set("Authorization", "Bearer "+c.config.Token)
	req.Header.Set("Content-Type", "application/json")

	// Add org ID header if provided
	if c.config.OrgID != "" {
		req.Header.Set("X-AXIOM-ORG-ID", c.config.OrgID)
	}

	// Use the axiom client's Do method which handles retries and error parsing
	var result []StarredQuery
	_, err = c.client.Do(req, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch starred queries: %w", err)
	}

	return result, nil
}
