from typing import Dict, List, Optional

from apis.opensearch_api import OpenSearchApi
from apis.ssh_api import SSHApi
from services.base_service import BaseService


class SSHService(BaseService):
    def __init__(
        self,
        opensearch: Optional[OpenSearchApi] = None,
        log_index: str = "ssh-logs",
    ):
        """
        Initialize SSH service
        Args:
            opensearch: OpenSearch API for logging
            log_index: Index name for logs
        """
        super().__init__(opensearch=opensearch, log_index=log_index)

    def execute_command(
        self,
        hostname: str,
        command: str,
        username: str,
        password: Optional[str] = None,
        key_filename: Optional[str] = None,
    ) -> Dict:
        """
        Execute command on remote host with logging
        Args:
            hostname: Remote host
            command: Command to execute
            username: SSH username
            password: Optional password
            key_filename: Optional key file
        """
        try:
            # Initialize SSH client
            ssh = SSHApi(
                hostname=hostname,
                username=username,
                password=password,
                key_filename=key_filename,
            )

            # Connect and execute
            ssh.connect()
            try:
                result = ssh.execute_command(command)

                # Log successful execution
                self.log_action(
                    action="execute_command",
                    status="success",
                    service="ssh",
                    details={
                        "hostname": hostname,
                        "command": command,
                        "username": username,
                        "exit_code": result["exit_code"],
                        "output_length": len(result["output"]),
                    },
                )

                return result
            finally:
                ssh.disconnect()

        except Exception as e:
            # Log failed execution
            self.log_action(
                action="execute_command",
                status="failed",
                service="ssh",
                details={
                    "hostname": hostname,
                    "command": command,
                    "username": username,
                },
                error=str(e),
            )
            raise

    def copy_file(
        self,
        hostname: str,
        local_path: str,
        remote_path: str,
        username: str,
        password: Optional[str] = None,
        key_filename: Optional[str] = None,
    ) -> Dict:
        """
        Copy file to remote host with logging
        Args:
            hostname: Remote host
            local_path: Local file path
            remote_path: Remote file path
            username: SSH username
            password: Optional password
            key_filename: Optional key file
        """
        try:
            # Initialize SSH client
            ssh = SSHApi(
                hostname=hostname,
                username=username,
                password=password,
                key_filename=key_filename,
            )

            # Connect and copy
            ssh.connect()
            try:
                result = ssh.copy_file_to_remote(local_path, remote_path)

                # Log successful copy
                self.log_action(
                    action="copy_file",
                    status="success",
                    service="ssh",
                    details={
                        "hostname": hostname,
                        "local_path": local_path,
                        "remote_path": remote_path,
                        "username": username,
                        "bytes_copied": result.get("bytes_copied", 0),
                    },
                )

                return result
            finally:
                ssh.disconnect()

        except Exception as e:
            # Log failed copy
            self.log_action(
                action="copy_file",
                status="failed",
                service="ssh",
                details={
                    "hostname": hostname,
                    "local_path": local_path,
                    "remote_path": remote_path,
                    "username": username,
                },
                error=str(e),
            )
            raise
