import json

import typer
from rich.console import Console

from core.client import get_client

app = typer.Typer(no_args_is_help=True)
console = Console()


def create_issue(title: str, description: str, project: str):
    """Create Jira issue"""
    client = get_client()
    try:
        issue = client.jira.create_issue(title, description, project)
        console.print(f"[green]Created issue:[/] {issue['key']}")
    except Exception as e:
        console.print(f"[red]Error: {str(e)}[/]")


@app.command()
def jql(query: str):
    """Search Jira issues using JQL"""
    client = get_client()
    try:
        issues = client.jira.search_issues(jql=f"{query}")
        print(json.dumps(issues, indent=4))
    except Exception as e:
        console.print(f"[red]Error: {str(e)}[/]")
