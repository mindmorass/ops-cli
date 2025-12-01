package jira

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	jira "github.com/andygrunwald/go-jira"
)

// Client wraps the go-jira client
type Client struct {
	client     *jira.Client
	httpClient *http.Client
	baseURL    string
	ctx        context.Context
}

// GetClient returns the underlying go-jira client
func (c *Client) GetClient() *jira.Client {
	return c.client
}

// NewClient creates a new Jira API client
func NewClient(baseURL, username, token string) (*Client, error) {
	tp := jira.BasicAuthTransport{
		Username: username,
		Password: token,
	}

	httpClient := tp.Client()
	client, err := jira.NewClient(httpClient, baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create Jira client: %w", err)
	}
	
	// Set API version to v3 if supported
	// The go-jira library should handle API version automatically, but we can try to force it

	// Remove trailing slash from baseURL
	if len(baseURL) > 0 && baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}

	return &Client{
		client:     client,
		httpClient: httpClient,
		baseURL:    baseURL,
		ctx:        context.Background(),
	}, nil
}

// SearchIssues searches for issues using JQL
// Uses the /rest/api/3/search/jql endpoint (new API v3 endpoint)
func (c *Client) SearchIssues(jql string, maxResults int, fields []string) ([]jira.Issue, error) {
	// Use the new API v3 search/jql endpoint as required by Atlassian
	searchURL := fmt.Sprintf("%s/rest/api/3/search/jql", c.baseURL)

	// Build request body for /rest/api/3/search/jql endpoint
	// According to Atlassian docs, this endpoint expects: jql, maxResults, fields (optional), nextPageToken (optional)
	requestBody := map[string]interface{}{
		"jql":        jql,
		"maxResults": maxResults,
	}
	// Fields are optional - only include if specified
	if len(fields) > 0 {
		requestBody["fields"] = fields
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create HTTP request - try POST first
	req, err := http.NewRequestWithContext(c.ctx, "POST", searchURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	
	// Add basic auth header manually (the httpClient should handle this, but ensure it's set)
	// The BasicAuthTransport should handle this, but let's make sure

	// Make request using stored HTTP client
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var searchResult struct {
		Issues []jira.Issue `json:"issues"`
		Total  int          `json:"total"`
	}
	if err := json.Unmarshal(body, &searchResult); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return searchResult.Issues, nil
}

// GetIssue retrieves a specific issue by key
func (c *Client) GetIssue(issueKey string) (*jira.Issue, error) {
	issue, _, err := c.client.Issue.Get(issueKey, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue: %w", err)
	}
	return issue, nil
}

// CreateIssue creates a new issue
func (c *Client) CreateIssue(issue *jira.Issue) (*jira.Issue, error) {
	created, _, err := c.client.Issue.Create(issue)
	if err != nil {
		return nil, fmt.Errorf("failed to create issue: %w", err)
	}
	return created, nil
}

// GetCurrentUser returns the current authenticated user
func (c *Client) GetCurrentUser() (*jira.User, error) {
	user, _, err := c.client.User.GetSelf()
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}
	return user, nil
}

// AddWorklog adds a worklog entry to an issue
func (c *Client) AddWorklog(issueKey string, worklog *jira.WorklogRecord) (*jira.WorklogRecord, error) {
	created, _, err := c.client.Issue.AddWorklogRecord(issueKey, worklog)
	if err != nil {
		return nil, fmt.Errorf("failed to add worklog: %w", err)
	}
	return created, nil
}

// GetWorklogs retrieves worklogs for an issue
func (c *Client) GetWorklogs(issueKey string) ([]jira.WorklogRecord, error) {
	worklogs, _, err := c.client.Issue.GetWorklogs(issueKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get worklogs: %w", err)
	}
	return worklogs.Worklogs, nil
}

// SearchUserWorklogs searches for worklogs by user within a date range
func (c *Client) SearchUserWorklogs(startDate, endDate string) ([]jira.Issue, error) {
	// Build JQL to find issues with worklogs in date range
	jql := fmt.Sprintf("worklogAuthor = currentUser() AND worklogDate >= %s AND worklogDate <= %s", startDate, endDate)

	issues, _, err := c.client.Issue.Search(jql, &jira.SearchOptions{
		MaxResults: 1000,
		Fields:     []string{"summary", "status", "worklog"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to search worklogs: %w", err)
	}

	// For now, return all issues found by JQL
	// The JQL filter should handle the date range
	return issues, nil
}

// AddComment adds a comment to an issue
func (c *Client) AddComment(issueKey string, comment *jira.Comment) (*jira.Comment, error) {
	created, _, err := c.client.Issue.AddComment(issueKey, comment)
	if err != nil {
		return nil, fmt.Errorf("failed to add comment: %w", err)
	}
	return created, nil
}
