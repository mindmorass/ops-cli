from typing import Dict, List, Optional, Union

from github import Github as PyGithub
from github.Organization import Organization
from github.PaginatedList import PaginatedList
from github.Repository import Repository


class GithubApi:
    def __init__(self, token: str):
        """
        Initialize GitHub API client
        Args:
            token: GitHub personal access token
        """
        self.client = PyGithub(token)

    def get_user_repos(self, username: Optional[str] = None) -> List[Dict]:
        """
        Get repositories for a user or authenticated user
        Args:
            username: Optional GitHub username. If None, returns authenticated user's repos
        """
        try:
            if username:
                user = self.client.get_user(username)
            else:
                user = self.client.get_user()

            return [
                {
                    "name": repo.name,
                    "full_name": repo.full_name,
                    "description": repo.description,
                    "private": repo.private,
                    "url": repo.html_url,
                    "default_branch": repo.default_branch,
                    "stars": repo.stargazers_count,
                    "forks": repo.forks_count,
                }
                for repo in user.get_repos()
            ]
        except Exception as e:
            raise Exception(f"Error fetching repositories: {str(e)}")

    def get_org_repos(self, org_name: str) -> List[Dict]:
        """
        Get repositories for an organization
        Args:
            org_name: Organization name
        """
        try:
            org = self.client.get_organization(org_name)
            return [
                {
                    "name": repo.name,
                    "full_name": repo.full_name,
                    "description": repo.description,
                    "private": repo.private,
                    "url": repo.html_url,
                    "default_branch": repo.default_branch,
                }
                for repo in org.get_repos()
            ]
        except Exception as e:
            raise Exception(f"Error fetching organization repositories: {str(e)}")

    def create_repo(
        self,
        name: str,
        private: bool = False,
        description: Optional[str] = None,
        org_name: Optional[str] = None,
    ) -> Dict:
        """
        Create a new repository
        Args:
            name: Repository name
            private: Whether the repository should be private
            description: Optional repository description
            org_name: Optional organization name. If provided, creates repo under org
        """
        try:
            if org_name:
                org = self.client.get_organization(org_name)
                repo = org.create_repo(
                    name=name, private=private, description=description
                )
            else:
                user = self.client.get_user()
                repo = user.create_repo(
                    name=name, private=private, description=description
                )

            return {
                "name": repo.name,
                "full_name": repo.full_name,
                "description": repo.description,
                "private": repo.private,
                "url": repo.html_url,
                "clone_url": repo.clone_url,
            }
        except Exception as e:
            raise Exception(f"Error creating repository: {str(e)}")

    def delete_repo(self, repo_name: str, org_name: Optional[str] = None) -> Dict:
        """
        Delete a repository
        Args:
            repo_name: Repository name
            org_name: Optional organization name
        """
        try:
            if org_name:
                full_name = f"{org_name}/{repo_name}"
            else:
                user = self.client.get_user()
                full_name = f"{user.login}/{repo_name}"

            repo = self.client.get_repo(full_name)
            repo.delete()
            return {"message": f"Repository {full_name} deleted successfully"}
        except Exception as e:
            raise Exception(f"Error deleting repository: {str(e)}")

    def list_branches(
        self, repo_name: str, org_name: Optional[str] = None
    ) -> List[Dict]:
        """
        List branches in a repository
        Args:
            repo_name: Repository name
            org_name: Optional organization name
        """
        try:
            if org_name:
                full_name = f"{org_name}/{repo_name}"
            else:
                user = self.client.get_user()
                full_name = f"{user.login}/{repo_name}"

            repo = self.client.get_repo(full_name)
            return [
                {
                    "name": branch.name,
                    "protected": branch.protected,
                    "commit_sha": branch.commit.sha,
                }
                for branch in repo.get_branches()
            ]
        except Exception as e:
            raise Exception(f"Error listing branches: {str(e)}")

    def get_repo_issues(
        self, repo_name: str, org_name: Optional[str] = None, state: str = "open"
    ) -> List[Dict]:
        """
        Get issues for a repository
        Args:
            repo_name: Repository name
            org_name: Optional organization name
            state: Issue state (open/closed/all)
        """
        try:
            if org_name:
                full_name = f"{org_name}/{repo_name}"
            else:
                user = self.client.get_user()
                full_name = f"{user.login}/{repo_name}"

            repo = self.client.get_repo(full_name)
            return [
                {
                    "number": issue.number,
                    "title": issue.title,
                    "state": issue.state,
                    "created_at": issue.created_at.isoformat(),
                    "updated_at": issue.updated_at.isoformat(),
                    "creator": issue.user.login,
                    "assignee": issue.assignee.login if issue.assignee else None,
                    "labels": [label.name for label in issue.labels],
                }
                for issue in repo.get_issues(state=state)
            ]
        except Exception as e:
            raise Exception(f"Error fetching issues: {str(e)}")

    def update_file(
        self,
        repo_name: str,
        file_path: str,
        content: str,
        commit_message: str,
        branch: str = "main",
        org_name: Optional[str] = None,
    ) -> Dict:
        """
        Update or create a file in a repository
        Args:
            repo_name: Repository name
            file_path: Path to the file in the repository
            content: New content of the file
            commit_message: Commit message
            branch: Branch name (default: main)
            org_name: Optional organization name
        """
        try:
            if org_name:
                full_name = f"{org_name}/{repo_name}"
            else:
                user = self.client.get_user()
                full_name = f"{user.login}/{repo_name}"

            repo = self.client.get_repo(full_name)

            # Try to get the file first to check if it exists
            try:
                contents = repo.get_contents(file_path, ref=branch)
                repo.update_file(
                    path=file_path,
                    message=commit_message,
                    content=content,
                    sha=contents.sha,
                    branch=branch,
                )
                operation = "updated"
            except Exception:
                # File doesn't exist, create it
                repo.create_file(
                    path=file_path,
                    message=commit_message,
                    content=content,
                    branch=branch,
                )
                operation = "created"

            return {
                "message": f"File {file_path} {operation} successfully",
                "path": file_path,
                "branch": branch,
            }
        except Exception as e:
            raise Exception(f"Error updating file: {str(e)}")

    def update_files(
        self,
        repo_name: str,
        files: List[Dict[str, str]],
        commit_message: str,
        branch: str = "main",
        org_name: Optional[str] = None,
    ) -> Dict:
        """
        Update multiple files in a repository
        Args:
            repo_name: Repository name
            files: List of dictionaries containing file paths and contents
                  [{"path": "path/to/file", "content": "file content"}]
            commit_message: Commit message
            branch: Branch name (default: main)
            org_name: Optional organization name
        """
        try:
            if org_name:
                full_name = f"{org_name}/{repo_name}"
            else:
                user = self.client.get_user()
                full_name = f"{user.login}/{repo_name}"

            repo = self.client.get_repo(full_name)
            ref = repo.get_git_ref(f"heads/{branch}")
            base_tree = repo.get_git_tree(ref.object.sha)

            # Create blob for each file
            element_list = []
            for file in files:
                blob = repo.create_git_blob(file["content"], "utf-8")
                element = github.InputGitTreeElement(
                    path=file["path"],
                    mode="100644",
                    type="blob",
                    sha=blob.sha,
                )
                element_list.append(element)

            # Create tree and commit
            tree = repo.create_git_tree(element_list, base_tree)
            parent = repo.get_git_commit(ref.object.sha)
            commit = repo.create_git_commit(commit_message, tree, [parent])
            ref.edit(sha=commit.sha)

            return {
                "message": f"Updated {len(files)} files successfully",
                "branch": branch,
                "commit_sha": commit.sha,
            }
        except Exception as e:
            raise Exception(f"Error updating files: {str(e)}")

    def create_pull_request(
        self,
        repo_name: str,
        title: str,
        body: str,
        head_branch: str,
        base_branch: str = "main",
        org_name: Optional[str] = None,
    ) -> Dict:
        """
        Create a pull request
        Args:
            repo_name: Repository name
            title: Pull request title
            body: Pull request description
            head_branch: Source branch
            base_branch: Target branch (default: main)
            org_name: Optional organization name
        """
        try:
            if org_name:
                full_name = f"{org_name}/{repo_name}"
            else:
                user = self.client.get_user()
                full_name = f"{user.login}/{repo_name}"

            repo = self.client.get_repo(full_name)
            pr = repo.create_pull(
                title=title,
                body=body,
                head=head_branch,
                base=base_branch,
            )

            return {
                "number": pr.number,
                "title": pr.title,
                "url": pr.html_url,
                "state": pr.state,
                "created_at": pr.created_at.isoformat(),
            }
        except Exception as e:
            raise Exception(f"Error creating pull request: {str(e)}")

    def list_pull_requests(
        self,
        repo_name: str,
        state: str = "open",
        org_name: Optional[str] = None,
    ) -> List[Dict]:
        """
        List pull requests in a repository
        Args:
            repo_name: Repository name
            state: PR state (open/closed/all)
            org_name: Optional organization name
        """
        try:
            if org_name:
                full_name = f"{org_name}/{repo_name}"
            else:
                user = self.client.get_user()
                full_name = f"{user.login}/{repo_name}"

            repo = self.client.get_repo(full_name)
            return [
                {
                    "number": pr.number,
                    "title": pr.title,
                    "state": pr.state,
                    "created_at": pr.created_at.isoformat(),
                    "updated_at": pr.updated_at.isoformat(),
                    "user": pr.user.login,
                    "head_branch": pr.head.ref,
                    "base_branch": pr.base.ref,
                    "url": pr.html_url,
                }
                for pr in repo.get_pulls(state=state)
            ]
        except Exception as e:
            raise Exception(f"Error listing pull requests: {str(e)}")

    def list_packages(
        self,
        org_name: Optional[str] = None,
        package_type: Optional[str] = None,
    ) -> List[Dict]:
        """
        List packages for a user or organization
        Args:
            org_name: Optional organization name
            package_type: Optional package type filter (container, maven, npm, etc.)
        """
        try:
            if org_name:
                owner = self.client.get_organization(org_name)
            else:
                owner = self.client.get_user()

            packages = owner.get_packages()
            if package_type:
                packages = [
                    p
                    for p in packages
                    if p.package_type.lower() == package_type.lower()
                ]

            return [
                {
                    "name": package.name,
                    "package_type": package.package_type,
                    "version_count": package.version_count,
                    "visibility": package.visibility,
                    "url": package.html_url,
                }
                for package in packages
            ]
        except Exception as e:
            raise Exception(f"Error listing packages: {str(e)}")

    def list_tags(
        self,
        repo_name: str,
        org_name: Optional[str] = None,
        latest: bool = False,
    ) -> Union[List[Dict], Dict]:
        """
        List tags in a repository
        Args:
            repo_name: Repository name
            org_name: Optional organization name
            latest: If True, return only the latest tag
        """
        try:
            if org_name:
                full_name = f"{org_name}/{repo_name}"
            else:
                user = self.client.get_user()
                full_name = f"{user.login}/{repo_name}"

            repo = self.client.get_repo(full_name)
            tags = [
                {
                    "name": tag.name,
                    "commit_sha": tag.commit.sha,
                    "commit_url": tag.commit.html_url,
                    "zipball_url": tag.zipball_url,
                    "tarball_url": tag.tarball_url,
                }
                for tag in repo.get_tags()
            ]

            if latest and tags:
                return tags[0]  # GitHub returns tags in descending order
            return tags
        except Exception as e:
            raise Exception(f"Error listing tags: {str(e)}")

    def list_releases(
        self,
        repo_name: str,
        org_name: Optional[str] = None,
        latest: bool = False,
        include_drafts: bool = False,
        include_prereleases: bool = False,
    ) -> Union[List[Dict], Dict]:
        """
        List releases in a repository
        Args:
            repo_name: Repository name
            org_name: Optional organization name
            latest: If True, return only the latest release
            include_drafts: Include draft releases
            include_prereleases: Include pre-releases
        """
        try:
            if org_name:
                full_name = f"{org_name}/{repo_name}"
            else:
                user = self.client.get_user()
                full_name = f"{user.login}/{repo_name}"

            repo = self.client.get_repo(full_name)

            if latest:
                # Get latest release based on filters
                if include_prereleases:
                    release = repo.get_latest_release()
                else:
                    release = repo.get_latest_release()
                    if release.prerelease:
                        releases = repo.get_releases()
                        for r in releases:
                            if not r.prerelease:
                                release = r
                                break

                if release.draft and not include_drafts:
                    raise Exception("Latest release is a draft")

                return self._format_release(release)

            # Get all releases
            releases = [
                self._format_release(release)
                for release in repo.get_releases()
                if (include_drafts or not release.draft)
                and (include_prereleases or not release.prerelease)
            ]
            return releases
        except Exception as e:
            raise Exception(f"Error listing releases: {str(e)}")

    def _format_release(self, release) -> Dict:
        """Helper method to format release data"""
        return {
            "id": release.id,
            "tag_name": release.tag_name,
            "name": release.title,
            "body": release.body,
            "draft": release.draft,
            "prerelease": release.prerelease,
            "created_at": release.created_at.isoformat(),
            "published_at": (
                release.published_at.isoformat() if release.published_at else None
            ),
            "author": {
                "login": release.author.login,
                "url": release.author.html_url,
            },
            "url": release.html_url,
            "assets": [
                {
                    "name": asset.name,
                    "size": asset.size,
                    "download_count": asset.download_count,
                    "content_type": asset.content_type,
                    "url": asset.browser_download_url,
                }
                for asset in release.get_assets()
            ],
        }
