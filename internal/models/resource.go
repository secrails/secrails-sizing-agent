package models

import "time"

// Resource represents a cloud resource
type Resource struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Type      ResourceType      `json:"type"`
	Provider  string            `json:"provider"`
	Region    string            `json:"region"`
	Tags      map[string]string `json:"tags,omitempty"`
	CreatedAt *time.Time        `json:"created_at,omitempty"`
	Status    string            `json:"status"`
	Account   string            `json:"account,omitempty"`
}

// ResourceCount represents count statistics for resources
type ResourceCount struct {
	Provider       string         `json:"provider"`
	Type           ResourceType   `json:"type"`
	DisplayName    string         `json:"display_name"`
	TotalResources int            `json:"total_resources"`
	ByLocation     map[string]int `json:"by_location"`
	ByAccount      map[string]int `json:"by_account"`
}

// AccountCount represents Azure|AWS account resource count
type AccountCount struct {
	ID            string               `json:"id"`
	Name          string               `json:"name"`
	Status        string               `json:"status"`
	ResourceCount int                  `json:"resource_count"`
	ByType        map[ResourceType]int `json:"by_type"`
}

type SizingResult struct {
	// Metadata
	Provider  string
	Timestamp time.Time

	// Your existing models
	ResourceCounts []*ResourceCount
	AccountCounts  []AccountCount

	// Totals (calculated from above)
	TotalResources int
	TotalAccounts  int
}

type ResourceDefinition struct {
	Type             string // Azure resource type (e.g., "microsoft.compute/virtualmachines")
	DisplayName      string // Human-friendly name
	Category         string // Category for grouping
	UseResourceGraph bool   // Whether to use Resource Graph for counting
}
