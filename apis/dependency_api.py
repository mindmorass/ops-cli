import os
import subprocess
from pathlib import Path
from typing import Dict, List, Optional, Tuple

import yaml


class DependencyApi:
    def __init__(self, config_path: Optional[str] = None):
        """
        Initialize dependency manager
        Args:
            config_path: Path to dependencies.yaml (default: configs/dependencies.yaml)
        """
        self.config_path = config_path or str(Path("configs") / "dependencies.yaml")
        self.dependencies = self._load_dependencies()

    def _load_dependencies(self) -> Dict:
        """Load dependencies from YAML config"""
        try:
            with open(self.config_path) as f:
                return yaml.safe_load(f)
        except Exception as e:
            raise Exception(f"Failed to load dependencies config: {str(e)}")

    def check_dependencies(self) -> List[Dict]:
        """
        Check all configured dependencies
        Returns list of dependency status dicts
        """
        results = []
        for dep_type, deps in self.dependencies.items():
            if dep_type == "homebrew":
                for dep in deps:
                    if dep.get("cask", False):
                        status = self.check_brew_cask(dep["name"])
                    else:
                        status = self.check_brew_package(dep["name"])

                    results.append(
                        {
                            "type": "homebrew",
                            "name": dep["name"],
                            "version": status[0],
                            "installed": status[1],
                            "required_version": dep.get("version"),
                            "is_cask": dep.get("cask", False),
                        }
                    )
            # Add other dependency types here (apt, pip, etc.)

        return results

    # =========================================================================
    # Homebrew Operations
    # =========================================================================

    def check_brew_package(self, package_name: str) -> Tuple[Optional[str], bool]:
        """
        Check if Homebrew package is installed
        Returns (version, installed) tuple
        """
        try:
            result = subprocess.run(
                ["brew", "list", "--versions", package_name],
                capture_output=True,
                text=True,
                check=False,
            )
            if result.returncode == 0:
                # Parse version from output (format: "package_name 1.2.3")
                version = result.stdout.strip().split()[-1]
                return version, True
            return None, False
        except Exception:
            return None, False

    def install_brew_package(
        self,
        package_name: str,
        version: Optional[str] = None,
        force: bool = False,
    ) -> Dict:
        """
        Install Homebrew package
        Args:
            package_name: Package to install
            version: Specific version to install
            force: Force reinstall if already installed
        """
        try:
            cmd = ["brew", "install"]
            if force:
                cmd.append("--force")

            if version:
                cmd.extend(["--version", version])

            cmd.append(package_name)

            result = subprocess.run(
                cmd,
                capture_output=True,
                text=True,
                check=True,
            )

            return {
                "success": True,
                "output": result.stdout,
                "package": package_name,
                "version": version,
            }
        except subprocess.CalledProcessError as e:
            return {
                "success": False,
                "error": e.stderr,
                "package": package_name,
                "version": version,
            }

    def uninstall_brew_package(self, package_name: str) -> Dict:
        """Uninstall Homebrew package"""
        try:
            result = subprocess.run(
                ["brew", "uninstall", package_name],
                capture_output=True,
                text=True,
                check=True,
            )

            return {
                "success": True,
                "output": result.stdout,
                "package": package_name,
            }
        except subprocess.CalledProcessError as e:
            return {
                "success": False,
                "error": e.stderr,
                "package": package_name,
            }

    def update_brew_package(self, package_name: str) -> Dict:
        """Update Homebrew package to latest version"""
        try:
            result = subprocess.run(
                ["brew", "upgrade", package_name],
                capture_output=True,
                text=True,
                check=True,
            )

            return {
                "success": True,
                "output": result.stdout,
                "package": package_name,
            }
        except subprocess.CalledProcessError as e:
            return {
                "success": False,
                "error": e.stderr,
                "package": package_name,
            }

    def list_brew_packages(self) -> List[Dict]:
        """List installed Homebrew packages with versions"""
        try:
            result = subprocess.run(
                ["brew", "list", "--versions"],
                capture_output=True,
                text=True,
                check=True,
            )

            packages = []
            for line in result.stdout.splitlines():
                if line:
                    parts = line.split()
                    packages.append(
                        {
                            "name": parts[0],
                            "version": parts[1] if len(parts) > 1 else None,
                        }
                    )

            return packages
        except subprocess.CalledProcessError as e:
            raise Exception(f"Failed to list packages: {str(e)}")

    # =========================================================================
    # Homebrew Cask Operations
    # =========================================================================

    def check_brew_cask(self, cask_name: str) -> Tuple[Optional[str], bool]:
        """
        Check if Homebrew cask is installed
        Returns (version, installed) tuple
        """
        try:
            result = subprocess.run(
                ["brew", "list", "--cask", "--versions", cask_name],
                capture_output=True,
                text=True,
                check=False,
            )
            if result.returncode == 0:
                # Parse version from output (format: "cask_name 1.2.3")
                version = result.stdout.strip().split()[-1]
                return version, True
            return None, False
        except Exception:
            return None, False

    def install_brew_cask(
        self,
        cask_name: str,
        force: bool = False,
    ) -> Dict:
        """
        Install Homebrew cask
        Args:
            cask_name: Cask to install
            force: Force reinstall if already installed
        """
        try:
            cmd = ["brew", "install", "--cask"]
            if force:
                cmd.append("--force")

            cmd.append(cask_name)

            result = subprocess.run(
                cmd,
                capture_output=True,
                text=True,
                check=True,
            )

            return {
                "success": True,
                "output": result.stdout,
                "cask": cask_name,
            }
        except subprocess.CalledProcessError as e:
            return {
                "success": False,
                "error": e.stderr,
                "cask": cask_name,
            }

    def uninstall_brew_cask(self, cask_name: str) -> Dict:
        """Uninstall Homebrew cask"""
        try:
            result = subprocess.run(
                ["brew", "uninstall", "--cask", cask_name],
                capture_output=True,
                text=True,
                check=True,
            )

            return {
                "success": True,
                "output": result.stdout,
                "cask": cask_name,
            }
        except subprocess.CalledProcessError as e:
            return {
                "success": False,
                "error": e.stderr,
                "cask": cask_name,
            }

    def list_brew_casks(self) -> List[Dict]:
        """List installed Homebrew casks with versions"""
        try:
            result = subprocess.run(
                ["brew", "list", "--cask", "--versions"],
                capture_output=True,
                text=True,
                check=True,
            )

            casks = []
            for line in result.stdout.splitlines():
                if line:
                    parts = line.split()
                    casks.append(
                        {
                            "name": parts[0],
                            "version": parts[1] if len(parts) > 1 else None,
                        }
                    )

            return casks
        except subprocess.CalledProcessError as e:
            raise Exception(f"Failed to list casks: {str(e)}")
