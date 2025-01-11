from typing import Dict, List, Union

from google.oauth2.credentials import Credentials
from googleapiclient.discovery import build
from googleapiclient.errors import HttpError


class GoogleSheetsApi:
    def __init__(self, credentials: Credentials):
        """
        Initialize Google Sheets API client
        Args:
            credentials: Google OAuth2 credentials
        """
        self.service = build("sheets", "v4", credentials=credentials)

    def create_spreadsheet(self, title: str) -> Dict:
        """
        Create a new spreadsheet
        Args:
            title: Spreadsheet title
        """
        try:
            spreadsheet = {"properties": {"title": title}}
            request = self.service.spreadsheets().create(body=spreadsheet)
            response = request.execute()
            return {
                "id": response["spreadsheetId"],
                "title": response["properties"]["title"],
                "url": f"https://docs.google.com/spreadsheets/d/{response['spreadsheetId']}",
            }
        except HttpError as e:
            raise Exception(f"Failed to create spreadsheet: {str(e)}")

    def get_values(
        self,
        spreadsheet_id: str,
        range_name: str,
        value_render_option: str = "FORMATTED_VALUE",
    ) -> List[List[str]]:
        """
        Get values from a spreadsheet range
        Args:
            spreadsheet_id: Spreadsheet ID
            range_name: A1 notation range (e.g., 'Sheet1!A1:B2')
            value_render_option: How values should be rendered
        """
        try:
            result = (
                self.service.spreadsheets()
                .values()
                .get(
                    spreadsheetId=spreadsheet_id,
                    range=range_name,
                    valueRenderOption=value_render_option,
                )
                .execute()
            )
            return result.get("values", [])
        except HttpError as e:
            raise Exception(f"Failed to get values: {str(e)}")

    def update_values(
        self,
        spreadsheet_id: str,
        range_name: str,
        values: List[List[Union[str, int, float]]],
        value_input_option: str = "RAW",
    ) -> Dict:
        """
        Update values in a spreadsheet range
        Args:
            spreadsheet_id: Spreadsheet ID
            range_name: A1 notation range
            values: 2D array of values to update
            value_input_option: How input should be interpreted
        """
        try:
            body = {"values": values}
            result = (
                self.service.spreadsheets()
                .values()
                .update(
                    spreadsheetId=spreadsheet_id,
                    range=range_name,
                    valueInputOption=value_input_option,
                    body=body,
                )
                .execute()
            )
            return {
                "updated_range": result["updatedRange"],
                "updated_rows": result["updatedRows"],
                "updated_columns": result["updatedColumns"],
                "updated_cells": result["updatedCells"],
            }
        except HttpError as e:
            raise Exception(f"Failed to update values: {str(e)}")

    def append_values(
        self,
        spreadsheet_id: str,
        range_name: str,
        values: List[List[Union[str, int, float]]],
        value_input_option: str = "RAW",
    ) -> Dict:
        """
        Append values to a spreadsheet
        Args:
            spreadsheet_id: Spreadsheet ID
            range_name: A1 notation range
            values: 2D array of values to append
            value_input_option: How input should be interpreted
        """
        try:
            body = {"values": values}
            result = (
                self.service.spreadsheets()
                .values()
                .append(
                    spreadsheetId=spreadsheet_id,
                    range=range_name,
                    valueInputOption=value_input_option,
                    body=body,
                )
                .execute()
            )
            return {
                "updates": {
                    "spreadsheet_id": result["spreadsheetId"],
                    "updated_range": result["updates"]["updatedRange"],
                    "updated_rows": result["updates"]["updatedRows"],
                    "updated_columns": result["updates"]["updatedColumns"],
                }
            }
        except HttpError as e:
            raise Exception(f"Failed to append values: {str(e)}")

    def clear_values(self, spreadsheet_id: str, range_name: str) -> Dict:
        """
        Clear values in a spreadsheet range
        Args:
            spreadsheet_id: Spreadsheet ID
            range_name: A1 notation range
        """
        try:
            result = (
                self.service.spreadsheets()
                .values()
                .clear(spreadsheetId=spreadsheet_id, range=range_name)
                .execute()
            )
            return {
                "spreadsheet_id": result["spreadsheetId"],
                "cleared_range": result["clearedRange"],
            }
        except HttpError as e:
            raise Exception(f"Failed to clear values: {str(e)}")

    def get_spreadsheet_metadata(self, spreadsheet_id: str) -> Dict:
        """
        Get spreadsheet metadata
        Args:
            spreadsheet_id: Spreadsheet ID
        """
        try:
            result = (
                self.service.spreadsheets().get(spreadsheetId=spreadsheet_id).execute()
            )
            return {
                "title": result["properties"]["title"],
                "locale": result["properties"]["locale"],
                "time_zone": result["properties"]["timeZone"],
                "sheets": [
                    {
                        "id": sheet["properties"]["sheetId"],
                        "title": sheet["properties"]["title"],
                        "index": sheet["properties"]["index"],
                    }
                    for sheet in result["sheets"]
                ],
            }
        except HttpError as e:
            raise Exception(f"Failed to get spreadsheet metadata: {str(e)}")
