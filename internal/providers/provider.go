package providers

import (
	"context"

	"github.com/secrails/secrails-sizing-agent/internal/models"
)

// Provider defines the interface for cloud providers
type Provider interface {
	// Name returns the provider name
	Name() string

	// Connect establishes connection to the cloud provider
	Connect(ctx context.Context) error

	// ListResources lists all resources from the provider
	ListResources(ctx context.Context) ([]models.Resource, error)

	CountResources(ctx context.Context) (*models.ResourceCount, error)

	// Close closes any open connections
	Close() error
}
