package config

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/roessland/curated-axiom-mcp/pkg/utils/iferr"
	"github.com/spf13/viper"
)

//go:embed embedded_config.yaml
var embeddedConfigTemplate string

//go:embed embedded_queries.yaml
var embeddedQueriesTemplate string

const (
	configDirName  = "curated-axiom-mcp"
	configFileName = "config"
)

// LoadConfig loads configuration with precedence: CLI flags > Env vars > Config file > Defaults
func LoadConfig(configFile string, portFlag int) (*AppConfig, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Setup config file locations
	if configFile != "" {
		v.SetConfigFile(configFile)
	} else {
		// Look in standard locations
		v.SetConfigName(configFileName)
		v.SetConfigType("yaml")

		// Add search paths
		if configDir := getConfigDir(); configDir != "" {
			v.AddConfigPath(configDir)
		}
		v.AddConfigPath(".")
		v.AddConfigPath("./configs")
	}

	// Setup environment variable handling
	v.SetEnvPrefix("AXIOM_MCP")
	v.AutomaticEnv()

	// Specific env var mappings
	iferr.Panic(v.BindEnv("axiom.token", "AXIOM_TOKEN"))
	iferr.Panic(v.BindEnv("axiom.org_id", "AXIOM_ORG_ID"))
	iferr.Panic(v.BindEnv("axiom.url", "AXIOM_URL"))
	iferr.Panic(v.BindEnv("server.port", "PORT")) // Use simple PORT env var

	// Read config file (optional)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found is OK, we can use env vars + defaults
	}

	// Handle port precedence: PORT env > --port flag > config file > default
	if portFlag > 0 {
		// Flag was explicitly set, override config file but not env var
		if os.Getenv("PORT") == "" {
			v.Set("server.port", portFlag)
		}
	}

	// Unmarshal into struct
	var config AppConfig
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate required fields
	if err := validateConfig(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("axiom.url", "https://app.axiom.co")
	v.SetDefault("server.host", "127.0.0.1")
	v.SetDefault("server.port", 5111)
	v.SetDefault("queries.file", "queries.yaml")
	v.SetDefault("queries.cache_ttl", "5m")
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "text")
}

func getConfigDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".config", configDirName)
}

func validateConfig(config *AppConfig) error {
	if config.Axiom.Token == "" {
		return fmt.Errorf("AXIOM_TOKEN is required (set via environment variable or config file)")
	}
	return nil
}

// CreateExampleConfig creates an example config file and queries file
func CreateExampleConfig() error {
	configDir := getConfigDir()
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	queriesPath := filepath.Join(configDir, "queries.yaml")

	// Check if files already exist and return errors if they do
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("config file already exists at %s", configPath)
	}

	if _, err := os.Stat(queriesPath); err == nil {
		return fmt.Errorf("queries file already exists at %s", queriesPath)
	}

	// Prepare the config content by replacing the placeholder
	configContent := strings.ReplaceAll(embeddedConfigTemplate, "{{QUERIES_PATH}}", queriesPath)

	// Write the config file
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// Write the queries file
	if err := os.WriteFile(queriesPath, []byte(embeddedQueriesTemplate), 0644); err != nil {
		return fmt.Errorf("failed to write queries file: %w", err)
	}

	fmt.Printf("Example config created at: %s\n", configPath)
	fmt.Printf("Example queries created at: %s\n", queriesPath)
	fmt.Println("Please edit the config file and add your AXIOM_TOKEN.")
	return nil
}
