package newrelic

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

// Client wraps the New Relic API
type Client struct {
	baseURL    string
	apiKey     string
	accountID  string
	httpClient *http.Client
	ctx        context.Context
}

// NewClient creates a new New Relic API client
func NewClient(apiKey, accountID, region string) *Client {
	baseURL := "https://api.newrelic.com"
	if region == "EU" {
		baseURL = "https://api.eu.newrelic.com"
	}

	return &Client{
		baseURL:    baseURL,
		apiKey:     apiKey,
		accountID:  accountID,
		httpClient: &http.Client{},
		ctx:        context.Background(),
	}
}

// makeRequest makes an HTTP request to the New Relic API
func (c *Client) makeRequest(method, endpoint string, body interface{}) (*http.Response, error) {
	// Build URL
	apiURL := fmt.Sprintf("%s%s", c.baseURL, endpoint)
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
	req.Header.Set("Api-Key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// TestConnection tests the connection to New Relic
func (c *Client) TestConnection() (bool, error) {
	query := `{
		actor {
			user {
				name
			}
		}
	}`

	resp, err := c.makeGraphQLRequest(query)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result GraphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Data.Actor.User.Name != "", nil
}

// makeGraphQLRequest makes a GraphQL request
func (c *Client) makeGraphQLRequest(query string) (*http.Response, error) {
	payload := map[string]interface{}{
		"query": query,
	}
	return c.makeRequest("POST", "/graphql", payload)
}

// QueryLogs queries logs using NRQL
func (c *Client) QueryLogs(options LogsQueryOptions) (*LogResponse, error) {
	// Build NRQL query
	nrqlQuery := options.Query

	if options.Since != "" {
		nrqlQuery += fmt.Sprintf(" SINCE '%s'", options.Since)
	} else if options.StartTime > 0 {
		// Convert timestamp to ISO string
		nrqlQuery += fmt.Sprintf(" SINCE %d", options.StartTime)
	}

	if options.Until != "" {
		nrqlQuery += fmt.Sprintf(" UNTIL '%s'", options.Until)
	} else if options.EndTime > 0 {
		nrqlQuery += fmt.Sprintf(" UNTIL %d", options.EndTime)
	}

	if options.Limit > 0 {
		nrqlQuery += fmt.Sprintf(" LIMIT %d", options.Limit)
	}

	// Escape quotes in query
	nrqlQuery = strings.ReplaceAll(nrqlQuery, `"`, `\"`)

	// Build GraphQL query
	graphQLQuery := fmt.Sprintf(`{
		actor {
			account(id: %s) {
				nrql(query: "%s") {
					results
				}
			}
		}
	}`, c.accountID, nrqlQuery)

	resp, err := c.makeGraphQLRequest(graphQLQuery)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to query logs: %s", string(body))
	}

	var result GraphQLNRQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Errors) > 0 {
		errorMessages := make([]string, len(result.Errors))
		for i, err := range result.Errors {
			errorMessages[i] = err.Message
		}
		return nil, fmt.Errorf("GraphQL errors: %s", strings.Join(errorMessages, ", "))
	}

	if result.Data.Actor.Account.NRQL == nil {
		return nil, fmt.Errorf("no NRQL data returned")
	}

	// Convert results to LogEntry format
	logEntries := make([]LogEntry, len(result.Data.Actor.Account.NRQL.Results))
	for i, r := range result.Data.Actor.Account.NRQL.Results {
		logEntries[i] = convertToLogEntry(r)
	}

	return &LogResponse{
		Results: logEntries,
		Metadata: LogMetadata{
			Count: len(logEntries),
		},
	}, nil
}

// GetEntities retrieves New Relic entities
func (c *Client) GetEntities(options EntitiesOptions) ([]Entity, error) {
	// Build GraphQL query with filters
	filter := ""
	if options.Filter != "" {
		filter = fmt.Sprintf(`domain: "%s"`, strings.ToUpper(options.Filter))
	}

	limit := options.MaxResults
	if limit == 0 {
		limit = 100
	}

	query := fmt.Sprintf(`{
		actor {
			entitySearch(query: "%s") {
				results {
					entities {
						guid
						name
						entityType
						domain
						alertSeverity
						permalink
					}
				}
			}
		}
	}`, filter)

	if filter == "" {
		query = `{
			actor {
				entitySearch(query: "") {
					results {
						entities {
							guid
							name
							entityType
							domain
							alertSeverity
							permalink
						}
					}
				}
			}
		}`
	}

	resp, err := c.makeGraphQLRequest(query)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get entities: %s", string(body))
	}

	var result GraphQLEntitiesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Errors) > 0 {
		errorMessages := make([]string, len(result.Errors))
		for i, err := range result.Errors {
			errorMessages[i] = err.Message
		}
		return nil, fmt.Errorf("GraphQL errors: %s", strings.Join(errorMessages, ", "))
	}

	entities := result.Data.Actor.EntitySearch.Results.Entities
	if len(entities) > limit {
		entities = entities[:limit]
	}

	return entities, nil
}

// convertToLogEntry converts a generic map to LogEntry
func convertToLogEntry(data map[string]interface{}) LogEntry {
	entry := LogEntry{
		Attributes: make(map[string]interface{}),
	}

	// Extract common fields
	if timestamp, ok := data["timestamp"].(float64); ok {
		entry.Timestamp = int64(timestamp)
	}
	if message, ok := data["message"].(string); ok {
		entry.Message = message
	}
	if level, ok := data["level"].(string); ok {
		entry.Level = level
	}
	if service, ok := data["service"].(string); ok {
		entry.Service = service
	}
	if hostname, ok := data["hostname"].(string); ok {
		entry.Hostname = hostname
	}

	// Copy all other fields to attributes
	for k, v := range data {
		if k != "timestamp" && k != "message" && k != "level" && k != "service" && k != "hostname" {
			entry.Attributes[k] = v
		}
	}

	return entry
}
