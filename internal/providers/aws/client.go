package aws

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConf "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
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
	config    config.ProviderConfig
	awsConfig aws.Config

	// AWS SDK clients
	stsClient      *sts.Client
	orgClient      *organizations.Client
	taggingClients map[string]*resourcegroupstaggingapi.Client

	// Account information
	currentAccount *CallerIdentity
	accounts       []models.AccountCount
	regions        []string

	// Resource collector
	collector *ResourceCollector
}

// NewAWSProvider creates a new AWS provider
func NewAWSProvider(cfg config.ProviderConfig) (*AWSProvider, error) {
	provider := &AWSProvider{
		config:         cfg,
		taggingClients: make(map[string]*resourcegroupstaggingapi.Client),
		accounts:       []models.AccountCount{},
		collector:      &ResourceCollector{},
	}

	return provider, nil
}

// Name returns the provider name
func (p *AWSProvider) Name() string {
	return "aws"
}

// Connect establishes connection to AWS
func (p *AWSProvider) Connect(ctx context.Context) error {
	// Step 1: Load AWS configuration
	if err := p.loadAWSConfig(ctx); err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Step 2: Create STS client for identity verification
	p.stsClient = sts.NewFromConfig(p.awsConfig)

	// Step 3: Verify credentials by getting caller identity
	if err := p.verifyCredentials(ctx); err != nil {
		return fmt.Errorf("failed to verify AWS credentials: %w", err)
	}

	// Step 4: Initialize Organizations client (for multi-account)
	p.orgClient = organizations.NewFromConfig(p.awsConfig)

	// Step 5: Discover accounts (if using Organizations)
	if err := p.discoverAccounts(ctx); err != nil {
		// Not fatal - might be a single account setup
		logging.Debug("Could not discover organization accounts (might be single account)", zap.Error(err))
	}

	// Step 6: Get regions to scan
	if err := p.setupRegions(ctx); err != nil {
		return fmt.Errorf("failed to setup regions: %w", err)
	}

	// Step 7: Initialize tagging clients for each region
	if err := p.initializeClients(); err != nil {
		return fmt.Errorf("failed to initialize tagging clients: %w", err)
	}

	logging.Info("✓ Connected to AWS successfully")
	logging.Info("  Account ID", zap.String("account_id", p.currentAccount.AccountID))
	logging.Info("  Regions to scan", zap.Strings("regions", p.regions))
	if len(p.accounts) > 1 {
		logging.Info("  Organization accounts found", zap.Int("count", len(p.accounts)))
	}

	return nil
}

func (p *AWSProvider) loadAWSConfig(ctx context.Context) error {
	logging.Debug("Loading AWS configuration...")

	var opts []func(*awsConf.LoadOptions) error

	// Set region
	opts = append(opts, awsConf.WithRegion(p.config.Region))

	// Use specific profile if provided
	if p.config.Profile != "" {
		logging.Debug("Using AWS profile", zap.String("profile", p.config.Profile))
		opts = append(opts, awsConf.WithSharedConfigProfile(p.config.Profile))
	}

	// Load the configuration
	cfg, err := awsConf.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return fmt.Errorf("unable to load AWS SDK config: %w", err)
	}

	p.awsConfig = cfg
	return nil
}

// verifyCredentials verifies AWS credentials are valid
func (p *AWSProvider) verifyCredentials(ctx context.Context) error {
	logging.Debug("Verifying AWS credentials...")

	result, err := p.stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return fmt.Errorf("failed to get caller identity: %w", err)
	}

	p.currentAccount = &CallerIdentity{
		AccountID: *result.Account,
		UserID:    *result.UserId,
		Arn:       *result.Arn,
	}

	logging.Debug("Authenticated as", zap.String("arn", p.currentAccount.Arn))
	return nil
}

func (p *AWSProvider) initializeClients() error {
	logging.Debug("Initializing tagging clients for each region...")

	for _, region := range p.regions {
		// Create a new config for this region
		regionalConfig := p.awsConfig.Copy()
		regionalConfig.Region = region

		// Create tagging client for this region
		p.taggingClients[region] = resourcegroupstaggingapi.NewFromConfig(regionalConfig)

		logging.Debug("Initialized tagging client", zap.String("region", region))
	}

	return nil
}

