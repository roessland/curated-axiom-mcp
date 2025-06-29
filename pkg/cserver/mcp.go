package cserver

import (
	"log/slog"
	"sync"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/roessland/curated-axiom-mcp/pkg/config"
)

type MCPManager struct {
	server      *server.MCPServer
	appConfig   *config.AppConfig
	registry    *config.Registry
	toolsLoaded bool
	mu          sync.RWMutex
}

func NewMCP(appConfig *config.AppConfig, registry *config.Registry) *MCPManager {
	s := server.NewMCPServer("curated-axiom-mcp", "1.0.0")

	// Add static tools
	s.AddTool(runQueryTool, RunQueryHandler(registry, appConfig))
	s.AddTool(starredQueriesTool, DebugStarredQueriesHandler(appConfig))

	manager := &MCPManager{
		server:    s,
		appConfig: appConfig,
		registry:  registry,
	}

	return manager
}

// LoadDynamicTools loads dynamic tools from the registry
func (m *MCPManager) LoadDynamicTools() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.toolsLoaded {
		return nil // Already loaded
	}

	// Load queries from Axiom
	if err := m.registry.LoadFromAxiom(); err != nil {
		return err
	}

	// Register each dynamic query as an MCP tool
	dynamicQueries := m.registry.ListDynamicQueries()
	for toolName, query := range dynamicQueries {
		tool := createDynamicTool(toolName, query)
		handler := CreateDynamicQueryHandler(toolName, m.registry, m.appConfig)
		m.server.AddTool(tool, handler)
		slog.Info("Registered dynamic tool", "name", toolName)
	}

	m.toolsLoaded = true
	
	if len(dynamicQueries) == 0 {
		slog.Warn("No dynamic tools registered - no CuratedAxiomMCP queries found")
	} else {
		slog.Info("Successfully loaded dynamic tools", "count", len(dynamicQueries))
	}
	
	return nil
}

// GetServer returns the underlying MCP server
func (m *MCPManager) GetServer() *server.MCPServer {
	return m.server
}

// AreToolsLoaded returns whether dynamic tools have been loaded
func (m *MCPManager) AreToolsLoaded() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.toolsLoaded
}

// createDynamicTool creates an MCP tool definition from a dynamic query
func createDynamicTool(toolName string, query *config.DynamicQuery) mcp.Tool {
	// Build options array for tool creation
	opts := []mcp.ToolOption{
		mcp.WithDescription(query.Description),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithReadOnlyHintAnnotation(true),
	}

	// Add parameters from the dynamic query
	for _, param := range query.Parameters {
		if param.Required {
			opts = append(opts, mcp.WithString(param.Name, mcp.Required()))
		} else {
			opts = append(opts, mcp.WithString(param.Name))
		}
	}

	return mcp.NewTool(toolName, opts...)
}
