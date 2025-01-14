from typing import Dict, List, Optional

import typer

from apis.core.plugin_base import PluginBase


class PluginRegistry:
    """Plugin registry singleton"""

    _instance = None
    _app = None  # Store app at class level

    def __init__(self):
        """Initialize plugin registry"""
        self._plugins: Dict[str, PluginBase] = {}

    @classmethod
    def initialize(cls, app: typer.Typer) -> None:
        """
        Initialize the registry with a Typer app
        Args:
            app: Typer app instance
        """
        cls._app = app

    @classmethod
    def get_instance(cls) -> "PluginRegistry":
        """Get or create singleton instance"""
        if cls._instance is None:
            if cls._app is None:
                raise RuntimeError("PluginRegistry not initialized with app")
            cls._instance = cls()
        return cls._instance

    @property
    def app(self) -> typer.Typer:
        """Get the Typer app"""
        if self._app is None:
            raise RuntimeError("PluginRegistry not initialized with app")
        return self._app

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
        """Get plugin by name"""
        return self._plugins.get(name)

    def list_plugins(self) -> List[str]:
        """Get list of registered plugin names"""
        return list(self._plugins.keys())
