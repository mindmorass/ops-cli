from datetime import datetime
from typing import Dict, List, Optional, Union

from opensearchpy import OpenSearch, RequestsHttpConnection
from opensearchpy.exceptions import NotFoundError, RequestError


class OpenSearchApi:
    def __init__(
        self,
        hosts: List[Dict[str, Union[str, int]]] = None,
        username: str = "admin",
        password: str = "admin",
        use_ssl: bool = True,
        verify_certs: bool = False,
        ssl_show_warn: bool = False,
    ):
        """
        Initialize OpenSearch client
        Args:
            hosts: List of host dictionaries [{"host": "localhost", "port": 9200}]
            username: OpenSearch username
            password: OpenSearch password
            use_ssl: Whether to use SSL/TLS
            verify_certs: Whether to verify SSL certificates
            ssl_show_warn: Whether to show SSL warnings
        """
        self.client = OpenSearch(
            hosts=hosts or [{"host": "localhost", "port": 9200}],
            http_auth=(username, password),
            use_ssl=use_ssl,
            verify_certs=verify_certs,
            ssl_show_warn=ssl_show_warn,
            connection_class=RequestsHttpConnection,
        )

    def create_index(
        self,
        index_name: str,
        mappings: Optional[Dict] = None,
        settings: Optional[Dict] = None,
    ) -> Dict:
        """
        Create a new index
        Args:
            index_name: Name of the index
            mappings: Index mappings configuration
            settings: Index settings configuration
        """
        try:
            # Default mappings for logs if none provided
            if not mappings:
                mappings = {
                    "properties": {
                        "timestamp": {"type": "date"},
                        "level": {"type": "keyword"},
                        "service": {"type": "keyword"},
                        "message": {"type": "text"},
                        "metadata": {"type": "object"},
                    }
                }

            body = {
                "mappings": mappings,
            }

            if settings:
                body["settings"] = settings

            response = self.client.indices.create(
                index=index_name,
                body=body,
            )
            return response
        except RequestError as e:
            raise Exception(f"Failed to create index: {str(e)}")

    def write_log(
        self,
        index_name: str,
        level: str,
        message: str,
        service: str,
        metadata: Optional[Dict] = None,
        timestamp: Optional[datetime] = None,
    ) -> Dict:
        """
        Write a log entry
        Args:
            index_name: Target index name
            level: Log level (INFO, WARNING, ERROR, etc.)
            message: Log message
            service: Service/component name
            metadata: Additional log metadata
            timestamp: Log timestamp (defaults to current time)
        """
        try:
            document = {
                "timestamp": timestamp or datetime.utcnow().isoformat(),
                "level": level.upper(),
                "service": service,
                "message": message,
                "metadata": metadata or {},
            }

            response = self.client.index(
                index=index_name,
                body=document,
            )
            return response
        except Exception as e:
            raise Exception(f"Failed to write log: {str(e)}")

    def write_logs(
        self,
        index_name: str,
        logs: List[Dict],
    ) -> Dict:
        """
        Write multiple log entries in bulk
        Args:
            index_name: Target index name
            logs: List of log dictionaries
        """
        try:
            operations = []
            for log in logs:
                # Add index operation
                operations.append({"index": {"_index": index_name}})
                # Add log document
                operations.append(
                    {
                        "timestamp": log.get("timestamp")
                        or datetime.utcnow().isoformat(),
                        "level": log["level"].upper(),
                        "service": log["service"],
                        "message": log["message"],
                        "metadata": log.get("metadata", {}),
                    }
                )

            response = self.client.bulk(
                operations=operations,
                refresh=True,
            )
            return response
        except Exception as e:
            raise Exception(f"Failed to write logs: {str(e)}")

    def search_logs(
        self,
        index_name: str,
        query: Optional[Dict] = None,
        service: Optional[str] = None,
        level: Optional[str] = None,
        start_time: Optional[datetime] = None,
        end_time: Optional[datetime] = None,
        size: int = 100,
        sort_order: str = "desc",
    ) -> List[Dict]:
        """
        Search logs with filters
        Args:
            index_name: Index to search
            query: Custom query DSL
            service: Filter by service name
            level: Filter by log level
            start_time: Filter logs after this time
            end_time: Filter logs before this time
            size: Maximum number of results
            sort_order: Sort order (asc/desc)
        """
        try:
            if query is None:
                # Build query from parameters
                must_conditions = []

                if service:
                    must_conditions.append({"term": {"service": service}})
                if level:
                    must_conditions.append({"term": {"level": level.upper()}})
                if start_time or end_time:
                    time_range = {}
                    if start_time:
                        time_range["gte"] = start_time.isoformat()
                    if end_time:
                        time_range["lte"] = end_time.isoformat()
                    must_conditions.append({"range": {"timestamp": time_range}})

                query = {
                    "bool": (
                        {"must": must_conditions}
                        if must_conditions
                        else {"match_all": {}}
                    )
                }

            body = {
                "query": query,
                "size": size,
                "sort": [{"timestamp": {"order": sort_order}}],
            }

            response = self.client.search(
                index=index_name,
                body=body,
            )

            return [hit["_source"] for hit in response["hits"]["hits"]]
        except Exception as e:
            raise Exception(f"Failed to search logs: {str(e)}")

    def delete_old_logs(
        self,
        index_name: str,
        older_than: datetime,
    ) -> Dict:
        """
        Delete logs older than specified time
        Args:
            index_name: Index name
            older_than: Delete logs older than this time
        """
        try:
            query = {"range": {"timestamp": {"lt": older_than.isoformat()}}}

            response = self.client.delete_by_query(
                index=index_name,
                body={"query": query},
            )
            return response
        except Exception as e:
            raise Exception(f"Failed to delete logs: {str(e)}")

    def get_index_stats(self, index_name: str) -> Dict:
        """
        Get index statistics
        Args:
            index_name: Index name
        """
        try:
            return self.client.indices.stats(index=index_name)
        except NotFoundError:
            raise Exception(f"Index {index_name} not found")
        except Exception as e:
            raise Exception(f"Failed to get index stats: {str(e)}")

    def create_index_template(
        self,
        template_name: str,
        index_patterns: List[str],
        mappings: Dict,
        settings: Optional[Dict] = None,
    ) -> Dict:
        """
        Create an index template
        Args:
            template_name: Template name
            index_patterns: List of index patterns
            mappings: Template mappings
            settings: Template settings
        """
        try:
            body = {
                "index_patterns": index_patterns,
                "mappings": mappings,
            }

            if settings:
                body["settings"] = settings

            return self.client.indices.put_template(
                name=template_name,
                body=body,
            )
        except Exception as e:
            raise Exception(f"Failed to create template: {str(e)}")
