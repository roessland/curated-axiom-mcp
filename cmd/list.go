package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available queries",
	Long:  "Display all available queries with their descriptions.",
	RunE: func(cmd *cobra.Command, args []string) error {
		queries, err := registry.ListQueries()
		if err != nil {
			return fmt.Errorf("failed to list queries: %w", err)
		}

		if len(queries) == 0 {
			fmt.Println("No queries available.")
			return nil
		}

		// Sort queries by name
		names := make([]string, 0, len(queries))
		for name := range queries {
			names = append(names, name)
		}
		sort.Strings(names)

		// Display queries
		fmt.Printf("📋 Available Queries (%d total):\n\n", len(queries))

		// Calculate max name length for alignment
		maxNameLen := 0
		for _, name := range names {
			if len(name) > maxNameLen {
				maxNameLen = len(name)
			}
		}

		for _, name := range names {
			description := queries[name]
			// Truncate long descriptions
			if len(description) > 80 {
				description = description[:77] + "..."
			}
			fmt.Printf("  %-*s  %s\n", maxNameLen, name, description)
		}

		fmt.Printf("\nUse 'curated-axiom-mcp describe <query-name>' for detailed information.\n")
		return nil
	},
}

var describeCmd = &cobra.Command{
	Use:   "describe <query-name>",
	Short: "Show detailed information about a query",
	Long:  "Display detailed information about a specific query including parameters.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		queryName := args[0]

		query, err := registry.GetQuery(queryName)
		if err != nil {
			return err
		}

		fmt.Printf("📝 Query: %s\n\n", query.Name)
		fmt.Printf("Description: %s\n\n", query.Description)

		if len(query.Tags) > 0 {
			fmt.Printf("Tags: %s\n\n", strings.Join(query.Tags, ", "))
		}

		if len(query.Parameters) > 0 {
			fmt.Println("Parameters:")
			for _, param := range query.Parameters {
				var flags []string
				if param.Required {
					flags = append(flags, "required")
				}
				if param.Default != nil {
					flags = append(flags, fmt.Sprintf("default: %v", param.Default))
				}
				if len(param.Enum) > 0 {
					enumValues := make([]string, len(param.Enum))
					for i, v := range param.Enum {
						enumValues[i] = fmt.Sprintf("%v", v)
					}
					flags = append(flags, fmt.Sprintf("enum: [%s]", strings.Join(enumValues, ", ")))
				}

				flagStr := ""
				if len(flags) > 0 {
					flagStr = fmt.Sprintf(" (%s)", strings.Join(flags, ", "))
				}

				fmt.Printf("  • %s (%s)%s\n", param.Name, param.Type, flagStr)
				if param.Description != "" {
					fmt.Printf("    %s\n", param.Description)
				}
			}
			fmt.Println()
		}

		fmt.Printf("Output Format: %s\n", query.OutputFormat)
		fmt.Printf("LLM Friendly: %t\n\n", query.LLMFriendly)

		fmt.Println("APL Query:")
		fmt.Printf("```\n%s\n```\n\n", query.APLQuery)

		fmt.Printf("Example usage:\n")
		fmt.Printf("  curated-axiom-mcp run %s", queryName)
		for _, param := range query.Parameters {
			if param.Required {
				fmt.Printf(" %s=<value>", param.Name)
			}
		}
		fmt.Println()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(describeCmd)
}
