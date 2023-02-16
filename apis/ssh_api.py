import os
from typing import Dict, List, Optional, Union

import paramiko
from paramiko.ssh_exception import SSHException


class SSHApi:
    def __init__(
        self,
        hostname: str,
        username: str,
        password: Optional[str] = None,
        key_filename: Optional[str] = None,
        port: int = 22,
    ):
        """
        Initialize SSH client
        Args:
            hostname: Remote host to connect to
            username: SSH username
            password: Optional password for authentication
            key_filename: Optional path to private key file
            port: SSH port (default: 22)
        """
        self.hostname = hostname
        self.username = username
        self.password = password
        self.key_filename = key_filename
        self.port = port
        self.client = paramiko.SSHClient()
        self.client.set_missing_host_key_policy(paramiko.AutoAddPolicy())

    def connect(self) -> None:
        """Establish SSH connection"""
        try:
            self.client.connect(
                hostname=self.hostname,
                username=self.username,
                password=self.password,
                key_filename=self.key_filename,
                port=self.port,
            )
        except Exception as e:
            raise Exception(f"Failed to connect to {self.hostname}: {str(e)}")

    def disconnect(self) -> None:
        """Close SSH connection"""
        self.client.close()

    def execute_command(
        self, command: str, timeout: int = 30, get_pty: bool = False
    ) -> Dict[str, Union[int, str]]:
        """
        Execute command on remote host
        Args:
            command: Command to execute
            timeout: Command timeout in seconds
            get_pty: Whether to allocate a pseudo-terminal
        """
        try:
            stdin, stdout, stderr = self.client.exec_command(
                command, timeout=timeout, get_pty=get_pty
            )

            exit_status = stdout.channel.recv_exit_status()
            output = stdout.read().decode().strip()
            error = stderr.read().decode().strip()

            return {
                "exit_code": exit_status,
                "output": output,
                "error": error,
                "success": exit_status == 0,
            }
        except Exception as e:
            raise Exception(f"Failed to execute command: {str(e)}")

    def execute_commands(
        self, commands: List[str], timeout: int = 30
    ) -> List[Dict[str, Union[int, str]]]:
        """
        Execute multiple commands sequentially
        Args:
            commands: List of commands to execute
            timeout: Command timeout in seconds
        """
        results = []
        for cmd in commands:
            result = self.execute_command(cmd, timeout)
            results.append({"command": cmd, **result})
            if not result["success"]:
                break
        return results

    def copy_file_to_remote(
        self,
        local_path: str,
        remote_path: str,
        preserve_mode: bool = True,
    ) -> Dict[str, Union[bool, str]]:
        """
        Copy local file to remote host
        Args:
            local_path: Path to local file
            remote_path: Destination path on remote host
            preserve_mode: Preserve file permissions
        """
        try:
            sftp = self.client.open_sftp()
            try:
                sftp.put(local_path, remote_path, preserve_mode=preserve_mode)
                return {
                    "success": True,
                    "message": f"File copied to {remote_path} successfully",
                }
            finally:
                sftp.close()
        except Exception as e:
            return {
                "success": False,
                "message": f"Failed to copy file: {str(e)}",
            }

    def copy_file_from_remote(
        self,
        remote_path: str,
        local_path: str,
        preserve_mode: bool = True,
    ) -> Dict[str, Union[bool, str]]:
        """
        Copy remote file to local machine
        Args:
            remote_path: Path to remote file
            local_path: Destination path on local machine
            preserve_mode: Preserve file permissions
        """
        try:
            sftp = self.client.open_sftp()
            try:
                sftp.get(remote_path, local_path, preserve_mode=preserve_mode)
                return {
                    "success": True,
                    "message": f"File copied to {local_path} successfully",
                }
            finally:
                sftp.close()
        except Exception as e:
            return {
                "success": False,
                "message": f"Failed to copy file: {str(e)}",
            }

    def copy_directory_to_remote(
        self,
        local_path: str,
        remote_path: str,
        preserve_mode: bool = True,
    ) -> Dict[str, Union[bool, str, List[str]]]:
        """
        Copy local directory to remote host recursively
        Args:
            local_path: Path to local directory
            remote_path: Destination path on remote host
            preserve_mode: Preserve file permissions
        """
        try:
            sftp = self.client.open_sftp()
            try:
                copied_files = []
                for root, _, files in os.walk(local_path):
                    for file in files:
                        local_file = os.path.join(root, file)
                        relative_path = os.path.relpath(local_file, local_path)
                        remote_file = os.path.join(remote_path, relative_path)

                        # Create remote directory if it doesn't exist
                        remote_dir = os.path.dirname(remote_file)
                        try:
                            sftp.stat(remote_dir)
                        except FileNotFoundError:
                            self.execute_command(f"mkdir -p {remote_dir}")

                        sftp.put(local_file, remote_file, preserve_mode=preserve_mode)
                        copied_files.append(relative_path)

                return {
                    "success": True,
                    "message": f"Directory copied to {remote_path} successfully",
                    "copied_files": copied_files,
                }
            finally:
                sftp.close()
        except Exception as e:
            return {
                "success": False,
                "message": f"Failed to copy directory: {str(e)}",
                "copied_files": [],
            }

    def list_directory(self, remote_path: str = ".") -> List[Dict[str, str]]:
        """
        List contents of remote directory
        Args:
            remote_path: Remote directory path
        """
        try:
            sftp = self.client.open_sftp()
            try:
                files = []
                for entry in sftp.listdir_attr(remote_path):
                    files.append(
                        {
                            "name": entry.filename,
                            "size": entry.st_size,
                            "mode": oct(entry.st_mode)[-4:],
                            "mtime": entry.st_mtime,
                            "type": "directory" if entry.st_mode & 0o40000 else "file",
                        }
                    )
                return files
            finally:
                sftp.close()
        except Exception as e:
            raise Exception(f"Failed to list directory: {str(e)}")

    def create_directory(self, remote_path: str, mode: int = 0o755) -> Dict[str, bool]:
        """
        Create directory on remote host
        Args:
            remote_path: Remote directory path
            mode: Directory permissions (octal)
        """
        try:
            sftp = self.client.open_sftp()
            try:
                sftp.mkdir(remote_path, mode)
                return {
                    "success": True,
                    "message": f"Directory {remote_path} created successfully",
                }
            finally:
                sftp.close()
        except Exception as e:
            return {
                "success": False,
                "message": f"Failed to create directory: {str(e)}",
            }
