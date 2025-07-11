package generator

import (
	"encoding/json"
	"fmt"
	"strings"

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

	// Generate resource-based policies automatically
	policyStatements := g.generateLambdaResourcePolicies(resource.Metadata.Name, lambda)

	if len(policyStatements) > 0 {
		moduleBody.SetAttributeValue("lambda_resource_policy_statements", cty.ListVal(policyStatements))
	}

	body.AppendNewline()

	g.logger.WithField("lambda", resource.Metadata.Name).Info("Generated lambda module")
	return nil
}

// generateLambdaResourcePolicies creates resource-based policies for Lambda functions
func (g *HCLGenerator) generateLambdaResourcePolicies(lambdaName string, lambda models.LambdaSpec) []cty.Value {
	var policyStatements []cty.Value

	// Add custom statements if provided (for advanced users)
	if lambda.ResourcePolicy != nil && len(lambda.ResourcePolicy.Statements) > 0 {
		for _, stmt := range lambda.ResourcePolicy.Statements {
			stmtValues := make(map[string]cty.Value)
			stmtValues["sid"] = cty.StringVal(stmt.Sid)
			stmtValues["effect"] = cty.StringVal(stmt.Effect)

			// Handle principals
			if len(stmt.Principal) > 0 {
				principalList := make([]cty.Value, 0)
				for pType, pValues := range stmt.Principal {
					switch values := pValues.(type) {
					case string:
						principalList = append(principalList, cty.ObjectVal(map[string]cty.Value{
							"type":        cty.StringVal(pType),
							"identifiers": cty.ListVal([]cty.Value{cty.StringVal(values)}),
						}))
					case []interface{}:
						identifiers := make([]cty.Value, 0)
						for _, v := range values {
							if str, ok := v.(string); ok {
								identifiers = append(identifiers, cty.StringVal(str))
							}
						}
						principalList = append(principalList, cty.ObjectVal(map[string]cty.Value{
							"type":        cty.StringVal(pType),
							"identifiers": cty.ListVal(identifiers),
						}))
					}
				}
				stmtValues["principals"] = cty.ListVal(principalList)
			}

			// Handle actions
			switch actions := stmt.Action.(type) {
			case string:
				stmtValues["actions"] = cty.ListVal([]cty.Value{cty.StringVal(actions)})
			case []interface{}:
				actionList := make([]cty.Value, 0)
				for _, action := range actions {
					if str, ok := action.(string); ok {
						actionList = append(actionList, cty.StringVal(str))
					}
				}
				stmtValues["actions"] = cty.ListVal(actionList)
			}

			// Handle conditions
			if len(stmt.Condition) > 0 {
				conditionValues := make(map[string]cty.Value)
				for k, v := range stmt.Condition {
					if str, ok := v.(string); ok {
						conditionValues[k] = cty.StringVal(str)
					}
				}
				stmtValues["condition"] = cty.ObjectVal(conditionValues)
			}

			policyStatements = append(policyStatements, cty.ObjectVal(stmtValues))
		}
	}

	// Find all agents that reference this Lambda function
	referencingAgents := g.findAgentsReferencingLambda(lambdaName)

	if len(referencingAgents) > 0 {
		// Create agent-specific permissions
		for _, agentName := range referencingAgents {
			agentResourceName := g.sanitizeResourceName(agentName)

			agentStmt := cty.ObjectVal(map[string]cty.Value{
				"sid":    cty.StringVal(fmt.Sprintf("AllowBedrockAgent_%s", agentResourceName)),
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
				"condition": cty.ObjectVal(map[string]cty.Value{
					"StringEquals": cty.ObjectVal(map[string]cty.Value{
						"aws:SourceArn": cty.StringVal(fmt.Sprintf("${module.%s.agent_arn}", agentResourceName)),
					}),
				}),
			})

			policyStatements = append(policyStatements, agentStmt)

			g.logger.WithField("lambda", lambdaName).WithField("agent", agentName).Debug("Generated agent-specific Lambda permission")
		}
	} else {
		// If no agents reference this Lambda, add general Bedrock permission (unless explicitly disabled)
		if lambda.ResourcePolicy == nil || lambda.ResourcePolicy.AllowBedrockAgents {
			defaultStmt := cty.ObjectVal(map[string]cty.Value{
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
			})
			policyStatements = append(policyStatements, defaultStmt)
		}
	}

	return policyStatements
}

// findAgentsReferencingLambda finds all agents that reference the given Lambda function
func (g *HCLGenerator) findAgentsReferencingLambda(lambdaName string) []string {
	var referencingAgents []string

	// Iterate through all resources to find agents
	for _, resource := range g.registry.GetResourcesByKind(models.AgentKind) {
		if agent, ok := resource.Resource.(*models.Agent); ok {
			// Check inline action groups
			for _, ag := range agent.Spec.ActionGroups {
				if ag.ActionGroupExecutor != nil && ag.ActionGroupExecutor.Lambda.String() == lambdaName {
					referencingAgents = append(referencingAgents, resource.Metadata.Name)
					break // Found one reference, no need to check more action groups for this agent
				}
			}
		}
	}

	// Also check standalone ActionGroup resources that reference this Lambda
	for _, resource := range g.registry.GetResourcesByKind(models.ActionGroupKind) {
		if actionGroup, ok := resource.Resource.(*models.ActionGroup); ok {
			if actionGroup.Spec.ActionGroupExecutor != nil && actionGroup.Spec.ActionGroupExecutor.Lambda.String() == lambdaName {
				// For standalone action groups, we need to find the agent they belong to
				// Parse agentId to extract agent name if it's a module reference
				if !actionGroup.Spec.AgentId.IsEmpty() {
					agentName := extractAgentNameFromId(actionGroup.Spec.AgentId.String())
					if agentName != "" {
						referencingAgents = append(referencingAgents, agentName)
					}
				}
			}
		}
	}

	// Remove duplicates
	seen := make(map[string]bool)
	var uniqueAgents []string
	for _, agent := range referencingAgents {
		if !seen[agent] {
			seen[agent] = true
			uniqueAgents = append(uniqueAgents, agent)
		}
	}

	return uniqueAgents
}

// extractAgentNameFromId extracts the agent name from an agentId field
// Handles module references like "${module.agent_name.agent_id}" and returns the agent name
// For direct ARNs, returns empty string since we can't extract agent name
func extractAgentNameFromId(agentId string) string {
	// Check if it's a module reference: ${module.agent_name.agent_id}
	if strings.HasPrefix(agentId, "${module.") && strings.HasSuffix(agentId, ".agent_id}") {
		// Extract agent name from ${module.agent_name.agent_id}
		withoutPrefix := strings.TrimPrefix(agentId, "${module.")
		withoutSuffix := strings.TrimSuffix(withoutPrefix, ".agent_id}")
		return withoutSuffix
	}

	// For direct ARNs or other formats, we can't reliably extract agent name
	// In practice, standalone ActionGroups with direct agent ARNs won't be used
	// for Lambda permission generation since we can't determine the agent name
	return ""
}
