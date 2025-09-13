package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
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

	defer func() {
		if err := cloudProvider.Close(); err != nil {
			fmt.Printf("⚠️  Warning: failed to close provider connection: %v\n", err)
		}
	}()

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
	fmt.Printf("Total Resources: %d\n", result.TotalResources)
	fmt.Printf("Accounts/Subscriptions: %d\n", len(result.AccountCounts))

	// Show per-account breakdown
	if len(result.AccountCounts) > 0 {
		fmt.Println("---------------------------------")
		fmt.Println("Per Account/Subscription:")
		for _, account := range result.AccountCounts {
			fmt.Printf("  %-30s: %d resources\n", account.Name, account.ResourceCount)
		}
	}

	// Show resource breakdown with better formatting
	fmt.Println("---------------------------------")
	fmt.Println("Resource Breakdown:")
	for _, rc := range result.ResourceCounts {
		if rc.TotalResources > 0 {
			fmt.Printf("  %-30s: %d\n", rc.DisplayName, rc.TotalResources)
			// Optionally show top regions
			if len(rc.ByLocation) > 0 && a.config.Verbose {
				fmt.Printf("    Regions: ")
				count := 0
				for loc, cnt := range rc.ByLocation {
					if count > 0 {
						fmt.Printf(", ")
					}
					fmt.Printf("%s(%d)", loc, cnt)
					count++
					if count >= 3 {
						break
					}
				}
				fmt.Println()
			}
		}
	}

	fmt.Println("=================================")
	fmt.Printf("Timestamp: %s\n", result.Timestamp)

	// Don't claim file is saved if it's not
	if a.config.OutputFile != "" {
		// Actually implement file saving or remove this
		// return saveTableToFile(a.config.OutputFile, result)
	}

	return nil
}

// outputJSON outputs results in JSON format
func (a *Agent) outputJSON(result *models.SizingResult) error {
	// Marshal the result to JSON with indentation
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal results to JSON: %w", err)
	}

	// If output file is specified, write to file
	if a.config.OutputFile != "" {
		err = os.WriteFile(a.config.OutputFile, jsonData, 0644)
		if err != nil {
			return fmt.Errorf("failed to write JSON to file: %w", err)
		}
		fmt.Printf("\n✓ Results saved to: %s\n", a.config.OutputFile)
	} else {
		// Otherwise print to stdout
		fmt.Println(string(jsonData))
	}

	return nil
}
