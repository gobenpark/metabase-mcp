# metabase-mcp

A Model Context Protocol (MCP) server for [Metabase](https://www.metabase.com), built in Go. Connect Claude (or any MCP client) directly to your Metabase instance to query data, browse dashboards, and manage collections through natural language.

## Features

| Category | Tool | Description |
|----------|------|-------------|
| **Query** | `execute_query` | Run native SQL queries against any connected database |
| | `run_question` | Execute a saved Metabase question by ID |
| **Browse** | `list_databases` | List all connected databases |
| | `list_dashboards` | List all dashboards |
| | `get_dashboard` | Get dashboard details including cards |
| | `list_collections` | List all collections (folders) |
| | `search` | Full-text search across questions, dashboards, and collections |
| **Create** | `create_dashboard` | Create a new dashboard |
| | `create_collection` | Create a new collection (folder) |
| | `add_card_to_dashboard` | Add a saved question to a dashboard |
| **Manage** | `move_dashboard` | Move a dashboard to a different collection |
| | `delete_dashboard` | Permanently delete a dashboard |
| | `archive_collection` | Archive a collection |

## Installation

### npx (Recommended)

No prerequisites required. The binary is downloaded automatically on first run:

```bash
npx metabase-mcp
```

For Claude Code, use `npx` in your MCP config:

```json
{
  "mcpServers": {
    "metabase": {
      "command": "npx",
      "args": ["-y", "metabase-mcp"],
      "env": {
        "METABASE_URL": "https://your-metabase.example.com",
        "METABASE_API_KEY": "mb_your_api_key_here"
      }
    }
  }
}
```

### Pre-built binaries

Download from [GitHub Releases](https://github.com/gobenpark/metabase-mcp/releases):

```bash
# macOS (Apple Silicon)
curl -L https://github.com/gobenpark/metabase-mcp/releases/latest/download/metabase-mcp_darwin_arm64.tar.gz | tar xz
mv metabase-mcp /usr/local/bin/

# macOS (Intel)
curl -L https://github.com/gobenpark/metabase-mcp/releases/latest/download/metabase-mcp_darwin_amd64.tar.gz | tar xz
mv metabase-mcp /usr/local/bin/

# Linux (amd64)
curl -L https://github.com/gobenpark/metabase-mcp/releases/latest/download/metabase-mcp_linux_amd64.tar.gz | tar xz
mv metabase-mcp /usr/local/bin/
```

### From source

```bash
go install github.com/gobenpark/metabase-mcp@latest
```

## Configuration

### 1. Get a Metabase API Key

Go to **Metabase Admin** > **Settings** > **Authentication** > **API Keys** and create a new key.

### 2. Configure Claude Code

Using the CLI:

```bash
# Add to current project
claude mcp add metabase -e METABASE_URL=https://your-metabase.example.com -e METABASE_API_KEY=mb_your_api_key_here -- npx -y metabase-mcp

# Add globally (available in all projects)
claude mcp add metabase -s user -e METABASE_URL=https://your-metabase.example.com -e METABASE_API_KEY=mb_your_api_key_here -- npx -y metabase-mcp
```

Or manually add to your `.mcp.json` (project-level) or `~/.claude/settings.json` (global):

```json
{
  "mcpServers": {
    "metabase": {
      "command": "npx",
      "args": ["-y", "metabase-mcp"],
      "env": {
        "METABASE_URL": "https://your-metabase.example.com",
        "METABASE_API_KEY": "mb_your_api_key_here"
      }
    }
  }
}
```

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `METABASE_URL` | Yes | Your Metabase instance URL |
| `METABASE_API_KEY` | Yes | Metabase API key |

## Usage Examples

Once configured, use natural language with Claude:

```
> List all databases connected to Metabase
> Run a SQL query: SELECT count(*) FROM orders WHERE created_at > '2025-01-01'
> Show me all dashboards
> Search for anything related to "revenue"
> Create a new dashboard called "Weekly KPIs" in the Reports collection
> Add question #12 to dashboard #5
```

## Development

```bash
# Clone
git clone https://github.com/gobenpark/metabase-mcp.git
cd metabase-mcp

# Build
go build -o metabase-mcp .

# Test locally
export METABASE_URL="http://localhost:3000"
export METABASE_API_KEY="mb_your_key"
./metabase-mcp
```

## Release

Releases are automated via GitHub Actions + [GoReleaser](https://goreleaser.com). To create a release:

```bash
git tag v0.1.0
git push origin v0.1.0
```

This builds binaries for Linux, macOS, and Windows (amd64/arm64) automatically.

## License

MIT
