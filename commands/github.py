import json
import sys

import typer
from halo import Halo
from rich.console import Console

from apis.core.client import get_client

app = typer.Typer(no_args_is_help=True)
console = Console()


@app.command()
def branches(repo: str):
    """List GitHub issues"""
    client = get_client()
    org_name = repo.split("/")[0]
    repo_name = repo.split("/")[1]
    try:
        with Halo(
            text="Listing GitHub branches\n", spinner="dots", stream=sys.stderr
        ) as spinner:
            branches = client.github.list_branches(
                org_name=org_name, repo_name=repo_name
            )
            spinner.stop()
            for branch in branches:
                console.print(json.dumps(branch, indent=4))
    except Exception as e:
        console.print(f"[red]Error: {str(e)}[/]")
