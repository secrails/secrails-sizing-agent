package config

type ProviderConfig struct {
	Provider       string   `json:"provider" yaml:"provider"`
	Profile        string   `json:"profile" yaml:"profile"` // AWS profile or Azure credentials
	Region         string   `json:"region" yaml:"region"`
	Regions        []string `json:"regions" yaml:"regions"`
	Resources      []string `json:"resources" yaml:"resources"` // Resource types to count
	SubscriptionID string   `json:"subscription_id" yaml:"subscription_id"`
}
