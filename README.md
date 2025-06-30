# curated-axiom-mcp

An MCP (Model Context Protocol) server that provides LLM-friendly access to curated Axiom queries with intelligent formatting and column statistics.

## Features

- 🔍 **Dynamic Tool Generation**: Automatically creates MCP tools from Axiom starred queries
- 🤖 **LLM-Optimized Formatting**: Results formatted as markdown with CSV data and comprehensive column statistics
- 📊 **Smart Data Analysis**: Automatic column stats including unique values, examples, and frequency analysis
- 🚀 **Dual MCP Modes**: Supports both stdio and HTTP server modes
- ⚡ **Response Size Management**: Warns when responses exceed 20KB to optimize LLM context usage
- 🛡️ **Parameter Validation**: Type checking and validation for query parameters

## Quick Start

### 1. Installation

```bash
go install github.com/roessland/curated-axiom-mcp@latest
```

### 2. Configuration

Set your Axiom token and URL via environment variables:

```bash
export AXIOM_TOKEN="your-axiom-token-here"
export AXIOM_URL="https://api.axiom.co"  # or https://api.eu.axiom.co for EU
```

Or create a config file at `~/.config/curated-axiom-mcp/config.yaml` (no environment variables needed if using config file):

```yaml
axiom:
  token: "your-axiom-token"
  url: "https://api.axiom.co"
```

### 3. Start MCP Server

```bash
# Stdio mode (for MCP clients like Claude Desktop)
curated-axiom-mcp --stdio

# HTTP server mode (for testing with mcptools)
curated-axiom-mcp --port 8080
```

## Testing with mcptools

Install mcptools for testing:
```bash
npm install -g @modelcontextprotocol/cli
```

Test the server:
```bash
# List available tools
mcptools tools go run main.go --stdio

# Test a manual query
mcptools call run_query --params '{"apl": "[\"activities\"] | where athlete_id == \"12345\" | where _time > ago(24h) | limit 10"}' go run main.go --stdio

# Test a dynamic tool (if you have starred queries)
mcptools call athlete_activity_summary --params '{"athlete_id": "12345", "time_range": "7d"}' go run main.go --stdio
```

## Creating Dynamic Tools from Starred Queries

The server automatically converts Axiom starred queries into MCP tools when they contain special metadata comments:

### Example Starred Query

Create a starred query in Axiom with this APL and metadata:

```apl
// CuratedAxiomMCP:
//   ToolName: athlete_activity_summary
//   Description: Get activity summary for a specific athlete
//   Params:
//     - Name: athlete_id
//       Type: string
//       Example: "12345"
//       Description: The athlete's unique identifier
//     - Name: time_range
//       Type: string
//       Example: "7d"
//       Description: Time range (e.g., 1h, 24h, 7d, 30d)
//   Constraints:
//     - Only returns activities from the last 90 days
//     - Limited to 1000 results for performance

athlete_id:string = "12345", ///param=athlete_id,
time_range:string = "7d" ///param=time_range

["strava_activities"]
| where athlete_id == athlete_id
| where _time > ago(time_range)
| summarize 
    total_activities = count(),
    total_distance = sum(distance),
    avg_speed = avg(average_speed),
    activity_types = dcount(sport_type)
  by athlete_id
| extend
    total_distance_km = round(total_distance / 1000, 2),
    avg_speed_kmh = round(avg_speed * 3.6, 2)
| project athlete_id, total_activities, total_distance_km, avg_speed_kmh, activity_types
```

### Metadata Format

The starred query must contain YAML metadata in comments:

```yaml
// CuratedAxiomMCP:
//   ToolName: your_tool_name           # Required: MCP tool name
//   Description: Tool description      # Optional: Tool description
//   Params:                           # Required: Parameter definitions
//     - Name: param_name              # Parameter name
//       Type: string                  # Type: string, int, float, bool
//       Example: "example_value"      # Example value
//       Description: Parameter desc   # Parameter description
//   Constraints:                      # Optional: Usage constraints
//     - Constraint description
```

### Parameter Annotations

Parameters must be annotated in the APL query:

```apl
param_name:type = default_value, ///param=template_replacement,
final_param:type = default_value ///param=template_replacement
```

## Output Format

The server returns structured markdown with:

### Query Summary
```
Found 42 records in 0.8 s with 6 fields.
```

### APL Query Section
```markdown
## APL
["strava_activities"] | where athlete_id == "12345" | where _time > ago(7d)
```

