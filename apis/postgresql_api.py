from typing import Dict, List, Optional, Union

import psycopg2
from psycopg2.extras import RealDictCursor


class PostgreSQLApi:
    def __init__(
        self,
        host: str,
        port: int,
        database: str,
        user: str,
        password: str,
    ):
        """
        Initialize PostgreSQL client
        Args:
            host: Database host
            port: Database port
            database: Database name
            user: Username
            password: Password
        """
        self.connection_params = {
            "host": host,
            "port": port,
            "database": database,
            "user": user,
            "password": password,
        }

    def _get_connection(self):
        """Create a new database connection"""
        return psycopg2.connect(
            **self.connection_params,
            cursor_factory=RealDictCursor,
        )

    def get_blocked_queries(self) -> List[Dict]:
        """
        Get queries that are being blocked by other queries
        Returns list of blocked queries with details about the blocking query
        """
        try:
            query = """
                SELECT
                    blocked.pid AS blocked_pid,
                    blocked.application_name AS blocked_application,
                    blocked.query AS blocked_query,
                    blocked.state AS blocked_state,
                    blocked.wait_event_type AS blocked_wait_type,
                    blocked.wait_event AS blocked_wait_event,
                    blocked.xact_start AS blocked_transaction_start,
                    blocked.query_start AS blocked_query_start,
                    blocked.state_change AS blocked_state_change,
                    blocking.pid AS blocking_pid,
                    blocking.application_name AS blocking_application,
                    blocking.query AS blocking_query,
                    blocking.state AS blocking_state,
                    blocking.wait_event_type AS blocking_wait_type,
                    blocking.wait_event AS blocking_wait_event,
                    blocking.xact_start AS blocking_transaction_start,
                    blocking.query_start AS blocking_query_start,
                    blocking.state_change AS blocking_state_change
                FROM pg_stat_activity blocked
                JOIN pg_stat_activity blocking 
                    ON blocking.pid = ANY(pg_blocking_pids(blocked.pid))
                WHERE blocked.pid != blocking.pid
                ORDER BY blocked.xact_start;
            """
            with self._get_connection() as conn:
                with conn.cursor() as cur:
                    cur.execute(query)
                    return cur.fetchall()
        except Exception as e:
            raise Exception(f"Failed to get blocked queries: {str(e)}")

    def get_blocking_queries(self) -> List[Dict]:
        """
        Get queries that are blocking other queries
        Returns list of blocking queries with count of queries they're blocking
        """
        try:
            query = """
                SELECT
                    blocking.pid,
                    blocking.application_name,
                    blocking.query,
                    blocking.state,
                    blocking.wait_event_type,
                    blocking.wait_event,
                    blocking.xact_start,
                    blocking.query_start,
                    blocking.state_change,
                    COUNT(blocked.pid) AS blocked_process_count,
                    array_agg(blocked.pid) AS blocked_pids
                FROM pg_stat_activity blocking
                JOIN pg_stat_activity blocked 
                    ON blocking.pid = ANY(pg_blocking_pids(blocked.pid))
                WHERE blocked.pid != blocking.pid
                GROUP BY
                    blocking.pid,
                    blocking.application_name,
                    blocking.query,
                    blocking.state,
                    blocking.wait_event_type,
                    blocking.wait_event,
                    blocking.xact_start,
                    blocking.query_start,
                    blocking.state_change
                ORDER BY COUNT(blocked.pid) DESC;
            """
            with self._get_connection() as conn:
                with conn.cursor() as cur:
                    cur.execute(query)
                    return cur.fetchall()
        except Exception as e:
            raise Exception(f"Failed to get blocking queries: {str(e)}")

    def get_locks(self) -> List[Dict]:
        """
        Get information about current locks
        Returns detailed information about locks and the queries holding them
        """
        try:
            query = """
                WITH RECURSIVE lock_chain AS (
                    SELECT DISTINCT
                        lock.locktype,
                        lock.relation::regclass::text AS table_name,
                        lock.mode,
                        lock.granted,
                        lock.pid,
                        activity.query,
                        activity.application_name,
                        activity.client_addr,
                        activity.query_start,
                        activity.xact_start,
                        activity.wait_event_type,
                        activity.wait_event,
                        ARRAY[lock.pid] AS pid_chain
                    FROM pg_locks lock
                    JOIN pg_stat_activity activity ON lock.pid = activity.pid
                    WHERE lock.pid != pg_backend_pid()
                    
                    UNION ALL
                    
                    SELECT DISTINCT
                        lock.locktype,
                        lock.relation::regclass::text AS table_name,
                        lock.mode,
                        lock.granted,
                        lock.pid,
                        activity.query,
                        activity.application_name,
                        activity.client_addr,
                        activity.query_start,
                        activity.xact_start,
                        activity.wait_event_type,
                        activity.wait_event,
                        pid_chain || lock.pid
                    FROM pg_locks lock
                    JOIN pg_stat_activity activity ON lock.pid = activity.pid
                    JOIN lock_chain ON lock.pid = ANY(pg_blocking_pids(lock_chain.pid))
                    WHERE NOT lock.pid = ANY(lock_chain.pid_chain)
                )
                SELECT
                    locktype,
                    table_name,
                    mode,
                    granted,
                    pid,
                    query,
                    application_name,
                    client_addr,
                    query_start,
                    xact_start,
                    wait_event_type,
                    wait_event,
                    pid_chain,
                    array_length(pid_chain, 1) AS chain_length
                FROM lock_chain
                ORDER BY pid_chain;
            """
            with self._get_connection() as conn:
                with conn.cursor() as cur:
                    cur.execute(query)
                    return cur.fetchall()
        except Exception as e:
            raise Exception(f"Failed to get locks: {str(e)}")

    def kill_process(self, pid: int) -> Dict:
        """
        Kill a specific database process
        Args:
            pid: Process ID to kill
        """
        try:
            with self._get_connection() as conn:
                with conn.cursor() as cur:
                    # First try to terminate gracefully
                    cur.execute("SELECT pg_terminate_backend(%s);", (pid,))
                    result = cur.fetchone()
                    return {
                        "pid": pid,
                        "terminated": result["pg_terminate_backend"],
                    }
        except Exception as e:
            raise Exception(f"Failed to kill process {pid}: {str(e)}")

    def kill_processes(self, pids: List[int]) -> List[Dict]:
        """
        Kill multiple database processes
        Args:
            pids: List of Process IDs to kill
        """
        results = []
        for pid in pids:
            try:
                result = self.kill_process(pid)
                results.append(result)
            except Exception as e:
                results.append(
                    {
                        "pid": pid,
                        "terminated": False,
                        "error": str(e),
                    }
                )
        return results

    def kill_blocking_queries(self, min_age_minutes: int = 10) -> List[Dict]:
        """
        Kill all queries that are blocking other queries and older than specified age
        Args:
            min_age_minutes: Minimum age in minutes for queries to be killed
        """
        try:
            query = """
                SELECT DISTINCT blocking.pid
                FROM pg_stat_activity blocking
                JOIN pg_stat_activity blocked 
                    ON blocking.pid = ANY(pg_blocking_pids(blocked.pid))
                WHERE 
                    blocked.pid != blocking.pid
                    AND blocking.query_start < NOW() - INTERVAL '%s minutes';
            """
            with self._get_connection() as conn:
                with conn.cursor() as cur:
                    cur.execute(query, (min_age_minutes,))
                    blocking_pids = [row["pid"] for row in cur.fetchall()]
                    return self.kill_processes(blocking_pids)
        except Exception as e:
            raise Exception(f"Failed to kill blocking queries: {str(e)}")
