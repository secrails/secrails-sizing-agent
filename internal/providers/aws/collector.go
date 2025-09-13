// aws/collector.go
package aws

import (
	"context"
	"fmt"

	awsSdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	"github.com/secrails/secrails-sizing-agent/internal/models"
	"github.com/secrails/secrails-sizing-agent/pkg/logging"
	"go.uber.org/zap"
)

type ResourceCollector struct {
}

func (c *ResourceCollector) GetResourceTypesToCount() []models.ResourceDefinition {
	return []models.ResourceDefinition{
		// Compute
		{Type: "ec2:instance", DisplayName: "EC2 Instances", Category: "Compute", UseResourceGraph: false},
		{Type: "lambda:function", DisplayName: "Lambda Functions", Category: "Compute", UseResourceGraph: false},
		{Type: "ecs:cluster", DisplayName: "ECS Clusters", Category: "Containers", UseResourceGraph: false},
		{Type: "ecs:service", DisplayName: "ECS Services", Category: "Containers", UseResourceGraph: false},
		{Type: "ec2:autoscaling", DisplayName: "Auto Scaling Groups", Category: "Compute", UseResourceGraph: false},
		{Type: "lightsail:instance", DisplayName: "Lightsail Instances", Category: "Compute", UseResourceGraph: false},
		{Type: "eks:cluster", DisplayName: "EKS Clusters", Category: "Containers", UseResourceGraph: false},

		// Messaging
		{Type: "sqs:queue", DisplayName: "SQS Queues", Category: "Messaging", UseResourceGraph: false},
		{Type: "sns:topic", DisplayName: "SNS Topics", Category: "Messaging", UseResourceGraph: false},

		// Analytics
		{Type: "kinesis:stream", DisplayName: "Kinesis Streams", Category: "Analytics", UseResourceGraph: false},
		{Type: "firehose:delivery-stream", DisplayName: "Kinesis Firehose Delivery Streams", Category: "Analytics", UseResourceGraph: false},

		// Monitoring
		{Type: "cloudwatch:alarm", DisplayName: "CloudWatch Alarms", Category: "Monitoring", UseResourceGraph: false},

		// Identity & Access Management
		{Type: "iam:user", DisplayName: "IAM Users", Category: "IAM", UseResourceGraph: false},
		{Type: "iam:role", DisplayName: "IAM Roles", Category: "IAM", UseResourceGraph: false},
		{Type: "iam:group", DisplayName: "IAM Groups", Category: "IAM", UseResourceGraph: false},
		{Type: "iam:policy", DisplayName: "IAM Policies", Category: "IAM", UseResourceGraph: false},

		// Application Integration
		{Type: "stepfunctions:state-machine", DisplayName: "Step Functions State Machines", Category: "Application Integration", UseResourceGraph: false},

		// Developer Tools
		{Type: "codecommit:repository", DisplayName: "CodeCommit Repositories", Category: "Developer Tools", UseResourceGraph: false},
		{Type: "codebuild:project", DisplayName: "CodeBuild Projects", Category: "Developer Tools", UseResourceGraph: false},
		{Type: "codedeploy:application", DisplayName: "CodeDeploy Applications", Category: "Developer Tools", UseResourceGraph: false},
		{Type: "codepipeline:pipeline", DisplayName: "CodePipeline Pipelines", Category: "Developer Tools", UseResourceGraph: false},

		// Machine Learning
		{Type: "sagemaker:notebook-instance", DisplayName: "SageMaker Notebook Instances", Category: "Machine Learning", UseResourceGraph: false},
		{Type: "sagemaker:endpoint", DisplayName: "SageMaker Endpoints", Category: "Machine Learning", UseResourceGraph: false},

		// Storage
		{Type: "s3:bucket", DisplayName: "S3 Buckets", Category: "Storage", UseResourceGraph: false},
		{Type: "rds:db", DisplayName: "RDS Databases", Category: "Databases", UseResourceGraph: false},
		{Type: "dynamodb:table", DisplayName: "DynamoDB Tables", Category: "Databases", UseResourceGraph: false},
		{Type: "ebs:volume", DisplayName: "EBS Volumes", Category: "Storage", UseResourceGraph: false},
		{Type: "efs:file-system", DisplayName: "EFS File Systems", Category: "Storage", UseResourceGraph: false},
		{Type: "backup:backup-vault", DisplayName: "Backup Vaults", Category: "Storage", UseResourceGraph: false},
		{Type: "elasticache:cluster", DisplayName: "ElastiCache Clusters", Category: "Databases", UseResourceGraph: false},
		{Type: "redshift:cluster", DisplayName: "Redshift Clusters", Category: "Databases", UseResourceGraph: false},
		{Type: "neptune:db-cluster", DisplayName: "Neptune Clusters", Category: "Databases", UseResourceGraph: false},

		// Networking & Content Delivery
		{Type: "cloudfront:distribution", DisplayName: "CloudFront Distributions", Category: "Networking", UseResourceGraph: false},
		{Type: "route53:hosted-zone", DisplayName: "Route 53 Hosted Zones", Category: "Networking", UseResourceGraph: false},
		{Type: "apigateway:rest-api", DisplayName: "API Gateway REST APIs", Category: "Networking", UseResourceGraph: false},
		{Type: "apigatewayv2:api", DisplayName: "API Gateway HTTP/WebSocket APIs", Category: "Networking", UseResourceGraph: false},
		{Type: "directconnect:connection", DisplayName: "Direct Connect Connections", Category: "Networking", UseResourceGraph: false},
		{Type: "vpn:connection", DisplayName: "VPN Connections", Category: "Networking", UseResourceGraph: false},

		// Migration & Transfer
		{Type: "dms:replication-instance", DisplayName: "DMS Replication Instances", Category: "Migration & Transfer", UseResourceGraph: false},

		// Business Applications
		{Type: "workspaces:workspace", DisplayName: "WorkSpaces", Category: "Business Applications", UseResourceGraph: false},

		// Networking
		{Type: "ec2:vpc", DisplayName: "VPCs", Category: "Networking", UseResourceGraph: false},
		{Type: "elasticloadbalancing:loadbalancer", DisplayName: "Load Balancers", Category: "Networking", UseResourceGraph: false},
		{Type: "ec2:nat-gateway", DisplayName: "NAT Gateways", Category: "Networking", UseResourceGraph: false},
		{Type: "ec2:internet-gateway", DisplayName: "Internet Gateways", Category: "Networking", UseResourceGraph: false},
		{Type: "ec2:security-group", DisplayName: "Security Groups", Category: "Networking", UseResourceGraph: false},

		// Security
		{Type: "kms:key", DisplayName: "KMS Keys", Category: "Security", UseResourceGraph: false},
		{Type: "secretsmanager:secret", DisplayName: "Secrets Manager Secrets", Category: "Security", UseResourceGraph: false},
		{Type: "acm:certificate", DisplayName: "ACM Certificates", Category: "Security", UseResourceGraph: false},
		{Type: "cloudhsm:v2-cluster", DisplayName: "CloudHSM Clusters", Category: "Security", UseResourceGraph: false},
	}
}

