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

	// CountResources counts all resources and returns complete results
	CountResources(ctx context.Context) (*models.SizingResult, error)

	// Close closes any open connections
	Close() error
}
