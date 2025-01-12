import json
from typing import Any, Dict, List, Optional, Union

import psycopg2
from psycopg2.extras import RealDictCursor


class PostgreSQLApi:
    def __init__(
        self,
        host: str,
        database: str,
        user: str,
        password: str,
        port: int = 5432,
    ):
        """
        Initialize PostgreSQL client
        Args:
            host: Database host
            database: Database name
            user: Username
            password: Password
            port: Port number (default: 5432)
        """
        try:
            self.conn = psycopg2.connect(
                host=host,
                database=database,
                user=user,
                password=password,
                port=port,
            )
            self.conn.autocommit = False
        except Exception as e:
            raise Exception(f"Failed to connect to PostgreSQL: {str(e)}")

    def __del__(self):
        """Close connection on cleanup"""
        if hasattr(self, "conn"):
            self.conn.close()

    # =========================================================================
    # Connection Management
    # =========================================================================

    def commit(self) -> None:
        """Commit current transaction"""
        self.conn.commit()

    def rollback(self) -> None:
        """Rollback current transaction"""
        self.conn.rollback()

    # =========================================================================
    # Table Operations
    # =========================================================================

    def list_tables(self, schema: str = "public") -> List[str]:
        """
        List all tables in schema
        Args:
            schema: Schema name (default: public)
        """
        try:
            with self.conn.cursor() as cur:
                cur.execute(
                    """
                    SELECT table_name 
                    FROM information_schema.tables 
                    WHERE table_schema = %s
                    """,
                    (schema,),
                )
                return [row[0] for row in cur.fetchall()]
        except Exception as e:
            raise Exception(f"Failed to list tables: {str(e)}")

    def table_exists(self, table_name: str, schema: str = "public") -> bool:
        """
        Check if table exists
        Args:
            table_name: Table name
            schema: Schema name (default: public)
        """
        try:
            with self.conn.cursor() as cur:
                cur.execute(
                    """
                    SELECT EXISTS (
                        SELECT FROM information_schema.tables 
                        WHERE table_schema = %s 
                        AND table_name = %s
                    )
                    """,
                    (schema, table_name),
                )
                return cur.fetchone()[0]
        except Exception as e:
            raise Exception(f"Failed to check table existence: {str(e)}")

    # =========================================================================
    # CRUD Operations
    # =========================================================================

    def create(
        self,
        table_name: str,
        data: Union[Dict, List[Dict]],
        returning: Optional[str] = None,
    ) -> List[Dict]:
        """
        Insert new record(s)
        Args:
            table_name: Target table
            data: Record data (dict or list of dicts)
            returning: Optional columns to return
        """
        try:
            if not isinstance(data, list):
                data = [data]

            with self.conn.cursor(cursor_factory=RealDictCursor) as cur:
                # Build query dynamically
                columns = data[0].keys()
                placeholders = ", ".join(["%s"] * len(columns))
                column_names = ", ".join(columns)

                query = f"""
                    INSERT INTO {table_name} ({column_names})
                    VALUES ({placeholders})
                """

                if returning:
                    query += f" RETURNING {returning}"

                # Execute for each record
                results = []
                for record in data:
                    values = [record[col] for col in columns]
                    cur.execute(query, values)
                    if returning:
                        results.extend(cur.fetchall())

                self.commit()
                return results
        except Exception as e:
            self.rollback()
            raise Exception(f"Failed to create record(s): {str(e)}")

    def read(
        self,
        table_name: str,
        columns: Optional[List[str]] = None,
        where: Optional[Dict] = None,
        order_by: Optional[str] = None,
        limit: Optional[int] = None,
        offset: Optional[int] = None,
    ) -> List[Dict]:
        """
        Read records
        Args:
            table_name: Source table
            columns: Columns to select (default: all)
            where: Where conditions
            order_by: Order by clause
            limit: Result limit
            offset: Result offset
        """
        try:
            with self.conn.cursor(cursor_factory=RealDictCursor) as cur:
                # Build query
                query = (
                    f"SELECT {', '.join(columns) if columns else '*'} FROM {table_name}"
                )
                params = []

                if where:
                    conditions = []
                    for key, value in where.items():
                        conditions.append(f"{key} = %s")
                        params.append(value)
                    query += f" WHERE {' AND '.join(conditions)}"

                if order_by:
                    query += f" ORDER BY {order_by}"

                if limit:
                    query += f" LIMIT {limit}"

                if offset:
                    query += f" OFFSET {offset}"

                cur.execute(query, params)
                return cur.fetchall()
        except Exception as e:
            raise Exception(f"Failed to read records: {str(e)}")

    def update(
        self,
        table_name: str,
        data: Dict,
        where: Dict,
        returning: Optional[str] = None,
    ) -> List[Dict]:
        """
        Update records
        Args:
            table_name: Target table
            data: Update data
            where: Where conditions
            returning: Optional columns to return
        """
        try:
            with self.conn.cursor(cursor_factory=RealDictCursor) as cur:
                # Build query
                set_items = [f"{k} = %s" for k in data.keys()]
                where_items = [f"{k} = %s" for k in where.keys()]

                query = f"""
                    UPDATE {table_name} 
                    SET {', '.join(set_items)}
                    WHERE {' AND '.join(where_items)}
                """

                if returning:
                    query += f" RETURNING {returning}"

                # Combine parameters
                params = list(data.values()) + list(where.values())

                cur.execute(query, params)
                results = cur.fetchall() if returning else []

                self.commit()
                return results
        except Exception as e:
            self.rollback()
            raise Exception(f"Failed to update records: {str(e)}")

    def delete(
        self,
        table_name: str,
        where: Dict,
        returning: Optional[str] = None,
    ) -> List[Dict]:
        """
        Delete records
        Args:
            table_name: Target table
            where: Where conditions
            returning: Optional columns to return
        """
        try:
            with self.conn.cursor(cursor_factory=RealDictCursor) as cur:
                # Build query
                conditions = [f"{k} = %s" for k in where.keys()]
                query = f"""
                    DELETE FROM {table_name}
                    WHERE {' AND '.join(conditions)}
                """

                if returning:
                    query += f" RETURNING {returning}"

                cur.execute(query, list(where.values()))
                results = cur.fetchall() if returning else []

                self.commit()
                return results
        except Exception as e:
            self.rollback()
            raise Exception(f"Failed to delete records: {str(e)}")

    # =========================================================================
    # Query Operations
    # =========================================================================

    def execute(
        self,
        query: str,
        params: Optional[tuple] = None,
        fetch: bool = True,
    ) -> Union[List[Dict], None]:
        """
        Execute raw SQL query
        Args:
            query: SQL query
            params: Query parameters
            fetch: Fetch results (default: True)
        """
        try:
            with self.conn.cursor(cursor_factory=RealDictCursor) as cur:
                cur.execute(query, params)
                if fetch:
                    return cur.fetchall()
                self.commit()
                return None
        except Exception as e:
            self.rollback()
            raise Exception(f"Query execution failed: {str(e)}")

    def execute_many(
        self,
        query: str,
        params: List[tuple],
    ) -> None:
        """
        Execute same query with multiple parameter sets
        Args:
            query: SQL query
            params: List of parameter tuples
        """
        try:
            with self.conn.cursor() as cur:
                cur.executemany(query, params)
                self.commit()
        except Exception as e:
            self.rollback()
            raise Exception(f"Batch execution failed: {str(e)}")

    # =========================================================================
    # Lock and Blocking Detection
    # =========================================================================

    def get_blocked_queries(self) -> List[Dict]:
        """
        Get currently blocked queries with details about blockers
        Returns information about:
            - Blocked query
            - Blocking query
            - Duration of block
            - Application names
            - Users involved
        """
        try:
            query = """
                SELECT 
                    blocked_activity.pid AS blocked_pid,
                    blocked_activity.application_name AS blocked_application,
                    blocked_activity.usename AS blocked_user,
                    blocked_activity.query AS blocked_query,
                    blocking_activity.pid AS blocking_pid,
                    blocking_activity.application_name AS blocking_application,
                    blocking_activity.usename AS blocking_user,
                    blocking_activity.query AS blocking_query,
                    extract(epoch from now() - blocked_activity.xact_start)::integer AS blocked_duration_seconds
                FROM pg_stat_activity blocked_activity
                JOIN pg_stat_activity blocking_activity 
                    ON blocked_activity.pid != blocking_activity.pid
                JOIN pg_locks blocked_locks 
                    ON blocked_activity.pid = blocked_locks.pid
                JOIN pg_locks blocking_locks 
                    ON blocking_locks.locktype = blocked_locks.locktype
                    AND blocking_locks.database IS NOT DISTINCT FROM blocked_locks.database
                    AND blocking_locks.relation IS NOT DISTINCT FROM blocked_locks.relation
                    AND blocking_locks.page IS NOT DISTINCT FROM blocked_locks.page
                    AND blocking_locks.tuple IS NOT DISTINCT FROM blocked_locks.tuple
                    AND blocking_locks.virtualxid IS NOT DISTINCT FROM blocked_locks.virtualxid
                    AND blocking_locks.transactionid IS NOT DISTINCT FROM blocked_locks.transactionid
                    AND blocking_locks.classid IS NOT DISTINCT FROM blocked_locks.classid
                    AND blocking_locks.objid IS NOT DISTINCT FROM blocked_locks.objid
                    AND blocking_locks.objsubid IS NOT DISTINCT FROM blocked_locks.objsubid
                    AND blocking_locks.pid = blocking_activity.pid
                WHERE NOT blocked_locks.granted
                AND blocking_locks.granted;
            """
            return self.execute(query)
        except Exception as e:
            raise Exception(f"Failed to get blocked queries: {str(e)}")

    def get_lock_conflicts(self) -> List[Dict]:
        """
        Get current lock conflicts and waiting transactions
        Returns information about:
            - Lock types
            - Relations involved
            - Transaction details
            - Wait duration
        """
        try:
            query = """
                SELECT DISTINCT
                    waiting.locktype AS waiting_locktype,
                    waiting.relation::regclass AS waiting_table,
                    waiting_stm.usename AS waiting_user,
                    waiting_stm.query AS waiting_query,
                    waiting_stm.pid AS waiting_pid,
                    blocking_stm.usename AS blocking_user,
                    blocking_stm.query AS blocking_query,
                    blocking_stm.pid AS blocking_pid,
                    extract(epoch from now() - waiting_stm.xact_start)::integer AS wait_duration_seconds
                FROM pg_locks AS waiting
                JOIN pg_stat_activity AS waiting_stm
                    ON waiting.pid = waiting_stm.pid
                JOIN pg_locks AS blocking
                    ON waiting.transactionid = blocking.transactionid
                    AND waiting.pid != blocking.pid
                JOIN pg_stat_activity blocking_stm
                    ON blocking.pid = blocking_stm.pid
                WHERE NOT waiting.granted;
            """
            return self.execute(query)
        except Exception as e:
            raise Exception(f"Failed to get lock conflicts: {str(e)}")

    def get_deadlocks(self, minutes: int = 60) -> List[Dict]:
        """
        Get deadlock events from the past period
        Args:
            minutes: Look back period in minutes
        Returns information about:
            - Deadlock time
            - Processes involved
            - Queries causing deadlock
        """
        try:
            query = """
                SELECT 
                    event_time,
                    process_id,
                    client_application,
                    client_hostname,
                    deadlock_details
                FROM pg_log
                WHERE error_severity = 'ERROR'
                AND message LIKE '%deadlock%'
                AND event_time > now() - interval '%s minutes'
                ORDER BY event_time DESC;
            """
            return self.execute(query, (minutes,))
        except Exception as e:
            raise Exception(f"Failed to get deadlocks: {str(e)}")

    def get_long_running_transactions(self, seconds: int = 300) -> List[Dict]:
        """
        Get transactions running longer than specified duration
        Args:
            seconds: Minimum transaction duration in seconds
        Returns information about:
            - Transaction duration
            - Query
            - Application
            - State
        """
        try:
            query = """
                SELECT 
                    pid,
                    usename,
                    application_name,
                    client_addr,
                    backend_start,
                    xact_start,
                    query_start,
                    state,
                    query,
                    extract(epoch from now() - xact_start)::integer AS duration_seconds
                FROM pg_stat_activity
                WHERE xact_start IS NOT NULL
                AND extract(epoch from now() - xact_start) > %s
                ORDER BY xact_start;
            """
            return self.execute(query, (seconds,))
        except Exception as e:
            raise Exception(f"Failed to get long running transactions: {str(e)}")

    def get_table_locks(self, schema: str = "public") -> List[Dict]:
        """
        Get current locks on tables
        Args:
            schema: Schema to check (default: public)
        Returns information about:
            - Table name
            - Lock type
            - Lock mode
            - Process holding lock
        """
        try:
            query = """
                SELECT 
                    pg_class.relname AS table_name,
                    pg_locks.locktype,
                    pg_locks.mode,
                    pg_locks.granted,
                    pg_stat_activity.usename,
                    pg_stat_activity.query,
                    pg_stat_activity.pid
                FROM pg_locks
                JOIN pg_class 
                    ON pg_locks.relation = pg_class.oid
                JOIN pg_stat_activity 
                    ON pg_locks.pid = pg_stat_activity.pid
                WHERE pg_class.relnamespace = (
                    SELECT oid 
                    FROM pg_namespace 
                    WHERE nspname = %s
                )
                AND pg_locks.locktype = 'relation'
                ORDER BY pg_class.relname;
            """
            return self.execute(query, (schema,))
        except Exception as e:
            raise Exception(f"Failed to get table locks: {str(e)}")

    def kill_blocked_queries(self, older_than_seconds: int = 300) -> List[Dict]:
        """
        Kill blocked queries running longer than specified duration
        Args:
            older_than_seconds: Minimum blocked duration in seconds
        Returns:
            List of terminated queries
        """
        try:
            query = """
                SELECT pg_terminate_backend(blocked_activity.pid) AS terminated,
                    blocked_activity.pid,
                    blocked_activity.query
                FROM pg_stat_activity blocked_activity
                JOIN pg_stat_activity blocking_activity 
                    ON blocked_activity.pid != blocking_activity.pid
                WHERE blocked_activity.state = 'active'
                AND blocked_activity.wait_event_type = 'Lock'
                AND extract(epoch from now() - blocked_activity.xact_start) > %s;
            """
            return self.execute(query, (older_than_seconds,))
        except Exception as e:
            raise Exception(f"Failed to kill blocked queries: {str(e)}")
