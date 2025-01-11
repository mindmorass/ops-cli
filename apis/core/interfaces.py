from typing import Optional, Protocol


class ClientInterface(Protocol):
    """Interface for client API"""

    @property
    def github(self): ...

    @property
    def jira(self): ...

    @property
    def confluence(self): ...

    @property
    def kubernetes(self): ...

    @property
    def docker(self): ...

    def ssh(self, hostname: str, **kwargs): ...
