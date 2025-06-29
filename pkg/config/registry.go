package config

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/roessland/curated-axiom-mcp/pkg/axiom"
	"github.com/roessland/curated-axiom-mcp/pkg/caxiom"
)

// Registry holds the loaded queries and provides thread-safe access
type Registry struct {
	queries        *QueryRegistry
	dynamicQueries map[string]*DynamicQuery
	loadedAt       time.Time
	cacheTTL       time.Duration
	filePath       string
	axiomClient    *axiom.Client
	mu             sync.RWMutex
}

// NewRegistry creates a new query registry
func NewRegistry(filePath string, cacheTTL time.Duration) *Registry {
	return &Registry{
		filePath:       filePath,
		cacheTTL:       cacheTTL,
		dynamicQueries: make(map[string]*DynamicQuery),
	}
}

// NewRegistryWithAxiom creates a new registry with Axiom client for dynamic loading
func NewRegistryWithAxiom(axiomConfig *AxiomConfig, cacheTTL time.Duration) *Registry {
	// Convert config to avoid import cycle
	clientConfig := &axiom.AxiomConfig{
		Token:   axiomConfig.Token,
		OrgID:   axiomConfig.OrgID,
		Dataset: axiomConfig.Dataset,
		URL:     axiomConfig.URL,
	}
	
	return &Registry{
		cacheTTL:       cacheTTL,
		axiomClient:    axiom.NewClient(clientConfig),
		dynamicQueries: make(map[string]*DynamicQuery),
	}
}

// Load loads or reloads the queries from the file
func (r *Registry) Load() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	queries, err := LoadQueries(r.filePath)
	if err != nil {
		return err
	}

	r.queries = queries
	r.loadedAt = time.Now()
	return nil
}

// GetQuery retrieves a query by name, reloading if cache is expired
func (r *Registry) GetQuery(name string) (*Query, error) {
	r.mu.RLock()

	// Check if we need to reload
	if r.queries == nil || time.Since(r.loadedAt) > r.cacheTTL {
		r.mu.RUnlock()
		if err := r.Load(); err != nil {
			return nil, err
		}
		r.mu.RLock()
	}

	query, exists := r.queries.Queries[name]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("query %s not found", name)
	}

	return &query, nil
}

// ListQueries returns all available query names and descriptions
func (r *Registry) ListQueries() (map[string]string, error) {
	r.mu.RLock()

	// Check if we need to reload
	if r.queries == nil || time.Since(r.loadedAt) > r.cacheTTL {
		r.mu.RUnlock()
		if err := r.Load(); err != nil {
			return nil, err
		}
		r.mu.RLock()
	}

	result := make(map[string]string)
	for name, query := range r.queries.Queries {
		result[name] = query.Description
	}
	r.mu.RUnlock()

	return result, nil
}

// LoadFromAxiom loads queries from Axiom starred queries with 10-second timeout
func (r *Registry) LoadFromAxiom() error {
	if r.axiomClient == nil {
		return fmt.Errorf("no Axiom client configured")
	}

	// Create context with 10-second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Run the loading in a goroutine to respect timeout
	errChan := make(chan error, 1)
	go func() {
		errChan <- r.loadFromAxiomInternal()
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return fmt.Errorf("timeout loading queries from Axiom (exceeded 10 seconds)")
	}
}

// loadFromAxiomInternal performs the actual loading logic
func (r *Registry) loadFromAxiomInternal() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Fetch starred queries from Axiom
	starredQueries, err := r.axiomClient.StarredQueries()
	if err != nil {
		return fmt.Errorf("failed to fetch starred queries from Axiom: %w", err)
	}

	// Clear existing dynamic queries
	r.dynamicQueries = make(map[string]*DynamicQuery)

	// Parse each starred query
	for _, sq := range starredQueries {
		// Try to parse the query for MCP usage
		parsed, err := caxiom.ParseStarredQuery(sq.Name, sq.Query.APL)
		if err != nil {
			// Check if this query contains CuratedAxiomMCP marker but failed to parse
			if strings.Contains(sq.Query.APL, "CuratedAxiomMCP") {
				// This is a parsing error for a query that should be processed - log as error
				slog.Error("Failed to parse CuratedAxiomMCP query", "name", sq.Name, "error", err, "full_apl", sq.Query.APL)
			} else {
				// This query doesn't have the marker - log at debug level to reduce noise
				slog.Debug("Skipping starred query (no CuratedAxiomMCP marker)", "name", sq.Name)
			}
			continue
		}

		// Convert to DynamicQuery
		dynamicQuery := &DynamicQuery{
			Name:        sq.Name,
			OriginalAPL: parsed.OriginalAPL,
			TemplateAPL: parsed.TemplateAPL,
			ToolName:    parsed.Metadata.CuratedAxiomMCP.ToolName,
			Description: parsed.Metadata.CuratedAxiomMCP.Description,
			Constraints: parsed.Metadata.CuratedAxiomMCP.Constraints,
			Parameters:  make([]DynamicParameter, len(parsed.Metadata.CuratedAxiomMCP.Params)),
		}

		// Convert parameters
		for i, param := range parsed.Metadata.CuratedAxiomMCP.Params {
			dynamicQuery.Parameters[i] = DynamicParameter{
				Name:        param.Name,
				Type:        param.Type,
				Example:     param.Example,
				Description: param.Description,
				Required:    true, // For now, all template parameters are required
			}
		}

		// Use ToolName as key if specified, otherwise use query name
		key := dynamicQuery.ToolName
		if key == "" {
			key = sq.Name
		}

		r.dynamicQueries[key] = dynamicQuery
		slog.Info("Loaded dynamic query", "name", sq.Name, "tool_name", key)
	}

	r.loadedAt = time.Now()
	
	if len(r.dynamicQueries) == 0 {
		slog.Warn("No CuratedAxiomMCP queries found in starred queries", "total_starred_queries", len(starredQueries))
	} else {
		slog.Info("Successfully loaded dynamic queries from Axiom", "count", len(r.dynamicQueries))
	}
	
	return nil
}

// GetDynamicQuery retrieves a dynamic query by tool name
func (r *Registry) GetDynamicQuery(toolName string) (*DynamicQuery, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query, exists := r.dynamicQueries[toolName]
	if !exists {
		return nil, fmt.Errorf("dynamic query %s not found", toolName)
	}

	return query, nil
}

// ListDynamicQueries returns all available dynamic queries
func (r *Registry) ListDynamicQueries() map[string]*DynamicQuery {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy to avoid concurrent access issues
	result := make(map[string]*DynamicQuery)
	for name, query := range r.dynamicQueries {
		result[name] = query
	}
	return result
}
