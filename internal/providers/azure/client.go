package azure

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resourcegraph/armresourcegraph"
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
	tenantsClient       *armsubscriptions.TenantsClient
	subscriptionClient  *armsubscriptions.Client
	resourceGraphClient *armresourcegraph.Client
	resourceClients     map[string]*armresources.Client

	// Account information
	tenantID      string
	locations     []string
	subscriptions []models.AccountCount

	// Resource collector
	collector *ResourceCollector

	mu sync.RWMutex
}

// NewAzureProvider creates a new Azure provider
func NewAzureProvider(cfg config.ProviderConfig) (*AzureProvider, error) {
	provider := &AzureProvider{
		config:        cfg,
		subscriptions: []models.AccountCount{},
		collector:     &ResourceCollector{},
	}

	return provider, nil
}

// Name returns the provider name
func (p *AzureProvider) Name() string {
	return "azure"
}

func (p *AzureProvider) Connect(ctx context.Context) error {
	logging.Info("Connecting to Azure...")

	// Step 1: Setup Azure credentials
	if err := p.setupCredentials(); err != nil {
		return fmt.Errorf("failed to setup Azure credentials: %w", err)
	}

	// Step 2: Initialize clients
	if err := p.initializeClients(); err != nil {
		return fmt.Errorf("failed to initialize Azure clients: %w", err)
	}

	// Step 3: Verify credentials and get tenant info
	if err := p.verifyCredentials(ctx); err != nil {
		return fmt.Errorf("failed to verify Azure credentials: %w", err)
	}

	// Step 4: Discover subscriptions
	if err := p.discoverSubscriptions(ctx); err != nil {
		return fmt.Errorf("failed to discover Azure subscriptions: %w", err)
	}

	logging.Info("Connected to Azure successfully")
	logging.Info("Tenant ID", zap.String("tenant_id", p.tenantID))
	logging.Info("Subscriptions found", zap.Int("count", len(p.subscriptions)))
	if len(p.locations) > 0 {
		logging.Info("Locations to scan", zap.Strings("locations", p.locations))
	}

	return nil
}

// setupCredentials sets up Azure authentication
func (p *AzureProvider) setupCredentials() error {
	logging.Debug("Setting up Azure credentials...")

	var credential azcore.TokenCredential
	var err error

	// Try different authentication methods in order of preference

	// 1. First, check for Service Principal credentials in environment
	tenantID := os.Getenv("AZURE_TENANT_ID")
	clientID := os.Getenv("AZURE_CLIENT_ID")
	clientSecret := os.Getenv("AZURE_CLIENT_SECRET")

	if tenantID != "" && clientID != "" && clientSecret != "" {
		logging.Debug("Using Service Principal authentication from environment variables")
		credential, err = azidentity.NewClientSecretCredential(tenantID, clientID, clientSecret, nil)
		if err == nil {
			p.tenantID = tenantID
			p.credential = credential
			return nil
		}
		logging.Debug("Service Principal authentication failed", zap.Error(err))
	}

	// 2. Try Managed Identity (for Azure VMs, App Service, etc.)
	if os.Getenv("AZURE_USE_MANAGED_IDENTITY") == "true" {
		logging.Debug("Attempting Managed Identity authentication")
		credential, err = azidentity.NewManagedIdentityCredential(nil)
		if err == nil {
			p.credential = credential
			// Tenant ID will be discovered during verification
			return nil
		}
		logging.Debug("Managed Identity authentication failed: ", zap.Error(err))
	}

	// 3. Try Azure CLI authentication (for local development)
	logging.Debug("Attempting Azure CLI authentication")
	credential, err = azidentity.NewAzureCLICredential(nil)
	if err == nil {
		p.credential = credential
		// Tenant ID will be discovered during verification
		return nil
	}
	logging.Debug("Azure CLI authentication failed:", zap.Error(err))

	// 4. Try DefaultAzureCredential (tries multiple methods)
	logging.Debug("Attempting DefaultAzureCredential authentication")
	credential, err = azidentity.NewDefaultAzureCredential(nil)
	if err == nil {
		p.credential = credential
		return nil
	}

	return fmt.Errorf("failed to authenticate with Azure. Please ensure you have valid credentials set up. " +
		"You can use: 1) Service Principal (set AZURE_TENANT_ID, AZURE_CLIENT_ID, AZURE_CLIENT_SECRET), " +
		"2) Azure CLI (run 'az login'), or 3) Managed Identity (set AZURE_USE_MANAGED_IDENTITY=true)")
}

func (p *AzureProvider) initializeClients() error {
	// Initialize subscription client
	var err error
	p.subscriptionClient, err = armsubscriptions.NewClient(p.credential, nil)
	if err != nil {
		return fmt.Errorf("failed to create subscription client: %w", err)
	}

	// Initialize Resource Graph client for efficient querying
	p.resourceGraphClient, err = armresourcegraph.NewClient(p.credential, nil)
	if err != nil {
		return fmt.Errorf("failed to create resource graph client: %w", err)
	}

	// Initialize Tenants client
	p.tenantsClient, err = armsubscriptions.NewTenantsClient(p.credential, nil)
	if err != nil {
		return fmt.Errorf("failed to create tenants client: %w", err)
	}

	// Initialize map for resource clients
	p.resourceClients = make(map[string]*armresources.Client)

	return nil
}

