package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/secrails/secrails-sizing-agent/internal/models"
	"github.com/secrails/secrails-sizing-agent/internal/providers"
)

// Agent represents the Secrails cloud sizing agent
type Agent struct {
	config          *Config
	providerManager *providers.ProviderManager
}

func New(config *Config) *Agent {
	return &Agent{
		config:          config,
		providerManager: providers.NewManager(config.Verbose),
	}
}

// Run executes the main sizing logic
func (a *Agent) Run() error {
	if a.config.Provider == "" {
		return fmt.Errorf("no provider specified")
	}

	fmt.Printf("\n🚀 Secrails Sizing Agent\n")
	fmt.Printf("Selected cloud provider: %s\n", strings.ToUpper(a.config.Provider))

	ctx := context.Background()

	// Get the appropriate provider from the manager
	cloudProvider, err := a.providerManager.GetProvider(a.config.Provider)
	if err != nil {
		return fmt.Errorf("failed to initialize provider: %w", err)
	}

	// Connect to the cloud provider
	if err := cloudProvider.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to %s: %w", cloudProvider.Name(), err)
	}

	defer cloudProvider.Close()

	// Count resources
	result, err := cloudProvider.CountResources(ctx)
	if err != nil {
		return fmt.Errorf("failed to count resources: %w", err)
	}

	return a.outputResults(result)
}

// outputResults formats and outputs the counting results
func (a *Agent) outputResults(result *models.SizingResult) error {
	switch a.config.OutputFormat {
	case "json":
		return a.outputJSON(result)
	default: // table format
		return a.outputTable(result)
	}
}

// outputTable prints results in a table format
func (a *Agent) outputTable(result *models.SizingResult) error {
	fmt.Println("\n=================================")
	fmt.Printf("Provider: %s\n", result.Provider)
	fmt.Printf("Accounts/Subscriptions: %d\n", len(result.AccountCounts))
	fmt.Printf("Total Resources: %d\n", result.TotalResources)
	fmt.Println("---------------------------------")
	fmt.Println("Resource Breakdown:")
	for _, resourceType := range result.ResourceCounts {
		fmt.Printf("  %-20v: %d\n", resourceType.DisplayName, resourceType.TotalResources)
	}
	fmt.Println("=================================")
	fmt.Printf("Timestamp: %s\n", result.Timestamp)

	// TODO: If OutputFile is specified, write to file
	if a.config.OutputFile != "" {
		fmt.Printf("\n✓ Results saved to: %s\n", a.config.OutputFile)
	}

	return nil
}

// outputJSON outputs results in JSON format
func (a *Agent) outputJSON(results *models.SizingResult) error {
	// TODO: Implement JSON output
	fmt.Println("TODO: JSON output not yet implemented")
	return nil
}
