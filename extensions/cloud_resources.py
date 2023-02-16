from dataclasses import dataclass
from typing import Dict, List, Optional

from apis.client_api import Client


@dataclass
class ResourceInfo:
    name: str
    type: str
    status: str
    metadata: Dict


class CloudResourcesExtension:
    """Extension for managing cloud resources across different services"""

    def __init__(self, client: Client):
        self.client = client

    def get_kubernetes_resources(self, namespace: str) -> List[ResourceInfo]:
        """Get Kubernetes resources"""
        resources = []

        # Get deployments
        deployments = self.client.kubernetes.list_deployments(namespace)
        for dep in deployments:
            resources.append(
                ResourceInfo(
                    name=dep["name"],
                    type="deployment",
                    status=dep["status"],
                    metadata={
                        "replicas": dep["replicas"],
                        "image": dep["image"],
                    },
                )
            )

        return resources

    def get_github_resources(self, org: Optional[str] = None) -> List[ResourceInfo]:
        """Get GitHub resources"""
        resources = []

        # Get repositories
        if org:
            repos = self.client.github.get_org_repos(org)
        else:
            repos = self.client.github.get_user_repos()

        for repo in repos:
            resources.append(
                ResourceInfo(
                    name=repo["name"],
                    type="repository",
                    status="active" if not repo["private"] else "private",
                    metadata={
                        "description": repo["description"],
                        "stars": repo["stars"],
                        "forks": repo["forks"],
                    },
                )
            )

        return resources

    def get_all_resources(
        self,
        namespace: Optional[str] = None,
        org: Optional[str] = None,
    ) -> Dict[str, List[ResourceInfo]]:
        """Get all cloud resources"""
        resources = {}

        if namespace:
            resources["kubernetes"] = self.get_kubernetes_resources(namespace)

        resources["github"] = self.get_github_resources(org)

        return resources


def setup_extension(client: Client) -> CloudResourcesExtension:
    """Extension setup function"""
    return CloudResourcesExtension(client)
