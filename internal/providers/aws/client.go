package aws

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConf "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
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
	currentAccount      *CallerIdentity
	currentAccountAlias string
	accounts            []models.AccountCount
	regions             []string

	mu sync.RWMutex
}

// NewAWSProvider creates a new AWS provider
func NewAWSProvider(cfg config.ProviderConfig) (*AWSProvider, error) {
	provider := &AWSProvider{
		config:         cfg,
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

func (p *AWSProvider) discoverAccounts(ctx context.Context) error {
	logging.Debug("Discovering AWS accounts in the organization...")

	// Check if we're in an organization
	orgInfo, err := p.orgClient.DescribeOrganization(ctx, &organizations.DescribeOrganizationInput{})
	if err != nil {
		// Not in an organization, just use current account
		p.accounts = append(p.accounts, models.AccountCount{
			ID:   p.currentAccount.AccountID,
			Name: "Current Account",
		})
		return err
	}

	logging.Debug("Organization ID", zap.String("organization_id", *orgInfo.Organization.Id))

	// List all accounts in the organization
	paginator := organizations.NewListAccountsPaginator(p.orgClient, &organizations.ListAccountsInput{})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list organization accounts: %w", err)
		}

		for _, account := range page.Accounts {
			p.accounts = append(p.accounts, models.AccountCount{
				ID:   *account.Id,
				Name: *account.Name,
			})
		}
	}

	logging.Debug("Found accounts in organization", zap.Int("count", len(p.accounts)))
	return nil
}

func (p *AWSProvider) setupRegions(ctx context.Context) error {
	if len(p.config.Regions) > 0 {
		p.regions = p.config.Regions
	}

	// use EC2.DescribeRegions to get all regions
	ec2Client := ec2.NewFromConfig(p.awsConfig)
	output, err := ec2Client.DescribeRegions(ctx, &ec2.DescribeRegionsInput{
		AllRegions: aws.Bool(true),
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

	// List all accounts

	return &models.SizingResult{}, nil
}

// Close closes any open connections
func (p *AWSProvider) Close() error {
	logging.Info("Closing AWS provider connections")
	// AWS SDK clients don't require explicit closing
	return nil
}