### Results Section
```markdown
## Results
athlete_id,total_activities,total_distance_km,avg_speed_kmh,activity_types
string,int,float,float,int
---
12345,15,247.8,18.5,3
67890,8,156.2,16.7,2
```

### Column Statistics
```markdown
## Column Stats

**athlete_id** (string)
- First: "12345", Last: "67890"
- Nulls: 0
- Unique: 2 total
- Top values: "12345" (15), "67890" (8)
- Examples: "12345", "67890"

**total_activities** (int)
- First: "15", Last: "8"
- Nulls: 0
- Unique: 2 total
- Top values: "15" (1), "8" (1)
- Examples: "15", "8"
```

## Configuration

### Environment Variables

| Variable       | Description                | Default                |
| -------------- | -------------------------- | ---------------------- |
| `AXIOM_TOKEN`  | Axiom API token (required) | -                      |
| `AXIOM_ORG_ID` | Axiom organization ID      | -                      |
| `AXIOM_URL`    | Axiom base URL             | `https://api.axiom.co` |
| `PORT`         | Server port                | 5111                   |
| `LOG_LEVEL`    | Logging level              | info                   |

### Configuration File

Location: `~/.config/curated-axiom-mcp/config.yaml`

```yaml
axiom:
  token: "your-axiom-token"
  org_id: "your-org-id" # optional
  url: "https://api.axiom.co" # or https://api.eu.axiom.co for EU

server:
  host: "127.0.0.1"
  port: 5111

queries:
  cache_ttl: "5m"

logging:
  level: "info"
  format: "text"
```

### Regions

| Region | Base URL                  | Environment Variable Setting        |
| ------ | ------------------------- | ----------------------------------- |
| US     | `https://api.axiom.co`    | `AXIOM_URL=https://api.axiom.co`    |
| EU     | `https://api.eu.axiom.co` | `AXIOM_URL=https://api.eu.axiom.co` |

## Example Use Cases

### Fitness Activity Analysis
```apl
// Track athlete performance metrics
["strava_activities"] 
| where athlete_id == "12345"
| where sport_type == "running"
| where _time > ago(30d)
| summarize avg_pace = avg(average_speed), total_distance = sum(distance)
```

### Event Monitoring
```apl
// Monitor application events
["app_events"]
| where user_id == "user123"
| where event_type == "workout_completed"
| where _time >= ago(7d)
| summarize events = count() by bin(1d, _time)
```

### Performance Metrics
```apl
// Analyze API response times
["api_logs"]
| where endpoint contains "activities"
| where _time >= ago(1h)
| summarize avg_response_time = avg(response_time_ms), error_rate = countif(status_code >= 400) * 100.0 / count()
```

## Development

### Building
```bash
go build -o curated-axiom-mcp ./main.go
```

### Testing
```bash
# Run tests
go test ./...

# Run with quality checks
just check  # requires justfile

# Integration testing (requires Axiom config)
just integration-test
```

### Debug Logging

Debug logs are written to `~/.config/curated-axiom-mcp/stderr.log`. Tool calls and responses are also logged to `stdout.log` for debugging.

```bash
# Monitor debug logs
tail -f ~/.config/curated-axiom-mcp/stderr.log

# Monitor tool execution
tail -f ~/.config/curated-axiom-mcp/stdout.log
```

## MCP Integration

### Claude Desktop

Add to your Claude Desktop MCP configuration:

```json
{
  "mcpServers": {
    "curated-axiom-mcp": {
      "command": "curated-axiom-mcp",
      "args": ["--stdio"],
      "env": {
        "AXIOM_TOKEN": "your-token-here",
        "AXIOM_URL": "https://api.axiom.co"
      }
    }
  }
}
```

Note: The `env` section is only needed if you're not using a config file. If you have `~/.config/curated-axiom-mcp/config.yaml` with your credentials, you can omit the environment variables.

### Other MCP Clients

The server supports standard MCP protocol over both stdio and HTTP transports.

## Features in Detail

### Column Statistics
- **First/Last Values**: Shows first and last values in the dataset
- **Null Counts**: Tracks empty/null values
- **Unique Value Analysis**: Counts unique values with frequency analysis
- **Smart Sampling**: For large datasets, intelligently samples data for statistics
- **High Cardinality Handling**: Summarizes columns with many unique values

### Response Size Management
- Warns when responses exceed 20KB to help optimize LLM context usage
- Automatically truncates debug logs to manageable sizes
- Provides row limiting options for large result sets

### Error Handling
- Detailed error logging for debugging
- Graceful handling of malformed starred queries
- Clear error messages for missing parameters or invalid queries

## License

MIT License - see LICENSE file for details.