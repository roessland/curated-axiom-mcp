package config

import (
	"fmt"
	"sync"
	"time"
)

// Registry holds the loaded queries and provides thread-safe access
type Registry struct {
	queries  *QueryRegistry
	loadedAt time.Time
	cacheTTL time.Duration
	filePath string
	mu       sync.RWMutex
}

// NewRegistry creates a new query registry
func NewRegistry(filePath string, cacheTTL time.Duration) *Registry {
	return &Registry{
		filePath: filePath,
		cacheTTL: cacheTTL,
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

// GetAllQueries returns all queries (used for MCP tool generation)
func (r *Registry) GetAllQueries() (map[string]Query, error) {
	r.mu.RLock()

	// Check if we need to reload
	if r.queries == nil || time.Since(r.loadedAt) > r.cacheTTL {
		r.mu.RUnlock()
		if err := r.Load(); err != nil {
			return nil, err
		}
		r.mu.RLock()
	}

	// Return a copy to avoid concurrent access issues
	result := make(map[string]Query)
	for name, query := range r.queries.Queries {
		result[name] = query
	}
	r.mu.RUnlock()

	return result, nil
}
