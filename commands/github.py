import typer
from rich.console import Console

from core.client import get_client

app = typer.Typer(no_args_is_help=True)
console = Console()


@app.command()
def branches(repo: str):
    """List GitHub issues"""
    client = get_client()
    org_name = repo.split("/")[0]
    repo_name = repo.split("/")[1]
    try:
        branches = client.github.list_branches(org_name=org_name, repo_name=repo_name)
        for branch in branches:
            console.print(f"[green]{branch}[/green]")
    except Exception as e:
        console.print(f"[red]Error: {str(e)}[/]")
