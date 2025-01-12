from typing import Dict, List, Optional, Union

from opensearchpy import OpenSearch, RequestsHttpConnection
from opensearchpy.exceptions import NotFoundError, RequestError


class OpenSearchApi:
    def __init__(
        self,
        hosts: List[str],
        username: Optional[str] = None,
        password: Optional[str] = None,
        use_ssl: bool = True,
        verify_certs: bool = True,
        ca_certs: Optional[str] = None,
    ):
        """
        Initialize OpenSearch client
        Args:
            hosts: List of OpenSearch hosts
            username: Optional username for authentication
            password: Optional password for authentication
            use_ssl: Use SSL/TLS
            verify_certs: Verify SSL certificates
            ca_certs: Path to CA certificate
        """
        try:
            auth = (username, password) if username and password else None
            self.client = OpenSearch(
                hosts=hosts,
                http_auth=auth,
                use_ssl=use_ssl,
                verify_certs=verify_certs,
                ca_certs=ca_certs,
                connection_class=RequestsHttpConnection,
            )
        except Exception as e:
            raise Exception(f"Failed to connect to OpenSearch: {str(e)}")

    # Index operations
    def list_indices(self, pattern: str = "*") -> List[Dict]:
        """
        List indices matching pattern
        Args:
            pattern: Index pattern to match
        """
        try:
            response = self.client.indices.get(pattern)
            return [
                {
                    "name": name,
                    "settings": info["settings"],
                    "mappings": info["mappings"],
                }
                for name, info in response.items()
            ]
        except Exception as e:
            raise Exception(f"Failed to list indices: {str(e)}")

    def create_index(
        self,
        index_name: str,
        mappings: Optional[Dict] = None,
        settings: Optional[Dict] = None,
    ) -> Dict:
        """
        Create new index
        Args:
            index_name: Name of index to create
            mappings: Optional field mappings
            settings: Optional index settings
        """
        try:
            body = {}
            if mappings:
                body["mappings"] = mappings
            if settings:
                body["settings"] = settings

            response = self.client.indices.create(index=index_name, body=body)
            return response
        except Exception as e:
            raise Exception(f"Failed to create index: {str(e)}")

    def delete_index(self, index_name: str) -> Dict:
        """
        Delete index
        Args:
            index_name: Name of index to delete
        """
        try:
            response = self.client.indices.delete(index=index_name)
            return response
        except Exception as e:
            raise Exception(f"Failed to delete index: {str(e)}")

    # Document operations
    def index_document(
        self,
        index_name: str,
        document: Dict,
        doc_id: Optional[str] = None,
        refresh: bool = False,
    ) -> Dict:
        """
        Index a document
        Args:
            index_name: Target index name
            document: Document to index
            doc_id: Optional document ID
            refresh: Refresh index immediately
        """
        try:
            response = self.client.index(
                index=index_name,
                body=document,
                id=doc_id,
                refresh=refresh,
            )
            return response
        except Exception as e:
            raise Exception(f"Failed to index document: {str(e)}")

    def get_document(self, index_name: str, doc_id: str) -> Dict:
        """
        Get document by ID
        Args:
            index_name: Index name
            doc_id: Document ID
        """
        try:
            response = self.client.get(index=index_name, id=doc_id)
            return response
        except NotFoundError:
            raise Exception(f"Document not found: {doc_id}")
        except Exception as e:
            raise Exception(f"Failed to get document: {str(e)}")

    def update_document(
        self,
        index_name: str,
        doc_id: str,
        document: Dict,
        refresh: bool = False,
    ) -> Dict:
        """
        Update document
        Args:
            index_name: Index name
            doc_id: Document ID
            document: Updated document fields
            refresh: Refresh index immediately
        """
        try:
            response = self.client.update(
                index=index_name,
                id=doc_id,
                body={"doc": document},
                refresh=refresh,
            )
            return response
        except NotFoundError:
            raise Exception(f"Document not found: {doc_id}")
        except Exception as e:
            raise Exception(f"Failed to update document: {str(e)}")

    def delete_document(
        self, index_name: str, doc_id: str, refresh: bool = False
    ) -> Dict:
        """
        Delete document
        Args:
            index_name: Index name
            doc_id: Document ID
            refresh: Refresh index immediately
        """
        try:
            response = self.client.delete(
                index=index_name,
                id=doc_id,
                refresh=refresh,
            )
            return response
        except NotFoundError:
            raise Exception(f"Document not found: {doc_id}")
        except Exception as e:
            raise Exception(f"Failed to delete document: {str(e)}")

    # Search operations
    def search(
        self,
        index_name: str,
        query: Dict,
        size: int = 10,
        from_: int = 0,
        sort: Optional[List[Dict]] = None,
        source: Optional[Union[List[str], bool]] = None,
    ) -> Dict:
        """
        Search documents
        Args:
            index_name: Index to search
            query: Search query
            size: Number of results
            from_: Starting offset
            sort: Optional sort criteria
            source: Fields to return
        """
        try:
            body = {"query": query}
            if sort:
                body["sort"] = sort
            if source is not None:
                body["_source"] = source

            response = self.client.search(
                index=index_name,
                body=body,
                size=size,
                from_=from_,
            )
            return response
        except Exception as e:
            raise Exception(f"Search failed: {str(e)}")

    def count(self, index_name: str, query: Optional[Dict] = None) -> int:
        """
        Count documents
        Args:
            index_name: Index name
            query: Optional query to filter documents
        """
        try:
            body = {"query": query} if query else None
            response = self.client.count(index=index_name, body=body)
            return response["count"]
        except Exception as e:
            raise Exception(f"Count failed: {str(e)}")

    # Bulk operations
    def bulk_index(
        self,
        index_name: str,
        documents: List[Dict],
        refresh: bool = False,
    ) -> Dict:
        """
        Bulk index documents
        Args:
            index_name: Target index
            documents: List of documents to index
            refresh: Refresh index immediately
        """
        try:
            body = []
            for doc in documents:
                body.extend(
                    [
                        {"index": {"_index": index_name}},
                        doc,
                    ]
                )
            response = self.client.bulk(body=body, refresh=refresh)
            return response
        except Exception as e:
            raise Exception(f"Bulk indexing failed: {str(e)}")

    # =========================================================================
    # Dashboard Index Pattern Operations
    # =========================================================================

    def create_index_pattern(
        self,
        title: str,
        time_field_name: str = "@timestamp",
        allow_hidden_indices: bool = False,
    ) -> Dict:
        """
        Create an index pattern in OpenSearch Dashboards
        Args:
            title: Pattern title (e.g., 'logs-*')
            time_field_name: Default timestamp field
            allow_hidden_indices: Include hidden indices
        """
        try:
            # Index patterns are stored in .kibana index
            body = {
                "type": "index-pattern",
                "index-pattern": {
                    "title": title,
                    "timeFieldName": time_field_name,
                    "allowHiddenIndices": allow_hidden_indices,
                    "fields": "[]",  # Will be populated by OpenSearch
                },
            }

            response = self.client.index(
                index=".opensearch_dashboards",
                body=body,
                id=f"index-pattern:{title}",
                refresh=True,
            )
            return response
        except Exception as e:
            raise Exception(f"Failed to create index pattern: {str(e)}")

    def get_index_pattern(self, pattern_id: str) -> Dict:
        """
        Get index pattern details
        Args:
            pattern_id: Index pattern ID
        """
        try:
            response = self.client.get(
                index=".opensearch_dashboards",
                id=f"index-pattern:{pattern_id}",
            )
            return response["_source"]["index-pattern"]
        except Exception as e:
            raise Exception(f"Failed to get index pattern: {str(e)}")

    def update_index_pattern(
        self,
        pattern_id: str,
        title: Optional[str] = None,
        time_field_name: Optional[str] = None,
        allow_hidden_indices: Optional[bool] = None,
    ) -> Dict:
        """
        Update index pattern settings
        Args:
            pattern_id: Index pattern ID
            title: New pattern title
            time_field_name: New timestamp field
            allow_hidden_indices: Include hidden indices
        """
        try:
            # Get current pattern
            current = self.get_index_pattern(pattern_id)

            # Update fields if provided
            update_body = {
                "index-pattern": {
                    **current,
                    **({"title": title} if title else {}),
                    **({"timeFieldName": time_field_name} if time_field_name else {}),
                    **(
                        {"allowHiddenIndices": allow_hidden_indices}
                        if allow_hidden_indices is not None
                        else {}
                    ),
                }
            }

            response = self.client.update(
                index=".opensearch_dashboards",
                id=f"index-pattern:{pattern_id}",
                body={"doc": update_body},
                refresh=True,
            )
            return response
        except Exception as e:
            raise Exception(f"Failed to update index pattern: {str(e)}")

    def delete_index_pattern(self, pattern_id: str) -> Dict:
        """
        Delete index pattern
        Args:
            pattern_id: Index pattern ID
        """
        try:
            response = self.client.delete(
                index=".opensearch_dashboards",
                id=f"index-pattern:{pattern_id}",
                refresh=True,
            )
            return response
        except Exception as e:
            raise Exception(f"Failed to delete index pattern: {str(e)}")

    def list_index_patterns(self) -> List[Dict]:
        """List all index patterns"""
        try:
            query = {"query": {"term": {"type": "index-pattern"}}}

            response = self.client.search(
                index=".opensearch_dashboards",
                body=query,
                size=1000,  # Adjust if you have more patterns
            )

            patterns = []
            for hit in response["hits"]["hits"]:
                pattern = hit["_source"]["index-pattern"]
                pattern["id"] = hit["_id"].replace("index-pattern:", "")
                patterns.append(pattern)

            return patterns
        except Exception as e:
            raise Exception(f"Failed to list index patterns: {str(e)}")
