from typing import Any, Dict, List

from pydantic import BaseModel, Field


class Environment(BaseModel):
    """Class representing an environment configuration."""

    name: str
    env_type: str = Field(
        ..., description="Type of the environment (e.g., production, staging)"
    )
    region: str
    cloud_provider: str


class EnvironmentTypeManager:
    """Class to manage dynamic environment types."""

    def __init__(self):
        self.environment_types: List[str] = []

    def add_environment_type(self, env_type: str) -> None:
        if env_type not in self.environment_types:
            self.environment_types.append(env_type)

    def remove_environment_type(self, env_type: str) -> None:
        if env_type in self.environment_types:
            self.environment_types.remove(env_type)

    def get_environment_types(self) -> List[str]:
        return self.environment_types
