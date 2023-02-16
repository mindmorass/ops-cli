import os
from pathlib import Path
from typing import Dict, List, Optional, Union

from compose.cli.main import project_from_options
from compose.project import Project
from compose.service import ImageType


class DockerComposeApi:
    def __init__(
        self,
        project_dir: Union[str, Path],
        compose_files: Optional[List[str]] = None,
        project_name: Optional[str] = None,
    ):
        """
        Initialize Docker Compose API
        Args:
            project_dir: Directory containing compose file(s)
            compose_files: List of compose file names (default: docker-compose.yml)
            project_name: Project name (default: directory name)
        """
        self.project_dir = str(project_dir)
        self.compose_files = compose_files or ["docker-compose.yml"]
        self.project_name = project_name

        # Build options dictionary for compose
        self.options = {
            "--file": self.compose_files,
            "--project-name": self.project_name,
            "--project-directory": self.project_dir,
        }

        # Initialize project
        self.project = project_from_options(self.project_dir, self.options)

    def up(
        self,
        services: Optional[List[str]] = None,
        detach: bool = True,
        scale: Optional[Dict[str, int]] = None,
    ) -> Dict:
        """
        Start services
        Args:
            services: List of service names (default: all services)
            detach: Run in background
            scale: Dictionary of service names and their replica count
        """
        try:
            self.project.up(
                service_names=services,
                detach=detach,
                scale=scale,
            )
            return {
                "status": "success",
                "services": services or self.list_services(),
                "scale": scale,
            }
        except Exception as e:
            raise Exception(f"Failed to start services: {str(e)}")

    def down(
        self,
        remove_images: Optional[ImageType] = None,
        remove_volumes: bool = False,
    ) -> Dict:
        """
        Stop and remove services
        Args:
            remove_images: Remove images (local, all, or None)
            remove_volumes: Remove named volumes
        """
        try:
            self.project.down(
                remove_image_type=remove_images,
                include_volumes=remove_volumes,
            )
            return {
                "status": "success",
                "removed_images": bool(remove_images),
                "removed_volumes": remove_volumes,
            }
        except Exception as e:
            raise Exception(f"Failed to stop services: {str(e)}")

    def start(self, services: Optional[List[str]] = None) -> Dict:
        """
        Start existing containers
        Args:
            services: List of service names (default: all services)
        """
        try:
            self.project.start(service_names=services)
            return {
                "status": "success",
                "services": services or self.list_services(),
            }
        except Exception as e:
            raise Exception(f"Failed to start services: {str(e)}")

    def stop(self, services: Optional[List[str]] = None) -> Dict:
        """
        Stop containers
        Args:
            services: List of service names (default: all services)
        """
        try:
            self.project.stop(service_names=services)
            return {
                "status": "success",
                "services": services or self.list_services(),
            }
        except Exception as e:
            raise Exception(f"Failed to stop services: {str(e)}")

    def restart(self, services: Optional[List[str]] = None) -> Dict:
        """
        Restart containers
        Args:
            services: List of service names (default: all services)
        """
        try:
            self.project.restart(service_names=services)
            return {
                "status": "success",
                "services": services or self.list_services(),
            }
        except Exception as e:
            raise Exception(f"Failed to restart services: {str(e)}")

    def logs(
        self,
        services: Optional[List[str]] = None,
        tail: Optional[int] = None,
        follow: bool = False,
    ) -> Union[str, Dict]:
        """
        Get service logs
        Args:
            services: List of service names (default: all services)
            tail: Number of lines to show from end
            follow: Follow log output
        """
        try:
            logs = self.project.logs(
                service_names=services,
                tail=tail,
                follow=follow,
            )
            if follow:
                return logs
            return {
                "logs": logs.decode() if isinstance(logs, bytes) else logs,
                "services": services or self.list_services(),
            }
        except Exception as e:
            raise Exception(f"Failed to get logs: {str(e)}")

    def pull(self, services: Optional[List[str]] = None) -> Dict:
        """
        Pull service images
        Args:
            services: List of service names (default: all services)
        """
        try:
            self.project.pull(service_names=services)
            return {
                "status": "success",
                "services": services or self.list_services(),
            }
        except Exception as e:
            raise Exception(f"Failed to pull images: {str(e)}")

    def build(self, services: Optional[List[str]] = None) -> Dict:
        """
        Build service images
        Args:
            services: List of service names (default: all services)
        """
        try:
            self.project.build(service_names=services)
            return {
                "status": "success",
                "services": services or self.list_services(),
            }
        except Exception as e:
            raise Exception(f"Failed to build services: {str(e)}")

    def list_services(self) -> List[str]:
        """Get list of service names"""
        return [service.name for service in self.project.services]

    def get_service_status(self, service_name: str) -> Dict:
        """
        Get status of a specific service
        Args:
            service_name: Service name
        """
        try:
            service = self.project.get_service(service_name)
            containers = service.containers()

            return {
                "name": service_name,
                "running": bool(containers),
                "containers": [
                    {
                        "id": container.id,
                        "name": container.name,
                        "state": container.get("State", {}),
                        "status": container.get("Status", ""),
                    }
                    for container in containers
                ],
            }
        except Exception as e:
            raise Exception(f"Failed to get service status: {str(e)}")

    def get_config(self) -> Dict:
        """Get compose configuration"""
        try:
            config = self.project.get_config()
            return {
                "version": config.version,
                "services": config.services,
                "volumes": config.volumes,
                "networks": config.networks,
            }
        except Exception as e:
            raise Exception(f"Failed to get config: {str(e)}")
