package generator

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"

	"bedrock-forge/internal/models"
)

// generateOpenSearchServerlessModule creates resources for OpenSearch Serverless collection with security policies
func (g *HCLGenerator) generateOpenSearchServerlessModule(body *hclwrite.Body, resource models.BaseResource) error {
	opensearchServerless, ok := resource.Spec.(models.OpenSearchServerlessSpec)
	if !ok {
		// Try to parse as map and convert to OpenSearchServerlessSpec
		specMap, mapOk := resource.Spec.(map[string]interface{})
		if !mapOk {
			return fmt.Errorf("invalid opensearch serverless spec format")
		}

		// Convert map to OpenSearchServerlessSpec
		specJSON, err := json.Marshal(specMap)
		if err != nil {
			return fmt.Errorf("failed to marshal opensearch serverless spec: %w", err)
		}

		if err := json.Unmarshal(specJSON, &opensearchServerless); err != nil {
			return fmt.Errorf("failed to unmarshal opensearch serverless spec: %w", err)
		}
	}

	resourceName := g.sanitizeResourceName(resource.Metadata.Name)
	collectionName := opensearchServerless.CollectionName
	if collectionName == "" {
		collectionName = resource.Metadata.Name
	}

	// Generate encryption policy
	if err := g.generateEncryptionPolicy(body, resourceName, collectionName, opensearchServerless.EncryptionPolicy); err != nil {
		return fmt.Errorf("failed to generate encryption policy: %w", err)
	}

	// Generate network policy
	if err := g.generateNetworkPolicy(body, resourceName, collectionName, opensearchServerless.NetworkPolicy); err != nil {
		return fmt.Errorf("failed to generate network policy: %w", err)
	}

	// Generate access policy
	if err := g.generateAccessPolicy(body, resourceName, collectionName, opensearchServerless.AccessPolicy); err != nil {
		return fmt.Errorf("failed to generate access policy: %w", err)
	}

	// Generate collection
	if err := g.generateCollection(body, resourceName, collectionName, opensearchServerless); err != nil {
		return fmt.Errorf("failed to generate collection: %w", err)
	}

	// Generate vector index if specified
	if opensearchServerless.VectorIndex != nil {
		if err := g.generateVectorIndex(body, resourceName, collectionName, opensearchServerless.VectorIndex); err != nil {
			return fmt.Errorf("failed to generate vector index: %w", err)
		}
	}

	g.logger.WithField("opensearch_serverless", resource.Metadata.Name).Info("Generated OpenSearch Serverless resources")
	return nil
}

// generateEncryptionPolicy creates the encryption policy for the collection
func (g *HCLGenerator) generateEncryptionPolicy(body *hclwrite.Body, resourceName, collectionName string, policy *models.EncryptionPolicy) error {
	policyName := fmt.Sprintf("%s-encryption-policy", resourceName)
	if policy != nil && policy.Name != "" {
		policyName = policy.Name
	}

	// Create encryption policy resource
	policyBlock := body.AppendNewBlock("resource", []string{"aws_opensearchserverless_security_policy", fmt.Sprintf("%s_encryption_policy", resourceName)})
	policyBody := policyBlock.Body()

	policyBody.SetAttributeValue("name", cty.StringVal(policyName))
	policyBody.SetAttributeValue("type", cty.StringVal("encryption"))

	// Description
	description := fmt.Sprintf("Encryption policy for %s collection", collectionName)
	if policy != nil && policy.Description != "" {
		description = policy.Description
	}
	policyBody.SetAttributeValue("description", cty.StringVal(description))

	// Policy document
	policyDoc := map[string]interface{}{
		"Rules": []map[string]interface{}{
			{
				"Resource":     []string{fmt.Sprintf("collection/%s", collectionName)},
				"ResourceType": "collection",
			},
		},
		"AWSOwnedKey": true,
	}

	// Use custom KMS key if provided
	if policy != nil && policy.KmsKeyId != "" {
		policyDoc["AWSOwnedKey"] = false
		policyDoc["KmsKeyId"] = policy.KmsKeyId
	}

	policyJSON, err := json.Marshal(policyDoc)
	if err != nil {
		return fmt.Errorf("failed to marshal encryption policy: %w", err)
	}

	policyBody.SetAttributeValue("policy", cty.StringVal(string(policyJSON)))

	body.AppendNewline()
	return nil
}

