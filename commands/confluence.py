import sys
from typing import Optional

import typer
from halo import Halo
from rich.console import Console

from apis.core.client import get_client

app = typer.Typer(no_args_is_help=True)
console = Console()


@app.command()
def create(space: str, title: str, content: str, parent: Optional[str] = None):
    """Create Confluence page"""
    client = get_client()
    try:
        with Halo(
            text="Creating Confluence page\n", spinner="dots", stream=sys.stderr
        ) as spinner:
            page = client.confluence.create_page(
                space_key=space, title=title, body=content, parent_id=parent
            )
            spinner.stop()
        console.print(f"[green]Created page:[/green] {page['url']}")
    except Exception as e:
        console.print(f"[red]Error: {str(e)}[/red]")


@app.command()
def export_pdf(
    page_id: str,
    output: Optional[str] = typer.Option(None, help="Output file path"),
    expand: bool = typer.Option(True, help="Include expanded content"),
):
    """Export Confluence page as PDF by page id"""
    client = get_client()
    try:
        with Halo(
            text="Exporting page as PDF\n", spinner="dots", stream=sys.stderr
        ) as spinner:
            pdf_path = client.confluence.export_page_as_pdf(
                page_id=page_id,
                output_path=output,
            )
            spinner.stop()
        console.print(f"[green]PDF exported to:[/green] {pdf_path}")
    except Exception as e:
        console.print(f"[red]Error: {str(e)}[/red]")
