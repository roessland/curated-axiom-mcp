package cmd

import (
	"fmt"
	"os"

	"github.com/roessland/curated-axiom-mcp/pkg/axiom"
	"github.com/spf13/cobra"
)

var starredQueriesCmd = &cobra.Command{
	Use:   "starred-queries",
	Short: "List all starred queries from Axiom",
	Long:  "Fetch and display all starred queries for the authenticated user from Axiom.",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := axiom.NewClient(&appConfig.Axiom)
		queries, err := client.StarredQueries()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to fetch starred queries: %v\n", err)
			return err
		}
		if len(queries) == 0 {
			fmt.Println("No starred queries found.")
			return nil
		}
		fmt.Printf("Starred Queries (%d):\n\n", len(queries))
		for _, q := range queries {
			fmt.Printf("- %s: %s\n", q.Name, q.Description)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(starredQueriesCmd)
}
