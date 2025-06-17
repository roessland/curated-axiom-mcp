# curated-axiom-mcp justfile
# Use `just <recipe>` to run a recipe
# Use `just --list` to see all available recipes

# Default recipe - show available commands
default:
  @just --list

# Build the project
build:
  go build -o curated-axiom-mcp ./main.go

# Install dependencies and tidy go.mod
deps:
  go mod download
  go mod tidy

# Run tests
test:
  go test ./...

# Run integration tests
# (Uses private Axiom instance)
integration-test:
  cd private-integration-tests && just integration-test

# Run tests with coverage
test-coverage:
  go test -cover ./...

# Run tests with verbose output
test-verbose:
  go test -v ./...

# Run benchmarks
bench:
  go test -bench=. ./...

# Format code
fmt:
  go fmt ./...

# Run linter (requires golangci-lint)
lint:
  @command -v golangci-lint >/dev/null 2>&1 || (echo "golangci-lint not installed. Install with: brew install golangci-lint" && exit 1)
  golangci-lint run

# Vet code
vet:
  go vet ./...

# Run all quality checks
[group('quality')]
check: fmt vet lint test

# Clean build artifacts
clean:
  rm -f curated-axiom-mcp