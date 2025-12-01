package confluence

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Client wraps the Confluence REST API
type Client struct {
	baseURL    string
	username   string
	token      string
	httpClient *http.Client
	ctx        context.Context
}

// NewClient creates a new Confluence API client
func NewClient(baseURL, username, token string) (*Client, error) {
	// Remove trailing slash
	baseURL = strings.TrimSuffix(baseURL, "/")

	return &Client{
		baseURL:  baseURL,
		username: username,
		token:    token,
		httpClient: &http.Client{
			Timeout: 0, // Use default timeout
		},
		ctx: context.Background(),
	}, nil
}

// makeRequest makes an HTTP request to the Confluence API
func (c *Client) makeRequest(method, endpoint string, body interface{}) (*http.Response, error) {
	// Build URL
	apiURL := fmt.Sprintf("%s/rest/api%s", c.baseURL, endpoint)
	reqURL, err := url.Parse(apiURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Create request body if needed
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(c.ctx, method, reqURL.String(), reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.SetBasicAuth(c.username, c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// TestConnection tests the connection to Confluence
func (c *Client) TestConnection() (bool, error) {
	resp, err := c.makeRequest("GET", "/user/current", nil)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

// GetCurrentUser returns the current authenticated user
func (c *Client) GetCurrentUser() (*User, error) {
	resp, err := c.makeRequest("GET", "/user/current", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get current user: %s", string(body))
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode user: %w", err)
	}

	return &user, nil
}

// GetPage retrieves a page by ID
func (c *Client) GetPage(pageID string, expand []string) (*Page, error) {
	endpoint := fmt.Sprintf("/content/%s", pageID)
	if len(expand) > 0 {
		endpoint += "?expand=" + strings.Join(expand, ",")
	}

	resp, err := c.makeRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get page: %s", string(body))
	}

	var page Page
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		return nil, fmt.Errorf("failed to decode page: %w", err)
	}

	return &page, nil
}

// GetPageContent retrieves page content in a specific representation
func (c *Client) GetPageContent(pageID, representation string) (string, error) {
	expand := []string{"body." + representation}
	page, err := c.GetPage(pageID, expand)
	if err != nil {
		return "", err
	}

	if page.Body == nil {
		return "", nil
	}

	switch representation {
	case "storage":
		if page.Body.Storage != nil {
			return page.Body.Storage.Value, nil
		}
	case "view":
		if page.Body.View != nil {
			return page.Body.View.Value, nil
		}
	case "export_view":
		if page.Body.ExportView != nil {
			return page.Body.ExportView.Value, nil
		}
	}

	return "", nil
}

// GetSpaces retrieves all spaces
func (c *Client) GetSpaces(expand []string, limit int) (*SpacesResponse, error) {
	endpoint := "/space"
	params := url.Values{}
	if len(expand) > 0 {
		params.Set("expand", strings.Join(expand, ","))
	}
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}

	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	resp, err := c.makeRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get spaces: %s", string(body))
	}

	var spacesResp SpacesResponse
	if err := json.NewDecoder(resp.Body).Decode(&spacesResp); err != nil {
		return nil, fmt.Errorf("failed to decode spaces: %w", err)
	}

	return &spacesResp, nil
}

// GetChildPages retrieves child pages of a parent page
func (c *Client) GetChildPages(pageID string, expand []string, limit int) (*PagesResponse, error) {
	endpoint := fmt.Sprintf("/content/%s/child/page", pageID)
	params := url.Values{}
	if len(expand) > 0 {
		params.Set("expand", strings.Join(expand, ","))
	}
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}

	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	resp, err := c.makeRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get child pages: %s", string(body))
	}

	var pagesResp PagesResponse
	if err := json.NewDecoder(resp.Body).Decode(&pagesResp); err != nil {
		return nil, fmt.Errorf("failed to decode child pages: %w", err)
	}

	return &pagesResp, nil
}

// GetComments retrieves comments for a page
func (c *Client) GetComments(pageID string, expand []string, limit int) (*CommentsResponse, error) {
	endpoint := fmt.Sprintf("/content/%s/child/comment", pageID)
	params := url.Values{}
	if len(expand) > 0 {
		params.Set("expand", strings.Join(expand, ","))
	}
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}

	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	resp, err := c.makeRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get comments: %s", string(body))
	}

	var commentsResp CommentsResponse
	if err := json.NewDecoder(resp.Body).Decode(&commentsResp); err != nil {
		return nil, fmt.Errorf("failed to decode comments: %w", err)
	}

	return &commentsResp, nil
}

// GetAttachments retrieves attachments for a page
func (c *Client) GetAttachments(pageID string, expand []string, limit int) (*AttachmentsResponse, error) {
	endpoint := fmt.Sprintf("/content/%s/child/attachment", pageID)
	params := url.Values{}
	if len(expand) > 0 {
		params.Set("expand", strings.Join(expand, ","))
	}
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}

	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	resp, err := c.makeRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get attachments: %s", string(body))
	}

	var attachmentsResp AttachmentsResponse
	if err := json.NewDecoder(resp.Body).Decode(&attachmentsResp); err != nil {
		return nil, fmt.Errorf("failed to decode attachments: %w", err)
	}

	return &attachmentsResp, nil
}

// BuildPageURL builds a URL to view a page in the browser
func (c *Client) BuildPageURL(pageID string) string {
	return fmt.Sprintf("%s/pages/viewpage.action?pageId=%s", c.baseURL, pageID)
}
