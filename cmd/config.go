package cmd

import (
	"fmt"

	"github.com/roessland/curated-axiom-mcp/pkg/axiom"
	"github.com/roessland/curated-axiom-mcp/pkg/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management",
	Long:  "Manage curated-axiom-mcp configuration",
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create example config file",
	Long: `Create an example configuration file at ~/.config/curated-axiom-mcp/config.yaml.
This file can then be edited to add your AXIOM_TOKEN and other settings.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Don't use the global config validation for init
		return config.CreateExampleConfig()
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long:  "Display the current configuration (with sensitive values hidden).",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Hide sensitive values
		displayConfig := *appConfig
		displayConfig.Axiom.Token = "***HIDDEN***"

		data, err := yaml.Marshal(displayConfig)
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}

		fmt.Printf("Current configuration:\n%s", data)
		return nil
	},
}

var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration",
	Long:  "Validate the current configuration and test connection to Axiom.",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("✓ Configuration is valid")
		fmt.Printf("✓ Axiom token: %s\n", maskToken(appConfig.Axiom.Token))

		// Test connection to Axiom
		fmt.Print("Testing connection to Axiom... ")
		client := axiom.NewClient(&appConfig.Axiom)
		if err := client.TestConnection(); err != nil {
			fmt.Printf("❌ Failed: %v\n", err)
			return fmt.Errorf("axiom connection test failed: %w", err)
		}
		fmt.Println("✓ Success")

		return nil
	},
}

func maskToken(token string) string {
	if len(token) <= 8 {
		return "***"
	}
	return token[:4] + "***" + token[len(token)-4:]
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configValidateCmd)
}
