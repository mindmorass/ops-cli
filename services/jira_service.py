from typing import Dict, List, Optional

from apis.jira_api import JiraApi
from apis.opensearch_api import OpenSearchApi
from services.base_service import BaseService


class JiraService(BaseService):
    def __init__(
        self,
        url: str,
        username: str,
        token: str,
        opensearch: Optional[OpenSearchApi] = None,
        log_index: str = "jira-logs",
    ):
        """
        Initialize Jira service
        Args:
            url: Jira URL
            username: Username
            token: API token
            opensearch: OpenSearch API for logging
            log_index: Index name for logs
        """
        super().__init__(opensearch=opensearch, log_index=log_index)
        self.client = JiraApi(
            server_url=url,
            username=username,
            token=token,
        )

    def create_issue(
        self,
        project_key: str,
        summary: str,
        description: str,
        issue_type: str = "Task",
        priority: Optional[str] = None,
        assignee: Optional[str] = None,
        labels: Optional[List[str]] = None,
    ) -> Dict:
        """Create Jira issue with logging"""
        try:
            result = self.client.create_issue(
                project_key=project_key,
                summary=summary,
                description=description,
                issue_type=issue_type,
                priority=priority,
                assignee=assignee,
                labels=labels,
            )

            self.log_action(
                action="create_issue",
                status="success",
                service="jira",
                details={
                    "project": project_key,
                    "summary": summary,
                    "type": issue_type,
                    "priority": priority,
                    "assignee": assignee,
                    "labels": labels,
                    "issue_key": result.get("key"),
                },
            )
            return result
        except Exception as e:
            self.log_action(
                action="create_issue",
                status="failed",
                service="jira",
                details={
                    "project": project_key,
                    "summary": summary,
                    "type": issue_type,
                    "priority": priority,
                    "assignee": assignee,
                    "labels": labels,
                },
                error=str(e),
            )
            raise

    def transition_issue(
        self,
        issue_key: str,
        status: str,
    ) -> Dict:
        """Transition Jira issue with logging"""
        try:
            result = self.client.update_issue(
                issue_key=issue_key,
                status=status,
            )

            self.log_action(
                action="transition_issue",
                status="success",
                service="jira",
                details={
                    "issue_key": issue_key,
                    "new_status": status,
                },
            )
            return result
        except Exception as e:
            self.log_action(
                action="transition_issue",
                status="failed",
                service="jira",
                details={
                    "issue_key": issue_key,
                    "new_status": status,
                },
                error=str(e),
            )
            raise
