import json
import sys

import typer
from halo import Halo
from rich.console import Console

from apis.core.client import get_client

app = typer.Typer(no_args_is_help=True)
console = Console()


@app.command()
def jql(query: str):
    """Search Jira issues using JQL"""
    client = get_client()
    try:
        with Halo(
            text="Searching Jira issues\n", spinner="dots", stream=sys.stderr
        ) as spinner:
            issues = client.jira.search_issues(jql=f"{query}")
            spinner.stop()
        print(json.dumps(issues, indent=4))
    except Exception as e:
        console.print(f"[red]Error: {str(e)}[/]")
