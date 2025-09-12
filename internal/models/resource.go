package models

import "time"

// Resource represents a cloud resource
type Resource struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Type         ResourceType      `json:"type"`
	Provider     string            `json:"provider"`
	Region       string            `json:"region"`
	Tags         map[string]string `json:"tags,omitempty"`
	CreatedAt    *time.Time        `json:"created_at,omitempty"`
	Status       string            `json:"status"`
	Subscription string            `json:"subscription,omitempty"` // For Azure
	Account      string            `json:"account,omitempty"`      // For AWS
}

// ResourceCount represents count statistics for resources
type ResourceCount struct {
	Provider          string               `json:"provider"`
	TotalResources    int                  `json:"total_resources"`
	ResourcesByType   map[ResourceType]int `json:"resources_by_type"`
	ResourcesByRegion map[string]int       `json:"resources_by_region"`
	Subscriptions     []SubscriptionCount  `json:"subscriptions,omitempty"` // For Azure
	Accounts          []AccountCount       `json:"accounts,omitempty"`      // For AWS
	Timestamp         time.Time            `json:"timestamp"`
}

// SubscriptionCount represents Azure subscription resource count
type SubscriptionCount struct {
	ID              string               `json:"id"`
	Name            string               `json:"name"`
	ResourceCount   int                  `json:"resource_count"`
	ResourcesByType map[ResourceType]int `json:"resources_by_type"`
}

// AccountCount represents AWS account resource count
type AccountCount struct {
	ID              string               `json:"id"`
	Name            string               `json:"name"`
	ResourceCount   int                  `json:"resource_count"`
	ResourcesByType map[ResourceType]int `json:"resources_by_type"`
}
