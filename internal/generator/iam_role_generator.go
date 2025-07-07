package generator

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"

	"bedrock-forge/internal/models"
)

// generateAutoIAMRole creates an auto-generated IAM role for Bedrock agents
func (g *HCLGenerator) generateAutoIAMRole(body *hclwrite.Body, agentName string, iamConfig *models.IAMRoleConfig) error {
	roleName := fmt.Sprintf("%s_execution_role", g.sanitizeResourceName(agentName))

	g.logger.WithField("agent", agentName).Debug("Generating auto IAM role")

	// Create module block
	moduleBlock := body.AppendNewBlock("module", []string{roleName})
	moduleBody := moduleBlock.Body()

	// Set module source
	moduleSource := fmt.Sprintf("%s//modules/iam-role", g.config.ModuleRegistry)
	if g.config.ModuleVersion != "" {
		moduleSource += fmt.Sprintf("?ref=%s", g.config.ModuleVersion)
	}
	moduleBody.SetAttributeValue("source", cty.StringVal(moduleSource))

	// Set basic attributes
	moduleBody.SetAttributeValue("role_name", cty.StringVal(fmt.Sprintf("%s-execution-role", agentName)))
	moduleBody.SetAttributeValue("description", cty.StringVal(fmt.Sprintf("Auto-generated execution role for Bedrock agent %s", agentName)))

	// Set assume role policy for Bedrock service
	assumeRolePolicy := cty.ObjectVal(map[string]cty.Value{
		"version": cty.StringVal("2012-10-17"),
		"statement": cty.ListVal([]cty.Value{
			cty.ObjectVal(map[string]cty.Value{
				"effect": cty.StringVal("Allow"),
				"principal": cty.ObjectVal(map[string]cty.Value{
					"service": cty.StringVal("bedrock.amazonaws.com"),
				}),
				"action": cty.StringVal("sts:AssumeRole"),
			}),
		}),
	})
	moduleBody.SetAttributeValue("assume_role_policy", assumeRolePolicy)

	// Set basic managed policies for Bedrock agents
	basicPolicies := []cty.Value{
		cty.ObjectVal(map[string]cty.Value{
			"policy_arn": cty.StringVal("arn:aws:iam::aws:policy/AmazonBedrockFullAccess"),
		}),
	}

	// Add additional policies if specified
	if iamConfig != nil && len(iamConfig.AdditionalPolicies) > 0 {
		for _, policy := range iamConfig.AdditionalPolicies {
			if policy.PolicyArn != "" {
				basicPolicies = append(basicPolicies, cty.ObjectVal(map[string]cty.Value{
					"policy_arn": cty.StringVal(policy.PolicyArn),
				}))
			}
		}
	}

	moduleBody.SetAttributeValue("managed_policies", cty.ListVal(basicPolicies))

	// Add inline policy for foundation model access and Lambda invocation
	inlinePolicies := []cty.Value{
		cty.ObjectVal(map[string]cty.Value{
			"name": cty.StringVal("BedrockAgentExecutionPolicy"),
			"policy": cty.ObjectVal(map[string]cty.Value{
				"version": cty.StringVal("2012-10-17"),
				"statement": cty.ListVal([]cty.Value{
					// Foundation model access
					cty.ObjectVal(map[string]cty.Value{
						"effect": cty.StringVal("Allow"),
						"action": cty.ListVal([]cty.Value{
							cty.StringVal("bedrock:InvokeModel"),
							cty.StringVal("bedrock:InvokeModelWithResponseStream"),
						}),
						"resource": cty.StringVal("arn:aws:bedrock:*::foundation-model/*"),
					}),
					// Lambda invocation for action groups
					cty.ObjectVal(map[string]cty.Value{
						"effect": cty.StringVal("Allow"),
						"action": cty.ListVal([]cty.Value{
							cty.StringVal("lambda:InvokeFunction"),
						}),
						"resource": cty.StringVal("arn:aws:lambda:*:*:function:*"),
					}),
					// Knowledge base access
					cty.ObjectVal(map[string]cty.Value{
						"effect": cty.StringVal("Allow"),
						"action": cty.ListVal([]cty.Value{
							cty.StringVal("bedrock:Retrieve"),
							cty.StringVal("bedrock:RetrieveAndGenerate"),
						}),
						"resource": cty.StringVal("arn:aws:bedrock:*:*:knowledge-base/*"),
					}),
					// CloudWatch Logs
					cty.ObjectVal(map[string]cty.Value{
						"effect": cty.StringVal("Allow"),
						"action": cty.ListVal([]cty.Value{
							cty.StringVal("logs:CreateLogGroup"),
							cty.StringVal("logs:CreateLogStream"),
							cty.StringVal("logs:PutLogEvents"),
						}),
						"resource": cty.StringVal("arn:aws:logs:*:*:*"),
					}),
				}),
			}),
		}),
	}

	moduleBody.SetAttributeValue("inline_policies", cty.ListVal(inlinePolicies))

	// Set default tags
	tags := map[string]cty.Value{
		"CreatedBy": cty.StringVal("bedrock-forge"),
		"Agent":     cty.StringVal(agentName),
		"Purpose":   cty.StringVal("BedrockAgentExecution"),
	}
	moduleBody.SetAttributeValue("tags", cty.ObjectVal(tags))

	body.AppendNewline()

	g.logger.WithField("agent", agentName).Info("Generated auto IAM role")
	return nil
}

