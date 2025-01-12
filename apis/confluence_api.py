import os
from typing import Dict, List, Optional

import requests
from atlassian import Confluence
from atlassian.errors import ApiError


class ConfluenceApi:
    def __init__(
        self,
        url: str,
        username: Optional[str] = None,
        password: Optional[str] = None,
        api_token: Optional[str] = None,
        cloud: bool = True,
    ):
        """
        Initialize Confluence API client
        Args:
            url: Confluence URL
            username: Username or email
            password: Password (for server)
            api_token: API token (for cloud)
            cloud: Whether this is a cloud instance
        """
        self.url = url
        self.username = username
        self.password = password
        self.api_token = api_token
        self.cloud = cloud

        try:
            if cloud:
                if not api_token:
                    raise ValueError("API token required for cloud instance")
                self.client = Confluence(
                    url=url,
                    username=username,
                    password=api_token,
                    api_version="cloud",
                    cloud=True,
                )
            else:
                if not password:
                    raise ValueError("Password required for server instance")
                self.client = Confluence(
                    url=url,
                    username=username,
                    password=password,
                    cloud=False,
                )
        except Exception as e:
            raise Exception(f"Failed to connect to Confluence: {str(e)}")

    def create_page(
        self,
        space_key: str,
        title: str,
        body: str,
        parent_id: Optional[str] = None,
        representation: str = "storage",
    ) -> Dict:
        """
        Create a new page
        Args:
            space_key: Space key
            title: Page title
            body: Page content
            parent_id: Parent page ID
            representation: Content format (wiki, storage, editor, view, export_view, styled_view)
        """
        try:
            page = self.client.create_page(
                space=space_key,
                title=title,
                body=body,
                parent_id=parent_id,
                representation=representation,
            )
            return self._format_page(page)
        except ApiError as e:
            raise Exception(f"Failed to create page: {str(e)}")

    def get_page(
        self,
        page_id: str,
        expand: Optional[List[str]] = None,
    ) -> Dict:
        """
        Get page by ID
        Args:
            page_id: Page ID
            expand: List of properties to expand
        """
        try:
            expand = expand or ["body.storage", "version", "space"]
            page = self.client.get_page_by_id(
                page_id=page_id,
                expand=",".join(expand),
            )
            return self._format_page(page)
        except ApiError as e:
            raise Exception(f"Failed to get page: {str(e)}")

    def update_page(
        self,
        page_id: str,
        title: str,
        body: str,
        representation: str = "storage",
    ) -> Dict:
        """
        Update existing page
        Args:
            page_id: Page ID
            title: New title
            body: New content
            representation: Content format
        """
        try:
            page = self.client.update_page(
                page_id=page_id,
                title=title,
                body=body,
                representation=representation,
            )
            return self._format_page(page)
        except ApiError as e:
            raise Exception(f"Failed to update page: {str(e)}")

    def delete_page(self, page_id: str) -> Dict:
        """
        Delete page
        Args:
            page_id: Page ID
        """
        try:
            self.client.remove_page(page_id=page_id)
            return {"message": f"Page {page_id} deleted successfully"}
        except ApiError as e:
            raise Exception(f"Failed to delete page: {str(e)}")

    def get_page_by_title(
        self,
        space_key: str,
        title: str,
        expand: Optional[List[str]] = None,
    ) -> Dict:
        """
        Get page by title
        Args:
            space_key: Space key
            title: Page title
            expand: List of properties to expand
        """
        try:
            expand = expand or ["body.storage", "version", "space"]
            page = self.client.get_page_by_title(
                space=space_key,
                title=title,
                expand=",".join(expand),
            )
            return self._format_page(page)
        except ApiError as e:
            raise Exception(f"Failed to get page: {str(e)}")

    def search_pages(
        self,
        cql: str,
        limit: int = 50,
        expand: Optional[List[str]] = None,
    ) -> List[Dict]:
        """
        Search pages using CQL
        Args:
            cql: Confluence Query Language string
            limit: Maximum number of results
            expand: List of properties to expand
        """
        try:
            expand = expand or ["body.storage", "version", "space"]
            results = self.client.cql(
                cql=cql,
                limit=limit,
                expand=",".join(expand),
            )
            return [self._format_page(page) for page in results.get("results", [])]
        except ApiError as e:
            raise Exception(f"Failed to search pages: {str(e)}")

    def get_space_content(
        self,
        space_key: str,
        content_type: str = "page",
        expand: Optional[List[str]] = None,
        limit: int = 50,
    ) -> List[Dict]:
        """
        Get all content in a space
        Args:
            space_key: Space key
            content_type: Content type (page, blogpost, comment)
            expand: List of properties to expand
            limit: Maximum number of results
        """
        try:
            expand = expand or ["body.storage", "version", "space"]
            content = self.client.get_all_pages_from_space(
                space=space_key,
                start=0,
                limit=limit,
                expand=",".join(expand),
                content_type=content_type,
            )
            return [self._format_page(page) for page in content]
        except ApiError as e:
            raise Exception(f"Failed to get space content: {str(e)}")

    def add_attachment(
        self,
        page_id: str,
        file_path: str,
        name: Optional[str] = None,
        content_type: Optional[str] = None,
    ) -> Dict:
        """
        Add attachment to page
        Args:
            page_id: Page ID
            file_path: Path to file
            name: Attachment name (default: filename)
            content_type: File content type
        """
        try:
            result = self.client.attach_file(
                filename=file_path,
                page_id=page_id,
                title=name,
                content_type=content_type,
            )
            return {
                "id": result["results"][0]["id"],
                "title": result["results"][0]["title"],
                "size": result["results"][0]["size"],
                "content_type": result["results"][0]["metadata"]["mediaType"],
            }
        except ApiError as e:
            raise Exception(f"Failed to add attachment: {str(e)}")

    def get_attachments(self, page_id: str) -> List[Dict]:
        """
        Get page attachments
        Args:
            page_id: Page ID
        """
        try:
            attachments = self.client.get_attachments_from_content(page_id)
            return [
                {
                    "id": att["id"],
                    "title": att["title"],
                    "size": att["size"],
                    "content_type": att["metadata"]["mediaType"],
                    "download_url": att["_links"]["download"],
                }
                for att in attachments.get("results", [])
            ]
        except ApiError as e:
            raise Exception(f"Failed to get attachments: {str(e)}")

    def get_space_info(self, space_key: str) -> Dict:
        """
        Get space information
        Args:
            space_key: Space key
        """
        try:
            space = self.client.get_space(space_key)
            return {
                "id": space["id"],
                "key": space["key"],
                "name": space["name"],
                "type": space["type"],
                "status": space["status"],
                "description": space.get("description", {})
                .get("plain", {})
                .get("value", ""),
                "homepage_id": space.get("homepage", {}).get("id"),
            }
        except ApiError as e:
            raise Exception(f"Failed to get space info: {str(e)}")

    def _format_page(self, page: Dict) -> Dict:
        """Helper method to format page data"""
        formatted = {
            "id": page["id"],
            "type": page["type"],
            "title": page["title"],
            "space_key": page["space"]["key"],
            "version": page["version"]["number"],
            "created_by": page["history"]["createdBy"]["displayName"],
            "created_date": page["history"]["createdDate"],
            "last_modified": page["version"]["when"],
            "last_modified_by": page["version"]["by"]["displayName"],
        }

        # Add body content if available
        if "body" in page and "storage" in page["body"]:
            formatted["content"] = page["body"]["storage"]["value"]
            formatted["content_format"] = page["body"]["storage"]["representation"]

        # Add parent page info if available
        if "ancestors" in page and page["ancestors"]:
            parent = page["ancestors"][-1]
            formatted["parent"] = {
                "id": parent["id"],
                "title": parent["title"],
            }

        # Add URL if available
        if "_links" in page:
            formatted["url"] = page["_links"].get("base") + page["_links"].get(
                "webui", ""
            )

        return formatted

    def export_page_as_pdf(
        self,
        page_id: str,
        output_path: Optional[str] = None,
    ) -> str:
        """
        Export a Confluence page as PDF
        Args:
            page_id: Page ID to export
            output_path: Optional path to save PDF (default: page_title.pdf)
        Returns:
            Path to saved PDF file
        """
        if not self.client.api_version == "cloud":
            raise Exception("This method is only supported for Confluence Cloud.")

        try:
            # Get page info for title
            page = self.client.get_page_by_id(page_id)
            title = page["title"]

            if not output_path:
                download_path = f"{os.path.expanduser("~")}/Downloads"
                output_path = f"{title.replace(' ', '_')}.pdf"
                pdf_full_path = os.path.join(download_path, output_path)

            pdf = self.client.export_page(page_id)

            # Check if the PDF response is valid
            if not pdf:
                raise Exception("No PDF data returned.")

            with open(pdf_full_path, "wb") as pdf_file:
                pdf_file.write(pdf)

            return output_path  # Return the path to the saved PDF

        except Exception as e:
            raise Exception(f"Failed to export page as PDF: {str(e)}")
