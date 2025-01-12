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
        console.print(f"[red]Error: {str(e)}[/red]")


@app.command()
def issues(repo: str):
    """List GitHub issues"""
    client = get_client()
    org_name = repo.split("/")[0]
    repo_name = repo.split("/")[1]
    try:
        with Halo(
            text="Listing GitHub issues\n", spinner="dots", stream=sys.stderr
        ) as spinner:
            issues = client.github.get_repo_issues(
                org_name=org_name, repo_name=repo_name
            )
            spinner.stop()
            for issue in issues:
                console.print(json.dumps(issue, indent=4))
    except Exception as e:
        console.print(f"[red]Error: {str(e)}[/red]")


@app.command()
def list_open_prs(repo: str):
    """List GitHub pull requests"""
    client = get_client()
    org_name = repo.split("/")[0]
    repo_name = repo.split("/")[1]
    try:
        with Halo(
            text="Listing GitHub pull requests\n", spinner="dots", stream=sys.stderr
        ) as spinner:
            pull_requests = client.github.get_repo_pull_requests(
                org_name=org_name, repo_name=repo_name, state="open"
            )
            spinner.stop()
            for pull_request in pull_requests:
                console.print(json.dumps(pull_request, indent=4))
    except Exception as e:
        console.print(f"[red]Error: {str(e)}[/red]")


@app.command()
def create_pr(
    repo: str,
    title: str = typer.Option(..., help="PR title"),
    body: str = typer.Option(..., help="PR description"),
    head: str = typer.Option(..., help="Head branch"),
    base: str = typer.Option("main", help="Base branch"),
    draft: bool = typer.Option(False, help="Create as draft PR"),
):
    """Create a GitHub pull request"""
    client = get_client()
    org_name = repo.split("/")[0]
    repo_name = repo.split("/")[1]
    try:
        with Halo(
            text="Creating pull request\n", spinner="dots", stream=sys.stderr
        ) as spinner:
            pr = client.github.create_pull_request(
                org_name=org_name,
                repo_name=repo_name,
                title=title,
                body=body,
                head=head,
                base=base,
                draft=draft,
            )
            spinner.stop()
            console.print(json.dumps(pr, indent=4))
    except Exception as e:
        console.print(f"[red]Error: {str(e)}[/red]")


@app.command()
def get_file_contents(repo: str, file_path: str, branch: str = "main"):
    """Get the contents of a file in a GitHub repository"""
    client = get_client()
    org_name = repo.split("/")[0]
    repo_name = repo.split("/")[1]
    try:
        contents = client.github.get_file_contents(
            repo_name, file_path, branch, org_name
        )
        console.print(contents)
    except Exception as e:
        console.print(f"[red]Error: {str(e)}[/red]")
