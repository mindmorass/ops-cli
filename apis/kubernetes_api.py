from typing import Dict, List, Optional, Union

from kubernetes import client, config
from kubernetes.client import ApiClient
from kubernetes.client.rest import ApiException


class KubernetesApi:
    def __init__(self, context_name: Optional[str] = None):
        """
        Initialize Kubernetes API client
        Args:
            context_name: Optional Kubernetes context name
        """
        try:
            if context_name:
                config.load_kube_config(context=context_name)
            else:
                config.load_kube_config()

            self.core_v1 = client.CoreV1Api()
            self.apps_v1 = client.AppsV1Api()
            self.batch_v1 = client.BatchV1Api()
            self.networking_v1 = client.NetworkingV1Api()
            self.api_client = ApiClient()

        except Exception as e:
            raise Exception(f"Failed to initialize Kubernetes client: {str(e)}")

    # Pod operations
    def list_pods(
        self, namespace: str = "default", label_selector: Optional[str] = None
    ) -> List[Dict]:
        """List pods in namespace"""
        try:
            pods = self.core_v1.list_namespaced_pod(
                namespace=namespace, label_selector=label_selector
            )
            return self._format_k8s_list(pods)
        except ApiException as e:
            raise Exception(f"Failed to list pods: {str(e)}")

    def get_pod(self, name: str, namespace: str = "default") -> Dict:
        """Get pod details"""
        try:
            pod = self.core_v1.read_namespaced_pod(name=name, namespace=namespace)
            return self._format_k8s_obj(pod)
        except ApiException as e:
            raise Exception(f"Failed to get pod: {str(e)}")

    def create_pod(self, pod_manifest: Dict, namespace: str = "default") -> Dict:
        """Create pod from manifest"""
        try:
            pod = self.core_v1.create_namespaced_pod(
                namespace=namespace, body=pod_manifest
            )
            return self._format_k8s_obj(pod)
        except ApiException as e:
            raise Exception(f"Failed to create pod: {str(e)}")

    def delete_pod(self, name: str, namespace: str = "default") -> Dict:
        """Delete pod"""
        try:
            result = self.core_v1.delete_namespaced_pod(name=name, namespace=namespace)
            return self._format_k8s_obj(result)
        except ApiException as e:
            raise Exception(f"Failed to delete pod: {str(e)}")

    # Deployment operations
    def list_deployments(
        self, namespace: str = "default", label_selector: Optional[str] = None
    ) -> List[Dict]:
        """List deployments in namespace"""
        try:
            deployments = self.apps_v1.list_namespaced_deployment(
                namespace=namespace, label_selector=label_selector
            )
            return self._format_k8s_list(deployments)
        except ApiException as e:
            raise Exception(f"Failed to list deployments: {str(e)}")

    def get_deployment(self, name: str, namespace: str = "default") -> Dict:
        """Get deployment details"""
        try:
            deployment = self.apps_v1.read_namespaced_deployment(
                name=name, namespace=namespace
            )
            return self._format_k8s_obj(deployment)
        except ApiException as e:
            raise Exception(f"Failed to get deployment: {str(e)}")

    def create_deployment(
        self, deployment_manifest: Dict, namespace: str = "default"
    ) -> Dict:
        """Create deployment from manifest"""
        try:
            deployment = self.apps_v1.create_namespaced_deployment(
                namespace=namespace, body=deployment_manifest
            )
            return self._format_k8s_obj(deployment)
        except ApiException as e:
            raise Exception(f"Failed to create deployment: {str(e)}")

    def update_deployment(
        self, name: str, deployment_manifest: Dict, namespace: str = "default"
    ) -> Dict:
        """Update deployment"""
        try:
            deployment = self.apps_v1.patch_namespaced_deployment(
                name=name, namespace=namespace, body=deployment_manifest
            )
            return self._format_k8s_obj(deployment)
        except ApiException as e:
            raise Exception(f"Failed to update deployment: {str(e)}")

    def delete_deployment(self, name: str, namespace: str = "default") -> Dict:
        """Delete deployment"""
        try:
            result = self.apps_v1.delete_namespaced_deployment(
                name=name, namespace=namespace
            )
            return self._format_k8s_obj(result)
        except ApiException as e:
            raise Exception(f"Failed to delete deployment: {str(e)}")

    # Service operations
    def list_services(
        self, namespace: str = "default", label_selector: Optional[str] = None
    ) -> List[Dict]:
        """List services in namespace"""
        try:
            services = self.core_v1.list_namespaced_service(
                namespace=namespace, label_selector=label_selector
            )
            return self._format_k8s_list(services)
        except ApiException as e:
            raise Exception(f"Failed to list services: {str(e)}")

    def get_service(self, name: str, namespace: str = "default") -> Dict:
        """Get service details"""
        try:
            service = self.core_v1.read_namespaced_service(
                name=name, namespace=namespace
            )
            return self._format_k8s_obj(service)
        except ApiException as e:
            raise Exception(f"Failed to get service: {str(e)}")

    def create_service(
        self, service_manifest: Dict, namespace: str = "default"
    ) -> Dict:
        """Create service from manifest"""
        try:
            service = self.core_v1.create_namespaced_service(
                namespace=namespace, body=service_manifest
            )
            return self._format_k8s_obj(service)
        except ApiException as e:
            raise Exception(f"Failed to create service: {str(e)}")

    def delete_service(self, name: str, namespace: str = "default") -> Dict:
        """Delete service"""
        try:
            result = self.core_v1.delete_namespaced_service(
                name=name, namespace=namespace
            )
            return self._format_k8s_obj(result)
        except ApiException as e:
            raise Exception(f"Failed to delete service: {str(e)}")

    # ConfigMap operations
    def list_configmaps(
        self, namespace: str = "default", label_selector: Optional[str] = None
    ) -> List[Dict]:
        """List configmaps in namespace"""
        try:
            configmaps = self.core_v1.list_namespaced_config_map(
                namespace=namespace, label_selector=label_selector
            )
            return self._format_k8s_list(configmaps)
        except ApiException as e:
            raise Exception(f"Failed to list configmaps: {str(e)}")

    def get_configmap(self, name: str, namespace: str = "default") -> Dict:
        """Get configmap details"""
        try:
            configmap = self.core_v1.read_namespaced_config_map(
                name=name, namespace=namespace
            )
            return self._format_k8s_obj(configmap)
        except ApiException as e:
            raise Exception(f"Failed to get configmap: {str(e)}")

    def create_configmap(
        self, configmap_manifest: Dict, namespace: str = "default"
    ) -> Dict:
        """Create configmap from manifest"""
        try:
            configmap = self.core_v1.create_namespaced_config_map(
                namespace=namespace, body=configmap_manifest
            )
            return self._format_k8s_obj(configmap)
        except ApiException as e:
            raise Exception(f"Failed to create configmap: {str(e)}")

    def delete_configmap(self, name: str, namespace: str = "default") -> Dict:
        """Delete configmap"""
        try:
            result = self.core_v1.delete_namespaced_config_map(
                name=name, namespace=namespace
            )
            return self._format_k8s_obj(result)
        except ApiException as e:
            raise Exception(f"Failed to delete configmap: {str(e)}")

    # Secret operations
    def list_secrets(
        self, namespace: str = "default", label_selector: Optional[str] = None
    ) -> List[Dict]:
        """List secrets in namespace"""
        try:
            secrets = self.core_v1.list_namespaced_secret(
                namespace=namespace, label_selector=label_selector
            )
            return self._format_k8s_list(secrets)
        except ApiException as e:
            raise Exception(f"Failed to list secrets: {str(e)}")

    def get_secret(self, name: str, namespace: str = "default") -> Dict:
        """Get secret details"""
        try:
            secret = self.core_v1.read_namespaced_secret(name=name, namespace=namespace)
            return self._format_k8s_obj(secret)
        except ApiException as e:
            raise Exception(f"Failed to get secret: {str(e)}")

    def create_secret(self, secret_manifest: Dict, namespace: str = "default") -> Dict:
        """Create secret from manifest"""
        try:
            secret = self.core_v1.create_namespaced_secret(
                namespace=namespace, body=secret_manifest
            )
            return self._format_k8s_obj(secret)
        except ApiException as e:
            raise Exception(f"Failed to create secret: {str(e)}")

    def delete_secret(self, name: str, namespace: str = "default") -> Dict:
        """Delete secret"""
        try:
            result = self.core_v1.delete_namespaced_secret(
                name=name, namespace=namespace
            )
            return self._format_k8s_obj(result)
        except ApiException as e:
            raise Exception(f"Failed to delete secret: {str(e)}")

    def _format_k8s_obj(self, obj: object) -> Dict:
        """Convert Kubernetes object to dictionary"""
        return self.api_client.sanitize_for_serialization(obj)

    def _format_k8s_list(self, obj_list: object) -> List[Dict]:
        """Convert list of Kubernetes objects to list of dictionaries"""
        items = []
        for item in obj_list.items:
            items.append(self._format_k8s_obj(item))
        return items
