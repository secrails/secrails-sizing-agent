# Azure Setup Guide for Secrails Sizing Agent

## Prerequisites

1. Go 1.21 or later
2. Azure subscription(s)
3. One of the following authentication methods:
   - Azure CLI
   - Service Principal
   - Managed Identity (if running in Azure)

## Authentication Methods

### Option 1: Azure CLI (Easiest for Development)

```bash
# Install Azure CLI if not already installed
# macOS
brew install azure-cli

# Linux
curl -sL https://aka.ms/InstallAzureCLIDeb | sudo bash

# Windows
winget install Microsoft.AzureCLI

# Login to Azure
az login

# Verify you're logged in
az account show

# List available subscriptions
az account list --output table
```

### Option 2: Service Principal

```bash
# Create a Service Principal with Reader role
az ad sp create-for-rbac \
  --name "secrails-sizing-agent" \
  --role Reader \
  --scopes /subscriptions/{subscription-id}

# Output will look like:
# {
#   "appId": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
#   "displayName": "secrails-sizing-agent",
#   "password": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
#   "tenant": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
# }

# Set environment variables
export AZURE_TENANT_ID="your-tenant-id"
export AZURE_CLIENT_ID="your-app-id"
export AZURE_CLIENT_SECRET="your-password"

# Optional: Specific subscription
export AZURE_SUBSCRIPTION_ID="your-subscription-id"
```

### Option 3: Using .env File

Create a `.env` file in your project root:

```bash
# .env
AZURE_TENANT_ID=your-tenant-id
AZURE_CLIENT_ID=your-client-id
AZURE_CLIENT_SECRET=your-client-secret
AZURE_SUBSCRIPTION_ID=your-subscription-id  # Optional
```

Then load it:
```bash
source .env
```

## Testing the Connection

### Full Application
```bash
# Run with Azure provider
go run cmd/main.go --provider azure --verbose
```

## Troubleshooting

### Common Issues

1. **"no active Azure subscriptions found"**
   - Ensure your credentials have access to at least one subscription
   - Check subscription state: `az account list --output table`

2. **"failed to authenticate with Azure"**
   - Verify credentials are set correctly
   - For Service Principal: check all three environment variables are set
   - For Azure CLI: run `az login` again

3. **"failed to list tenants"**
   - This is often normal and non-fatal
   - Some credential types don't have permission to list tenants

### Verify Permissions

```bash
# Check current account
az account show

# List accessible subscriptions
az account list --output table

# Check Service Principal permissions (if using SP)
az role assignment list --assignee {client-id} --output table
```

## Required Permissions

The sizing agent needs **Reader** role at minimum:

```bash
# Grant Reader role to Service Principal for all subscriptions
az ad sp create-for-rbac \
  --name "secrails-sizing-agent" \
  --role Reader \
  --scopes /subscriptions/{subscription-id}

# Or grant for a specific resource group
az ad sp create-for-rbac \
  --name "secrails-sizing-agent" \
  --role Reader \
  --scopes /subscriptions/{sub-id}/resourceGroups/{rg-name}
```

## Environment Variables Reference

| Variable | Required | Description |
|----------|----------|-------------|
| `AZURE_TENANT_ID` | For SP auth | Azure AD tenant ID |
| `AZURE_CLIENT_ID` | For SP auth | Service Principal client/app ID |
| `AZURE_CLIENT_SECRET` | For SP auth | Service Principal password/secret |
| `AZURE_SUBSCRIPTION_ID` | Optional | Specific subscription to scan |
| `AZURE_USE_MANAGED_IDENTITY` | Optional | Set to "true" for Managed Identity |