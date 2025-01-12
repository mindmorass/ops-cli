import json

import typer
from halo import Halo
from rich.console import Console

from apis.core.client import get_client
from apis.core.plugin import PluginRegistry
from apis.dependency_api import DependencyApi

app = typer.Typer(no_args_is_help=True)
console = Console()


@app.command()
def list():
    """list dependencies"""
    dependency_api = DependencyApi()
    with Halo(text="Checking dependencies", spinner="dots") as spinner:
        dep_list = dependency_api.check_dependencies()
        spinner.stop()
    console.print(json.dumps(dep_list, indent=4))


@app.command()
def install():
    """install dependencies"""
    dependency_api = DependencyApi()
    with Halo(text="Checking dependencies", spinner="dots") as spinner:
        dep_list = dependency_api.check_dependencies()
        spinner.stop()

    for dep in dep_list:
        with Halo(text=f"Installing {dep['name']}", spinner="dots") as spinner:
            if not dep["installed"]:
                dependency_api.install_brew_package(dep["name"])
                spinner.succeed(f"{dep['name']} installed")
