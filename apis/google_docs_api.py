from typing import Dict, List, Optional, Union

from google.oauth2.credentials import Credentials
from googleapiclient.discovery import build
from googleapiclient.errors import HttpError


class GoogleDocsApi:
    def __init__(self, credentials: Credentials):
        """
        Initialize Google Docs API client
        Args:
            credentials: Google OAuth2 credentials
        """
        self.service = build("docs", "v1", credentials=credentials)

    def create_document(self, title: str) -> Dict:
        """
        Create a new document
        Args:
            title: Document title
        """
        try:
            body = {"title": title}
            doc = self.service.documents().create(body=body).execute()
            return {
                "id": doc["documentId"],
                "title": doc["title"],
                "url": f"https://docs.google.com/document/d/{doc['documentId']}/edit",
            }
        except HttpError as e:
            raise Exception(f"Failed to create document: {str(e)}")

    def get_document(self, document_id: str) -> Dict:
        """
        Get document content and metadata
        Args:
            document_id: Document ID
        """
        try:
            doc = self.service.documents().get(documentId=document_id).execute()
            return {
                "title": doc["title"],
                "body": doc["body"],
                "revision_id": doc["revisionId"],
            }
        except HttpError as e:
            raise Exception(f"Failed to get document: {str(e)}")

    def insert_text(
        self,
        document_id: str,
        text: str,
        index: Optional[int] = None,
    ) -> Dict:
        """
        Insert text into document
        Args:
            document_id: Document ID
            text: Text to insert
            index: Position to insert text (end of document if None)
        """
        try:
            if index is None:
                doc = self.service.documents().get(documentId=document_id).execute()
                index = self._get_end_index(doc)

            requests = [{"insertText": {"location": {"index": index}, "text": text}}]

            result = (
                self.service.documents()
                .batchUpdate(documentId=document_id, body={"requests": requests})
                .execute()
            )
            return {
                "document_id": document_id,
                "revision_id": result["writeControl"]["requiredRevisionId"],
            }
        except HttpError as e:
            raise Exception(f"Failed to insert text: {str(e)}")

    def replace_text(
        self,
        document_id: str,
        find_text: str,
        replace_text: str,
        match_case: bool = True,
    ) -> Dict:
        """
        Replace all occurrences of text
        Args:
            document_id: Document ID
            find_text: Text to find
            replace_text: Text to replace with
            match_case: Whether to match case
        """
        try:
            requests = [
                {
                    "replaceAllText": {
                        "containsText": {
                            "text": find_text,
                            "matchCase": match_case,
                        },
                        "replaceText": replace_text,
                    }
                }
            ]

            result = (
                self.service.documents()
                .batchUpdate(documentId=document_id, body={"requests": requests})
                .execute()
            )
            return {
                "document_id": document_id,
                "revision_id": result["writeControl"]["requiredRevisionId"],
                "replacements": result.get("replies", [{}])[0]
                .get("replaceAllText", {})
                .get("occurrencesChanged", 0),
            }
        except HttpError as e:
            raise Exception(f"Failed to replace text: {str(e)}")

    def insert_image(
        self,
        document_id: str,
        image_uri: str,
        index: Optional[int] = None,
        width: Optional[float] = None,
        height: Optional[float] = None,
    ) -> Dict:
        """
        Insert image into document
        Args:
            document_id: Document ID
            image_uri: URI of the image
            index: Position to insert image (end of document if None)
            width: Optional width in points
            height: Optional height in points
        """
        try:
            if index is None:
                doc = self.service.documents().get(documentId=document_id).execute()
                index = self._get_end_index(doc)

            requests = [
                {
                    "insertInlineImage": {
                        "location": {"index": index},
                        "uri": image_uri,
                        "objectSize": (
                            {
                                "width": (
                                    {"magnitude": width, "unit": "PT"}
                                    if width
                                    else None
                                ),
                                "height": (
                                    {"magnitude": height, "unit": "PT"}
                                    if height
                                    else None
                                ),
                            }
                            if width or height
                            else None
                        ),
                    }
                }
            ]

            result = (
                self.service.documents()
                .batchUpdate(documentId=document_id, body={"requests": requests})
                .execute()
            )
            return {
                "document_id": document_id,
                "revision_id": result["writeControl"]["requiredRevisionId"],
            }
        except HttpError as e:
            raise Exception(f"Failed to insert image: {str(e)}")

    @staticmethod
    def _get_end_index(document: Dict) -> int:
        """Get the index of the end of the document"""
        return document["body"]["content"][-1]["endIndex"] - 1
