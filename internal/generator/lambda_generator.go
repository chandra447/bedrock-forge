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

	// Handle new attributes
	g.setLambdaAdvancedAttributes(moduleBody, lambda)

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

// createPolicyStatement creates a consistent policy statement with all required fields
func (g *HCLGenerator) createPolicyStatement(sid, effect string, principals []cty.Value, actions []cty.Value, condition cty.Value) cty.Value {
	return cty.ObjectVal(map[string]cty.Value{
		"sid":        cty.StringVal(sid),
		"effect":     cty.StringVal(effect),
		"principals": cty.ListVal(principals),
		"actions":    cty.ListVal(actions),
		"condition":  condition,
	})
}

// generateLambdaResourcePolicies creates resource-based policies for Lambda functions
func (g *HCLGenerator) generateLambdaResourcePolicies(lambdaName string, lambda models.LambdaSpec) []cty.Value {
	var policyStatements []cty.Value

	// Add custom statements if provided (for advanced users)
	if lambda.ResourcePolicy != nil && len(lambda.ResourcePolicy.Statements) > 0 {
		for _, stmt := range lambda.ResourcePolicy.Statements {
			// Handle principals
			var principalList []cty.Value
			if len(stmt.Principal) > 0 {
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
			}

			// Handle actions
			var actionList []cty.Value
			switch actions := stmt.Action.(type) {
			case string:
				actionList = []cty.Value{cty.StringVal(actions)}
			case []interface{}:
				for _, action := range actions {
					if str, ok := action.(string); ok {
						actionList = append(actionList, cty.StringVal(str))
					}
				}
			}

			// Handle conditions - ensure consistent structure
			var conditionVal cty.Value
			if len(stmt.Condition) > 0 {
				conditionValues := make(map[string]cty.Value)
				for k, v := range stmt.Condition {
					if objVal, ok := v.(map[string]interface{}); ok {
						// Handle nested condition objects, ensuring consistent structure
						nestedValues := make(map[string]cty.Value)

						// Normalize to have consistent schema with agent conditions
						nestedValues["aws:SourceArn"] = cty.StringVal("")
						nestedValues["aws:SourceAccount"] = cty.StringVal("")

						for nk, nv := range objVal {
							if nstr, ok := nv.(string); ok {
								nestedValues[nk] = cty.StringVal(nstr)
							}
						}
						conditionValues[k] = cty.ObjectVal(nestedValues)
						continue
					}

					conditionValues[k] = cty.StringVal(fmt.Sprintf("%v", v))
				}
				conditionVal = cty.ObjectVal(conditionValues)
			} else {
				// Create empty condition with consistent structure
				conditionVal = cty.EmptyObjectVal
			}

			policyStatements = append(policyStatements, g.createPolicyStatement(stmt.Sid, stmt.Effect, principalList, actionList, conditionVal))
		}
	}

	// Find all agents that reference this Lambda function
	referencingAgents := g.findAgentsReferencingLambda(lambdaName)

	if len(referencingAgents) > 0 {
		// Create agent-specific permissions
		for _, agentName := range referencingAgents {
			agentResourceName := g.sanitizeResourceName(agentName)

			// Create principals for agent permission
			agentPrincipals := []cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"type":        cty.StringVal("Service"),
					"identifiers": cty.ListVal([]cty.Value{cty.StringVal("bedrock.amazonaws.com")}),
				}),
			}

			// Create actions for agent permission
			agentActions := []cty.Value{
				cty.StringVal("lambda:InvokeFunction"),
			}

			// Create condition for agent permission with consistent structure
			// Normalize all conditions to have consistent field schemas
			agentCondition := cty.ObjectVal(map[string]cty.Value{
				"StringEquals": cty.ObjectVal(map[string]cty.Value{
					"aws:SourceArn":     cty.StringVal(fmt.Sprintf("${module.%s.agent_arn}", agentResourceName)),
					"aws:SourceAccount": cty.StringVal(""), // Add empty field for consistency
				}),
			})

			agentStmt := g.createPolicyStatement(
				fmt.Sprintf("AllowBedrockAgent_%s", agentResourceName),
				"Allow",
				agentPrincipals,
				agentActions,
				agentCondition,
			)

			policyStatements = append(policyStatements, agentStmt)

			g.logger.WithField("lambda", lambdaName).WithField("agent", agentName).Debug("Generated agent-specific Lambda permission")
		}
	} else {
		// If no agents reference this Lambda, add general Bedrock permission (unless explicitly disabled)
		if lambda.ResourcePolicy == nil || lambda.ResourcePolicy.AllowBedrockAgents {
			// Create principals for default permission
			defaultPrincipals := []cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"type":        cty.StringVal("Service"),
					"identifiers": cty.ListVal([]cty.Value{cty.StringVal("bedrock.amazonaws.com")}),
				}),
			}

			// Create actions for default permission
			defaultActions := []cty.Value{
				cty.StringVal("lambda:InvokeFunction"),
			}

			defaultStmt := g.createPolicyStatement(
				"AllowBedrockAgentInvoke",
				"Allow",
				defaultPrincipals,
				defaultActions,
				cty.EmptyObjectVal,
			)
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

