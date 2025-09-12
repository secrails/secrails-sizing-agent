package aws

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/secrails/secrails-sizing-agent/internal/models"
	"github.com/secrails/secrails-sizing-agent/internal/providers/config"
	"github.com/secrails/secrails-sizing-agent/pkg/logging"

	"go.uber.org/zap"
)

// AWSProvider implements the Provider interface for AWS
type AWSProvider struct {
	awsConfig aws.Config

	// AWS SDK clients
	stsClient      *sts.Client
	orgClient      *organizations.Client
	taggingClients map[string]*resourcegroupstaggingapi.Client

	// Account information
	currentAccountID    string
	currentAccountAlias string
	accounts            []models.AccountCount

	mu sync.RWMutex
}

// NewAWSProvider creates a new AWS provider
func NewAWSProvider(config.ProviderConfig) (*AWSProvider, error) {
	provider := &AWSProvider{
		taggingClients: make(map[string]*resourcegroupstaggingapi.Client),
		accounts:       []models.AccountCount{},
	}

	return provider, nil
}

// Name returns the provider name
func (p *AWSProvider) Name() string {
	return "aws"
}

// Connect establishes connection to AWS
func (p *AWSProvider) Connect(ctx context.Context) error {
	logging.Info("Connecting to AWS...")
	return nil
}

// listAccounts lists AWS accounts in the organization
func (p *AWSProvider) listAccounts(ctx context.Context) {
	// Try to list organization accounts
	result, err := p.orgClient.ListAccounts(ctx, &organizations.ListAccountsInput{})
	if err != nil {
		// Not an org master account or no org access, use current account only
		logging.Debug("Could not list organization accounts, using current account only", zap.Error(err))
		p.accounts = []models.AccountCount{
			{
				ID:   p.currentAccountID,
				Name: p.currentAccountAlias,
			},
		}
		return
	}

	// Store organization accounts
	for _, account := range result.Accounts {
		if account.Id != nil && account.Name != nil {
			p.accounts = append(p.accounts, models.AccountCount{
				ID:   *account.Id,
				Name: *account.Name,
			})
		}
	}

	logging.Info("Found AWS accounts", zap.Int("count", len(p.accounts)))
}

func (p *AWSProvider) CountResources(ctx context.Context) (*models.ResourceCount, error) {
	logging.Info("Counting AWS resources...")

	// List all accounts

	return &models.ResourceCount{}, nil
}

// ListResources lists all resources from AWS
func (p *AWSProvider) ListResources(ctx context.Context) ([]models.Resource, error) {
	logging.Info("Listing AWS resources...")

	// Get regions to scan
	regions := p.getRegionsToScan()
	if len(regions) == 0 {
		regions = []string{"us-east-1", "us-west-2", "eu-west-1"} // Default regions
	}

	logging.Info("Scanning regions", zap.Strings("regions", regions))

	// Collect resources from all regions in parallel
	var wg sync.WaitGroup
	resourcesChan := make(chan []models.Resource, len(regions))
	errorsChan := make(chan error, len(regions))

	for _, region := range regions {
		wg.Add(1)
		go func(r string) {
			defer wg.Done()

			resources, err := p.listResourcesForRegion(ctx, r)
			if err != nil {
				errorsChan <- fmt.Errorf("failed to list resources for region %s: %w", r, err)
				return
			}

			resourcesChan <- resources
		}(region)
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

	logging.Info("AWS resource listing complete", zap.Int("total_resources", len(allResources)))
	return allResources, nil
}

// getRegionsToScan returns the list of regions to scan
func (p *AWSProvider) getRegionsToScan() []string {

	// Could dynamically fetch enabled regions using EC2 DescribeRegions
	// For now, return default regions
	return []string{
		"us-east-1",
		"us-west-2",
		"eu-west-1",
		"ap-southeast-1",
	}
}

// listResourcesForRegion lists resources for a specific region
func (p *AWSProvider) listResourcesForRegion(ctx context.Context, region string) ([]models.Resource, error) {
	logging.Debug("Listing resources for region", zap.String("region", region))

	// Initialize clients for this region
	if err := p.initializeClientsForRegion(region); err != nil {
		return nil, fmt.Errorf("failed to initialize clients: %w", err)
	}

	var resources []models.Resource
	var wg sync.WaitGroup

	wg.Wait()

	return resources, nil
}

// initializeClientsForRegion initializes AWS clients for a specific region
func (p *AWSProvider) initializeClientsForRegion(region string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Create region-specific config
	regionConfig := p.awsConfig.Copy()
	regionConfig.Region = region

	// Initialize tagging client
	if _, exists := p.taggingClients[region]; !exists {
		p.taggingClients[region] = resourcegroupstaggingapi.NewFromConfig(regionConfig)
	}

	return nil
}

// Close closes any open connections
func (p *AWSProvider) Close() error {
	logging.Info("Closing AWS provider connections")
	// AWS SDK clients don't require explicit closing
	return nil
}
