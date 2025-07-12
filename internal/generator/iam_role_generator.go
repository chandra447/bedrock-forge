package generator

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"

	"bedrock-forge/internal/models"
)

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
			if !policy.PolicyName.IsEmpty() {
				policyObj = cty.ObjectVal(map[string]cty.Value{
					"policy_arn":  cty.StringVal(policy.PolicyArn),
					"policy_name": cty.StringVal(policy.PolicyName.String()),
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
