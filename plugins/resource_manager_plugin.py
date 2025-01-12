from typing import Optional

import typer
from rich.console import Console
from rich.table import Table

from apis.client_api import ClientApi
from apis.core.plugin_base import PluginBase


class ResourceManagerPlugin(PluginBase):
    def setup(self) -> None:
        """Setup plugin commands"""
        self.register_command(
            name="list",
            callback=self.list_resources,
            help="List cloud resources",
        )

        self.register_command(
            name="github",
            callback=self.list_github,
            help="List GitHub resources",
        )

        self.register_command(
            name="kubernetes",
            callback=self.list_kubernetes,
            help="List Kubernetes resources",
        )

        self.register_command(
            name="pull-requests",
            callback=self.list_pull_requests,
            help="List GitHub pull requests",
        )

    def list_resources(
        self,
        namespace: str = typer.Option(None, help="Kubernetes namespace"),
        org: str = typer.Option(None, help="GitHub organization"),
    ):
        """List all cloud resources"""
        try:
            # Get cloud resources extension
            cloud = self.client.get_extension("cloud_resources")
            if not cloud:
                typer.echo("Cloud resources extension not loaded!", err=True)
                raise typer.Exit(1)

            # Get resources
            resources = cloud.get_all_resources(namespace=namespace, org=org)

            # Create rich console and table
            console = Console()

            for platform, items in resources.items():
                table = Table(title=f"{platform.upper()} Resources")
                table.add_column("Name")
                table.add_column("Type")
                table.add_column("Status")
                table.add_column("Details")

                for resource in items:
                    table.add_row(
                        resource.name,
                        resource.type,
                        resource.status,
                        str(resource.metadata),
                    )

                console.print(table)
                console.print()

        except Exception as e:
            typer.echo(f"Error: {str(e)}", err=True)
            raise typer.Exit(1)

    def list_github(
        self,
        org: str = typer.Option(None, help="GitHub organization"),
    ):
        """List GitHub resources"""
        try:
            cloud = self.client.get_extension("cloud_resources")
            if not cloud:
                typer.echo("Cloud resources extension not loaded!", err=True)
                raise typer.Exit(1)

            resources = cloud.get_github_resources(org)

            console = Console()
            table = Table(title="GitHub Resources")
            table.add_column("Repository")
            table.add_column("Status")
            table.add_column("Stars")
            table.add_column("Forks")
            table.add_column("Description")

            for resource in resources:
                table.add_row(
                    resource.name,
                    resource.status,
                    str(resource.metadata["stars"]),
                    str(resource.metadata["forks"]),
                    str(resource.metadata["description"]),
                )

            console.print(table)

        except Exception as e:
            typer.echo(f"Error: {str(e)}", err=True)
            raise typer.Exit(1)

    def list_kubernetes(
        self,
        namespace: str = typer.Argument(..., help="Kubernetes namespace"),
    ):
        """List Kubernetes resources"""
        try:
            cloud = self.client.get_extension("cloud_resources")
            if not cloud:
                typer.echo("Cloud resources extension not loaded!", err=True)
                raise typer.Exit(1)

            resources = cloud.get_kubernetes_resources(namespace)

            console = Console()
            table = Table(title=f"Kubernetes Resources - {namespace}")
            table.add_column("Name")
            table.add_column("Type")
            table.add_column("Status")
            table.add_column("Replicas")
            table.add_column("Image")

            for resource in resources:
                table.add_row(
                    resource.name,
                    resource.type,
                    resource.status,
                    str(resource.metadata["replicas"]),
                    resource.metadata["image"],
                )

            console.print(table)

        except Exception as e:
            typer.echo(f"Error: {str(e)}", err=True)
            raise typer.Exit(1)

    def list_pull_requests(
        self,
        repo: str = typer.Argument(..., help="Repository (org/repo format)"),
        state: str = typer.Option("open", help="PR state (open/closed/all)"),
        sort: str = typer.Option(
            "created", help="Sort field (created/updated/popularity)"
        ),
        direction: str = typer.Option("desc", help="Sort direction (asc/desc)"),
    ):
        """List GitHub pull requests"""
        try:
            org_name = repo.split("/")[0]
            repo_name = repo.split("/")[1]

            pull_requests = self.client.github.get_repo_pull_requests(
                org_name=org_name,
                repo_name=repo_name,
                state=state,
                sort=sort,
                direction=direction,
            )

            console = Console()
            table = Table(title=f"Pull Requests - {repo}")
            table.add_column("Number")
            table.add_column("Title")
            table.add_column("State")
            table.add_column("Author")
            table.add_column("Created")
            table.add_column("Updated")
            table.add_column("Mergeable")

            for pr in pull_requests:
                table.add_row(
                    f"#{pr['number']}",
                    pr["title"],
                    f"[{'green' if pr['state'] == 'open' else 'red'}]{pr['state']}[/]",
                    pr["user"]["login"],
                    pr["created_at"].split("T")[0],
                    pr["updated_at"].split("T")[0],
                    str(pr.get("mergeable", "unknown")),
                )

            console.print(table)

        except Exception as e:
            typer.echo(f"Error: {str(e)}", err=True)
            raise typer.Exit(1)


def setup_plugin(client: ClientApi) -> ResourceManagerPlugin:
    """Plugin setup function"""
    plugin = ResourceManagerPlugin(client)
    plugin.setup()
    return plugin
