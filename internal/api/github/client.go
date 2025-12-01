package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v55/github"
	"golang.org/x/oauth2"
)

// Client wraps the GitHub API client
type Client struct {
	client *github.Client
	ctx    context.Context
}

// NewClient creates a new GitHub API client
func NewClient(token string) *Client {
	ctx := context.Background()
	var client *github.Client

	if token != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc := oauth2.NewClient(ctx, ts)
		client = github.NewClient(tc)
	} else {
		client = github.NewClient(nil)
	}

	return &Client{
		client: client,
		ctx:    ctx,
	}
}

// TestConnection tests the connection to GitHub API
func (c *Client) TestConnection() (bool, string, error) {
	user, _, err := c.client.Users.Get(c.ctx, "")
	if err != nil {
		return false, "", fmt.Errorf("failed to authenticate: %w", err)
	}

	if user.Login == nil {
		return false, "", fmt.Errorf("no user information returned")
	}

	return true, *user.Login, nil
}

// ListRepositories lists repositories based on options
func (c *Client) ListRepositories(options RepositoryListOptions) ([]*github.Repository, error) {
	var repos []*github.Repository
	var err error

	listOpts := &github.RepositoryListOptions{
		Type:      options.Type,
		Sort:      options.Sort,
		Direction: options.Direction,
		ListOptions: github.ListOptions{
			PerPage: options.PerPage,
			Page:    options.Page,
		},
	}

	if options.User != "" {
		repos, _, err = c.client.Repositories.List(c.ctx, options.User, listOpts)
	} else if options.Org != "" {
		orgOpts := &github.RepositoryListByOrgOptions{
			Type:      options.Type,
			Sort:      options.Sort,
			Direction: options.Direction,
			ListOptions: github.ListOptions{
				PerPage: options.PerPage,
				Page:    options.Page,
			},
		}
		repos, _, err = c.client.Repositories.ListByOrg(c.ctx, options.Org, orgOpts)
	} else if options.Query != "" {
		// Search repositories
		searchOpts := &github.SearchOptions{
			Sort:  options.Sort,
			Order: options.Direction,
			ListOptions: github.ListOptions{
				PerPage: options.PerPage,
				Page:    options.Page,
			},
		}
		result, _, err := c.client.Search.Repositories(c.ctx, options.Query, searchOpts)
		if err != nil {
			return nil, err
		}
		repos = result.Repositories
	} else {
		// List authenticated user's repositories
		repos, _, err = c.client.Repositories.List(c.ctx, "", listOpts)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list repositories: %w", err)
	}

	return repos, nil
}

// GetRepository gets a specific repository
func (c *Client) GetRepository(owner, repo string) (*github.Repository, error) {
	repository, _, err := c.client.Repositories.Get(c.ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}
	return repository, nil
}

// SearchFiles searches for files in a repository
func (c *Client) SearchFiles(options FileSearchOptions) ([]*github.CodeResult, error) {
	// Build search query
	query := options.Query
	query += fmt.Sprintf(" repo:%s/%s", options.Owner, options.Repo)

	if options.Path != "" {
		query += fmt.Sprintf(" path:%s", options.Path)
	}
	if options.Filename != "" {
		query += fmt.Sprintf(" filename:%s", options.Filename)
	}
	if options.Extension != "" {
		query += fmt.Sprintf(" extension:%s", options.Extension)
	}

	searchOpts := &github.SearchOptions{
		ListOptions: github.ListOptions{
			PerPage: options.PerPage,
			Page:    options.Page,
		},
	}

	result, _, err := c.client.Search.Code(c.ctx, query, searchOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to search files: %w", err)
	}

	return result.CodeResults, nil
}

// GetFileContent gets the content of a file from a repository
func (c *Client) GetFileContent(owner, repo, path, ref string) (string, error) {
	opts := &github.RepositoryContentGetOptions{}
	if ref != "" {
		opts.Ref = ref
	}

	fileContent, _, _, err := c.client.Repositories.GetContents(c.ctx, owner, repo, path, opts)
	if err != nil {
		return "", fmt.Errorf("failed to get file content: %w", err)
	}

	if fileContent == nil {
		return "", fmt.Errorf("file not found or is a directory")
	}

	content, err := fileContent.GetContent()
	if err != nil {
		return "", fmt.Errorf("failed to decode file content: %w", err)
	}

	return content, nil
}

// ListRepositoryContents lists the contents of a directory
func (c *Client) ListRepositoryContents(owner, repo, path, ref string) ([]*github.RepositoryContent, error) {
	opts := &github.RepositoryContentGetOptions{}
	if ref != "" {
		opts.Ref = ref
	}

	_, directoryContent, _, err := c.client.Repositories.GetContents(c.ctx, owner, repo, path, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list contents: %w", err)
	}

	if directoryContent == nil {
		return nil, fmt.Errorf("path not found or is a file")
	}

	return directoryContent, nil
}

// ListReleases lists releases for a repository
func (c *Client) ListReleases(owner, repo string, perPage, page int) ([]*github.RepositoryRelease, error) {
	opts := &github.ListOptions{
		PerPage: perPage,
		Page:    page,
	}

	releases, _, err := c.client.Repositories.ListReleases(c.ctx, owner, repo, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list releases: %w", err)
	}

	return releases, nil
}

// ListTags lists tags for a repository
func (c *Client) ListTags(owner, repo string, perPage, page int) ([]*github.RepositoryTag, error) {
	opts := &github.ListOptions{
		PerPage: perPage,
		Page:    page,
	}

	tags, _, err := c.client.Repositories.ListTags(c.ctx, owner, repo, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}

	return tags, nil
}

// ListWorkflows lists workflows for a repository
func (c *Client) ListWorkflows(owner, repo string) ([]*github.Workflow, error) {
	workflows, _, err := c.client.Actions.ListWorkflows(c.ctx, owner, repo, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list workflows: %w", err)
	}

	return workflows.Workflows, nil
}

// ListWorkflowRuns lists workflow runs for a repository
func (c *Client) ListWorkflowRuns(owner, repo string, perPage, page int) ([]*github.WorkflowRun, error) {
	opts := &github.ListWorkflowRunsOptions{
		ListOptions: github.ListOptions{
			PerPage: perPage,
			Page:    page,
		},
	}

	runs, _, err := c.client.Actions.ListRepositoryWorkflowRuns(c.ctx, owner, repo, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list workflow runs: %w", err)
	}

	return runs.WorkflowRuns, nil
}

// ListPackages lists packages for a user or organization
func (c *Client) ListPackages(owner string, packageType string, perPage, page int) ([]*github.Package, error) {
	opts := &github.PackageListOptions{
		PackageType: &packageType,
		ListOptions: github.ListOptions{
			PerPage: perPage,
			Page:    page,
		},
	}

	packages, _, err := c.client.Organizations.ListPackages(c.ctx, owner, opts)
	if err != nil {
		// Try as user if org fails
		packages, _, err = c.client.Users.ListPackages(c.ctx, owner, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list packages: %w", err)
		}
	}

	return packages, nil
}
