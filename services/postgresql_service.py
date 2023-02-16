from datetime import datetime
from typing import Dict, List, Optional

from apis.opensearch_api import OpenSearchApi
from apis.postgresql_api import PostgreSQLApi
from services.base_service import BaseService


class PostgreSQLService(BaseService):
    def __init__(
        self,
        host: str,
        port: int,
        database: str,
        user: str,
        password: str,
        opensearch: Optional[OpenSearchApi] = None,
        log_index: str = "postgresql-logs",
    ):
        """
        Initialize PostgreSQL service
        Args:
            host: Database host
            port: Database port
            database: Database name
            user: Username
            password: Password
            opensearch: OpenSearch API for logging
            log_index: Index name for logs
        """
        super().__init__(opensearch=opensearch, log_index=log_index)
        self.client = PostgreSQLApi(
            host=host,
            port=port,
            database=database,
            user=user,
            password=password,
        )

    def kill_blocking_queries(self, min_age_minutes: int = 10) -> List[Dict]:
        """Kill blocking queries with logging"""
        try:
            result = self.client.kill_blocking_queries(min_age_minutes)

            self.log_action(
                action="kill_blocking_queries",
                status="success",
                service="postgresql",
                details={
                    "min_age_minutes": min_age_minutes,
                    "queries_killed": len(result),
                    "pids": [r["pid"] for r in result],
                },
            )
            return result
        except Exception as e:
            self.log_action(
                action="kill_blocking_queries",
                status="failed",
                service="postgresql",
                details={"min_age_minutes": min_age_minutes},
                error=str(e),
            )
            raise

    def monitor_locks(self) -> Dict:
        """Monitor database locks with logging"""
        try:
            locks = self.client.get_locks()
            blocked = self.client.get_blocked_queries()
            blocking = self.client.get_blocking_queries()

            status = {
                "locks": locks,
                "blocked_queries": blocked,
                "blocking_queries": blocking,
            }

            self.log_action(
                action="monitor_locks",
                status="success",
                service="postgresql",
                details={
                    "lock_count": len(locks),
                    "blocked_count": len(blocked),
                    "blocking_count": len(blocking),
                },
            )
            return status
        except Exception as e:
            self.log_action(
                action="monitor_locks",
                status="failed",
                service="postgresql",
                error=str(e),
            )
            raise
