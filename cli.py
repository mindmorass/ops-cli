#!/usr/bin/env python

from pathlib import Path

import typer
from rich.console import Console

from apis.core.client import get_client
from apis.core.plugin import PluginRegistry
from commands import cli, confluence, deps, github, jira, logging

# Create main app and console
app = typer.Typer(no_args_is_help=True)
core_commands = typer.Typer(no_args_is_help=True)
console = Console()

# Initialize plugin registry with main app
# uncomment to enable plugins
# PluginRegistry.get_instance(app)

# core commands
core_commands.add_typer(github.app, name="github", help="Commands for GitHub")
core_commands.add_typer(jira.app, name="jira", help="Commands for Jira")
core_commands.add_typer(
    confluence.app, name="confluence", help="Commands for Confluence"
)
core_commands.add_typer(cli.app, name="client", help="Commands for Client")
core_commands.add_typer(logging.app, name="logging", help="Commands for Logging")
core_commands.add_typer(deps.app, name="deps", help="Commands for Dependencies")

# add core commands to main app
app.add_typer(core_commands, name="core", help="Commands for core APIs")


def main():
    """Main CLI entry point"""
    # Initialize client (which loads plugins)
    get_client()
    app()


if __name__ == "__main__":
    main()
