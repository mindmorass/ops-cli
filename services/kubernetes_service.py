from typing import Dict, List, Optional

from apis.kubernetes_api import KubernetesApi
from apis.opensearch_api import OpenSearchApi
from services.base_service import BaseService


class KubernetesService(BaseService):
    def __init__(
        self,
        context_name: Optional[str] = None,
        opensearch: Optional[OpenSearchApi] = None,
        log_index: str = "kubernetes-logs",
    ):
        """
        Initialize Kubernetes service
        Args:
            context_name: Kubernetes context name
            opensearch: OpenSearch API for logging
            log_index: Index name for logs
        """
        super().__init__(opensearch=opensearch, log_index=log_index)
        self.client = KubernetesApi(context_name=context_name)

    def scale_deployment(
        self,
        name: str,
        replicas: int,
        namespace: str = "default",
    ) -> Dict:
        """Scale deployment with logging"""
        try:
            result = self.client.scale_deployment(name, replicas, namespace)

            self.log_action(
                action="scale_deployment",
                status="success",
                service="kubernetes",
                details={
                    "deployment": name,
                    "namespace": namespace,
                    "replicas": replicas,
                },
            )
            return result
        except Exception as e:
            self.log_action(
                action="scale_deployment",
                status="failed",
                service="kubernetes",
                details={
                    "deployment": name,
                    "namespace": namespace,
                    "replicas": replicas,
                },
                error=str(e),
            )
            raise

    def create_deployment(
        self,
        name: str,
        image: str,
        replicas: int = 1,
        namespace: str = "default",
        labels: Optional[Dict[str, str]] = None,
    ) -> Dict:
        """Create deployment with logging"""
        try:
            result = self.client.create_deployment(
                name=name,
                image=image,
                replicas=replicas,
                namespace=namespace,
                labels=labels,
            )

            self.log_action(
                action="create_deployment",
                status="success",
                service="kubernetes",
                details={
                    "name": name,
                    "image": image,
                    "replicas": replicas,
                    "namespace": namespace,
                    "labels": labels,
                },
            )
            return result
        except Exception as e:
            self.log_action(
                action="create_deployment",
                status="failed",
                service="kubernetes",
                details={
                    "name": name,
                    "image": image,
                    "replicas": replicas,
                    "namespace": namespace,
                    "labels": labels,
                },
                error=str(e),
            )
            raise