func (p *AzureProvider) verifyCredentials(ctx context.Context) error {
	logging.Debug("Verifying Azure credentials...")

	// Get tenant information by listing tenants
	tenantPager := p.tenantsClient.NewListPager(nil)

	// Just get the first page to verify connection
	if tenantPager.More() {
		page, err := tenantPager.NextPage(ctx)
		if err != nil {
			// This might fail for some credential types, not fatal
			logging.Debug("Could not list tenants (may be normal): ", zap.Error(err))
			return nil
		}

		// Get the first tenant ID if available
		for _, tenant := range page.Value {
			if tenant.TenantID != nil && p.tenantID == "" {
				p.tenantID = *tenant.TenantID
				logging.Debug("Found tenant: ", zap.String("tenant_id", p.tenantID))
				break
			}
		}
	}

	return nil
}

func (p *AzureProvider) discoverSubscriptions(ctx context.Context) error {
	logging.Debug("Discovering Azure subscriptions...")

	// Check if a specific subscription is configured
	specificSubID := os.Getenv("AZURE_SUBSCRIPTION_ID")
	if p.config.SubscriptionID != "" {
		specificSubID = p.config.SubscriptionID
	}

	// List all accessible subscriptions
	pager := p.subscriptionClient.NewListPager(nil)

	subscriptionCount := 0
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list subscriptions: %w", err)
		}

		for _, sub := range page.Value {
			// Skip if we're looking for a specific subscription
			if specificSubID != "" && sub.SubscriptionID != nil && *sub.SubscriptionID != specificSubID {
				continue
			}

			// Only include enabled subscriptions
			if sub.State != nil && (*sub.State == armsubscriptions.SubscriptionStateEnabled ||
				*sub.State == armsubscriptions.SubscriptionStateWarned) {

				subscriptionCount++

				// Get safe string values
				subID := ""
				subName := ""
				subState := ""

				if sub.SubscriptionID != nil {
					subID = *sub.SubscriptionID
				}
				if sub.DisplayName != nil {
					subName = *sub.DisplayName
				}
				if sub.State != nil {
					subState = string(*sub.State)
				}

				account := models.AccountCount{
					ID:     subID,
					Name:   subName,
					Status: subState,
				}

				p.subscriptions = append(p.subscriptions, account)
				logging.Debug("Found subscription: ", zap.String("subscription_id", subID), zap.String("name", subName), zap.String("state", subState))
			}
		}
	}

	if subscriptionCount == 0 {
		return fmt.Errorf("no active Azure subscriptions found")
	}

	logging.Debug("Found active subscription(s)", zap.Int("count", subscriptionCount))
	return nil
}

func (p *AzureProvider) CountResources(ctx context.Context) (*models.SizingResult, error) {
	logging.Info("Counting Azure resources...")

	if len(p.subscriptions) == 0 {
		return nil, fmt.Errorf("no subscriptions available to scan")
	}

	// Initialize result
	result := &models.SizingResult{
		Provider:  "Azure",
		Timestamp: time.Now(),
	}

	// Create semaphore for concurrent operations
	maxConcurrency := 5
	semaphore := make(chan struct{}, maxConcurrency)

	// Get resource types to count
	resourceTypes := p.collector.GetResourceTypesToCount()
	logging.Debug("Resource types to count", zap.Int("count", len(resourceTypes)))

	// Get subscription IDs
	subscriptionIDs := make([]string, len(p.subscriptions))
	for i, sub := range p.subscriptions {
		subscriptionIDs[i] = sub.ID
	}

	var wg sync.WaitGroup
	resourceCounts := make([]*models.ResourceCount, 0)
	resultsMu := sync.Mutex{}

	// Count Resource Graph types
	for _, rt := range resourceTypes {
		if !rt.UseResourceGraph {
			continue
		}

		// Launch goroutine for each resource type
		wg.Add(1)
		go func(resourceDef models.ResourceDefinition) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Count this resource type
			count, err := p.collector.CountResourceType(ctx, resourceDef, subscriptionIDs, p.resourceGraphClient)
			if err != nil {
				logging.Error("Failed to count resource type",
					zap.String("type", resourceDef.Type),
					zap.Error(err))
				return
			}

			// Store result
			resultsMu.Lock()
			resourceCounts = append(resourceCounts, count)
			resultsMu.Unlock()
		}(rt)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Populate SizingResult
	result.ResourceCounts = resourceCounts
	result.AccountCounts = p.subscriptions // Already have this from Connect()

	// Calculate totals
	for _, rc := range resourceCounts {
		result.TotalResources += rc.TotalResources
	}
	result.TotalAccounts = len(p.subscriptions)

	logging.Info("Resource counting completed",
		zap.Int("total_resources", result.TotalResources),
		zap.Int("resource_types_counted", len(resourceCounts)),
		zap.Int("accounts", result.TotalAccounts))

	return result, nil
}

// Close closes any open connections
func (p *AzureProvider) Close() error {
	logging.Info("Closing Azure provider connections")
	// Azure SDK clients don't require explicit closing
	return nil
}
