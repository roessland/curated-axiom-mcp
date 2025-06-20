package cmd

import (
	"fmt"
	"os"

	"github.com/roessland/curated-axiom-mcp/pkg/config"
	"github.com/roessland/curated-axiom-mcp/pkg/server"
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

		// Setup logger based on configuration
		utils.SetupLogger(&appConfig.Logging)

		// Initialize query registry
		// Use queries file flag if provided, otherwise use config file setting
		actualQueriesFile := appConfig.Queries.File
		if queriesFile != "" {
			actualQueriesFile = queriesFile
		}
		registry = config.NewRegistry(actualQueriesFile, appConfig.Queries.CacheTTL)
		if err := registry.Load(); err != nil {
			return fmt.Errorf("failed to load queries: %w", err)
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		stdio, _ := cmd.Flags().GetBool("stdio")
		if stdio {
			return server.StartStdioServer(appConfig, registry)
		} else {
			return server.StartSSEServer(appConfig, registry)
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
