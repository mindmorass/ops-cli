import os
import sys

import typer
from halo import Halo
from rich.console import Console

from apis.docker_compose_api import DockerComposeApi
from apis.opensearch_api import OpenSearchApi

app = typer.Typer(no_args_is_help=True)
console = Console()


@app.command()
def start():
    """Start OpenSearch logging"""
    with Halo(
        text="Starting OpenSearch logging\n", spinner="dots", stream=sys.stderr
    ) as spinner:
        compose = DockerComposeApi(
            project_dir="configs", compose_files=["docker-compose.yml"]
        )
        compose.up()
        spinner.stop()

        console.print(
            f"[green]OpenSearch logging started with password: [/green]{os.getenv('OPENSEARCH_INITIAL_ADMIN_PASSWORD')}"
        )


@app.command()
def stop():
    """Stop OpenSearch logging"""
    with Halo(
        text="Stopping OpenSearch logging\n", spinner="dots", stream=sys.stderr
    ) as spinner:
        compose = DockerComposeApi(
            project_dir="configs", compose_files=["docker-compose.yml"]
        )
        compose.down(volumes=True)
        spinner.stop()

        console.print("[yellow]OpenSearch logging stopped[/yellow]")


@app.command()
def send():
    """Send a message to OpenSearch logging"""
    opensearch = OpenSearchApi(
        hosts=["http://localhost:9200"],
        username="admin",
        password=os.getenv("OPENSEARCH_INITIAL_ADMIN_PASSWORD"),
        verify_certs=False,
    )
    with Halo(
        text="Sending message to OpenSearch logging\n",
        spinner="dots",
        stream=sys.stderr,
    ) as spinner:
        opensearch.index_document(
            index_name="record",
            document={"message": "Hello, world!"},
        )
        spinner.stop()

        console.print("[green]Message sent to OpenSearch logging[/green]")
