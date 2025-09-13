package azure

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resourcegraph/armresourcegraph"
	"github.com/secrails/secrails-sizing-agent/internal/models"
	"github.com/secrails/secrails-sizing-agent/pkg/logging"
	"go.uber.org/zap"
)

type ResourceCollector struct {
}

func (c *ResourceCollector) GetResourceTypesToCount() []models.ResourceDefinition {
	return []models.ResourceDefinition{
		{Type: "microsoft.containerservice/managedclusters", DisplayName: "AKS Clusters", Category: "Containers", UseResourceGraph: true},
		{Type: "microsoft.apimanagement/service", DisplayName: "API Management", Category: "Developer Tools", UseResourceGraph: true},
		{Type: "microsoft.web/sites", DisplayName: "App Services", Category: "Compute", UseResourceGraph: true},
		{Type: "microsoft.network/applicationgateways", DisplayName: "Application Gateways", Category: "Networking", UseResourceGraph: true},
		{Type: "microsoft.insights/components", DisplayName: "Application Insights", Category: "Analytics", UseResourceGraph: true},
		{Type: "microsoft.automation/automationaccounts", DisplayName: "Automation Accounts", Category: "Developer Tools", UseResourceGraph: true},
		{Type: "microsoft.network/azurefirewalls", DisplayName: "Azure Firewalls", Category: "Networking", UseResourceGraph: true},
		{Type: "microsoft.recoveryservices/vaults/backuppolicies", DisplayName: "Backup Policies", Category: "Storage", UseResourceGraph: true},
		{Type: "microsoft.network/bastionhosts", DisplayName: "Bastion Hosts", Category: "Networking", UseResourceGraph: true},
		{Type: "microsoft.cognitiveservices/accounts", DisplayName: "Cognitive Services", Category: "Machine Learning", UseResourceGraph: true},
		{Type: "microsoft.network/connections", DisplayName: "Connections", Category: "Networking", UseResourceGraph: true},
		{Type: "microsoft.containerinstance/containergroups", DisplayName: "Container Instances", Category: "Containers", UseResourceGraph: true},
		{Type: "microsoft.containerregistry/registries", DisplayName: "Container Registries", Category: "Containers", UseResourceGraph: true},
		{Type: "microsoft.documentdb/databaseaccounts", DisplayName: "CosmosDB Accounts", Category: "Databases", UseResourceGraph: true},
		{Type: "microsoft.datafactory/factories", DisplayName: "Data Factories", Category: "Analytics", UseResourceGraph: true},
		{Type: "microsoft.datalakestore/accounts", DisplayName: "Data Lake Store Accounts", Category: "Storage", UseResourceGraph: true},
		{Type: "microsoft.visualstudio/account/project", DisplayName: "DevOps Projects", Category: "Developer Tools", UseResourceGraph: true},
		{Type: "microsoft.eventgrid/topics", DisplayName: "Event Grid Topics", Category: "Developer Tools", UseResourceGraph: true},
		{Type: "microsoft.eventhub/namespaces", DisplayName: "Event Hub Namespaces", Category: "Analytics", UseResourceGraph: true},
		{Type: "microsoft.hdinsight/clusters", DisplayName: "HDInsight Clusters", Category: "Analytics", UseResourceGraph: true},
		{Type: "microsoft.keyvault/vaults", DisplayName: "Key Vaults", Category: "Security", UseResourceGraph: true},
		{Type: "microsoft.network/loadbalancers", DisplayName: "Load Balancers", Category: "Networking", UseResourceGraph: true},
		{Type: "microsoft.network/localnetworkgateways", DisplayName: "Local Network Gateways", Category: "Networking", UseResourceGraph: true},
		{Type: "microsoft.machinelearningservices/workspaces", DisplayName: "Machine Learning Workspaces", Category: "Machine Learning", UseResourceGraph: true},
		{Type: "microsoft.cache/redisenterprise", DisplayName: "Managed Redis Cache", Category: "Databases", UseResourceGraph: true},
		{Type: "microsoft.dbformariadb/servers", DisplayName: "MariaDB Servers", Category: "Databases", UseResourceGraph: true},
		{Type: "microsoft.dbformysql/flexibleservers", DisplayName: "MySQL Servers", Category: "Databases", UseResourceGraph: true},
		{Type: "microsoft.network/networkinterfaces", DisplayName: "Network Interfaces", Category: "Networking", UseResourceGraph: true},
		{Type: "microsoft.network/networkwatchers", DisplayName: "Network Watchers", Category: "Networking", UseResourceGraph: true},
		{Type: "microsoft.dbforpostgresql/flexibleservers", DisplayName: "PostgreSQL Servers", Category: "Databases", UseResourceGraph: true},
		{Type: "microsoft.network/privateendpoints", DisplayName: "Private Endpoints", Category: "Networking", UseResourceGraph: true},
		{Type: "microsoft.network/publicipaddresses", DisplayName: "Public IP Addresses", Category: "Networking", UseResourceGraph: true},
		{Type: "microsoft.recoveryservices/vaults", DisplayName: "Recovery Services Vaults", Category: "Storage", UseResourceGraph: true},
		{Type: "microsoft.cache/redis", DisplayName: "Redis Cache", Category: "Databases", UseResourceGraph: true},
		{Type: "microsoft.network/routetables", DisplayName: "Route Tables", Category: "Networking", UseResourceGraph: true},
		{Type: "microsoft.sql/servers/databases", DisplayName: "SQL Databases", Category: "Databases", UseResourceGraph: true},
		{Type: "microsoft.sql/servers", DisplayName: "SQL Servers", Category: "Databases", UseResourceGraph: true},
		{Type: "microsoft.storage/storageaccounts", DisplayName: "Storage Accounts", Category: "Storage", UseResourceGraph: true},
		{Type: "microsoft.compute/virtualmachines", DisplayName: "Virtual Machines", Category: "Compute", UseResourceGraph: true},
		{Type: "microsoft.network/virtualnetworks", DisplayName: "Virtual Networks", Category: "Networking", UseResourceGraph: true},
		{Type: "microsoft.network/networksecuritygroups", DisplayName: "Network Security Groups", Category: "Networking", UseResourceGraph: true},
		{Type: "microsoft.network/vpngateways", DisplayName: "VPN Gateways", Category: "Networking", UseResourceGraph: true},
	}
}

