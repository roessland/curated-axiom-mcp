# curated-axiom-mcp

An MCP (Model Context Protocol) server that provides LLM-friendly access to curated Axiom queries with simplified, structured results.

## Features

- 🔍 **Curated Query Library**: Define and manage whitelisted Axiom queries in YAML
- 🤖 **LLM-Optimized**: Results formatted specifically for language model consumption
- 🚀 **Dual Modes**: Supports both stdio and SSE server modes
- 📊 **Smart Formatting**: Automatic detection of table, time-series, and summary data
- 🛡️ **Parameter Validation**: Type checking and validation for query parameters
- 🔧 **CLI Tools**: Test queries directly from command line

## Quick Start

### 1. Installation

```bash
go install github.com/roessland/curated-axiom-mcp@latest
```

### 2. Configuration

Set your Axiom token:

```bash
export AXIOM_TOKEN="your-axiom-token-here"
```

Or create a config file:

```bash
curated-axiom-mcp config init
# Edit ~/.config/curated-axiom-mcp/config.yaml
```

### 3. Define Queries

Create or modify `queries.yaml`:

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
        description: "User ID to filter events for"
      - name: "start_time"
        type: "datetime"
        required: true
        description: "Start time (ISO 8601 format)"
      - name: "limit"
        type: "int"
        required: false
        default: 100
        description: "Max number of events"
    output_format: "table"
    llm_friendly: true
    tags: ["user", "events"]
```

### 4. Test Queries

```bash
# Test a query directly
curated-axiom-mcp run user_events user_id="12345" start_time="2024-01-01T00:00:00Z"

# List available queries
curated-axiom-mcp list

# Get detailed query info
curated-axiom-mcp describe user_events
```

### 5. Start MCP Server

```bash
# Stdio mode (for direct MCP client connection)
curated-axiom-mcp --stdio

# SSE mode (HTTP server)
curated-axiom-mcp

# Custom port
PORT=8080 curated-axiom-mcp
# or
curated-axiom-mcp --port 8080
```

### 5. Optional: Set up shell completion

```
eval "$(./curated-axiom-mcp completion zsh)"
```

## Usage Modes

### MCP Server (Stdio)

For direct integration with MCP clients:

```bash
curated-axiom-mcp --stdio
```

### MCP Server (SSE)

For HTTP-based MCP connections:

```bash
curated-axiom-mcp
# Serves on http://127.0.0.1:5111/mcp by default
```

### CLI Query Execution

For testing and manual query execution:

```bash
curated-axiom-mcp run <query-name> [param1=value1] [param2=value2] ...
```

## Configuration

### Environment Variables

| Variable       | Description                | Default                |
| -------------- | -------------------------- | ---------------------- |
| `AXIOM_TOKEN`  | Axiom API token (required) | -                      |
| `AXIOM_ORG_ID` | Axiom organization ID      | -                      |
| `AXIOM_URL`    | Axiom base URL             | `https://api.axiom.co` |
| `PORT`         | Server port                | 5111                   |

### Configuration File

Location: `~/.config/curated-axiom-mcp/config.yaml`

```yaml
axiom:
  token: "your-axiom-token"
  org_id: "your-org-id" # optional
  dataset: "default-dataset" # optional
  url: "https://api.axiom.co" # optional, use "https://api.eu.axiom.co" for EU region

server:
  host: "127.0.0.1"
  port: 5111

queries:
  file: "queries.yaml"
  cache_ttl: "5m"

logging:
  level: "info"
  format: "text"
```

### Configuration Precedence

1. Command line flags (e.g., `--port`)
2. Environment variables (e.g., `PORT`)
3. Configuration file
4. Defaults

## Regions

Axiom supports different regions with different base URLs:

| Region | Base URL                  | Environment Variable Setting        |
| ------ | ------------------------- | ----------------------------------- |
| US     | `https://api.axiom.co`    | `AXIOM_URL=https://api.axiom.co`    |
| EU     | `https://api.eu.axiom.co` | `AXIOM_URL=https://api.eu.axiom.co` |

The `AXIOM_URL` is automatically converted to the corresponding API endpoint:

- US: `https://api.axiom.co` → `https://api.axiom.co/v1`
- EU: `https://api.eu.axiom.co` → `https://api.eu.axiom.co/v1`

### Setting Region

**Environment Variable:**

```bash
export AXIOM_URL="https://api.eu.axiom.co"  # For EU region
```

**Configuration File:**

```yaml
axiom:
  url: "https://api.eu.axiom.co" # For EU region
```

## Query Definition

### Query Structure

```yaml
queries:
  query_name:
    name: "query_name" # Query identifier
    description: "Human description" # Description for LLMs
    apl_query: | # APL query with {param} placeholders
      ['dataset']
      | where field == {param}
    parameters: # Parameter definitions
      - name: "param"
        type: "string" # string, int, float, datetime, duration, boolean
        required: true # Whether parameter is required
        default: "value" # Default value (optional)
        description: "Parameter desc" # Help text
        enum: ["val1", "val2"] # Allowed values (optional)
        pattern: "regex" # Validation pattern (optional)
    output_format: "table" # table, timeseries, summary
    llm_friendly: true # Optimize for LLM consumption
    tags: ["tag1", "tag2"] # Categorization tags
```

### Parameter Types

- **`string`**: Text values
- **`int`**: Integer numbers
- **`float`**: Decimal numbers
- **`boolean`**: true/false values
- **`datetime`**: ISO 8601 timestamps
- **`duration`**: Go duration strings (e.g., "1h", "30m")

### Parameter Placeholders

Use `{parameter_name}` in your APL queries:

```apl
['events']
| where user_id == {user_id}
| where _time >= {start_time}
| limit {limit}
```

## Commands

### Core Commands

- `curated-axiom-mcp` - Start SSE server
- `curated-axiom-mcp --stdio` - Start stdio server
- `curated-axiom-mcp run <query> [params...]` - Execute query
- `curated-axiom-mcp list` - List available queries
- `curated-axiom-mcp describe <query>` - Show query details

### Configuration Commands

- `curated-axiom-mcp config init` - Create example config
- `curated-axiom-mcp config show` - Display current config
- `curated-axiom-mcp config validate` - Test configuration

### Global Flags

- `--config <file>` - Custom config file path
- `--port <port>` - Override server port

## Example Queries

### Simple Event Query

```yaml
user_events:
  name: "user_events"
  description: "Get recent events for a user"
  apl_query: |
    ['events']
    | where user_id == {user_id}
    | where _time >= ago({timespan})
    | limit {limit}
  parameters:
    - name: "user_id"
      type: "string"
      required: true
    - name: "timespan"
      type: "duration"
      default: "1h"
    - name: "limit"
      type: "int"
      default: 100
```

### Error Analysis

```yaml
error_analysis:
  name: "error_analysis"
  description: "Analyze errors by service and time"
  apl_query: |
    ['logs']
    | where level == "error"
    | where service == {service}
    | where _time >= {start_time}
    | summarize errors = count() by bin(5m, _time), error_code
    | sort by _time desc
  parameters:
    - name: "service"
      type: "string"
      required: true
      enum: ["api", "web", "worker"]
    - name: "start_time"
      type: "datetime"
      default: "ago(1h)"
  output_format: "timeseries"
```

## Development

### Building

```bash
go build -o curated-axiom-mcp ./main.go
```

### Testing

```bash
# Test configuration
curated-axiom-mcp config validate

# Test a query
curated-axiom-mcp run user_events user_id="test" start_time="2024-01-01T00:00:00Z"

# Check available queries
curated-axiom-mcp list
```

## License

MIT License - see LICENSE file for details.