func (p *AWSProvider) discoverAccounts(ctx context.Context) error {
	logging.Info("Discovering AWS accounts in the organization...")

	// Check if we're in an organization
	orgInfo, err := p.orgClient.DescribeOrganization(ctx, &organizations.DescribeOrganizationInput{})
	if err != nil {
		// Not in an organization, just use current account
		p.accounts = append(p.accounts, models.AccountCount{
			ID:   p.currentAccount.AccountID,
			Name: "Current Account",
		})
		logging.Debug("Not in an organization, using single account")
		return nil
	}

	logging.Info("Organization ID", zap.String("organization_id", *orgInfo.Organization.Id))

	// Try to list all accounts in the organization (only works for management account)
	paginator := organizations.NewListAccountsPaginator(p.orgClient, &organizations.ListAccountsInput{})

	accountsFound := false
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			// If we can't list accounts (likely a member account, not management account)
			logging.Warn("Cannot list organization accounts (requires management account permissions)",
				zap.Error(err))
			break // Don't return error, just break the loop
		}

		for _, account := range page.Accounts {
			p.accounts = append(p.accounts, models.AccountCount{
				ID:   *account.Id,
				Name: *account.Name,
			})
			logging.Debug("Added account", zap.String("id", *account.Id), zap.String("name", *account.Name))
			accountsFound = true
		}
	}

	// If no accounts were found (member account scenario), just use current account
	if !accountsFound {
		p.accounts = append(p.accounts, models.AccountCount{
			ID:   p.currentAccount.AccountID,
			Name: "Current Account (Organization Member)",
		})
		logging.Info("Using current account only (member account in organization)")
	}

	logging.Info("Found accounts", zap.Int("count", len(p.accounts)))
	return nil
}

func (p *AWSProvider) setupRegions(ctx context.Context) error {
	ec2Client := ec2.NewFromConfig(p.awsConfig)
	output, err := ec2Client.DescribeRegions(ctx, &ec2.DescribeRegionsInput{
		AllRegions: aws.Bool(false), // Changed to false - only opted-in regions
		Filters: []types.Filter{
			{
				Name:   aws.String("opt-in-status"),
				Values: []string{"opt-in-not-required", "opted-in"},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to describe regions: %w", err)
	}

	var availableRegions []string
	for _, region := range output.Regions {
		if region.RegionName != nil {
			availableRegions = append(availableRegions, *region.RegionName)
		}
	}

	logging.Debug("Available AWS regions", zap.Strings("regions", availableRegions))
	if len(p.regions) == 0 {
		p.regions = availableRegions
	}

	return nil
}

func (p *AWSProvider) CountResources(ctx context.Context) (*models.SizingResult, error) {
	logging.Info("Counting AWS resources...")

	if len(p.accounts) == 0 {
		return nil, fmt.Errorf("no accounts available to scan")
	}

	// Initialize result
	result := &models.SizingResult{
		Provider:  "AWS",
		Timestamp: time.Now(),
	}

	// Create semaphore for concurrent operations
	maxConcurrency := 5
	semaphore := make(chan struct{}, maxConcurrency)

	// Get resource types to count
	resourceTypes := p.collector.GetResourceTypesToCount()
	logging.Debug("Resource types to count", zap.Int("count", len(resourceTypes)))

	var wg sync.WaitGroup
	resourceCounts := make([]*models.ResourceCount, 0)
	resultsMu := sync.Mutex{}

	// Count each resource type
	for _, rt := range resourceTypes {
		wg.Add(1)
		go func(resourceDef models.ResourceDefinition) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Count this resource type
			count, err := p.collector.CountResourceType(ctx, resourceDef, p.regions, p.taggingClients)
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
	result.AccountCounts = p.accounts

	// Calculate totals
	for _, rc := range resourceCounts {
		result.TotalResources += rc.TotalResources
	}
	result.TotalAccounts = len(p.accounts)

	logging.Info("Resource counting completed",
		zap.Int("total_resources", result.TotalResources),
		zap.Int("resource_types_counted", len(resourceCounts)),
		zap.Int("accounts", result.TotalAccounts))

	return result, nil
}

// Close closes any open connections
func (p *AWSProvider) Close() error {
	logging.Info("Closing AWS provider connections")
	// AWS SDK clients don't require explicit closing
	return nil
}