// CountResourceType counts resources for a specific resource type
func (c *ResourceCollector) CountResourceType(
	ctx context.Context,
	resourceDef models.ResourceDefinition,
	subscriptions []string,
	graphClient *armresourcegraph.Client,
) (*models.ResourceCount, error) {

	// Build query for this specific resource type
	query := fmt.Sprintf(`
		Resources
		| where type =~ "%s"
		| summarize count() by location, subscriptionId
		| project location, subscriptionId, count = count_
	`, resourceDef.Type)

	// Prepare subscription IDs
	subIDs := make([]*string, len(subscriptions))
	for i, sub := range subscriptions {
		subID := sub
		subIDs[i] = &subID
	}

	// Initialize result
	result := &models.ResourceCount{
		Provider:    "Azure",
		Type:        models.ResourceType(resourceDef.Type),
		DisplayName: resourceDef.DisplayName,
		ByLocation:  make(map[string]int),
		ByAccount:   make(map[string]int),
	}

	// Pagination loop
	var skipToken *string
	pageCount := 0
	maxPages := 10 // Safety limit

	for {
		// Create request with pagination
		resultFormat := armresourcegraph.ResultFormatObjectArray
		request := armresourcegraph.QueryRequest{
			Subscriptions: subIDs,
			Query:         &query,
			Options: &armresourcegraph.QueryRequestOptions{
				ResultFormat: &resultFormat,
				SkipToken:    skipToken,
			},
		}

		// Execute query
		response, err := graphClient.Resources(ctx, request, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to query %s (page %d): %w", resourceDef.Type, pageCount+1, err)
		}

		// Process response data
		if response.Data != nil {
			switch data := response.Data.(type) {
			case []interface{}:
				for _, item := range data {
					if row, ok := item.(map[string]interface{}); ok {
						location := ""
						subscriptionId := ""
						count := 0

						if v, ok := row["location"].(string); ok {
							location = v
						}
						if v, ok := row["subscriptionId"].(string); ok {
							subscriptionId = v
						}
						if v, ok := row["count"].(float64); ok {
							count = int(v)
						}

						// Update counts
						result.TotalResources += count
						if location != "" {
							result.ByLocation[location] += count
						}
						if subscriptionId != "" {
							result.ByAccount[subscriptionId] += count
						}
					}
				}
			}
		}

		pageCount++

		// Check for more pages
		if response.SkipToken == nil || *response.SkipToken == "" {
			break
		}
		if pageCount >= maxPages {
			logging.Warn("Reached max pages for resource type",
				zap.String("type", resourceDef.Type),
				zap.Int("pages", maxPages))
			break
		}

		skipToken = response.SkipToken
		logging.Debug("Fetching next page",
			zap.String("type", resourceDef.Type),
			zap.Int("page", pageCount+1))
	}

	logging.Debug("Completed counting",
		zap.String("type", resourceDef.Type),
		zap.Int("total", result.TotalResources),
		zap.Int("pages", pageCount))

	return result, nil
}
