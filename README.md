# ops-cli - Modular API Command Line Tool

A Go CLI tool that provides modular access to various APIs. The architecture follows the Cobra pattern with `<module> <subcommand>` structure for easy extensibility.

## Features

- **Modular Architecture**: Easy to add new API modules (Jira, GitHub, Docker, etc.)
- **Go**: Fast, compiled, single binary
- **Cobra Framework**: Powerful CLI framework with comprehensive help system
- **Cross-platform**: Works on macOS, Linux, and Windows
- **Comprehensive Help**: Context-aware help system at all levels

## Quick Start

```bash
# Clone and navigate to the project
cd /path/to/ops-cli

# Build the CLI
go build -o bin/ops-cli main.go

# Run the CLI
./bin/ops-cli --help

# Show module help
./bin/ops-cli jira --help
./bin/ops-cli github --help

# Run commands
./bin/ops-cli jira list --project ABC
./bin/ops-cli github repos --user octocat
```

## Available Modules

### Jira Module

- `ops-cli jira list [options]` - List Jira issues
- `ops-cli jira get <issue-key> [options]` - Get a specific issue
- `ops-cli jira create <summary> [options]` - Create a new issue
- `ops-cli jira time [options]` - Log time and generate reports
- `ops-cli jira config setup` - Configure Jira API credentials

### GitHub Module

- `ops-cli github repos [options]` - List repositories
- `ops-cli github repo <owner/repo> [options]` - Get repository details
- `ops-cli github search <owner/repo> <query>` - Search for files
- `ops-cli github releases <owner/repo>` - List releases
- `ops-cli github config setup` - Configure GitHub API credentials

### Docker Module

- `ops-cli docker ps` - List containers
- `ops-cli docker images` - List images
- `ops-cli docker logs <container>` - View container logs
- `ops-cli docker start` - Start Docker Desktop (macOS)
- `ops-cli docker stop` - Stop Docker Desktop (macOS)

### Confluence Module

- `ops-cli confluence get <page-id>` - Get a page
- `ops-cli confluence config setup` - Configure Confluence API

### New Relic Module

- `ops-cli newrelic logs stream` - Stream logs
- `ops-cli newrelic entities` - List entities
- `ops-cli newrelic config setup` - Configure New Relic API

### DevTools Module (macOS)

- `ops-cli devtools check` - Check installed tools
- `ops-cli devtools install` - Install development tools

### Startpage Module

- `ops-cli startpage init` - Initialize a new startpage
- `ops-cli startpage add` - Add a bookmark
- `ops-cli startpage dev` - Run development server

## Development

### Project Structure

```
ops-cli/
├── main.go                 # Main CLI entry point
├── cmd/                    # Command modules
│   ├── jira/              # Jira commands
│   ├── github/            # GitHub commands
│   ├── docker/            # Docker commands
│   └── ...
├── internal/              # Internal packages
│   ├── api/               # API client libraries
│   │   ├── base/         # Generic HTTP client
│   │   ├── jira/         # Jira API client
│   │   └── github/       # GitHub API client
│   ├── cli/               # CLI framework components
│   ├── config/            # Configuration management
│   └── ui/                # UI utilities (spinners, etc.)
├── go.mod                 # Go module definition
└── go.sum                 # Go module checksums
```

### Building

```bash
# Build for current platform
go build -o bin/ops-cli main.go

# Build for specific platforms
GOOS=linux GOARCH=amd64 go build -o bin/ops-cli-linux main.go
GOOS=darwin GOARCH=amd64 go build -o bin/ops-cli-macos main.go
GOOS=windows GOARCH=amd64 go build -o bin/ops-cli.exe main.go
```

### Testing

```bash
# Run tests
go test ./...

# Run with coverage
go test -cover ./...
```

## Configuration

The CLI supports configuration through:

1. **Environment Variables** (take precedence):

   ```bash
   export JIRA_BASE_URL="https://your-company.atlassian.net"
   export JIRA_USERNAME="your-email@company.com"
   export ATLASSIAN_TOKEN="your-atlassian-token"
   export GITHUB_TOKEN="your-github-token"
   ```

2. **Config Files**: `~/.config/ops-cli/config.toml`

   ```toml
   [jira]
   base_url = "https://your-company.atlassian.net"
   username = "your-email@company.com"
   atlassian_token = "your-token"

   [github]
   token = "your-github-token"
   default_owner = "your-username"
   ```

## Contributing

1. Follow Go best practices and conventions
2. Use Cobra for command structure
3. Add comprehensive help text and examples
4. Ensure commands work with both config files and environment variables
5. Use spinners for long-running operations (no icons)

## License

[Your License Here]
