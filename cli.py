#!/usr/bin/env python
import os
from pathlib import Path

import typer
from rich.console import Console

from apis.core.client import get_client
from apis.core.plugin import PluginRegistry
from commands import cli, confluence, deps, github, jira, logging

# generic client for reuse but allow for a specific client name
client_name = os.environ.get("OPS_CLI_CLIENT_NAME", "ops-cli")

# load config file if it exists name matches client_name
config_filename = os.path.expanduser(f"~/.config/{client_name}/config.yaml")

if os.path.exists(config_filename):
    client = get_client()
    client.load_config(config_filename)
    client.export_config()

# Create main app and console
app = typer.Typer(no_args_is_help=True)
core_commands = typer.Typer(no_args_is_help=True)
console = Console()

# Initialize plugin registry with plugins app
plugins_app = typer.Typer(name="plugins", help="Plugin commands", no_args_is_help=True)
app.add_typer(plugins_app)

# Initialize plugin registry
PluginRegistry.initialize(plugins_app)

# Core commands
core_commands.add_typer(github.app, name="github")
core_commands.add_typer(jira.app, name="jira")
core_commands.add_typer(confluence.app, name="confluence")
core_commands.add_typer(cli.app, name="client")
core_commands.add_typer(logging.app, name="logging")
core_commands.add_typer(deps.app, name="deps", help="Commands for Dependencies")

# Add core commands to main app
app.add_typer(core_commands, name="core", help="Commands for core APIs")


def main():
    """Main CLI entry point"""
    # Initialize client (which loads plugins)
    client = get_client()
    app()


if __name__ == "__main__":
    main()
