package providers

import (
	"fmt"
	"strings"

	"github.com/secrails/secrails-sizing-agent/internal/providers/aws"
	"github.com/secrails/secrails-sizing-agent/internal/providers/azure"
	"github.com/secrails/secrails-sizing-agent/internal/providers/config"
)

type ProviderManager struct {
	verbose bool
}

// NewManager creates a new provider manager
func NewManager(verbose bool) *ProviderManager {
	return &ProviderManager{
		verbose: verbose,
	}
}

// GetProvider returns the appropriate provider based on the name
func (m *ProviderManager) GetProvider(providerName string) (Provider, error) {
	// Normalize provider name
	providerName = strings.ToLower(strings.TrimSpace(providerName))

	config := config.ProviderConfig{
		Provider:    providerName,
		Credentials: make(map[string]interface{}),
		Regions:     []string{},
		Resources:   []string{},
	}
	switch providerName {
	case "aws":
		return aws.NewAWSProvider(config)
	case "azure":
		return azure.NewAzureProvider(config)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", providerName)
	}
}