// getAgentRoleReference returns the IAM role reference for an agent
func (g *HCLGenerator) getAgentRoleReference(agentName string, iamConfig *models.IAMRoleConfig) string {
	if iamConfig == nil {
		// Default to auto-created role
		roleName := fmt.Sprintf("%s_execution_role", g.sanitizeResourceName(agentName))
		return fmt.Sprintf("${module.%s.role_arn}", roleName)
	}

	if iamConfig.RoleArn != "" {
		// Use existing role ARN
		return iamConfig.RoleArn
	}

	if iamConfig.RoleName != "" {
		// Reference to manually defined IAMRole resource
		roleName := g.sanitizeResourceName(iamConfig.RoleName)
		return fmt.Sprintf("${module.%s.role_arn}", roleName)
	}

	if iamConfig.AutoCreate {
		// Auto-created role
		roleName := fmt.Sprintf("%s_execution_role", g.sanitizeResourceName(agentName))
		return fmt.Sprintf("${module.%s.role_arn}", roleName)
	}

	// Fallback to auto-created role
	roleName := fmt.Sprintf("%s_execution_role", g.sanitizeResourceName(agentName))
	return fmt.Sprintf("${module.%s.role_arn}", roleName)
}

// generateIAMRoleModule creates a Terraform module call for an IAM role
func (g *HCLGenerator) generateIAMRoleModule(body *hclwrite.Body, resource models.BaseResource) error {
	roleSpec, ok := resource.Spec.(models.IAMRoleSpec)
	if !ok {
		return fmt.Errorf("invalid IAMRole spec for resource %s", resource.Metadata.Name)
	}

	roleName := g.sanitizeResourceName(resource.Metadata.Name)

	g.logger.WithField("iam_role", resource.Metadata.Name).Debug("Generating IAM role module")

	// Create module block
	moduleBlock := body.AppendNewBlock("module", []string{roleName})
	moduleBody := moduleBlock.Body()

	// Set module source
	moduleBody.SetAttributeValue("source", cty.StringVal(fmt.Sprintf("%s//modules/iam-role?ref=%s",
		g.config.ModuleRegistry, g.config.ModuleVersion)))

	// Set basic attributes
	moduleBody.SetAttributeValue("role_name", cty.StringVal(resource.Metadata.Name))

	if roleSpec.Description != "" {
		moduleBody.SetAttributeValue("description", cty.StringVal(roleSpec.Description))
	}

	// Set assume role policy
	if roleSpec.AssumeRolePolicy != nil {
		assumeRolePolicy := g.buildAssumeRolePolicy(roleSpec.AssumeRolePolicy)
		moduleBody.SetAttributeValue("assume_role_policy", assumeRolePolicy)
	}

	// Set managed policies
	if len(roleSpec.Policies) > 0 {
		var policies []cty.Value
		for _, policy := range roleSpec.Policies {
			policyObj := cty.ObjectVal(map[string]cty.Value{
				"policy_arn": cty.StringVal(policy.PolicyArn),
			})
			if policy.PolicyName != "" {
				policyObj = cty.ObjectVal(map[string]cty.Value{
					"policy_arn":  cty.StringVal(policy.PolicyArn),
					"policy_name": cty.StringVal(policy.PolicyName),
				})
			}
			policies = append(policies, policyObj)
		}
		moduleBody.SetAttributeValue("managed_policies", cty.ListVal(policies))
	}

	// Set inline policies
	if len(roleSpec.InlinePolicies) > 0 {
		var inlinePolicies []cty.Value
		for _, inlinePolicy := range roleSpec.InlinePolicies {
			policyDoc := g.buildPolicyDocument(&inlinePolicy.Policy)
			inlinePolicyObj := cty.ObjectVal(map[string]cty.Value{
				"name":   cty.StringVal(inlinePolicy.Name),
				"policy": policyDoc,
			})
			inlinePolicies = append(inlinePolicies, inlinePolicyObj)
		}
		moduleBody.SetAttributeValue("inline_policies", cty.ListVal(inlinePolicies))
	}

	// Set tags
	if len(roleSpec.Tags) > 0 {
		tags := make(map[string]cty.Value)
		for k, v := range roleSpec.Tags {
			tags[k] = cty.StringVal(v)
		}
		moduleBody.SetAttributeValue("tags", cty.ObjectVal(tags))
	}

	body.AppendNewline()

	g.logger.WithField("iam_role", resource.Metadata.Name).Info("Generated IAM role module")
	return nil
}

