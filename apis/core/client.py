import os
from typing import Optional

from apis.client_api import ClientApi, ClientConfig


def get_client() -> ClientApi:
    """Get or create client instance"""
    global _client, _plugins_loaded
    if _client is None:
        # Load config from environment
        config = ClientConfig(
            github_token=os.getenv("GITHUB_TOKEN"),
            jira_url=os.getenv("JIRA_URL"),
            jira_username=os.getenv("JIRA_USERNAME"),
            jira_token=os.getenv("JIRA_TOKEN"),
            confluence_url=os.getenv("CONFLUENCE_URL"),
            confluence_username=os.getenv("CONFLUENCE_USERNAME"),
            confluence_token=os.getenv("CONFLUENCE_TOKEN"),
            kube_config_path=os.getenv("KUBE_CONFIG_PATH"),
            kube_context=os.getenv("KUBE_CONTEXT"),
        )
        _client = ClientApi(config)

        # Load plugins only once
        if not _plugins_loaded:
            _client._load_plugins()
            _plugins_loaded = True

    return _client


# Initialize globals
_client: Optional[ClientApi] = None
_plugins_loaded: bool = False