func (c *ResourceCollector) CountResourceType(
	ctx context.Context,
	resourceDef models.ResourceDefinition,
	regions []string,
	taggingClients map[string]*resourcegroupstaggingapi.Client,
) (*models.ResourceCount, error) {

	// Initialize result
	result := &models.ResourceCount{
		Provider:    "AWS",
		Type:        models.ResourceType(resourceDef.Type),
		DisplayName: resourceDef.DisplayName,
		ByLocation:  make(map[string]int),
		ByAccount:   make(map[string]int),
	}

	// Query each region
	for _, region := range regions {
		client, exists := taggingClients[region]
		if !exists {
			logging.Warn("No tagging client for region", zap.String("region", region))
			continue
		}

		// Count resources in this region - directly use resourceDef.Type
		count, err := c.countInRegion(ctx, client, resourceDef.Type)
		if err != nil {
			logging.Error("Failed to count in region",
				zap.String("region", region),
				zap.String("type", resourceDef.Type),
				zap.Error(err))
			continue
		}

		if count > 0 {
			result.ByLocation[region] = count
			result.TotalResources += count
		}
	}

	logging.Debug("Completed counting",
		zap.String("type", resourceDef.Type),
		zap.Int("total", result.TotalResources),
		zap.Int("regions", len(result.ByLocation)))

	return result, nil
}

// Count resources in a specific region
func (c *ResourceCollector) countInRegion(
	ctx context.Context,
	client *resourcegroupstaggingapi.Client,
	resourceType string,
) (int, error) {

	count := 0
	var paginationToken *string

	for {
		input := &resourcegroupstaggingapi.GetResourcesInput{
			ResourceTypeFilters: []string{resourceType},
			PaginationToken:     paginationToken,
			ResourcesPerPage:    awsSdk.Int32(100),
		}

		output, err := client.GetResources(ctx, input)
		if err != nil {
			return 0, fmt.Errorf("failed to get resources: %w", err)
		}

		count += len(output.ResourceTagMappingList)

		// Check for more pages
		if output.PaginationToken == nil || *output.PaginationToken == "" {
			break
		}
		paginationToken = output.PaginationToken
	}

	return count, nil
}
