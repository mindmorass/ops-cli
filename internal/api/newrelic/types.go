package newrelic

// LogEntry represents a New Relic log entry
type LogEntry struct {
	Timestamp  int64                  `json:"timestamp"`
	Message    string                 `json:"message"`
	Level      string                 `json:"level,omitempty"`
	Service    string                 `json:"service,omitempty"`
	Hostname   string                 `json:"hostname,omitempty"`
	Attributes map[string]interface{} `json:"attributes"`
}

// LogResponse represents the response from querying logs
type LogResponse struct {
	Results  []LogEntry  `json:"results"`
	Metadata LogMetadata `json:"metadata"`
}

// LogMetadata represents log query metadata
type LogMetadata struct {
	Count int `json:"count"`
}

// LogsQueryOptions represents options for querying logs
type LogsQueryOptions struct {
	Query     string
	StartTime int64
	EndTime   int64
	Since     string
	Until     string
	Limit     int
}

// Entity represents a New Relic entity
type Entity struct {
	GUID          string `json:"guid"`
	Name          string `json:"name"`
	EntityType    string `json:"entityType,omitempty"`
	Domain        string `json:"domain,omitempty"`
	AlertSeverity string `json:"alertSeverity,omitempty"`
	Permalink     string `json:"permalink,omitempty"`
}

// EntitiesOptions represents options for querying entities
type EntitiesOptions struct {
	Filter     string
	MaxResults int
}

// GraphQLResponse represents a GraphQL response
type GraphQLResponse struct {
	Data struct {
		Actor struct {
			User struct {
				Name string `json:"name"`
			} `json:"user"`
		} `json:"actor"`
	} `json:"data"`
	Errors []GraphQLError `json:"errors,omitempty"`
}

// GraphQLError represents a GraphQL error
type GraphQLError struct {
	Message string `json:"message"`
}

// GraphQLNRQLResponse represents a GraphQL NRQL query response
type GraphQLNRQLResponse struct {
	Data struct {
		Actor struct {
			Account struct {
				NRQL *struct {
					Results []map[string]interface{} `json:"results"`
				} `json:"nrql"`
			} `json:"account"`
		} `json:"actor"`
	} `json:"data"`
	Errors []GraphQLError `json:"errors,omitempty"`
}

// GraphQLEntitiesResponse represents a GraphQL entities query response
type GraphQLEntitiesResponse struct {
	Data struct {
		Actor struct {
			EntitySearch struct {
				Results struct {
					Entities []Entity `json:"entities"`
				} `json:"results"`
			} `json:"entitySearch"`
		} `json:"actor"`
	} `json:"data"`
	Errors []GraphQLError `json:"errors,omitempty"`
}
