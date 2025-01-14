"""Client initialization and management"""

import os
from typing import Optional

from apis.client_api import ClientApi, ClientConfig

_client: Optional[ClientApi] = None


def get_client() -> ClientApi:
    """Get or create client instance"""
    global _client

    if _client is None:
        # Load config if exists
        config = None
        config_file = os.environ.get("OPS_CLI_CONFIG")
        if config_file and os.path.exists(config_file):
            config = ClientConfig()
            config.load_config(config_file)

        _client = ClientApi(config)

    return _client
