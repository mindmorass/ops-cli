from typing import Dict, List, Optional

from jira import JIRA, JIRAError


class JiraApi:
    def __init__(
        self,
        server_url: str,
        auth_method: str = "token",
        username: Optional[str] = None,
        password: Optional[str] = None,
        token: Optional[str] = None,
    ):
        """
        Initialize Jira API client
        Args:
            server_url: Jira server URL
            auth_method: Authentication method ('basic', 'token', or 'oauth')
            username: Username for basic auth
            password: Password for basic auth
            token: API token for token auth
        """
        self.server_url = server_url

        if auth_method == "basic":
            if not username or not password:
                raise ValueError("Username and password required for basic auth")
            auth = (username, password)
        elif auth_method == "token":
            if not token:
                raise ValueError("Token required for token auth")
            auth = (username or "api-token", token)
        else:
            raise ValueError("Unsupported authentication method")

        try:
            self.client = JIRA(server=server_url, basic_auth=auth)
        except JIRAError as e:
            raise Exception(f"Failed to connect to Jira: {str(e)}")

    def create_issue(
        self,
        project_key: str,
        summary: str,
        description: str,
        issue_type: str = "Task",
        priority: Optional[str] = None,
        assignee: Optional[str] = None,
        labels: Optional[List[str]] = None,
        custom_fields: Optional[Dict] = None,
    ) -> Dict:
        """
        Create a new Jira issue
        Args:
            project_key: Project key (e.g., 'PROJ')
            summary: Issue summary
            description: Issue description
            issue_type: Issue type (default: Task)
            priority: Priority name
            assignee: Assignee username
            labels: List of labels
            custom_fields: Dictionary of custom field values
        """
        try:
            fields = {
                "project": {"key": project_key},
                "summary": summary,
                "description": description,
                "issuetype": {"name": issue_type},
            }

            if priority:
                fields["priority"] = {"name": priority}
            if assignee:
                fields["assignee"] = {"name": assignee}
            if labels:
                fields["labels"] = labels
            if custom_fields:
                fields.update(custom_fields)

            issue = self.client.create_issue(fields=fields)
            return self._format_issue(issue)
        except JIRAError as e:
            raise Exception(f"Failed to create issue: {str(e)}")

    def get_issue(self, issue_key: str) -> Dict:
        """
        Get issue details
        Args:
            issue_key: Issue key (e.g., 'PROJ-123')
        """
        try:
            issue = self.client.issue(issue_key)
            return self._format_issue(issue)
        except JIRAError as e:
            raise Exception(f"Failed to get issue: {str(e)}")

    def update_issue(
        self,
        issue_key: str,
        summary: Optional[str] = None,
        description: Optional[str] = None,
        priority: Optional[str] = None,
        assignee: Optional[str] = None,
        status: Optional[str] = None,
        labels: Optional[List[str]] = None,
        custom_fields: Optional[Dict] = None,
    ) -> Dict:
        """
        Update an existing issue
        Args:
            issue_key: Issue key
            summary: New summary
            description: New description
            priority: New priority name
            assignee: New assignee username
            status: New status name
            labels: New list of labels
            custom_fields: New custom field values
        """
        try:
            issue = self.client.issue(issue_key)
            update_fields = {}

            if summary:
                update_fields["summary"] = summary
            if description:
                update_fields["description"] = description
            if priority:
                update_fields["priority"] = {"name": priority}
            if assignee:
                update_fields["assignee"] = {"name": assignee}
            if labels:
                update_fields["labels"] = labels
            if custom_fields:
                update_fields.update(custom_fields)

            issue.update(fields=update_fields)

            if status:
                self._transition_issue(issue, status)

            return self._format_issue(issue)
        except JIRAError as e:
            raise Exception(f"Failed to update issue: {str(e)}")

    def delete_issue(self, issue_key: str) -> Dict:
        """
        Delete an issue
        Args:
            issue_key: Issue key
        """
        try:
            self.client.issue(issue_key).delete()
            return {"message": f"Issue {issue_key} deleted successfully"}
        except JIRAError as e:
            raise Exception(f"Failed to delete issue: {str(e)}")

    def validate_jql(self, jql: str) -> bool:
        """
        Validate JQL query using Jira's JQL validation endpoint.
        Args:
            jql: JQL query string to validate
        Returns:
            bool: True if valid, False otherwise
        """
        try:
            # Try cloud endpoint first
            cloud_endpoint = f"{self.server_url}/rest/api/3/jql/parse"
            response = self.client._session.post(
                cloud_endpoint,
                json={"queries": [jql]},
            )

            if response.status_code == 200:
                data = response.json()
                return all(not query.get("errors") for query in data.get("queries", []))

            # Fall back to server endpoint
            server_endpoint = f"{self.server_url}/rest/api/2/search"
            response = self.client._session.post(
                server_endpoint,
                json={
                    "jql": jql,
                    "maxResults": 1,  # Minimize data transfer
                    "fields": ["key"],  # Only request key field
                },
            )

            return response.status_code == 200

        except Exception as e:
            print(f"Error validating JQL: {str(e)}")
            # If we can't validate, assume it's valid and let the search handle errors
            return True

    def search_issues(
        self,
        jql: str,
        max_results: int = 50,
        fields: Optional[List[str]] = None,
    ) -> List[Dict]:
        """
        Search for issues using JQL
        Args:
            jql: JQL query string
            max_results: Maximum number of results
            fields: List of fields to return
        """
        if not self.validate_jql(jql):
            raise ValueError("Invalid JQL query")

        try:
            issues = self.client.search_issues(
                jql_str=jql,
                maxResults=max_results,
                fields=fields or ["summary", "status", "assignee", "priority"],
            )
            return [self._format_issue(issue) for issue in issues]
        except JIRAError as e:
            raise Exception(f"Failed to search issues: {str(e)}")

    def get_my_issues(
        self, status: Optional[str] = None, project_key: Optional[str] = None
    ) -> List[Dict]:
        """
        Get issues assigned to the authenticated user
        Args:
            status: Filter by status
            project_key: Filter by project
        """
        jql_parts = ["assignee = currentUser()"]
        if status:
            jql_parts.append(f'status = "{status}"')
        if project_key:
            jql_parts.append(f'project = "{project_key}"')

        jql = " AND ".join(jql_parts)
        return self.search_issues(jql)

    def get_user_info(self, username: Optional[str] = None) -> Dict:
        """
        Get user information
        Args:
            username: Username (current user if None)
        """
        try:
            user = (
                self.client.myself() if username is None else self.client.user(username)
            )
            return {
                "username": user.name,
                "email": user.emailAddress,
                "display_name": user.displayName,
                "active": user.active,
                "timezone": user.timeZone,
                "locale": user.locale,
                "groups": [group.name for group in user.groups],
            }
        except JIRAError as e:
            raise Exception(f"Failed to get user info: {str(e)}")

    def get_available_transitions(self, issue_key: str) -> List[Dict]:
        """
        Get available status transitions for an issue
        Args:
            issue_key: Issue key
        """
        try:
            transitions = self.client.transitions(issue_key)
            return [
                {
                    "id": t["id"],
                    "name": t["name"],
                    "to_status": t["to"]["name"],
                }
                for t in transitions
            ]
        except JIRAError as e:
            raise Exception(f"Failed to get transitions: {str(e)}")

    def _transition_issue(self, issue, target_status: str) -> None:
        """Helper method to transition issue status"""
        transitions = self.client.transitions(issue)
        for t in transitions:
            if t["to"]["name"].lower() == target_status.lower():
                self.client.transition_issue(issue, t["id"])
                return
        raise Exception(f"No transition found to status: {target_status}")

    def _format_issue(self, issue) -> Dict:
        """Helper method to format issue data"""
        fields = issue.fields
        formatted = {
            "key": issue.key,
            "summary": fields.summary,
            "status": fields.status.name,
        }

        # Optional fields
        if hasattr(fields, "issue_type") and fields.issue_type:
            formatted["issue_type"] = {
                "name": fields.issue_type.name,
                "subtask": fields.issue_type.subtask,
            }
        if hasattr(fields, "project") and fields.project:
            formatted["project"] = {
                "key": fields.project.key,
                "name": fields.project.name,
            }
        if hasattr(fields, "description") and fields.description:
            formatted["description"] = fields.description

        if hasattr(fields, "priority") and fields.priority:
            formatted["priority"] = fields.priority.name

        if hasattr(fields, "assignee") and fields.assignee:
            formatted["assignee"] = {
                "account_id": fields.assignee.accountId,
                "display_name": fields.assignee.displayName,
                "email": fields.assignee.emailAddress,
            }

        # Check for created and updated fields
        if hasattr(fields, "created"):
            formatted["created"] = fields.created
        if hasattr(fields, "updated"):
            formatted["updated"] = fields.updated

        return formatted
