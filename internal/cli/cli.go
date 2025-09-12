package cli

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/secrails/secrails-sizing-agent/internal/agent"
)

// CLI handles command-line interface interactions
type CLI struct {
	reader *bufio.Reader
}

// New creates a new CLI handler
func New() *CLI {
	return &CLI{
		reader: bufio.NewReader(os.Stdin),
	}
}

// GetConfig parses flags and/or prompts user to build configuration
func (c *CLI) GetConfig() (*agent.Config, error) {
	config := &agent.Config{
		OutputFormat: "table", // default
	}

	// Parse command-line flags
	flag.StringVar(&config.Provider, "provider", "", "Cloud provider (aws or azure)")
	flag.StringVar(&config.OutputFormat, "format", "table", "Output format (json, yaml, table, csv)")
	flag.StringVar(&config.OutputFile, "output", "", "Output file path")
	flag.BoolVar(&config.Verbose, "verbose", false, "Enable verbose output")
	flag.Parse()

	// Show debug info if verbose
	if config.Verbose {
		c.printDebugInfo(config)
	}

	// If no provider specified, prompt for it
	if config.Provider == "" {
		provider, err := c.promptForProvider()
		if err != nil {
			return nil, err
		}
		config.Provider = provider
	}

	return config, nil
}

// promptForProvider prompts the user to select a provider
func (c *CLI) promptForProvider() (string, error) {
	fmt.Println("=================================")
	fmt.Println("Secrails Sizing Agent")
	fmt.Println("=================================")
	fmt.Println("\nNo provider specified. Please select:")
	fmt.Println("1. AWS")
	fmt.Println("2. Azure")
	fmt.Print("\nEnter your choice (1/2) or type 'aws'/'azure': ")

	input, err := c.reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("error reading input: %w", err)
	}

	input = strings.TrimSpace(strings.ToLower(input))

	switch input {
	case "1", "aws":
		return "aws", nil
	case "2", "azure":
		return "azure", nil
	default:
		return "", fmt.Errorf("invalid choice '%s'", input)
	}
}

// printDebugInfo prints configuration in verbose mode
func (c *CLI) printDebugInfo(config *agent.Config) {
	fmt.Println("=================================")
	fmt.Println("Secrails Sizing Agent - Debug")
	fmt.Println("=================================")
	fmt.Printf("Provider: %s\n", config.Provider)
	fmt.Printf("Format: %s\n", config.OutputFormat)
	fmt.Printf("Output file: %s\n", config.OutputFile)
	fmt.Printf("Verbose: %v\n", config.Verbose)
	fmt.Println()
}
