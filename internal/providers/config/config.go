package config

type ProviderConfig struct {
	Provider    string         `json:"provider" yaml:"provider"`
	Credentials map[string]any `json:"credentials" yaml:"credentials"`
	Regions     []string       `json:"regions" yaml:"regions"`
	Resources   []string       `json:"resources" yaml:"resources"` // Resource types to count
}
