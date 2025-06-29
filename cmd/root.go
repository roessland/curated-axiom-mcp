package cmd

import (
	"fmt"
	"os"

	"log/slog"

	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/roessland/curated-axiom-mcp/pkg/config"
	"github.com/roessland/curated-axiom-mcp/pkg/cserver"
	"github.com/roessland/curated-axiom-mcp/pkg/utils"
	"github.com/spf13/cobra"
)

var (
	cfgFile     string
	portFlag    int
	queriesFile string
	appConfig   *config.AppConfig
	registry    *config.Registry
)

var rootCmd = &cobra.Command{
	Use:   "curated-axiom-mcp",
	Short: "MCP server for curated Axiom queries",
	Long: `An MCP server that provides LLM-friendly access to 
whitelisted Axiom queries with simplified results.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip config validation for config init command
		if cmd.Name() == "init" && cmd.Parent() != nil && cmd.Parent().Name() == "config" {
			return nil
		}

		var err error
		appConfig, err = config.LoadConfig(cfgFile, portFlag)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Check if stdio mode is enabled
		stdio, _ := cmd.Flags().GetBool("stdio")

		// Setup logger based on configuration
		utils.SetupLogger(&appConfig.Logging, stdio)

		// Initialize query registry with Axiom client for dynamic loading
		registry = config.NewRegistryWithAxiom(&appConfig.Axiom, appConfig.Queries.CacheTTL)

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		slog.Info("Initializing MCP server...")
		mcpManager := cserver.NewMCP(appConfig, registry)
		slog.Info("MCP server initialized")
		
		// Load dynamic tools from Axiom
		slog.Info("Loading dynamic tools from Axiom...")
		if err := mcpManager.LoadDynamicTools(); err != nil {
			slog.Error("Failed to load dynamic tools", "error", err)
			return fmt.Errorf("failed to load dynamic tools: %w", err)
		}
		slog.Info("Dynamic tools loaded successfully")

		stdio, _ := cmd.Flags().GetBool("stdio")
		if stdio {
			slog.Info("Starting stdio MCP server...")
			return mcpserver.ServeStdio(mcpManager.GetServer())
		} else {
			slog.Info("Starting HTTP MCP server...")
			server := mcpserver.NewStreamableHTTPServer(mcpManager.GetServer())
			slog.Info("Starting server", "url", fmt.Sprintf("http://localhost:%d/mcp", appConfig.Server.Port))
			return server.Start(fmt.Sprintf(":%d", appConfig.Server.Port))
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"config file (default: ~/.config/curated-axiom-mcp/config.yaml)")
	rootCmd.PersistentFlags().IntVar(&portFlag, "port", 0,
		"server port (overrides config file, but not PORT env var)")
	rootCmd.PersistentFlags().StringVar(&queriesFile, "queries", "",
		"queries file (default: ~/.config/curated-axiom-mcp/queries.yaml)")

	rootCmd.Flags().Bool("stdio", false, "run as stdio MCP server")
}
