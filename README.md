# Plugin Development Guide

This document provides instructions on how to add a new plugin to the tool. Plugins extend the functionality of the application and can be used to integrate with various services or add new features.

## Table of Contents

1. [Overview](#overview)
2. [Creating a New Plugin](#creating-a-new-plugin)
3. [Registering the Plugin](#registering-the-plugin)
4. [Implementing Plugin Commands](#implementing-plugin-commands)
5. [Testing the Plugin](#testing-the-plugin)
6. [Example Plugin](#example-plugin)

## Overview

Plugins are designed to be modular components that can be easily added or removed from the application. Each plugin can define its own commands and functionality, allowing for a flexible and extensible architecture.

## Creating a New Plugin

1. **Create a New Python File**:

   - Navigate to the `plugins` directory in your project
   - Create a new Python file for your plugin, e.g., `my_new_plugin.py`

2. **Define the Plugin Class**:

```python
from typing import Optional
import typer
from rich.console import Console
from rich.table import Table

from apis.client_api import ClientApi
from apis.core.plugin_base import PluginBase

class MyNewPlugin(PluginBase):
    def setup(self) -> None:
        """Register plugin commands"""
        self.register_command(
            name="list",
            callback=self.list_items,
            help="List items",
        )

    def list_items(
        self,
        format: str = typer.Option("table", help="Output format (table/json)"),
    ):
        """Example command implementation"""
        try:
            # Access APIs through self.client
            items = self.client.github.get_user_repos()

            console = Console()
            if format == "json":
                console.print_json(data=items)
                return

            table = Table(title="Items")
            table.add_column("Name")
            table.add_column("Status")

            for item in items:
                table.add_row(
                    item["name"],
                    f"[green]{item['status']}[/]"
                )

            console.print(table)

        except Exception as e:
            typer.echo(f"Error: {str(e)}", err=True)
            raise typer.Exit(1)

def setup_plugin(client: ClientApi) -> MyNewPlugin:
    """Required plugin setup function"""
    plugin = MyNewPlugin(client)
    plugin.setup()
    return plugin
```

## Registering the Plugin

Plugins are automatically registered when placed in the `plugins` directory. The CLI will:

1. Discover the plugin file
2. Look for the `setup_plugin` function
3. Call it with the client instance
4. Register the plugin's commands

## Implementing Plugin Commands

Each plugin command should:

1. Have clear help text
2. Use type hints
3. Handle errors gracefully
4. Provide user feedback

Example command implementation:

```python
def my_command(
    self,
    required_arg: str = typer.Argument(..., help="Required argument"),
    optional_arg: str = typer.Option(None, help="Optional argument"),
):
    """Command description"""
    try:
        with Halo(text="Working...\n", spinner="dots", stream=sys.stderr) as spinner:
            result = self.do_work(required_arg, optional_arg)
            spinner.stop()

        console = Console()
        console.print(f"[green]Success:[/] {result}")

    except Exception as e:
        console.print(f"[red]Error:[/] {str(e)}")
        raise typer.Exit(1)
```

## Testing the Plugin

Create tests in `tests/plugins/`:

```python
# tests/plugins/test_my_plugin.py
from plugins.my_new_plugin import MyNewPlugin

def test_list_items(mock_client):
    plugin = MyNewPlugin(mock_client)
    result = plugin.list_items()
    assert result is not None
```

## Example Plugin

See `plugins/example_plugin.py` for a complete example:

```python
class ExamplePlugin(PluginBase):
    def setup(self) -> None:
        """Setup plugin commands"""
        self.register_command(
            name="hello",
            callback=self.hello_command,
            help="Say hello",
        )

        self.register_command(
            name="create-issue",
            callback=self.create_issue_command,
            help="Create Jira issue with GitHub info",
        )

    def hello_command(
        self,
        name: str = typer.Argument(..., help="Name to greet"),
        count: int = typer.Option(1, help="Number of times to greet"),
    ):
        """Example command that says hello"""
        for _ in range(count):
            typer.echo(f"Hello {name}!")

    def create_issue_command(
        self,
        repo: str = typer.Argument(..., help="GitHub repository"),
        project: str = typer.Option(..., help="Jira project key"),
        issue_type: str = typer.Option("Task", help="Jira issue type"),
    ):
        """Create Jira issue from GitHub info"""
        try:
            # Get GitHub repo info
            repo_info = self.client.github.get_repo(repo)

            # Create Jira issue
            issue = self.client.jira.create_issue(
                project_key=project,
                summary=f"GitHub: {repo_info['name']}",
                description=repo_info["description"],
                issue_type=issue_type,
            )

            typer.echo(f"Created Jira issue: {issue['key']}")
        except Exception as e:
            typer.echo(f"Error: {str(e)}", err=True)
```

## Available APIs

Your plugin has access to these APIs through `self.client`:

- `client.github` - GitHub operations
- `client.jira` - Jira issue tracking
- `client.confluence` - Confluence wiki
- `client.kubernetes` - Kubernetes clusters
- `client.docker` - Docker operations
- `client.ssh(hostname)` - SSH connections

## Best Practices

1. Use descriptive command names
2. Add help text to all commands/options
3. Handle errors gracefully
4. Show progress for long operations
5. Use Rich tables for structured output
6. Follow existing plugin patterns

For more examples, check:

- `plugins/resource_manager_plugin.py`
- `plugins/example_plugin.py`
