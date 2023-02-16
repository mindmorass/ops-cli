from typing import Optional

import typer
from rich.console import Console

from core.client import get_client

app = typer.Typer(no_args_is_help=True)
console = Console()


@app.command()
def create(space: str, title: str, content: str, parent: Optional[str] = None):
    """Create Confluence page"""
    client = get_client()
    try:
        page = client.confluence.create_page(
            space_key=space, title=title, body=content, parent_id=parent
        )
        console.print(f"[green]Created page:[/] {page['url']}")
    except Exception as e:
        console.print(f"[red]Error: {str(e)}[/]")
