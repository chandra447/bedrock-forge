package generator

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"

	"bedrock-forge/internal/models"
)

// generateLambdaModule creates a module call for a Lambda resource
func (g *HCLGenerator) generateLambdaModule(body *hclwrite.Body, resource models.BaseResource) error {
	lambda, ok := resource.Spec.(models.LambdaSpec)
	if !ok {
		// Try to parse as map and convert to LambdaSpec
		specMap, mapOk := resource.Spec.(map[string]interface{})
		if !mapOk {
			return fmt.Errorf("invalid lambda spec format")
		}

		// Convert map to LambdaSpec
		specJSON, err := json.Marshal(specMap)
		if err != nil {
			return fmt.Errorf("failed to marshal lambda spec: %w", err)
		}

		if err := json.Unmarshal(specJSON, &lambda); err != nil {
			return fmt.Errorf("failed to unmarshal lambda spec: %w", err)
		}
	}

	resourceName := g.sanitizeResourceName(resource.Metadata.Name)

	// Create module block
	moduleBlock := body.AppendNewBlock("module", []string{resourceName})
	moduleBody := moduleBlock.Body()

	// Set module source
	moduleSource := fmt.Sprintf("%s//modules/lambda-function", g.config.ModuleRegistry)
	if g.config.ModuleVersion != "" {
		moduleSource += fmt.Sprintf("?ref=%s", g.config.ModuleVersion)
	}
	moduleBody.SetAttributeValue("source", cty.StringVal(moduleSource))

	// Set basic attributes
	moduleBody.SetAttributeValue("function_name", cty.StringVal(resource.Metadata.Name))
	moduleBody.SetAttributeValue("runtime", cty.StringVal(lambda.Runtime))
	moduleBody.SetAttributeValue("handler", cty.StringVal(lambda.Handler))

	// Optional description
	if resource.Metadata.Description != "" {
		moduleBody.SetAttributeValue("description", cty.StringVal(resource.Metadata.Description))
	}

	// Code configuration
	codeValues := make(map[string]cty.Value)
	codeValues["source"] = cty.StringVal(lambda.Code.Source)

	if lambda.Code.ZipFile != "" {
		codeValues["zip_file"] = cty.StringVal(lambda.Code.ZipFile)
	}

	if lambda.Code.S3Bucket != "" {
		codeValues["s3_bucket"] = cty.StringVal(lambda.Code.S3Bucket)
	}

	if lambda.Code.S3Key != "" {
		codeValues["s3_key"] = cty.StringVal(lambda.Code.S3Key)
	}

	if lambda.Code.S3ObjectVersion != "" {
		codeValues["s3_object_version"] = cty.StringVal(lambda.Code.S3ObjectVersion)
	}

	moduleBody.SetAttributeValue("code", cty.ObjectVal(codeValues))

	// Environment variables
	if len(lambda.Environment) > 0 {
		envValues := make(map[string]cty.Value)
		for key, value := range lambda.Environment {
			envValues[key] = cty.StringVal(value)
		}
		moduleBody.SetAttributeValue("environment_variables", cty.ObjectVal(envValues))
	}

	// Optional configuration
	if lambda.Timeout > 0 {
		moduleBody.SetAttributeValue("timeout", cty.NumberIntVal(int64(lambda.Timeout)))
	}

	if lambda.MemorySize > 0 {
		moduleBody.SetAttributeValue("memory_size", cty.NumberIntVal(int64(lambda.MemorySize)))
	}

	if lambda.ReservedConcurrency > 0 {
		moduleBody.SetAttributeValue("reserved_concurrency", cty.NumberIntVal(int64(lambda.ReservedConcurrency)))
	}

	// Tags
	if len(lambda.Tags) > 0 {
		tagValues := make(map[string]cty.Value)
		for key, value := range lambda.Tags {
			tagValues[key] = cty.StringVal(value)
		}
		moduleBody.SetAttributeValue("tags", cty.ObjectVal(tagValues))
	}

	// VPC configuration
	if lambda.VpcConfig != nil {
		vpcValues := make(map[string]cty.Value)

		if len(lambda.VpcConfig.SecurityGroupIds) > 0 {
			sgIds := make([]cty.Value, 0, len(lambda.VpcConfig.SecurityGroupIds))
			for _, sgId := range lambda.VpcConfig.SecurityGroupIds {
				sgIds = append(sgIds, cty.StringVal(sgId))
			}
			vpcValues["security_group_ids"] = cty.ListVal(sgIds)
		}

		if len(lambda.VpcConfig.SubnetIds) > 0 {
			subnetIds := make([]cty.Value, 0, len(lambda.VpcConfig.SubnetIds))
			for _, subnetId := range lambda.VpcConfig.SubnetIds {
				subnetIds = append(subnetIds, cty.StringVal(subnetId))
			}
			vpcValues["subnet_ids"] = cty.ListVal(subnetIds)
		}

		if len(vpcValues) > 0 {
			moduleBody.SetAttributeValue("vpc_config", cty.ObjectVal(vpcValues))
		}
	}

	// Add IAM role for Lambda execution
	moduleBody.SetAttributeValue("create_role", cty.BoolVal(true))

	// Add policy for Bedrock agent invocation
	policyStatements := []cty.Value{
		cty.ObjectVal(map[string]cty.Value{
			"sid":    cty.StringVal("AllowBedrockAgentInvoke"),
			"effect": cty.StringVal("Allow"),
			"principals": cty.ListVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"type":        cty.StringVal("Service"),
					"identifiers": cty.ListVal([]cty.Value{cty.StringVal("bedrock.amazonaws.com")}),
				}),
			}),
			"actions": cty.ListVal([]cty.Value{
				cty.StringVal("lambda:InvokeFunction"),
			}),
		}),
	}

	moduleBody.SetAttributeValue("lambda_resource_policy_statements", cty.ListVal(policyStatements))

	body.AppendNewline()

	g.logger.WithField("lambda", resource.Metadata.Name).Info("Generated lambda module")
	return nil
}