// generateNetworkPolicy creates the network policy for the collection
func (g *HCLGenerator) generateNetworkPolicy(body *hclwrite.Body, resourceName, collectionName string, policy *models.NetworkPolicy) error {
	policyName := fmt.Sprintf("%s-network-policy", resourceName)
	if policy != nil && policy.Name != "" {
		policyName = policy.Name
	}

	// Create network policy resource
	policyBlock := body.AppendNewBlock("resource", []string{"aws_opensearchserverless_security_policy", fmt.Sprintf("%s_network_policy", resourceName)})
	policyBody := policyBlock.Body()

	policyBody.SetAttributeValue("name", cty.StringVal(policyName))
	policyBody.SetAttributeValue("type", cty.StringVal("network"))

	// Description
	description := fmt.Sprintf("Network policy for %s collection", collectionName)
	if policy != nil && policy.Description != "" {
		description = policy.Description
	}
	policyBody.SetAttributeValue("description", cty.StringVal(description))

	// Policy document - default to public access
	policyDoc := []map[string]interface{}{
		{
			"Rules": []map[string]interface{}{
				{
					"Resource": []string{
						fmt.Sprintf("collection/%s", collectionName),
						fmt.Sprintf("dashboard/%s", collectionName),
					},
					"ResourceType": "collection",
				},
				{
					"Resource": []string{
						fmt.Sprintf("collection/%s", collectionName),
						fmt.Sprintf("dashboard/%s", collectionName),
					},
					"ResourceType": "dashboard",
				},
			},
			"AllowFromPublic": true,
		},
	}

	// Use custom access configuration if provided
	if policy != nil && len(policy.Access) > 0 {
		policyDoc[0]["AllowFromPublic"] = false
		for _, access := range policy.Access {
			if access.SourceType == "vpc" && len(access.SourceVPCEs) > 0 {
				policyDoc[0]["SourceVPCEs"] = access.SourceVPCEs
			}
		}
	}

	policyJSON, err := json.Marshal(policyDoc)
	if err != nil {
		return fmt.Errorf("failed to marshal network policy: %w", err)
	}

	policyBody.SetAttributeValue("policy", cty.StringVal(string(policyJSON)))

	body.AppendNewline()
	return nil
}

// generateAccessPolicy creates the data access policy for the collection
func (g *HCLGenerator) generateAccessPolicy(body *hclwrite.Body, resourceName, collectionName string, policy *models.AccessPolicy) error {
	policyName := fmt.Sprintf("%s-access-policy", resourceName)
	if policy != nil && policy.Name != "" {
		policyName = policy.Name
	}

	// Create access policy resource
	policyBlock := body.AppendNewBlock("resource", []string{"aws_opensearchserverless_access_policy", fmt.Sprintf("%s_access_policy", resourceName)})
	policyBody := policyBlock.Body()

	policyBody.SetAttributeValue("name", cty.StringVal(policyName))
	policyBody.SetAttributeValue("type", cty.StringVal("data"))

	// Description
	description := fmt.Sprintf("Data access policy for %s collection", collectionName)
	if policy != nil && policy.Description != "" {
		description = policy.Description
	}
	policyBody.SetAttributeValue("description", cty.StringVal(description))

	// Default principals and permissions for Bedrock
	principals := []string{}
	permissions := []string{
		"aoss:CreateCollectionItems",
		"aoss:DeleteCollectionItems",
		"aoss:UpdateCollectionItems",
		"aoss:DescribeCollectionItems",
	}

	// Add custom principals if provided
	if policy != nil && len(policy.Principals) > 0 {
		principals = append(principals, policy.Principals...)
	}

	// Add custom permissions if provided
	if policy != nil && len(policy.Permissions) > 0 {
		permissions = policy.Permissions
	}

	// Auto-configure for Bedrock if enabled
	if policy != nil && policy.AutoConfigureForBedrock {
		// Add Bedrock service principal
		principals = append(principals, "bedrock.amazonaws.com")

		// Add comprehensive permissions for Bedrock operations
		bedrockPermissions := []string{
			"aoss:CreateIndex",
			"aoss:DeleteIndex",
			"aoss:UpdateIndex",
			"aoss:DescribeIndex",
			"aoss:ReadDocument",
			"aoss:WriteDocument",
			"aoss:CreateCollectionItems",
			"aoss:DeleteCollectionItems",
			"aoss:UpdateCollectionItems",
			"aoss:DescribeCollectionItems",
		}

		// Merge with existing permissions
		permissionSet := make(map[string]bool)
		for _, perm := range permissions {
			permissionSet[perm] = true
		}
		for _, perm := range bedrockPermissions {
			permissionSet[perm] = true
		}

		permissions = make([]string, 0, len(permissionSet))
		for perm := range permissionSet {
			permissions = append(permissions, perm)
		}
	}

	// Policy document
	policyDoc := []map[string]interface{}{
		{
			"Rules": []map[string]interface{}{
				{
					"Resource": []string{
						fmt.Sprintf("collection/%s", collectionName),
						fmt.Sprintf("index/%s/*", collectionName),
					},
					"Permission":   permissions,
					"ResourceType": "collection",
				},
				{
					"Resource": []string{
						fmt.Sprintf("collection/%s", collectionName),
						fmt.Sprintf("index/%s/*", collectionName),
					},
					"Permission":   permissions,
					"ResourceType": "index",
				},
			},
			"Principal": principals,
		},
	}

	policyJSON, err := json.Marshal(policyDoc)
	if err != nil {
		return fmt.Errorf("failed to marshal access policy: %w", err)
	}

	policyBody.SetAttributeValue("policy", cty.StringVal(string(policyJSON)))

	body.AppendNewline()
	return nil
}

