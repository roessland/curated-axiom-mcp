package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/roessland/curated-axiom-mcp/pkg/config"
	"github.com/roessland/curated-axiom-mcp/pkg/utils/iferr"
)

// StartSSEServer starts the MCP server in HTTP mode
func StartSSEServer(appConfig *config.AppConfig, registry *config.Registry) error {
	addr := fmt.Sprintf("%s:%d", appConfig.Server.Host, appConfig.Server.Port)

	// Get query count for logging
	queryList, _ := registry.ListQueries()
	fmt.Printf("🚀 Starting MCP server (HTTP mode) on %s\n", addr)
	fmt.Printf("📋 Available queries: %d\n", len(queryList))
	fmt.Printf("💡 Note: Full MCP functionality requires --stdio mode\n")

	// Add health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		iferr.Log2(w.Write([]byte(`{"status":"healthy","mode":"http","note":"Use --stdio for full MCP functionality"}`)))
	})

	// Add info endpoint
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		queryListHTML := ""
		for name, desc := range queryList {
			queryListHTML += fmt.Sprintf("<li><strong>%s</strong>: %s</li>", name, desc)
		}

		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, err := fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>curated-axiom-mcp Server</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        h1 { color: #333; }
        .code { background: #f4f4f4; padding: 10px; border-radius: 4px; font-family: monospace; }
        .note { background: #fff3cd; padding: 10px; border-radius: 4px; margin: 10px 0; border-left: 4px solid #ffc107; }
    </style>
</head>
<body>
    <h1>🚀 curated-axiom-mcp Server</h1>
    <div class="note">
        <strong>Note:</strong> This server is running in HTTP mode. For full MCP functionality, use stdio mode.
    </div>
    
    <h2>Available queries (%d):</h2>
    <ul>%s</ul>
    
    <h2>For full MCP functionality:</h2>
    <div class="code">
        curated-axiom-mcp --stdio
    </div>
    
    <h2>Available endpoints:</h2>
    <ul>
        <li><a href="/health">/health</a> - Health check</li>
    </ul>
</body>
</html>
		`, len(queryList), queryListHTML)
		iferr.Log(err)
	})

	// Start HTTP server
	log.Printf("HTTP server listening on %s", addr)
	return http.ListenAndServe(addr, nil)
}