// buildAssumeRolePolicy converts AssumeRolePolicy to cty.Value
func (g *HCLGenerator) buildAssumeRolePolicy(policy *models.AssumeRolePolicy) cty.Value {
	statements := make([]cty.Value, len(policy.Statement))

	for i, stmt := range policy.Statement {
		statementObj := map[string]cty.Value{
			"effect": cty.StringVal(stmt.Effect),
		}

		// Handle principal
		if len(stmt.Principal) > 0 {
			principalMap := make(map[string]cty.Value)
			for k, v := range stmt.Principal {
				switch val := v.(type) {
				case string:
					principalMap[k] = cty.StringVal(val)
				case []interface{}:
					var values []cty.Value
					for _, item := range val {
						if str, ok := item.(string); ok {
							values = append(values, cty.StringVal(str))
						}
					}
					principalMap[k] = cty.ListVal(values)
				}
			}
			statementObj["principal"] = cty.ObjectVal(principalMap)
		}

		// Handle action
		switch action := stmt.Action.(type) {
		case string:
			statementObj["action"] = cty.StringVal(action)
		case []interface{}:
			var actions []cty.Value
			for _, act := range action {
				if str, ok := act.(string); ok {
					actions = append(actions, cty.StringVal(str))
				}
			}
			statementObj["action"] = cty.ListVal(actions)
		}

		// Handle condition if present
		if len(stmt.Condition) > 0 {
			conditionMap := make(map[string]cty.Value)
			for k, v := range stmt.Condition {
				// This is simplified - conditions can be complex
				if str, ok := v.(string); ok {
					conditionMap[k] = cty.StringVal(str)
				}
			}
			statementObj["condition"] = cty.ObjectVal(conditionMap)
		}

		statements[i] = cty.ObjectVal(statementObj)
	}

	return cty.ObjectVal(map[string]cty.Value{
		"version":   cty.StringVal(policy.Version),
		"statement": cty.ListVal(statements),
	})
}

// buildPolicyDocument converts IAMPolicyDocument to cty.Value
func (g *HCLGenerator) buildPolicyDocument(policy *models.IAMPolicyDocument) cty.Value {
	statements := make([]cty.Value, len(policy.Statement))

	for i, stmt := range policy.Statement {
		statementObj := map[string]cty.Value{
			"effect": cty.StringVal(stmt.Effect),
		}

		if stmt.Sid != "" {
			statementObj["sid"] = cty.StringVal(stmt.Sid)
		}

		// Handle action - normalize to list
		var actions []cty.Value
		switch action := stmt.Action.(type) {
		case string:
			actions = append(actions, cty.StringVal(action))
		case []interface{}:
			for _, act := range action {
				if str, ok := act.(string); ok {
					actions = append(actions, cty.StringVal(str))
				}
			}
		}
		if len(actions) > 0 {
			statementObj["action"] = cty.ListVal(actions)
		}

		// Handle resource - normalize to list
		var resources []cty.Value
		switch resource := stmt.Resource.(type) {
		case string:
			resources = append(resources, cty.StringVal(resource))
		case []interface{}:
			for _, res := range resource {
				if str, ok := res.(string); ok {
					resources = append(resources, cty.StringVal(str))
				}
			}
		}
		if len(resources) > 0 {
			statementObj["resource"] = cty.ListVal(resources)
		}

		// Handle condition if present
		if len(stmt.Condition) > 0 {
			conditionMap := make(map[string]cty.Value)
			for k, v := range stmt.Condition {
				if str, ok := v.(string); ok {
					conditionMap[k] = cty.StringVal(str)
				}
			}
			statementObj["condition"] = cty.ObjectVal(conditionMap)
		}

		statements[i] = cty.ObjectVal(statementObj)
	}

	return cty.ObjectVal(map[string]cty.Value{
		"version":   cty.StringVal(policy.Version),
		"statement": cty.ListVal(statements),
	})
}
