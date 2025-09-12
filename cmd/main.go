package main

import (
	"fmt"
	"os"

	"github.com/secrails/secrails-sizing-agent/internal/agent"
	"github.com/secrails/secrails-sizing-agent/internal/cli"
)

func main() {
	// Create CLI handler
	cliHandler := cli.New()

	// Get configuration from flags or prompts
	config, err := cliHandler.GetConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Create and run the agent with the configuration
	sizingAgent := agent.New(config)
	if err := sizingAgent.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
