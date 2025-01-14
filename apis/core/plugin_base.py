from typing import TYPE_CHECKING, Any, Callable, Dict

from .interfaces import ClientInterface

if TYPE_CHECKING:
    from apis.core.client import ClientApi


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

    def __init__(self, client: "ClientInterface"):
        self.client = client
        self._commands: Dict[str, PluginCommand] = {}

    @property
    def name(self) -> str:
        """Plugin name - derived from class name"""
        # Remove 'plugin' suffix if present and convert to lowercase
        name = self.__class__.__name__.lower()
        if name.endswith("plugin"):
            name = name[:-6]  # Remove 'plugin' suffix
        # Convert underscores to dashes for command naming
        return name.replace("_", "-")

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
