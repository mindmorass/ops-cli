import typer
from rich.console import Console

from core.client import get_client
from core.plugin import PluginRegistry

app = typer.Typer(no_args_is_help=True)
console = Console()


@app.command()
def list_plugins():
    """List available plugins and their commands"""
    registry = PluginRegistry.get_instance()
    plugins = registry.list_plugins()

    if not plugins:
        console.print("[yellow]No plugins installed[/]")
        return

    console.print("\n[bold green]Available Plugins:[/]")
    for plugin_name in plugins:
        plugin = registry.get_plugin(plugin_name)
        console.print(f"\n[bold blue]{plugin_name}[/]")
        for cmd_name, cmd in plugin.commands.items():
            console.print(f"  [cyan]{cmd_name}[/]: {cmd.help}")


@app.command()
def list_extensions():
    """List loaded extensions"""
    client = get_client()
    try:
        extensions = client.list_extensions()

        if not extensions:
            console.print("[yellow]No extensions loaded[/]")
            return

        console.print("\n[bold green]Loaded Extensions:[/]")
        for ext in extensions:
            console.print(f"  [blue]{ext}[/]")
    except Exception as e:
        console.print(f"[red]Error: {str(e)}[/]")
