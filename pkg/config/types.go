package config

import "time"

// AppConfig represents the application configuration
type AppConfig struct {
	Axiom   AxiomConfig   `yaml:"axiom" mapstructure:"axiom"`
	Server  ServerConfig  `yaml:"server" mapstructure:"server"`
	Queries QueriesConfig `yaml:"queries" mapstructure:"queries"`
	Logging LoggingConfig `yaml:"logging" mapstructure:"logging"`
}

type AxiomConfig struct {
	Token   string `yaml:"token" mapstructure:"token"`
	OrgID   string `yaml:"org_id" mapstructure:"org_id"`
	Dataset string `yaml:"dataset" mapstructure:"dataset"`
	URL     string `yaml:"url" mapstructure:"url"`
}

type ServerConfig struct {
	Host string `yaml:"host" mapstructure:"host"`
	Port int    `yaml:"port" mapstructure:"port"`
}

type QueriesConfig struct {
	File     string        `yaml:"file" mapstructure:"file"`
	CacheTTL time.Duration `yaml:"cache_ttl" mapstructure:"cache_ttl"`
}

type LoggingConfig struct {
	Level  string `yaml:"level" mapstructure:"level"`
	Format string `yaml:"format" mapstructure:"format"`
}

// QueryRegistry holds all available queries
type QueryRegistry struct {
	Queries map[string]Query `yaml:"queries"`
}

// Query definition
type Query struct {
	Name         string      `yaml:"name"`
	Description  string      `yaml:"description"`
	APLQuery     string      `yaml:"apl_query"`
	Parameters   []Parameter `yaml:"parameters"`
	OutputFormat string      `yaml:"output_format"`
	LLMFriendly  bool        `yaml:"llm_friendly"`
	Tags         []string    `yaml:"tags,omitempty"`
}

type Parameter struct {
	Name        string `yaml:"name"`
	Type        string `yaml:"type"` // string, int, float, datetime, duration
	Required    bool   `yaml:"required"`
	Default     any    `yaml:"default,omitempty"`
	Description string `yaml:"description"`
	Pattern     string `yaml:"pattern,omitempty"`
	Enum        []any  `yaml:"enum,omitempty"`
}
