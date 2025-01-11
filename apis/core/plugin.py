from typing import Dict, List, Optional

import typer

from apis.core.plugin_base import PluginBase


class PluginRegistry:
    """Plugin registry"""

    _instance = None

    def __init__(self, app: Optional[typer.Typer] = None):
        if app:
            self.app = app
        self._plugins: Dict[str, PluginBase] = {}
        self._registered_names: set[str] = set()  # Track registered names

    @classmethod
    def get_instance(cls, app: Optional[typer.Typer] = None) -> "PluginRegistry":
        """Get or create singleton instance"""
        if cls._instance is None:
            cls._instance = cls(app)
        return cls._instance

    def register_plugin(self, plugin: PluginBase) -> None:
        """
        Register a plugin and its commands
        Args:
            plugin: Plugin instance
        """
        name = plugin.name
        if name in self._registered_names:  # Check if already registered
            return  # Silently skip instead of raising error

        self._registered_names.add(name)  # Track registration

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
        if self.app:
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
