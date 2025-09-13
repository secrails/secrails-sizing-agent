# secrails-sizing-agent

A cloud resource counting and sizing tool for AWS and Azure environments. Efficiently counts and categorizes cloud resources across multiple accounts and subscriptions.

## Features

- **Multi-Cloud Support**: Azure and AWS resource enumeration
- **Parallel Processing**: Concurrent resource discovery across regions and subscriptions/accounts
- **Multi-Account/Subscription**: Scan across all accessible accounts
- **Flexible Output**: JSON, CSV, or formatted console output
- **Multiple Auth Methods**: Service principals, CLI, managed identities

## Installation

### Option 1: Download Pre-built Binary (Recommended)

Download the latest binary for your platform from [GitHub Actions](https://github.com/secrails/secrails-sizing-agent/actions):

1. Go to the [Actions tab](https://github.com/secrails/secrails-sizing-agent/actions)
2. Click on the latest successful workflow run
3. Download the artifact for your platform:
   - `secrails-sizing-agent-linux-amd64` - Linux x64
   - `secrails-sizing-agent-linux-arm64` - Linux ARM64
   - `secrails-sizing-agent-darwin-amd64` - macOS Intel
   - `secrails-sizing-agent-darwin-arm64` - macOS Apple Silicon
   - `secrails-sizing-agent-windows-amd64.exe` - Windows x64

```bash
# Make it executable (Linux/macOS)
chmod +x secrails-sizing-agent-*

# Run it
./secrails-sizing-agent-linux-amd64 --provider azure
```

### Option 2: Install with Go

```bash
go install github.com/secrails/secrails-sizing-agent/cmd@latest
```

### Option 3: Build from Source

```bash
# Clone the repository
git clone https://github.com/secrails/secrails-sizing-agent.git
cd secrails-sizing-agent

# Install dependencies
go mod download

# Build
go build -o sizing-agent ./cmd

# Run
./sizing-agent --provider azure
```
## Usage
```bash
# Basic usage
./sizing-agent --provider azure

# With options
./sizing-agent --provider azure --format json --output results.json --verbose

# Available flags
--provider string   Cloud provider (aws or azure) - required
--format string    Output format (json, csv, table, yaml) - default: table
--output string    Output file path - optional
--verbose          Enable verbose logging
```

## Supported Platforms

| Platform | Architecture  | Binary Name                             |
| -------- | ------------- | --------------------------------------- |
| Linux    | x64 (amd64)   | secrails-sizing-agent-linux-amd64       |
| Linux    | ARM64         | secrails-sizing-agent-linux-arm64       |
| macOS    | Intel (amd64) | secrails-sizing-agent-darwin-amd64      |
| macOS    | Apple Silicon | secrails-sizing-agent-darwin-arm64      |
| Windows  | x64 (amd64)   | secrails-sizing-agent-windows-amd64.exe |

## Configuration

### Azure Setup
See [Azure Setup](docs/AZURE_SETUP.md) for detailed instructions.

    **Quick Setup:**
```bash
# Using Azure CLI
az login

# OR using Service Principal
export AZURE_TENANT_ID="xxx"
export AZURE_CLIENT_ID="xxx"
export AZURE_CLIENT_SECRET="xxx"
```

### AWS Setup
See [AWS Setup](docs/AWS_SETUP.md) for detailed instructions.

**Quick Setup:**
```bash
# Using AWS CLI
aws configure

# OR using environment variables
export AWS_ACCESS_KEY_ID="xxx"
export AWS_SECRET_ACCESS_KEY="xxx"
export AWS_REGION="us-east-1"
```

## License
This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Support
For any issues or questions, please open an issue on the [GitHub Issues page](https://github.com/secrails/secrails-sizing-agent/issues).