package server

import (
	"fmt"
	"net/http"

	"github.com/roessland/curated-axiom-mcp/pkg/config"
	"github.com/roessland/curated-axiom-mcp/pkg/utils/iferr"
)

// StartSSEServer starts the MCP server in SSE mode
func StartSSEServer(appConfig *config.AppConfig, registry *config.Registry) error {
	// Setup HTTP server for SSE
	addr := fmt.Sprintf("%s:%d", appConfig.Server.Host, appConfig.Server.Port)

	fmt.Printf("🚀 Starting MCP server (HTTP mode) on %s\n", addr)

	// Get query count
	queries, _ := registry.ListQueries()
	fmt.Printf("📋 Available queries: %d\n", len(queries))
	fmt.Printf("💡 For now, use --stdio mode. HTTP/SSE support coming soon!\n")

	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		iferr.Log(w.Write([]byte(`{"status":"healthy","note":"Use --stdio mode for MCP functionality"}`)))
	})

	// Info endpoint
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		iferr.Log(w.Write([]byte(`
<!DOCTYPE html>
<html>
<head>
    <title>curated-axiom-mcp Server</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        h1 { color: #333; }
        .code { background: #f4f4f4; padding: 10px; border-radius: 4px; }
    </style>
</head>
<body>
    <h1>🚀 curated-axiom-mcp Server</h1>
    <p>This server is running in HTTP mode, but MCP functionality requires stdio mode.</p>
    
    <h2>To use MCP features:</h2>
    <div class="code">
        curated-axiom-mcp --stdio
    </div>
    
    <h2>Available endpoints:</h2>
    <ul>
        <li><a href="/health">/health</a> - Health check</li>
    </ul>
    
    <p>For more information, see the README.</p>
</body>
</html>
		`)))
	})

	// Start HTTP server
	return http.ListenAndServe(addr, nil)
}
