"""
Client API
This module provides a client API for interacting with various APIs.
"""

import importlib
import pkgutil
from pathlib import Path
from typing import Dict, List, Optional

from pydantic import BaseModel

from apis.confluence_api import ConfluenceApi
from apis.core.interfaces import ClientInterface
from apis.core.plugin import PluginRegistry
from apis.docker_api import DockerApi
from apis.github_api import GithubApi
from apis.jira_api import JiraApi
from apis.kubernetes_api import KubernetesApi
from apis.ssh_api import SSHApi


class ClientConfig(BaseModel):
    """Configuration model for client APIs"""

    # GitHub configuration
    github_token: Optional[str] = None

    # Jira configuration
    jira_url: Optional[str] = None
    jira_username: Optional[str] = None
    jira_token: Optional[str] = None

    # Confluence configuration
    confluence_url: Optional[str] = None
    confluence_username: Optional[str] = None
    confluence_token: Optional[str] = None

    # Google configuration
    google_credentials_file: Optional[str] = None

    # Kubernetes configuration
    kube_config_path: Optional[str] = None
    kube_context: Optional[str] = None


class ClientApi(ClientInterface):
    """
    Client API wrapper
    """

    def __init__(self, config: Optional[ClientConfig] = None):
        """
        Initialize client API wrapper
        Args:
            config: Optional configuration object
        """
        self.config = config or ClientConfig()
        self._apis: Dict[str, object] = {}
        self._extensions: Dict[str, object] = {}
        self._load_plugins()
        self._load_extensions()

    def register_extension(self, name: str, extension: object) -> None:
        """
        Register a new extension
        Args:
            name: Extension name
            extension: Extension instance
        """
        if name in self._extensions:
            raise ValueError(f"Extension {name} already registered")
        self._extensions[name] = extension

    def get_extension(self, name: str) -> Optional[object]:
        """
        Get a registered extension
        Args:
            name: Extension name
        """
        return self._extensions.get(name)

    def list_extensions(self) -> List[str]:
        """Get list of registered extension names"""
        return list(self._extensions.keys())

    def _load_extensions(self) -> None:
        """Load extensions from the extensions directory"""
        extensions_dir = Path(__file__).parent.parent / "extensions"
        if not extensions_dir.exists():
            return

        for finder, name, _ in pkgutil.iter_modules([str(extensions_dir)]):
            try:
                module = importlib.import_module(f"extensions.{name}")
                if hasattr(module, "setup_extension"):
                    extension = module.setup_extension(self)
                    self.register_extension(name, extension)
            except Exception as e:
                print(f"Failed to load extension {name}: {str(e)}")

    @property
    def github(self) -> GithubApi:
        """Get GitHub API client"""
        if "github" not in self._apis:
            if not self.config.github_token:
                raise ValueError("GitHub token not configured")
            self._apis["github"] = GithubApi(token=self.config.github_token)
        return self._apis["github"]

    @property
    def jira(self) -> JiraApi:
        """Get Jira API client"""
        if "jira" not in self._apis:
            if not all(
                [
                    self.config.jira_url,
                    self.config.jira_username,
                    self.config.jira_token,
                ]
            ):
                raise ValueError("Jira configuration incomplete")
            self._apis["jira"] = JiraApi(
                server_url=self.config.jira_url,
                username=self.config.jira_username,
                token=self.config.jira_token,
            )
        return self._apis["jira"]

    @property
    def confluence(self) -> ConfluenceApi:
        """Get Confluence API client"""
        if "confluence" not in self._apis:
            if not all(
                [
                    self.config.confluence_url,
                    self.config.confluence_username,
                    self.config.confluence_token,
                ]
            ):
                raise ValueError("Confluence configuration incomplete")
            self._apis["confluence"] = ConfluenceApi(
                url=self.config.confluence_url,
                username=self.config.confluence_username,
                api_token=self.config.confluence_token,
                cloud=True,
            )
        return self._apis["confluence"]

    @property
    def docker(self) -> DockerApi:
        """Get Docker API client"""
        if "docker" not in self._apis:
            self._apis["docker"] = DockerApi()
        return self._apis["docker"]

    @property
    def kubernetes(self) -> KubernetesApi:
        """Get Kubernetes API client"""
        if "kubernetes" not in self._apis:
            self._apis["kubernetes"] = KubernetesApi(
                context_name=self.config.kube_context
            )
        return self._apis["kubernetes"]

    def ssh(self, hostname: str, **kwargs) -> SSHApi:
        """
        Get SSH API client
        Args:
            hostname: Remote host to connect to
            **kwargs: Additional SSH configuration
        """
        return SSHApi(hostname=hostname, **kwargs)

    def _load_plugins(self) -> None:
        """Load plugins from the plugins directory"""
        plugins_dir = Path(__file__).parent.parent / "plugins"
        if not plugins_dir.exists():
            return

        registry = PluginRegistry.get_instance()
        for finder, name, _ in pkgutil.iter_modules([str(plugins_dir)]):
            try:
                module = importlib.import_module(f"plugins.{name}")
                if hasattr(module, "setup_plugin"):
                    plugin = module.setup_plugin(self)
                    registry.register_plugin(plugin)
            except Exception as e:
                print(f"Failed to load plugin {name}: {str(e)}")
