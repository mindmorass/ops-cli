# Workflow TOML Format Reference

Workflow files define sequences of terminal commands that can be executed using the `workflow` command. Workflow files are located at:

```
~/.config/<cli-name>/workflows/<workflow-name>.toml
```

## File Format

Workflow files use TOML format and support two syntax styles for defining commands.

### Format 1: Commands Section

```toml
[commands]
"git status"
"git add ."
"git commit -m 'Update files'"
"git push"
```

### Format 2: Array Syntax

```toml
commands = [
  "git status",
  "git add .",
  "git commit -m 'Update files'",
  "git push"
]
```

## Examples

### Simple Git Workflow

**File:** `~/.config/<cli-name>/workflows/git.toml`

```toml
[commands]
"git status"
"git add ."
"git commit -m 'Update'"
"git push"
```

**Usage:**
```bash
ops-cli workflow git
```

### Deployment Workflow

**File:** `~/.config/<cli-name>/workflows/deploy.toml`

```toml
commands = [
  "echo 'Starting deployment...'",
  "go build",
  "go test",
  "docker build -t myapp:latest .",
  "docker push myapp:latest",
  "kubectl apply -f k8s/deployment.yaml",
  "echo 'Deployment complete!'"
]
```

**Usage:**
```bash
ops-cli workflow deploy
```

### Cleanup Workflow with Comments

**File:** `~/.config/<cli-name>/workflows/cleanup.toml`

```toml
[commands]
# Remove old Docker containers
"docker container prune -f"

# Remove unused images
"docker image prune -a -f"

# Clean up build artifacts
"rm -rf dist/ build/ *.log"
```

### Complex Commands with Pipes

**File:** `~/.config/<cli-name>/workflows/logs.toml`

```toml
[commands]
# Get recent errors from logs
"tail -n 100 /var/log/app.log | grep ERROR"

# Count unique IP addresses
"cat /var/log/access.log | awk '{print $1}' | sort | uniq | wc -l"

# Find large files
"find . -type f -size +100M -exec ls -lh {} \\;"
```

### Testing Workflow

**File:** `~/.config/<cli-name>/workflows/test.toml`

```toml
commands = [
  "echo 'Running linter...'",
  "go fmt ./...",
  "echo 'Running vet...'",
  "go vet ./...",
  "echo 'Running unit tests...'",
  "go test ./...",
  "echo 'Running integration tests...'",
  "go test -tags=integration ./...",
  "echo 'All tests passed!'"
]
```

### Multi-Service Deployment

**File:** `~/.config/<cli-name>/workflows/deploy-all.toml`

```toml
[commands]
# Deploy frontend
"cd frontend && go build && ./deploy"

# Deploy backend
"cd backend && docker build -t backend:latest . && docker push backend:latest"

# Deploy API
"cd api && kubectl apply -f k8s/api.yaml"

# Run health checks
"curl -f https://api.example.com/health || exit 1"
"curl -f https://app.example.com/health || exit 1"
```

## Command Execution

Commands are executed sequentially in the order they appear in the file. Each command runs in a shell environment, so you can use:

- **Pipes:** `|` for chaining commands
- **Redirects:** `>`, `>>`, `<` for I/O redirection
- **Logical operators:** `&&`, `||` for conditional execution
- **Environment variables:** `$VAR` or `${VAR}`
- **Command substitution:** `` `command` `` or `$(command)`
- **Quotes:** Use quotes for commands with spaces or special characters

## Usage

### Execute a Workflow

```bash
ops-cli workflow <workflow-name>
```

### Execute with Verbose Output

```bash
ops-cli workflow <workflow-name> --verbose
```

Shows each command as it executes and displays output.

### Continue on Error

```bash
ops-cli workflow <workflow-name> --continue-on-error
```

Continues executing remaining commands even if one fails. Useful for cleanup workflows.

### List Available Workflows

```bash
ops-cli workflow --help
```

Shows all discovered workflows in the workflows directory.

## Best Practices

1. **Use quotes** for commands with spaces or special characters
2. **Add comments** to explain what each command does
3. **Test workflows** with `--verbose` flag first
4. **Use descriptive names** for workflow files (e.g., `git-commit.toml` not `gc.toml`)
5. **Keep workflows focused** on a single task
6. **Handle errors** appropriately - use `--continue-on-error` when needed
7. **Use environment variables** for configuration that varies between environments

## Workflow Discovery

Workflows are automatically discovered from the workflows directory. Simply create a `.toml` file and it will appear as a subcommand:

- `git.toml` → `ops-cli workflow git`
- `deploy-staging.toml` → `ops-cli workflow deploy-staging`
- `test-unit.toml` → `ops-cli workflow test-unit`

The workflow name is derived from the filename (without the `.toml` extension).

## Error Handling

By default, if any command fails (non-zero exit code), the workflow stops immediately. Use the `--continue-on-error` flag to continue executing remaining commands:

```bash
ops-cli workflow cleanup --continue-on-error
```

This is useful for cleanup workflows where some operations might fail but you want to continue with others.

## See Also

- [Workflow Usage Examples](../examples/workflow-usage.ts) - Comprehensive examples
- [Main README](../README.md) - General CLI documentation
- Run `ops-cli workflow --help` for command-line help

