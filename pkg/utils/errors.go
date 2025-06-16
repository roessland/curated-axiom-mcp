package utils

import (
	"fmt"
)

// QueryNotFoundError represents an error when a query is not found
type QueryNotFoundError struct {
	QueryName string
}

func (e *QueryNotFoundError) Error() string {
	return fmt.Sprintf("query '%s' not found", e.QueryName)
}

// ParameterError represents an error with query parameters
type ParameterError struct {
	Parameter string
	Message   string
}

func (e *ParameterError) Error() string {
	return fmt.Sprintf("parameter '%s': %s", e.Parameter, e.Message)
}

// AxiomError represents an error from the Axiom API
type AxiomError struct {
	StatusCode int
	Message    string
}

func (e *AxiomError) Error() string {
	return fmt.Sprintf("axiom error (status %d): %s", e.StatusCode, e.Message)
}

// ConfigError represents a configuration error
type ConfigError struct {
	Field   string
	Message string
}

func (e *ConfigError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("config error (%s): %s", e.Field, e.Message)
	}
	return fmt.Sprintf("config error: %s", e.Message)
}

// NewQueryNotFoundError creates a new QueryNotFoundError
func NewQueryNotFoundError(queryName string) *QueryNotFoundError {
	return &QueryNotFoundError{QueryName: queryName}
}

// NewParameterError creates a new ParameterError
func NewParameterError(parameter, message string) *ParameterError {
	return &ParameterError{Parameter: parameter, Message: message}
}

// NewAxiomError creates a new AxiomError
func NewAxiomError(statusCode int, message string) *AxiomError {
	return &AxiomError{StatusCode: statusCode, Message: message}
}

// NewConfigError creates a new ConfigError
func NewConfigError(field, message string) *ConfigError {
	return &ConfigError{Field: field, Message: message}
}
