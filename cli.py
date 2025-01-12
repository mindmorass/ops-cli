#!/usr/bin/env python

from pathlib import Path

import typer
from rich.console import Console

from apis.core.client import get_client
from apis.core.plugin import PluginRegistry
from commands import cli, confluence, github, jira

# Create main app and console
app = typer.Typer(no_args_is_help=True)
core_commands = typer.Typer(no_args_is_help=True)
console = Console()

# Initialize plugin registry with main app
PluginRegistry.get_instance(app)


# Add command groups
core_commands.add_typer(github.app, name="github")
core_commands.add_typer(jira.app, name="jira")
core_commands.add_typer(confluence.app, name="confluence")
core_commands.add_typer(cli.app, name="client")
app.add_typer(core_commands, name="core", help="Commands for core APIs")


def main():
    """Main CLI entry point"""
    # Initialize client (which loads plugins)
    get_client()
    app()


if __name__ == "__main__":
    main()
