import sys

import typer
from halo import Halo
from rich.console import Console

from apis.core.client import get_client
from apis.docker_compose_api import DockerComposeApi

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


@app.command()
def stop():
    """Stop OpenSearch logging"""
    with Halo(
        text="Stopping OpenSearch logging\n", spinner="dots", stream=sys.stderr
    ) as spinner:
        compose = DockerComposeApi(
            project_dir="configs", compose_files=["docker-compose.yml"]
        )
        compose.down()
        spinner.stop()
