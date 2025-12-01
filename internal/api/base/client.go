package base

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// AuthType represents the type of authentication
type AuthType string

const (
	AuthTypeBasic  AuthType = "basic"
	AuthTypeBearer AuthType = "bearer"
	AuthTypeAPIKey AuthType = "api-key"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Type         AuthType
	Username     string
	Password     string
	Token        string
	APIKey       string
	APIKeyHeader string
}

// ClientConfig holds client configuration
type ClientConfig struct {
	BaseURL       string
	Timeout       time.Duration
	RetryAttempts int
	RetryDelay    time.Duration
	Headers       map[string]string
}

// APIResponse represents a generic API response
type APIResponse[T any] struct {
	Data       T
	Status     int
	StatusText string
	Headers    http.Header
}

// APIError represents an API error
type APIError struct {
	Message    string
	Status     int
	StatusText string
	Code       string
	Details    interface{}
}

func (e *APIError) Error() string {
	return e.Message
}

// BaseClient is a generic HTTP client with retry logic and error handling
type BaseClient struct {
	config     ClientConfig
	auth       *AuthConfig
	httpClient *http.Client
}

// NewBaseClient creates a new base API client
func NewBaseClient(config ClientConfig, auth *AuthConfig) *BaseClient {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.RetryAttempts == 0 {
		config.RetryAttempts = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = 1 * time.Second
	}

	return &BaseClient{
		config: config,
		auth:   auth,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// Request makes an HTTP request with retry logic
func (c *BaseClient) Request(ctx context.Context, method, endpoint string, body interface{}, headers map[string]string) (*APIResponse[interface{}], error) {
	// Build URL
	baseURL, err := url.Parse(c.config.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	endpointURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint: %w", err)
	}

	fullURL := baseURL.ResolveReference(endpointURL)

	var lastError error
	for attempt := 0; attempt <= c.config.RetryAttempts; attempt++ {
		// Create request body
		var reqBody io.Reader
		if body != nil && method != http.MethodGet {
			bodyBytes, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal request body: %w", err)
			}
			reqBody = bytes.NewReader(bodyBytes)
		}

		// Create request
		req, err := http.NewRequestWithContext(ctx, method, fullURL.String(), reqBody)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Set headers
		req.Header.Set("Content-Type", "application/json")
		for k, v := range c.config.Headers {
			req.Header.Set(k, v)
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		// Add authentication
		c.addAuthHeaders(req)

		// Make request
		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastError = err
			if attempt < c.config.RetryAttempts {
				time.Sleep(c.config.RetryDelay * time.Duration(1<<uint(attempt)))
				continue
			}
			return nil, fmt.Errorf("request failed: %w", err)
		}

		// Read response body
		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastError = err
			if attempt < c.config.RetryAttempts {
				time.Sleep(c.config.RetryDelay * time.Duration(1<<uint(attempt)))
				continue
			}
			return nil, fmt.Errorf("failed to read response: %w", err)
		}

		// Check status code
		if resp.StatusCode >= 400 {
			// Don't retry on client errors (4xx)
			if resp.StatusCode >= 400 && resp.StatusCode < 500 {
				return nil, c.handleErrorResponse(resp, respBody)
			}

			// Retry on server errors (5xx)
			lastError = c.handleErrorResponse(resp, respBody)
			if attempt < c.config.RetryAttempts {
				time.Sleep(c.config.RetryDelay * time.Duration(1<<uint(attempt)))
				continue
			}
			return nil, lastError
		}

		// Parse response
		var data interface{}
		if len(respBody) > 0 {
			if err := json.Unmarshal(respBody, &data); err != nil {
				// If not JSON, return as string
				data = string(respBody)
			}
		}

		return &APIResponse[interface{}]{
			Data:       data,
			Status:     resp.StatusCode,
			StatusText: resp.Status,
			Headers:    resp.Header,
		}, nil
	}

	return nil, lastError
}

// Get makes a GET request
func (c *BaseClient) Get(ctx context.Context, endpoint string, headers map[string]string) (*APIResponse[interface{}], error) {
	return c.Request(ctx, http.MethodGet, endpoint, nil, headers)
}

// Post makes a POST request
func (c *BaseClient) Post(ctx context.Context, endpoint string, body interface{}, headers map[string]string) (*APIResponse[interface{}], error) {
	return c.Request(ctx, http.MethodPost, endpoint, body, headers)
}

// Put makes a PUT request
func (c *BaseClient) Put(ctx context.Context, endpoint string, body interface{}, headers map[string]string) (*APIResponse[interface{}], error) {
	return c.Request(ctx, http.MethodPut, endpoint, body, headers)
}

// Patch makes a PATCH request
func (c *BaseClient) Patch(ctx context.Context, endpoint string, body interface{}, headers map[string]string) (*APIResponse[interface{}], error) {
	return c.Request(ctx, http.MethodPatch, endpoint, body, headers)
}

// Delete makes a DELETE request
func (c *BaseClient) Delete(ctx context.Context, endpoint string, headers map[string]string) (*APIResponse[interface{}], error) {
	return c.Request(ctx, http.MethodDelete, endpoint, nil, headers)
}

// addAuthHeaders adds authentication headers to the request
func (c *BaseClient) addAuthHeaders(req *http.Request) {
	if c.auth == nil {
		return
	}

	switch c.auth.Type {
	case AuthTypeBasic:
		if c.auth.Username != "" && c.auth.Password != "" {
			credentials := base64.StdEncoding.EncodeToString([]byte(c.auth.Username + ":" + c.auth.Password))
			req.Header.Set("Authorization", "Basic "+credentials)
		}
	case AuthTypeBearer:
		if c.auth.Token != "" {
			req.Header.Set("Authorization", "Bearer "+c.auth.Token)
		}
	case AuthTypeAPIKey:
		if c.auth.APIKey != "" && c.auth.APIKeyHeader != "" {
			req.Header.Set(c.auth.APIKeyHeader, c.auth.APIKey)
		}
	}
}

// handleErrorResponse creates an APIError from an HTTP error response
func (c *BaseClient) handleErrorResponse(resp *http.Response, body []byte) *APIError {
	errorMsg := fmt.Sprintf("HTTP %d: %s", resp.StatusCode, resp.Status)
	var details interface{}

	if len(body) > 0 {
		var jsonData map[string]interface{}
		if err := json.Unmarshal(body, &jsonData); err == nil {
			details = jsonData
			if msg, ok := jsonData["message"].(string); ok {
				errorMsg = msg
			} else if err, ok := jsonData["error"].(string); ok {
				errorMsg = err
			}
		} else {
			details = string(body)
		}
	}

	return &APIError{
		Message:    errorMsg,
		Status:     resp.StatusCode,
		StatusText: resp.Status,
		Code:       fmt.Sprintf("HTTP_%d", resp.StatusCode),
		Details:    details,
	}
}

// UpdateConfig updates the client configuration
func (c *BaseClient) UpdateConfig(config ClientConfig) {
	c.config = config
	if config.Timeout > 0 {
		c.httpClient.Timeout = config.Timeout
	}
}

// UpdateAuth updates the authentication configuration
func (c *BaseClient) UpdateAuth(auth *AuthConfig) {
	c.auth = auth
}
