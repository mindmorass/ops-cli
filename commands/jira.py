import typer
from rich.console import Console

from core.client import get_client

app = typer.Typer(no_args_is_help=True)
console = Console()


@app.command()
def search(query: str):
    """Search Jira issues"""
    client = get_client()
    try:
        issues = client.jira.search_issues(query)
        for issue in issues:
            console.print(f"[blue]{issue['key']}[/]: {issue['summary']}")
    except Exception as e:
        console.print(f"[red]Error: {str(e)}[/]")
