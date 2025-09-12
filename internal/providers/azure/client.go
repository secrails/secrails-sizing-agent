package azure

import (
	"context"
	"fmt"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"

	"github.com/secrails/secrails-sizing-agent/internal/models"
	"github.com/secrails/secrails-sizing-agent/internal/providers/config"
	"github.com/secrails/secrails-sizing-agent/pkg/logging"

	"go.uber.org/zap"
)

// AzureProvider implements the Provider interface for Azure
type AzureProvider struct {
	config     config.ProviderConfig
	credential azcore.TokenCredential

	// Azure SDK clients
	subscriptionClient *armsubscriptions.Client
	resourceClient     *armresources.Client

	// Account information
	currentSubscriptionID    string
	currentSubscriptionAlias string
	subscriptions            []models.SubscriptionCount

	mu sync.RWMutex
}

// NewAzureProvider creates a new Azure provider
func NewAzureProvider(config.ProviderConfig) (*AzureProvider, error) {
	provider := &AzureProvider{
		subscriptions: []models.SubscriptionCount{},
	}

	return provider, nil
}

// Name returns the provider name
func (p *AzureProvider) Name() string {
	return "azure"
}

// Connect establishes connection to Azure
func (p *AzureProvider) Connect(ctx context.Context) error {
	logging.Info("Connecting to Azure...")

	// Extract credentials from config
	tenantID, _ := p.config.Credentials["tenant_id"].(string)
	clientID, _ := p.config.Credentials["client_id"].(string)
	clientSecret, _ := p.config.Credentials["client_secret"].(string)

	// Create credential
	cred, err := azidentity.NewClientSecretCredential(
		tenantID,
		clientID,
		clientSecret,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create Azure credential: %w", err)
	}

	p.credential = cred

	// Initialize subscription client
	subscriptionClient, err := armsubscriptions.NewClient(cred, nil)
	if err != nil {
		return fmt.Errorf("failed to create subscription client: %w", err)
	}
	p.subscriptionClient = subscriptionClient

	// Initialize resource client
	resourceClient, err := armresources.NewClient("", cred, nil)
	if err != nil {
		return fmt.Errorf("failed to create resource client: %w", err)
	}
	p.resourceClient = resourceClient

	logging.Info("Successfully connected to Azure")
	return nil
}

func (p *AzureProvider) CountResources(ctx context.Context) (*models.ResourceCount, error) {
	logging.Info("Counting Azure resources...")

	// List all subscriptions

	return &models.ResourceCount{}, nil
}

// ListResources lists all resources from Azure
func (p *AzureProvider) ListResources(ctx context.Context) ([]models.Resource, error) {
	logging.Info("Listing Azure resources...")

	// Get all subscriptions
	subscriptions, err := p.listSubscriptions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}

	logging.Info("Found subscriptions", zap.Int("count", len(subscriptions)))

	// Collect resources from all subscriptions in parallel
	var wg sync.WaitGroup
	resourcesChan := make(chan []models.Resource, len(subscriptions))
	errorsChan := make(chan error, len(subscriptions))

	for _, subscription := range subscriptions {
		wg.Add(1)
		go func(sub models.SubscriptionCount) {
			defer wg.Done()

			resources, err := p.listResourcesForSubscription(ctx, sub)
			if err != nil {
				errorsChan <- fmt.Errorf("failed to list resources for subscription %s: %w", sub.ID, err)
				return
			}

			resourcesChan <- resources
		}(subscription)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(resourcesChan)
	close(errorsChan)

	// Check for errors
	if len(errorsChan) > 0 {
		return nil, <-errorsChan
	}

	// Aggregate all resources
	var allResources []models.Resource
	for resources := range resourcesChan {
		allResources = append(allResources, resources...)
	}

	logging.Info("Azure resource listing complete", zap.Int("total_resources", len(allResources)))
	return allResources, nil
}

// listSubscriptions lists all Azure subscriptions
func (p *AzureProvider) listSubscriptions(ctx context.Context) ([]models.SubscriptionCount, error) {
	var subscriptions []models.SubscriptionCount

	pager := p.subscriptionClient.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, sub := range page.Value {
			if sub.SubscriptionID != nil && sub.DisplayName != nil {
				subscriptions = append(subscriptions, models.SubscriptionCount{
					ID:   *sub.SubscriptionID,
					Name: *sub.DisplayName,
				})
			}
		}
	}

	return subscriptions, nil
}

// listResourcesForSubscription lists resources for a specific subscription
func (p *AzureProvider) listResourcesForSubscription(ctx context.Context, subscription models.SubscriptionCount) ([]models.Resource, error) {
	logging.Debug("Listing resources for subscription",
		zap.String("subscription_id", subscription.ID),
		zap.String("subscription_name", subscription.Name))

	// Initialize clients for this subscription
	if err := p.initializeClientsForSubscription(subscription.ID); err != nil {
		return nil, fmt.Errorf("failed to initialize clients: %w", err)
	}

	var resources []models.Resource
	var wg sync.WaitGroup

	wg.Wait()

	return resources, nil
}

// initializeClientsForSubscription initializes resource clients for a subscription
func (p *AzureProvider) initializeClientsForSubscription(subscriptionID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	return nil
}

// Close closes any open connections
func (p *AzureProvider) Close() error {
	logging.Info("Closing Azure provider connections")
	// Azure SDK clients don't require explicit closing
	return nil
}
