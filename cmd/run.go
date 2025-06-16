package cmd

import (
	"fmt"

	"github.com/roessland/curated-axiom-mcp/pkg/axiom"
	"github.com/roessland/curated-axiom-mcp/pkg/formatter"
	"github.com/roessland/curated-axiom-mcp/pkg/utils"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run <query-name> [param1=value1] [param2=value2] ...",
	Short: "Execute a query directly from the command line",
	Long: `Execute a named query with parameters directly from the command line.
This is useful for testing queries before using them via the MCP server.

Example:
  curated-axiom-mcp run user-activity user_id=12345 start_time="2024-01-01"`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		queryName := args[0]
		paramArgs := args[1:]

		// Parse parameters
		params, err := utils.ParseParams(paramArgs)
		if err != nil {
			return fmt.Errorf("failed to parse parameters: %w", err)
		}

		// Get the query
		query, err := registry.GetQuery(queryName)
		if err != nil {
			return err
		}

		// Validate and convert parameters
		convertedParams, err := utils.ValidateAndConvertParams(params, query)
		if err != nil {
			return err
		}

		// Create Axiom client
		client := axiom.NewClient(&appConfig.Axiom)

		// Execute the query
		result, err := client.ExecuteQueryByName(queryName, convertedParams, registry)
		if err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}

		// Format and display results
		cliFormatter := formatter.NewCLIFormatter()
		output, err := cliFormatter.FormatForCLI(result, formatter.DefaultFormatOptions())
		if err != nil {
			return fmt.Errorf("failed to format results: %w", err)
		}

		fmt.Print(output)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
