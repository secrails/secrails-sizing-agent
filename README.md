# secrails-sizing-agent

A robust Go agent for counting and analyzing cloud resources across Azure and AWS.

## Features

- **Multi-Cloud Support**: Azure and AWS resource enumeration
- **Parallel Processing**: Concurrent resource discovery across regions and subscriptions/accounts
- **Flexible Configuration**: YAML/JSON configuration with environment variable support
- **Multiple Output Formats**: JSON, YAML, Table, CSV
- **Resource Filtering**: Filter by resource type and region
- **Mock Data Support**: Test without cloud credentials
- **Comprehensive Logging**: Structured logging with multiple levels
- **Docker Support**: Containerized deployment

## Installation

### From Source

```bash
git clone https://github.com/yourusername/cloud-resource-counter.git
cd cloud-resource-counter
make deps
make build
```

## Quick Start

### 1. Set up credentials:

```bash
export AZURE_TENANT_ID=xxx
export AZURE_CLIENT_ID=xxx
export AZURE_CLIENT_SECRET=xxx
export AWS_ACCESS_KEY_ID=xxx
export AWS_SECRET_ACCESS_KEY=xxx
```

### 2. Build the agent:

```bash
make build
```

### 3. Run with real cloud resources:

```bash
./bin/cloud-resource-counter -format table
./bin/cloud-resource-counter -format json
```

### 4. Run with mock data for testing:
```bash
./bin/cloud-resource-counter -mock -verbose
```

### 5. Run in Docker:
```bash
make docker-run
```

