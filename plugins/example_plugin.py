from typing import Optional

import typer

from apis.client_api import ClientApi
from core.plugin_base import PluginBase


class ExamplePlugin(PluginBase):
    def setup(self) -> None:
        """Setup plugin commands"""
        self.register_command(
            name="hello",
            callback=self.hello_command,
            help="Say hello",
        )

        self.register_command(
            name="create-issue",
            callback=self.create_issue_command,
            help="Create Jira issue with GitHub info",
        )

    def hello_command(
        self,
        name: str = typer.Argument(..., help="Name to greet"),
        count: int = typer.Option(1, help="Number of times to greet"),
    ):
        """Example command that says hello"""
        for _ in range(count):
            typer.echo(f"Hello {name}!")

    def create_issue_command(
        self,
        repo: str = typer.Argument(..., help="GitHub repository"),
        project: str = typer.Option(..., help="Jira project key"),
        issue_type: str = typer.Option("Task", help="Jira issue type"),
    ):
        """Create Jira issue from GitHub info"""
        try:
            # Get GitHub repo info
            repo_info = self.client.github.get_repo(repo)

            # Create Jira issue
            issue = self.client.jira.create_issue(
                project_key=project,
                summary=f"GitHub: {repo_info['name']}",
                description=repo_info["description"],
                issue_type=issue_type,
            )

            typer.echo(f"Created Jira issue: {issue['key']}")
        except Exception as e:
            typer.echo(f"Error: {str(e)}", err=True)


def setup_plugin(client: ClientApi) -> ExamplePlugin:
    """Plugin setup function"""
    plugin = ExamplePlugin(client)
    plugin.setup()
    return plugin
