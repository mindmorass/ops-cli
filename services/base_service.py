from datetime import datetime
from typing import Dict, Optional

from apis.opensearch_api import OpenSearchApi


class BaseService:
    """Base service class with logging capabilities"""

    def __init__(
        self,
        opensearch: Optional[OpenSearchApi] = None,
        log_index: str = "service-logs",
    ):
        """
        Initialize base service
        Args:
            opensearch: OpenSearch API instance for logging
            log_index: Index name for logs
        """
        self.opensearch = opensearch
        self.log_index = log_index

    def log_action(
        self,
        action: str,
        status: str,
        service: str,
        details: Optional[Dict] = None,
        error: Optional[str] = None,
    ) -> None:
        """
        Log an action to OpenSearch
        Args:
            action: Action being performed
            status: Status of the action (success/failed)
            service: Service name
            details: Additional action details
            error: Error message if any
        """
        if not self.opensearch:
            return

        self.opensearch.write_log(
            index_name=self.log_index,
            level="ERROR" if error else "INFO",
            message=f"{action} - {status}",
            service=service,
            metadata={
                "action": action,
                "status": status,
                "details": details or {},
                "error": error,
            },
        )
