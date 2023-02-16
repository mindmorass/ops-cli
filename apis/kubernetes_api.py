from typing import Dict, List, Optional

from kubernetes import client, config
from kubernetes.client.rest import ApiException
from kubernetes.config.kube_config import list_kube_config_contexts, load_kube_config


class KubernetesApi:
    def __init__(self, context_name: Optional[str] = None):
        """
        Initialize Kubernetes API client
        Args:
            context_name: Optional Kubernetes context name
        """
        if context_name:
            load_kube_config(context=context_name)
        else:
            load_kube_config()

        self.apps_v1 = client.AppsV1Api()
        self.core_v1 = client.CoreV1Api()

    @staticmethod
    def list_contexts() -> List[Dict]:
        """List available Kubernetes contexts from kubeconfig"""
        try:
            contexts, active_context = list_kube_config_contexts()
            return [
                {
                    "name": ctx["name"],
                    "cluster": ctx["context"]["cluster"],
                    "user": ctx["context"]["user"],
                    "is_active": ctx == active_context,
                }
                for ctx in contexts
            ]
        except Exception as e:
            raise Exception(f"Error listing contexts: {str(e)}")

    def list_pods(self, namespace: str = "default") -> List[Dict]:
        """List pods in a namespace"""
        try:
            pods = self.core_v1.list_namespaced_pod(namespace)
            return [
                {
                    "name": pod.metadata.name,
                    "namespace": pod.metadata.namespace,
                    "status": pod.status.phase,
                    "ip": pod.status.pod_ip,
                    "node": pod.spec.node_name,
                    "containers": [
                        {
                            "name": container.name,
                            "image": container.image,
                            "ready": any(
                                status.ready
                                for status in pod.status.container_statuses
                                if status.name == container.name
                            ),
                        }
                        for container in pod.spec.containers
                    ],
                }
                for pod in pods.items
            ]
        except ApiException as e:
            raise Exception(f"Error listing pods: {str(e)}")

    def delete_pod(self, name: str, namespace: str = "default") -> Dict:
        """Delete a pod"""
        try:
            self.core_v1.delete_namespaced_pod(name, namespace)
            return {"message": f"Pod {name} deleted successfully"}
        except ApiException as e:
            raise Exception(f"Error deleting pod: {str(e)}")

    def list_deployments(self, namespace: str = "default") -> List[Dict]:
        """List deployments in a namespace"""
        try:
            deployments = self.apps_v1.list_namespaced_deployment(namespace)
            return [
                {
                    "name": dep.metadata.name,
                    "namespace": dep.metadata.namespace,
                    "replicas": dep.spec.replicas,
                    "available": dep.status.available_replicas,
                    "ready": dep.status.ready_replicas,
                    "updated": dep.status.updated_replicas,
                    "image": dep.spec.template.spec.containers[0].image,
                }
                for dep in deployments.items
            ]
        except ApiException as e:
            raise Exception(f"Error listing deployments: {str(e)}")

    def create_deployment(
        self,
        name: str,
        image: str,
        namespace: str = "default",
        replicas: int = 1,
        labels: Optional[Dict[str, str]] = None,
    ) -> Dict:
        """Create a deployment"""
        try:
            labels = labels or {"app": name}
            deployment = client.V1Deployment(
                metadata=client.V1ObjectMeta(name=name, labels=labels),
                spec=client.V1DeploymentSpec(
                    replicas=replicas,
                    selector=client.V1LabelSelector(match_labels=labels),
                    template=client.V1PodTemplateSpec(
                        metadata=client.V1ObjectMeta(labels=labels),
                        spec=client.V1PodSpec(
                            containers=[
                                client.V1Container(
                                    name=name,
                                    image=image,
                                )
                            ]
                        ),
                    ),
                ),
            )

            self.apps_v1.create_namespaced_deployment(namespace, deployment)
            return {"message": f"Deployment {name} created successfully"}
        except ApiException as e:
            raise Exception(f"Error creating deployment: {str(e)}")

    def delete_deployment(self, name: str, namespace: str = "default") -> Dict:
        """Delete a deployment"""
        try:
            self.apps_v1.delete_namespaced_deployment(name, namespace)
            return {"message": f"Deployment {name} deleted successfully"}
        except ApiException as e:
            raise Exception(f"Error deleting deployment: {str(e)}")

    def scale_deployment(
        self, name: str, replicas: int, namespace: str = "default"
    ) -> Dict:
        """Scale a deployment"""
        try:
            scale = client.V1Scale(
                metadata=client.V1ObjectMeta(name=name, namespace=namespace),
                spec=client.V1ScaleSpec(replicas=replicas),
            )
            self.apps_v1.patch_namespaced_deployment_scale(name, namespace, scale)
            return {"message": f"Deployment {name} scaled to {replicas} replicas"}
        except ApiException as e:
            raise Exception(f"Error scaling deployment: {str(e)}")

    def list_statefulsets(self, namespace: str = "default") -> List[Dict]:
        """List StatefulSets in a namespace"""
        try:
            statefulsets = self.apps_v1.list_namespaced_stateful_set(namespace)
            return [
                {
                    "name": sts.metadata.name,
                    "namespace": sts.metadata.namespace,
                    "replicas": sts.spec.replicas,
                    "ready_replicas": sts.status.ready_replicas,
                    "current_replicas": sts.status.current_replicas,
                    "updated_replicas": sts.status.updated_replicas,
                    "image": sts.spec.template.spec.containers[0].image,
                }
                for sts in statefulsets.items
            ]
        except ApiException as e:
            raise Exception(f"Error listing StatefulSets: {str(e)}")

    def create_statefulset(
        self,
        name: str,
        image: str,
        namespace: str = "default",
        replicas: int = 1,
        labels: Optional[Dict[str, str]] = None,
    ) -> Dict:
        """Create a StatefulSet"""
        try:
            labels = labels or {"app": name}
            statefulset = client.V1StatefulSet(
                metadata=client.V1ObjectMeta(name=name, labels=labels),
                spec=client.V1StatefulSetSpec(
                    replicas=replicas,
                    selector=client.V1LabelSelector(match_labels=labels),
                    service_name=name,
                    template=client.V1PodTemplateSpec(
                        metadata=client.V1ObjectMeta(labels=labels),
                        spec=client.V1PodSpec(
                            containers=[
                                client.V1Container(
                                    name=name,
                                    image=image,
                                )
                            ]
                        ),
                    ),
                ),
            )

            self.apps_v1.create_namespaced_stateful_set(namespace, statefulset)
            return {"message": f"StatefulSet {name} created successfully"}
        except ApiException as e:
            raise Exception(f"Error creating StatefulSet: {str(e)}")

    def delete_statefulset(self, name: str, namespace: str = "default") -> Dict:
        """Delete a StatefulSet"""
        try:
            self.apps_v1.delete_namespaced_stateful_set(name, namespace)
            return {"message": f"StatefulSet {name} deleted successfully"}
        except ApiException as e:
            raise Exception(f"Error deleting StatefulSet: {str(e)}")

    def scale_statefulset(
        self, name: str, replicas: int, namespace: str = "default"
    ) -> Dict:
        """Scale a StatefulSet"""
        try:
            scale = client.V1Scale(
                metadata=client.V1ObjectMeta(name=name, namespace=namespace),
                spec=client.V1ScaleSpec(replicas=replicas),
            )
            self.apps_v1.patch_namespaced_stateful_set_scale(name, namespace, scale)
            return {"message": f"StatefulSet {name} scaled to {replicas} replicas"}
        except ApiException as e:
            raise Exception(f"Error scaling StatefulSet: {str(e)}")

    def list_daemonsets(self, namespace: str = "default") -> List[Dict]:
        """List DaemonSets in a namespace"""
        try:
            daemonsets = self.apps_v1.list_namespaced_daemon_set(namespace)
            return [
                {
                    "name": ds.metadata.name,
                    "namespace": ds.metadata.namespace,
                    "desired_number": ds.status.desired_number_scheduled,
                    "current_number": ds.status.current_number_scheduled,
                    "ready_number": ds.status.number_ready,
                    "available_number": ds.status.number_available,
                    "image": ds.spec.template.spec.containers[0].image,
                }
                for ds in daemonsets.items
            ]
        except ApiException as e:
            raise Exception(f"Error listing DaemonSets: {str(e)}")

    def create_daemonset(
        self,
        name: str,
        image: str,
        namespace: str = "default",
        labels: Optional[Dict[str, str]] = None,
    ) -> Dict:
        """Create a DaemonSet"""
        try:
            labels = labels or {"app": name}
            daemonset = client.V1DaemonSet(
                metadata=client.V1ObjectMeta(name=name, labels=labels),
                spec=client.V1DaemonSetSpec(
                    selector=client.V1LabelSelector(match_labels=labels),
                    template=client.V1PodTemplateSpec(
                        metadata=client.V1ObjectMeta(labels=labels),
                        spec=client.V1PodSpec(
                            containers=[
                                client.V1Container(
                                    name=name,
                                    image=image,
                                )
                            ]
                        ),
                    ),
                ),
            )

            self.apps_v1.create_namespaced_daemon_set(namespace, daemonset)
            return {"message": f"DaemonSet {name} created successfully"}
        except ApiException as e:
            raise Exception(f"Error creating DaemonSet: {str(e)}")

    def delete_daemonset(self, name: str, namespace: str = "default") -> Dict:
        """Delete a DaemonSet"""
        try:
            self.apps_v1.delete_namespaced_daemon_set(name, namespace)
            return {"message": f"DaemonSet {name} deleted successfully"}
        except ApiException as e:
            raise Exception(f"Error deleting DaemonSet: {str(e)}")
