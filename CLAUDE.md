# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

This is an MCP (Model Context Protocol) server that provides LLM-friendly access to curated Axiom queries. It's built in Go and provides both stdio and HTTP server modes for MCP communication.

## Key Architecture

- **Entry Point**: `main.go` → `cmd/root.go` - Uses Cobra for CLI structure
- **MCP Server**: `pkg/cserver/mcp.go` - Core MCP server implementation using mark3labs/mcp-go
- **Query Execution**: `pkg/cserver/tool-run-query.go` - Handles APL query execution
- **Configuration**: `pkg/config/` - Handles app config, query registry, and YAML loading
- **Formatting**: `pkg/formatter/` - Formats Axiom results for LLM consumption
- **Axiom Client**: `pkg/axiom/` - Wrapper around axiom-go library

## Development Commands

Build the project:
```bash
just build
# or
go build -o curated-axiom-mcp ./main.go
```

Run tests:
```bash
just test
# or
go test ./...
```

Run quality checks (format, vet, lint, test):
```bash
just check
```

Run integration tests (requires private Axiom instance):
```bash
just integration-test
```

## Running the MCP Server

Start in stdio mode (for MCP clients):
```bash
go run main.go --stdio
```

Start HTTP server (SSE mode):
```bash
go run main.go
# or with custom port
go run main.go --port 8080
```

## Testing MCP Tools

List available MCP tools:
```bash
mcpt tools go run main.go --stdio
```

Run APL query via MCP:
```bash
mcpt call run_query --params '{"apl": "[\"my-service\"] | where _time > ago(1h) | count"}' go run main.go --stdio
```

Debug starred queries:
```bash
mcpt call debug_starred_queries --params '{}' go run main.go --stdio
```

## Configuration

The app uses a layered configuration system:
1. Command line flags
2. Environment variables (`AXIOM_TOKEN`, `AXIOM_URL`, `PORT`)
3. Config file (`~/.config/curated-axiom-mcp/config.yaml`)
4. Defaults

Key config components:
- **AppConfig** (`pkg/config/types.go`): Main config structure
- **QueryRegistry** (`pkg/config/registry.go`): Manages query definitions
- **Embedded queries** (`pkg/config/embedded_queries.yaml`): Default query templates

## Query System

Queries are defined in YAML with:
- APL query templates with `{parameter}` placeholders
- Parameter definitions with types, validation, defaults
- Output formatting options
- LLM-friendly metadata

Example query structure:
```yaml
queries:
  user_events:
    name: "user_events"
    description: "Get events for a specific user"
    apl_query: |
      ['events']
      | where user_id == {user_id}
      | where _time >= {start_time}
      | limit {limit}
    parameters:
      - name: "user_id"
        type: "string"
        required: true
```

## Key Dependencies

- `github.com/axiomhq/axiom-go` - Axiom API client
- `github.com/mark3labs/mcp-go` - MCP protocol implementation
- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuration management

## Debugging MCP Tools

### Essential Debugging Commands

**1. Check tool registration and startup:**
```bash
LOG_LEVEL=debug mcptools call asdf go run main.go --stdio
# Returns "error: tool 'asdf' not found: tool not found" when working correctly
# Check stderr log for dynamic tool loading errors
tail ~/.config/curated-axiom-mcp/stderr.log
```

**2. List available tools:**
```bash
mcptools tools go run main.go --stdio
# Shows all registered tools with their parameters
# Dynamic tools should appear here after successful loading
```

**3. Test dynamic tool execution:**
```bash
mcptools call <tool_name> --params '{"param1": "value1", "param2": "value2"}' go run main.go --stdio
# Use actual parameter names from tool definition
# Check stderr log for template rendering and APL execution details
```

**4. Debug starred query parsing:**
- Look for these log patterns in stderr:
  - `"Successfully loaded dynamic tools" count=N` - N should be > 0
  - `"Failed to parse CuratedAxiomMCP query"` - YAML/template errors
  - `"Skipping starred query (no CuratedAxiomMCP marker)"` - Normal for non-MCP queries
  - `"Rendered APL query"` - Shows final templated APL sent to Axiom

**5. Common issues and fixes:**
- **No tools loaded**: Check YAML syntax in starred query comments
- **Template errors**: Verify `///param=` annotations (no space after ///)
- **Parameter errors**: Ensure proper YAML structure with `Name:`, `Type:`, etc.
- **Query execution errors**: Check rendered APL in debug logs

### Query Format Requirements

Starred queries must contain properly formatted YAML metadata:

```yaml
// CuratedAxiomMCP:
//   ToolName: my_tool_name
//   Params:
//     - Name: ParamName
//       Type: string
//       Example: example-value
//       Description: Parameter description
//   Constraints:
//     - Constraint description
```

And parameter annotations in APL:
```apl
param_name:type = default_value ///param=template_replacement
```

## Context7 Library IDs

- For Cobra: /spf13/cobra
- For axiom: /axiomhq/docs
- For go/golang: /golang/website or /llmstxt/mcp-go_dev-llms.txt
- For mcp-go: /mark3labs/mcp-go
