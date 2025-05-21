package schema

import (
	"context"
	"encoding/json"
	"fmt"

	ory "github.com/ory/client-go"
)

// Fetcher handles retrieving schemas from Ory Network
type Fetcher struct {
	client    *ory.APIClient
	projectID string
}

// NewFetcher creates a new schema fetcher
func NewFetcher(apiKey, projectID string) *Fetcher {
	config := ory.NewConfiguration()
	config.Servers = ory.ServerConfigurations{
		{
			URL: "https://api.ory.sh",
		},
	}
	config.AddDefaultHeader("Authorization", "Bearer "+apiKey)

	return &Fetcher{
		client:    ory.NewAPIClient(config),
		projectID: projectID,
	}
}

// FetchSchema retrieves the identity schema from Ory Network
func (f *Fetcher) FetchSchema() (*Schema, error) {
	ctx := context.Background()
	
	// Get the project
	project, resp, err := f.client.ProjectApi.GetProject(ctx, f.projectID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch project: %w", err)
	}
	defer resp.Body.Close()

	// Get the identity schema from the project services
	if project.Services.Identity == nil {
		return nil, fmt.Errorf("project has no identity service configuration")
	}

	// Extract the schema from the config
	schemaData, ok := project.Services.Identity.Config["identity_schema"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid schema format in project config")
	}

	// Convert the schema data to our Schema type
	schemaBytes, err := json.Marshal(schemaData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema: %w", err)
	}

	var schema Schema
	if err := json.Unmarshal(schemaBytes, &schema); err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema: %w", err)
	}

	return &schema, nil
} 