import subprocess
from pathlib import Path
from typing import Dict, List, Optional, Union


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
        self.project_dir = str(Path(project_dir).resolve())
        self.compose_files = compose_files or ["docker-compose.yml"]
        self.project_name = project_name or Path(project_dir).name

        # Validate compose files exist
        for file in self.compose_files:
            file_path = Path(self.project_dir) / file
            if not file_path.exists():
                raise FileNotFoundError(f"Compose file not found: {file_path}")

    def _run_compose_command(
        self,
        command: List[str],
        capture_output: bool = True,
        check: bool = True,
    ) -> subprocess.CompletedProcess:
        """Run docker-compose command"""
        cmd = ["docker", "compose"]

        # Add compose files
        for file in self.compose_files:
            cmd.extend(["-f", file])

        # Add project name
        cmd.extend(["--project-name", self.project_name])

        # Add the command
        cmd.extend(command)

        return subprocess.run(
            cmd,
            cwd=self.project_dir,
            capture_output=capture_output,
            text=True,
            check=check,
        )

    def up(
        self,
        services: Optional[List[str]] = None,
        detach: bool = True,
        build: bool = False,
        remove_orphans: bool = False,
        scale: Optional[Dict[str, int]] = None,
    ) -> Dict:
        """
        Start services
        Args:
            services: List of service names (default: all services)
            detach: Run in background
            build: Build images before starting
            remove_orphans: Remove containers for services not in compose file
            scale: Dictionary of service names and their replica count
        """
        try:
            cmd = ["up"]
            if detach:
                cmd.append("-d")
            if build:
                cmd.append("--build")
            if remove_orphans:
                cmd.append("--remove-orphans")

            # Add scale arguments
            if scale:
                for service, count in scale.items():
                    cmd.extend(["--scale", f"{service}={count}"])

            # Add services
            if services:
                cmd.extend(services)

            result = self._run_compose_command(cmd)

            return {
                "status": "success",
                "output": result.stdout,
                "services": services or self.list_services(),
                "scale": scale,
            }

        except subprocess.CalledProcessError as e:
            raise Exception(f"Failed to start services: {e.stderr}")

    def down(
        self,
        volumes: bool = False,
        remove_images: str = "none",
        remove_orphans: bool = False,
    ) -> Dict:
        """
        Stop and remove containers
        Args:
            volumes: Remove named volumes
            remove_images: Remove images (none, local, all)
            remove_orphans: Remove containers for services not in compose file
        """
        try:
            cmd = ["down"]
            if volumes:
                cmd.append("-v")
            if remove_images != "none":
                cmd.extend(["--rmi", remove_images])
            if remove_orphans:
                cmd.append("--remove-orphans")

            result = self._run_compose_command(cmd)

            return {
                "status": "success",
                "output": result.stdout,
            }

        except subprocess.CalledProcessError as e:
            raise Exception(f"Failed to stop services: {e.stderr}")

    def ps(self, services: Optional[List[str]] = None) -> List[Dict]:
        """
        List containers
        Args:
            services: List of service names (default: all services)
        """
        try:
            cmd = ["ps", "--format", "json"]
            if services:
                cmd.extend(services)

            result = self._run_compose_command(cmd)

            # Parse JSON output
            lines = result.stdout.strip().split("\n")
            containers = [eval(line) for line in lines if line.strip()]

            return containers

        except subprocess.CalledProcessError as e:
            raise Exception(f"Failed to list containers: {e.stderr}")

    def logs(
        self,
        services: Optional[List[str]] = None,
        follow: bool = False,
        tail: Optional[int] = None,
    ) -> str:
        """
        View output from containers
        Args:
            services: List of service names (default: all services)
            follow: Follow log output
            tail: Number of lines to show from end
        """
        try:
            cmd = ["logs"]
            if follow:
                cmd.append("-f")
            if tail is not None:
                cmd.extend(["--tail", str(tail)])
            if services:
                cmd.extend(services)

            result = self._run_compose_command(cmd, check=not follow)
            return result.stdout

        except subprocess.CalledProcessError as e:
            raise Exception(f"Failed to get logs: {e.stderr}")

    def list_services(self) -> List[str]:
        """List services defined in compose file"""
        try:
            result = self._run_compose_command(["config", "--services"])
            return result.stdout.strip().split("\n")

        except subprocess.CalledProcessError as e:
            raise Exception(f"Failed to list services: {e.stderr}")
