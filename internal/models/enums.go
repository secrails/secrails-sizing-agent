package models

type ResourceType string

const (
	// Azure Resource Types
	ResourceTypeAzureVM             ResourceType = "Azure::VM"
	ResourceTypeAzureStorageAccount ResourceType = "Azure::StorageAccount"
	ResourceTypeAzureDatabase       ResourceType = "Azure::Database"
	ResourceTypeAzureNetwork        ResourceType = "Azure::Network"
	ResourceTypeAzureKeyVault       ResourceType = "Azure::KeyVault"
	ResourceTypeAzureAppService     ResourceType = "Azure::AppService"

	// AWS Resource Types
	ResourceTypeAWSEC2      ResourceType = "AWS::EC2"
	ResourceTypeAWSS3       ResourceType = "AWS::S3"
	ResourceTypeAWSRDS      ResourceType = "AWS::RDS"
	ResourceTypeAWSVPC      ResourceType = "AWS::VPC"
	ResourceTypeAWSLambda   ResourceType = "AWS::Lambda"
	ResourceTypeAWSDynamoDB ResourceType = "AWS::DynamoDB"
)
