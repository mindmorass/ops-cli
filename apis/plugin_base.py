from typing import Any, Callable, Dict, List, Optional

import typer

from apis.client_api import ClientApi


class PluginCommand:
    """Represents a plugin command"""

    def __init__(
        self, name: str, callback: Callable, help: str = "", **typer_options: Any
    ):
        """
        Initialize plugin command
        Args:
            name: Command name
            callback: Command callback function
            help: Command help text
            typer_options: Additional Typer command options
        """
        self.name = name
        self.callback = callback
        self.help = help
        self.options = typer_options


class PluginBase:
    """Base class for plugins"""

    def __init__(self, client: ClientApi):
        self.client = client
        self._commands: Dict[str, PluginCommand] = {}

    @property
    def name(self) -> str:
        """Plugin name"""
        return self.__class__.__name__.lower().replace("plugin", "")

    @property
    def commands(self) -> Dict[str, PluginCommand]:
        """Registered commands"""
        return self._commands

    def register_command(
        self, name: str, callback: Callable, help: str = "", **typer_options: Any
    ) -> None:
        """
        Register a command
        Args:
            name: Command name
            callback: Command callback function
            help: Command help text
            typer_options: Additional Typer command options
        """
        self._commands[name] = PluginCommand(
            name=name, callback=callback, help=help, **typer_options
        )

    def setup(self) -> None:
        """Setup plugin - override in subclass"""
        pass


class PluginRegistry:
    """Plugin registry"""

    def __init__(self, app: typer.Typer):
        self.app = app
        self._plugins: Dict[str, PluginBase] = {}

    def register_plugin(self, plugin: PluginBase) -> None:
        """
        Register a plugin and its commands
        Args:
            plugin: Plugin instance
        """
        name = plugin.name
        if name in self._plugins:
            raise ValueError(f"Plugin {name} already registered")

        # Create plugin command group
        plugin_app = typer.Typer(
            name=name,
            help=f"Commands for {name} plugin",
        )

        # Register plugin commands
        for cmd in plugin.commands.values():
            plugin_app.command(name=cmd.name, help=cmd.help, **cmd.options)(
                cmd.callback
            )

        # Add plugin command group to main app
        self.app.add_typer(plugin_app, name=name)
        self._plugins[name] = plugin

    def get_plugin(self, name: str) -> Optional[PluginBase]:
        """
        Get plugin by name
        Args:
            name: Plugin name
        """
        return self._plugins.get(name)

    def list_plugins(self) -> List[str]:
        """Get list of registered plugin names"""
        return list(self._plugins.keys())
