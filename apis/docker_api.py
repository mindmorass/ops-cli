from typing import Dict, List, Optional

import docker


class DockerApi:
    def __init__(self):
        """Initialize Docker client"""
        self.client = docker.from_env()

    def create_container(
        self,
        image: str,
        name: Optional[str] = None,
        ports: Optional[Dict[str, str]] = None,
    ) -> Dict:
        """
        Create a new container
        Args:
            image: Docker image name
            name: Optional container name
            ports: Optional port mappings (e.g., {'8080/tcp': 8080})
        """
        try:
            container = self.client.containers.create(
                image=image, name=name, ports=ports, detach=True
            )
            return {
                "id": container.id,
                "name": container.name,
                "status": container.status,
            }
        except docker.errors.ImageNotFound:
            raise Exception(f"Image {image} not found")
        except docker.errors.APIError as e:
            raise Exception(f"Error creating container: {str(e)}")

    def list_containers(self, show_all: bool = True) -> List[Dict]:
        """
        List all containers
        Args:
            show_all: If True, show all containers (including stopped ones)
        """
        containers = self.client.containers.list(all=show_all)
        return [
            {
                "id": container.id,
                "name": container.name,
                "status": container.status,
                "image": container.image.tags[0] if container.image.tags else "none",
                "ports": container.ports,
            }
            for container in containers
        ]

    def start_container(self, container_id: str) -> Dict:
        """Start a container by ID"""
        try:
            container = self.client.containers.get(container_id)
            container.start()
            return {"id": container.id, "name": container.name, "status": "running"}
        except docker.errors.NotFound:
            raise Exception(f"Container {container_id} not found")

    def stop_container(self, container_id: str) -> Dict:
        """Stop a container by ID"""
        try:
            container = self.client.containers.get(container_id)
            container.stop()
            return {"id": container.id, "name": container.name, "status": "stopped"}
        except docker.errors.NotFound:
            raise Exception(f"Container {container_id} not found")

    def delete_container(self, container_id: str, force: bool = False) -> Dict:
        """
        Delete a container by ID
        Args:
            container_id: Container ID or name
            force: Force remove running container
        """
        try:
            container = self.client.containers.get(container_id)
            container.remove(force=force)
            return {"message": f"Container {container_id} deleted successfully"}
        except docker.errors.NotFound:
            raise Exception(f"Container {container_id} not found")

    def get_container_logs(self, container_id: str, lines: int = 100) -> str:
        """Get container logs"""
        try:
            container = self.client.containers.get(container_id)
            return container.logs(tail=lines).decode("utf-8")
        except docker.errors.NotFound:
            raise Exception(f"Container {container_id} not found")

    def get_container_stats(self, container_id: str) -> Dict:
        """Get container statistics"""
        try:
            container = self.client.containers.get(container_id)
            stats = container.stats(stream=False)
            return {
                "cpu_usage": stats["cpu_stats"]["cpu_usage"]["total_usage"],
                "memory_usage": stats["memory_stats"].get("usage", 0),
                "network_rx": (
                    stats["networks"]["eth0"]["rx_bytes"] if "networks" in stats else 0
                ),
                "network_tx": (
                    stats["networks"]["eth0"]["tx_bytes"] if "networks" in stats else 0
                ),
            }
        except docker.errors.NotFound:
            raise Exception(f"Container {container_id} not found")
