#!/usr/bin/env python

from pathlib import Path

import typer
from rich.console import Console

from commands import cli, confluence, github, jira
from core.client import get_client
from core.plugin import PluginRegistry

# Create main app and console
app = typer.Typer(no_args_is_help=True)
console = Console()

# Initialize plugin registry with main app
PluginRegistry.get_instance(app)

# Add command groups
app.add_typer(github.app, name="github")
app.add_typer(jira.app, name="jira")
app.add_typer(confluence.app, name="confluence")
app.add_typer(cli.app, name="client")


def main():
    """Main CLI entry point"""
    # Initialize client (which loads plugins)
    get_client()
    app()


if __name__ == "__main__":
    main()
