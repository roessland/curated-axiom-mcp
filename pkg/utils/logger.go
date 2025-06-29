package utils

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/roessland/curated-axiom-mcp/pkg/config"
)

// SetupLogger configures the global slog logger based on the configuration
func SetupLogger(cfg *config.LoggingConfig, stdioMode bool) {
	var level slog.Level
	switch strings.ToLower(cfg.Level) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	var output io.Writer = os.Stderr
	
	// If in stdio mode, log to file instead of stderr
	if stdioMode {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			// Fallback to stderr if we can't get home directory
			output = os.Stderr
		} else {
			logDir := filepath.Join(homeDir, ".config", "curated-axiom-mcp")
			logFile := filepath.Join(logDir, "stderr.log")
			
			// Create directory if it doesn't exist
			if err := os.MkdirAll(logDir, 0755); err != nil {
				// Fallback to stderr if we can't create directory
				output = os.Stderr
			} else {
				// Open log file for appending
				if file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err != nil {
					// Fallback to stderr if we can't open file
					output = os.Stderr
				} else {
					output = file
				}
			}
		}
	}

	var handler slog.Handler
	if strings.ToLower(cfg.Format) == "json" {
		handler = slog.NewJSONHandler(output, &slog.HandlerOptions{
			Level: level,
		})
	} else {
		handler = slog.NewTextHandler(output, &slog.HandlerOptions{
			Level: level,
		})
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
}