// setLambdaAdvancedAttributes sets the new advanced Lambda attributes
func (g *HCLGenerator) setLambdaAdvancedAttributes(moduleBody *hclwrite.Body, lambda models.LambdaSpec) {
	// Role handling - prefer direct ARN over reference
	if lambda.RoleArn != "" {
		moduleBody.SetAttributeValue("role", cty.StringVal(lambda.RoleArn))
	} else if !lambda.Role.IsEmpty() {
		// Handle reference to IAM role
		if roleRef, err := g.resolveReferenceToOutput(lambda.Role, models.IAMRoleKind, "role_arn"); err == nil {
			moduleBody.SetAttributeValue("role", cty.StringVal(roleRef))
		} else {
			g.logger.WithError(err).WithField("lambda", lambda.Role.String()).Warn("Failed to resolve IAM role reference")
		}
	}

	// Architectures
	if len(lambda.Architectures) > 0 {
		archVals := make([]cty.Value, 0, len(lambda.Architectures))
		for _, arch := range lambda.Architectures {
			archVals = append(archVals, cty.StringVal(arch))
		}
		moduleBody.SetAttributeValue("architectures", cty.ListVal(archVals))
	}

	// Code signing config
	if lambda.CodeSigningConfigArn != "" {
		moduleBody.SetAttributeValue("code_signing_config_arn", cty.StringVal(lambda.CodeSigningConfigArn))
	}

	// Dead letter config
	if lambda.DeadLetterConfig != nil {
		dlcValues := map[string]cty.Value{
			"target_arn": cty.StringVal(lambda.DeadLetterConfig.TargetArn),
		}
		moduleBody.SetAttributeValue("dead_letter_config", cty.ObjectVal(dlcValues))
	}

	// Ephemeral storage
	if lambda.EphemeralStorage != nil {
		esValues := map[string]cty.Value{
			"size": cty.NumberIntVal(int64(lambda.EphemeralStorage.Size)),
		}
		moduleBody.SetAttributeValue("ephemeral_storage", cty.ObjectVal(esValues))
	}

	// File system config
	if lambda.FileSystemConfig != nil {
		fscValues := map[string]cty.Value{
			"arn":              cty.StringVal(lambda.FileSystemConfig.Arn),
			"local_mount_path": cty.StringVal(lambda.FileSystemConfig.LocalMountPath),
		}
		moduleBody.SetAttributeValue("file_system_config", cty.ObjectVal(fscValues))
	}

	// Image config
	if lambda.ImageConfig != nil {
		imgValues := make(map[string]cty.Value)
		if len(lambda.ImageConfig.Command) > 0 {
			cmdVals := make([]cty.Value, 0, len(lambda.ImageConfig.Command))
			for _, cmd := range lambda.ImageConfig.Command {
				cmdVals = append(cmdVals, cty.StringVal(cmd))
			}
			imgValues["command"] = cty.ListVal(cmdVals)
		}
		if len(lambda.ImageConfig.EntryPoint) > 0 {
			epVals := make([]cty.Value, 0, len(lambda.ImageConfig.EntryPoint))
			for _, ep := range lambda.ImageConfig.EntryPoint {
				epVals = append(epVals, cty.StringVal(ep))
			}
			imgValues["entry_point"] = cty.ListVal(epVals)
		}
		if lambda.ImageConfig.WorkingDirectory != "" {
			imgValues["working_directory"] = cty.StringVal(lambda.ImageConfig.WorkingDirectory)
		}
		if len(imgValues) > 0 {
			moduleBody.SetAttributeValue("image_config", cty.ObjectVal(imgValues))
		}
	}

	// KMS key
	if lambda.KmsKeyArn != "" {
		moduleBody.SetAttributeValue("kms_key_arn", cty.StringVal(lambda.KmsKeyArn))
	}

	// Layers
	if len(lambda.Layers) > 0 {
		layerVals := make([]cty.Value, 0, len(lambda.Layers))
		for _, layer := range lambda.Layers {
			layerVals = append(layerVals, cty.StringVal(layer))
		}
		moduleBody.SetAttributeValue("layers", cty.ListVal(layerVals))
	}

	// Package type
	if lambda.PackageType != "" {
		moduleBody.SetAttributeValue("package_type", cty.StringVal(lambda.PackageType))
	}

	// Publish
	if lambda.Publish != nil {
		moduleBody.SetAttributeValue("publish", cty.BoolVal(*lambda.Publish))
	}

	// Replace security groups on destroy
	if lambda.ReplaceSecurityGroupsOnDestroy != nil {
		moduleBody.SetAttributeValue("replace_security_groups_on_destroy", cty.BoolVal(*lambda.ReplaceSecurityGroupsOnDestroy))
	}

	// Replacement security group IDs
	if len(lambda.ReplacementSecurityGroupIds) > 0 {
		rsgVals := make([]cty.Value, 0, len(lambda.ReplacementSecurityGroupIds))
		for _, rsgId := range lambda.ReplacementSecurityGroupIds {
			rsgVals = append(rsgVals, cty.StringVal(rsgId))
		}
		moduleBody.SetAttributeValue("replacement_security_group_ids", cty.ListVal(rsgVals))
	}

	// Skip destroy
	if lambda.SkipDestroy != nil {
		moduleBody.SetAttributeValue("skip_destroy", cty.BoolVal(*lambda.SkipDestroy))
	}

	// SnapStart
	if lambda.SnapStart != nil {
		ssValues := map[string]cty.Value{
			"apply_on": cty.StringVal(lambda.SnapStart.ApplyOn),
		}
		moduleBody.SetAttributeValue("snap_start", cty.ObjectVal(ssValues))
	}

	// Source code hash
	if lambda.SourceCodeHash != "" {
		moduleBody.SetAttributeValue("source_code_hash", cty.StringVal(lambda.SourceCodeHash))
	}

	// Timeouts
	if lambda.Timeouts != nil {
		timeoutValues := make(map[string]cty.Value)
		if lambda.Timeouts.Create != "" {
			timeoutValues["create"] = cty.StringVal(lambda.Timeouts.Create)
		}
		if lambda.Timeouts.Update != "" {
			timeoutValues["update"] = cty.StringVal(lambda.Timeouts.Update)
		}
		if lambda.Timeouts.Delete != "" {
			timeoutValues["delete"] = cty.StringVal(lambda.Timeouts.Delete)
		}
		if len(timeoutValues) > 0 {
			moduleBody.SetAttributeValue("timeouts", cty.ObjectVal(timeoutValues))
		}
	}

	// Tracing config
	if lambda.TracingConfig != nil {
		tcValues := map[string]cty.Value{
			"mode": cty.StringVal(lambda.TracingConfig.Mode),
		}
		moduleBody.SetAttributeValue("tracing_config", cty.ObjectVal(tcValues))
	}
}