// generateCollection creates the OpenSearch Serverless collection
func (g *HCLGenerator) generateCollection(body *hclwrite.Body, resourceName, collectionName string, spec models.OpenSearchServerlessSpec) error {
	// Create collection resource
	collectionBlock := body.AppendNewBlock("resource", []string{"aws_opensearchserverless_collection", resourceName})
	collectionBody := collectionBlock.Body()

	collectionBody.SetAttributeValue("name", cty.StringVal(collectionName))

	// Collection type
	collectionType := "VECTORSEARCH"
	if spec.Type != "" {
		collectionType = spec.Type
	}
	collectionBody.SetAttributeValue("type", cty.StringVal(collectionType))

	// Description
	if spec.Description != "" {
		collectionBody.SetAttributeValue("description", cty.StringVal(spec.Description))
	}

	// Dependencies on security policies
	depends_on := []string{
		fmt.Sprintf("aws_opensearchserverless_security_policy.%s_encryption_policy", resourceName),
		fmt.Sprintf("aws_opensearchserverless_security_policy.%s_network_policy", resourceName),
		fmt.Sprintf("aws_opensearchserverless_access_policy.%s_access_policy", resourceName),
	}

	dependsOnValues := make([]cty.Value, len(depends_on))
	for i, dep := range depends_on {
		dependsOnValues[i] = cty.StringVal(dep)
	}
	collectionBody.SetAttributeValue("depends_on", cty.ListVal(dependsOnValues))

	// Tags
	if len(spec.Tags) > 0 {
		tagValues := make(map[string]cty.Value)
		for key, value := range spec.Tags {
			tagValues[key] = cty.StringVal(value)
		}
		collectionBody.SetAttributeValue("tags", cty.ObjectVal(tagValues))
	}

	body.AppendNewline()
	return nil
}

// generateVectorIndex creates the vector index for the collection
func (g *HCLGenerator) generateVectorIndex(body *hclwrite.Body, resourceName, collectionName string, vectorIndex *models.VectorIndexConfig) error {
	// Note: Vector index creation is typically done via REST API after collection is created
	// For now, we'll add a local-exec provisioner to create the index

	// Create null resource for index creation
	indexBlock := body.AppendNewBlock("resource", []string{"null_resource", fmt.Sprintf("%s_vector_index", resourceName)})
	indexBody := indexBlock.Body()

	// Provisioner block
	provisionerBlock := indexBody.AppendNewBlock("provisioner", []string{"local-exec"})
	provisionerBody := provisionerBlock.Body()

	// Vector field mappings
	vectorField := "vector"
	textField := "text"
	metadataField := "metadata"

	if vectorIndex.FieldMapping.VectorField != "" {
		vectorField = vectorIndex.FieldMapping.VectorField
	}
	if vectorIndex.FieldMapping.TextField != "" {
		textField = vectorIndex.FieldMapping.TextField
	}
	if vectorIndex.FieldMapping.MetadataField != "" {
		metadataField = vectorIndex.FieldMapping.MetadataField
	}

	// Command to create vector index
	createIndexCommand := fmt.Sprintf(`
aws opensearchserverless batch-get-collection --names %s --query 'collectionDetails[0].collectionEndpoint' --output text | \
xargs -I {} curl -X PUT "{}/%s" -H "Content-Type: application/json" -d '{
  "settings": {
    "index": {
      "knn": true,
      "knn.algo_param.ef_search": 512
    }
  },
  "mappings": {
    "properties": {
      "%s": {
        "type": "knn_vector",
        "dimension": 1536,
        "method": {
          "name": "hnsw",
          "space_type": "l2",
          "engine": "nmslib"
        }
      },
      "%s": {
        "type": "text"
      },
      "%s": {
        "type": "text"
      }
    }
  }
}'`, collectionName, vectorIndex.Name, vectorField, textField, metadataField)

	provisionerBody.SetAttributeValue("command", cty.StringVal(createIndexCommand))

	// Dependencies
	depends_on := []string{
		fmt.Sprintf("aws_opensearchserverless_collection.%s", resourceName),
	}

	dependsOnValues := make([]cty.Value, len(depends_on))
	for i, dep := range depends_on {
		dependsOnValues[i] = cty.StringVal(dep)
	}
	indexBody.SetAttributeValue("depends_on", cty.ListVal(dependsOnValues))

	body.AppendNewline()
	return nil
}
