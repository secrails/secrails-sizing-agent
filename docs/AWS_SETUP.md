# AWS Setup Guide for Secrails Sizing Agent

## Prerequisites

1. Go 1.21 or later
2. AWS account(s)
3. One of the following authentication methods:
   - AWS CLI configured
   - IAM access keys
   - IAM role (if running on EC2/ECS/Lambda)

## Authentication Methods

### Option 1: AWS CLI (Easiest for Development)

```bash
# Install AWS CLI if not already installed

# macOS
brew install awscli

# Linux
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip
sudo ./aws/install

# Windows
msiexec.exe /i https://awscli.amazonaws.com/AWSCLIV2.msi

# Configure AWS CLI
aws configure

# You'll be prompted for:
AWS Access Key ID: [your-access-key]
AWS Secret Access Key: [your-secret-key]
Default region name: us-east-1
Default output format: json

# Verify configuration
aws sts get-caller-identity
```

### Option 2: Environment Variables

```bash
# Set AWS credentials
export AWS_ACCESS_KEY_ID="your-access-key-id"
export AWS_SECRET_ACCESS_KEY="your-secret-access-key"
export AWS_REGION="us-east-1"

# Optional: Session token for temporary credentials
export AWS_SESSION_TOKEN="your-session-token"
```

### Option 3: AWS Profile

```bash
# Create a named profile
aws configure --profile secrails-sizing

# Use the profile
export AWS_PROFILE="secrails-sizing"
```

#### Option 4: Using .env File

Create a `.env` file in your project root:

```bash
# .env
AWS_ACCESS_KEY_ID=your-access-key-id
AWS_SECRET_ACCESS_KEY=your-secret-access-key
AWS_REGION=us-east-1
AWS_PROFILE=default  # Optional
```

### Option 5: IAM Role (for EC2/ECS/Lambda)

If running on AWS infrastructure, the agent will automatically use the instance's IAM role.

## Required IAM Permissions
The sizing agent needs at least the following permissions:
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "tag:GetResources",
        "tag:GetTagKeys",
        "tag:GetTagValues",
        "ec2:DescribeRegions",
        "ec2:DescribeInstances",
        "organizations:DescribeOrganization",
        "organizations:ListAccounts",
        "sts:GetCallerIdentity"
      ],
      "Resource": "*"
    }
  ]
}
```
## For Organization-wide Scanning
If you want to scan all accounts in an AWS Organization, you'll need:

Run from the Organization's management account, OR
Set up cross-account roles with AssumeRole permissions


