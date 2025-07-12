package generator

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"

	"bedrock-forge/internal/models"
)

// generateAgentNative creates a native AWS Terraform resource for an Agent
func (g *HCLGenerator) generateAgentNative(body *hclwrite.Body, resource models.BaseResource) error {
	agent, ok := resource.Spec.(models.AgentSpec)
	if !ok {
		// Try to parse as map and convert to AgentSpec
		specMap, mapOk := resource.Spec.(map[string]interface{})
		if !mapOk {
			return fmt.Errorf("invalid agent spec format")
		}

		// Convert map to AgentSpec
		specJSON, err := json.Marshal(specMap)
		if err != nil {
			return fmt.Errorf("failed to marshal agent spec: %w", err)
		}

		if err := json.Unmarshal(specJSON, &agent); err != nil {
			return fmt.Errorf("failed to unmarshal agent spec: %w", err)
		}
	}

	resourceName := g.sanitizeResourceName(resource.Metadata.Name)

	// Generate IAM role for the agent if not provided by user
	if err := g.handleAgentExecutionRole(body, resource.Metadata.Name, agent); err != nil {
		return fmt.Errorf("failed to handle agent execution role: %w", err)
	}

	// Create native AWS resource block
	resourceBlock := body.AppendNewBlock("resource", []string{"aws_bedrockagent_agent", resourceName})
	resourceBody := resourceBlock.Body()

	// Set basic attributes according to AWS provider schema
	resourceBody.SetAttributeValue("agent_name", cty.StringVal(resource.Metadata.Name))
	resourceBody.SetAttributeValue("foundation_model", cty.StringVal(agent.FoundationModel))
	resourceBody.SetAttributeValue("instruction", cty.StringVal(agent.Instruction))

	// IAM role reference - handle both auto-generated and user-provided roles
	if err := g.setAgentRoleReference(resourceBody, resource.Metadata.Name, agent); err != nil {
		return fmt.Errorf("failed to set agent role reference: %w", err)
	}

	// Optional attributes according to AWS provider schema
	if agent.Description != "" {
		resourceBody.SetAttributeValue("description", cty.StringVal(agent.Description))
	}

	if agent.IdleSessionTTL > 0 {
		resourceBody.SetAttributeValue("idle_session_ttl_in_seconds", cty.NumberIntVal(int64(agent.IdleSessionTTL)))
	}

	if agent.CustomerEncryptionKey != "" {
		resourceBody.SetAttributeValue("customer_encryption_key_arn", cty.StringVal(agent.CustomerEncryptionKey))
	}

	// Tags
	if len(agent.Tags) > 0 {
		tagValues := make(map[string]cty.Value)
		for key, value := range agent.Tags {
			tagValues[key] = cty.StringVal(value)
		}
		resourceBody.SetAttributeValue("tags", cty.ObjectVal(tagValues))
	}

	// Terraform-specific attributes
	if agent.PrepareAgent != nil {
		resourceBody.SetAttributeValue("prepare_agent", cty.BoolVal(*agent.PrepareAgent))
	}

	if agent.SkipResourceInUseCheck != nil {
		resourceBody.SetAttributeValue("skip_resource_in_use_check", cty.BoolVal(*agent.SkipResourceInUseCheck))
	}

	// Timeouts configuration
	if agent.Timeouts != nil {
		timeoutValues := make(map[string]cty.Value)
		if agent.Timeouts.Create != "" {
			timeoutValues["create"] = cty.StringVal(agent.Timeouts.Create)
		}
		if agent.Timeouts.Update != "" {
			timeoutValues["update"] = cty.StringVal(agent.Timeouts.Update)
		}
		if agent.Timeouts.Delete != "" {
			timeoutValues["delete"] = cty.StringVal(agent.Timeouts.Delete)
		}
		if len(timeoutValues) > 0 {
			resourceBody.SetAttributeValue("timeouts", cty.ObjectVal(timeoutValues))
		}
	}

	body.AppendNewline()

	// Generate separate action group resources if specified
	if len(agent.ActionGroups) > 0 {
		if err := g.generateAgentActionGroups(body, resource.Metadata.Name, agent.ActionGroups); err != nil {
			return fmt.Errorf("failed to generate agent action groups: %w", err)
		}
	}

	// Generate agent aliases if specified
	if len(agent.Aliases) > 0 {
		if err := g.generateAgentAliases(body, resource.Metadata.Name, agent.Aliases); err != nil {
			return fmt.Errorf("failed to generate agent aliases: %w", err)
		}
	}

	g.logger.WithField("agent", resource.Metadata.Name).Info("Generated native agent resource")
	return nil
}

// generateAgentActionGroups creates separate aws_bedrockagent_agent_action_group resources
func (g *HCLGenerator) generateAgentActionGroups(body *hclwrite.Body, agentName string, actionGroups []models.InlineActionGroup) error {
	agentResourceName := g.sanitizeResourceName(agentName)

	for _, ag := range actionGroups {
		agResourceName := fmt.Sprintf("%s_%s", agentResourceName, g.sanitizeResourceName(ag.Name))

		// Create action group resource
		agBlock := body.AppendNewBlock("resource", []string{"aws_bedrockagent_agent_action_group", agResourceName})
		agBody := agBlock.Body()

		// Required attributes
		agBody.SetAttributeRaw("agent_id", hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(fmt.Sprintf("aws_bedrockagent_agent.%s.agent_id", agentResourceName))},
		})
		agBody.SetAttributeValue("agent_version", cty.StringVal("DRAFT"))
		agBody.SetAttributeValue("action_group_name", cty.StringVal(ag.Name))
		agBody.SetAttributeValue("skip_resource_in_use_check", cty.BoolVal(true))

		if ag.Description != "" {
			agBody.SetAttributeValue("description", cty.StringVal(ag.Description))
		}

		if ag.ActionGroupState != "" {
			agBody.SetAttributeValue("action_group_state", cty.StringVal(ag.ActionGroupState))
		} else {
			agBody.SetAttributeValue("action_group_state", cty.StringVal("ENABLED"))
		}

		if ag.ParentActionGroupSignature != "" {
			agBody.SetAttributeValue("parent_action_group_signature", cty.StringVal(ag.ParentActionGroupSignature))
		}

		// Action group executor configuration (using block syntax)
		if ag.ActionGroupExecutor != nil {
			executorBlock := agBody.AppendNewBlock("action_group_executor", nil)
			executorBody := executorBlock.Body()

			if !ag.ActionGroupExecutor.Lambda.IsEmpty() {
				// Reference to a Lambda resource
				lambdaResourceName := g.sanitizeResourceName(ag.ActionGroupExecutor.Lambda.String())
				executorBody.SetAttributeRaw("lambda", hclwrite.Tokens{
					{Type: hclsyntax.TokenIdent, Bytes: []byte(fmt.Sprintf("aws_lambda_function.%s.arn", lambdaResourceName))},
				})
			} else if ag.ActionGroupExecutor.LambdaArn != "" {
				// Direct Lambda ARN
				executorBody.SetAttributeValue("lambda", cty.StringVal(ag.ActionGroupExecutor.LambdaArn))
			} else if ag.ActionGroupExecutor.CustomControl != "" {
				executorBody.SetAttributeValue("custom_control", cty.StringVal(ag.ActionGroupExecutor.CustomControl))
			}
		}

		// API Schema configuration (using block syntax)
		if ag.APISchema != nil {
			apiSchemaBlock := agBody.AppendNewBlock("api_schema", nil)
			apiSchemaBody := apiSchemaBlock.Body()

			if ag.APISchema.S3 != nil {
				s3Block := apiSchemaBody.AppendNewBlock("s3", nil)
				s3Body := s3Block.Body()
				s3Body.SetAttributeValue("s3_bucket_name", cty.StringVal(ag.APISchema.S3.S3BucketName))
				s3Body.SetAttributeValue("s3_object_key", cty.StringVal(ag.APISchema.S3.S3ObjectKey))
			} else if ag.APISchema.Payload != "" {
				apiSchemaBody.SetAttributeValue("payload", cty.StringVal(ag.APISchema.Payload))
			}
		}

		// Function Schema configuration (using proper block syntax)
		if ag.FunctionSchema != nil {
			functionSchemaBlock := agBody.AppendNewBlock("function_schema", nil)
			functionSchemaBody := functionSchemaBlock.Body()

			// Create member_functions block
			memberFunctionsBlock := functionSchemaBody.AppendNewBlock("member_functions", nil)
			memberFunctionsBody := memberFunctionsBlock.Body()

			// Add functions
			for _, fn := range ag.FunctionSchema.Functions {
				functionBlock := memberFunctionsBody.AppendNewBlock("functions", nil)
				functionBody := functionBlock.Body()

				functionBody.SetAttributeValue("name", cty.StringVal(fn.Name))
				if fn.Description != "" {
					functionBody.SetAttributeValue("description", cty.StringVal(fn.Description))
				}

				// Add parameters
				for paramName, param := range fn.Parameters {
					paramBlock := functionBody.AppendNewBlock("parameters", nil)
					paramBody := paramBlock.Body()

					paramBody.SetAttributeValue("map_block_key", cty.StringVal(paramName))
					paramBody.SetAttributeValue("type", cty.StringVal(param.Type))
					paramBody.SetAttributeValue("required", cty.BoolVal(param.Required))
					if param.Description != "" {
						paramBody.SetAttributeValue("description", cty.StringVal(param.Description))
					}
				}
			}
		}

		body.AppendNewline()
	}

	return nil
}

// generateAgentExecutionRoleNative creates a native AWS IAM role for the agent
func (g *HCLGenerator) generateAgentExecutionRoleNative(body *hclwrite.Body, agentName string, agent models.AgentSpec) error {
	roleResourceName := fmt.Sprintf("%s_execution_role", g.sanitizeResourceName(agentName))

	// Create IAM role resource
	roleBlock := body.AppendNewBlock("resource", []string{"aws_iam_role", roleResourceName})
	roleBody := roleBlock.Body()

	roleBody.SetAttributeValue("name", cty.StringVal(fmt.Sprintf("%s-execution-role", agentName)))
	roleBody.SetAttributeValue("assume_role_policy", cty.StringVal(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Effect": "Allow",
      "Principal": {
        "Service": "bedrock.amazonaws.com"
      }
    }
  ]
}`))

	// Create IAM role policy attachment for Bedrock service
	bedrockPolicyAttachmentBlock := body.AppendNewBlock("resource", []string{"aws_iam_role_policy_attachment", fmt.Sprintf("%s_bedrock_policy", roleResourceName)})
	bedrockPolicyAttachmentBody := bedrockPolicyAttachmentBlock.Body()

	bedrockPolicyAttachmentBody.SetAttributeRaw("role", hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(fmt.Sprintf("aws_iam_role.%s.name", roleResourceName))},
	})
	bedrockPolicyAttachmentBody.SetAttributeValue("policy_arn", cty.StringVal("arn:aws:iam::aws:policy/AmazonBedrockFullAccess"))

	// Build specific Lambda ARNs from action groups
	lambdaArns := g.buildLambdaArnsFromActionGroups(agent.ActionGroups)

	// Create inline policy for specific Bedrock agent permissions
	inlinePolicyBlock := body.AppendNewBlock("resource", []string{"aws_iam_role_policy", fmt.Sprintf("%s_inline_policy", roleResourceName)})
	inlinePolicyBody := inlinePolicyBlock.Body()

	inlinePolicyBody.SetAttributeValue("name", cty.StringVal("BedrockAgentExecutionPolicy"))
	inlinePolicyBody.SetAttributeRaw("role", hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(fmt.Sprintf("aws_iam_role.%s.id", roleResourceName))},
	})

	// Generate policy with specific Lambda ARNs
	policyJson := g.buildAgentExecutionPolicy(lambdaArns)
	inlinePolicyBody.SetAttributeValue("policy", cty.StringVal(policyJson))

	body.AppendNewline()

	g.logger.WithField("agent", agentName).Info("Generated native agent execution role")
	return nil
}

// buildLambdaArnsFromActionGroups extracts Lambda function references from action groups
func (g *HCLGenerator) buildLambdaArnsFromActionGroups(actionGroups []models.InlineActionGroup) []string {
	var lambdaArns []string

	for _, ag := range actionGroups {
		if ag.ActionGroupExecutor != nil {
			if !ag.ActionGroupExecutor.Lambda.IsEmpty() {
				// Reference to a Lambda resource
				lambdaResourceName := g.sanitizeResourceName(ag.ActionGroupExecutor.Lambda.String())
				lambdaArn := fmt.Sprintf("aws_lambda_function.%s.arn", lambdaResourceName)
				lambdaArns = append(lambdaArns, lambdaArn)
			} else if ag.ActionGroupExecutor.LambdaArn != "" {
				// Direct Lambda ARN
				lambdaArns = append(lambdaArns, ag.ActionGroupExecutor.LambdaArn)
			}
		}
	}

	return lambdaArns
}

// buildAgentExecutionPolicy creates the IAM policy JSON with specific Lambda ARNs
func (g *HCLGenerator) buildAgentExecutionPolicy(lambdaArns []string) string {
	// Build Lambda resource array
	lambdaResourcesJson := ""
	if len(lambdaArns) > 0 {
		resources := make([]string, len(lambdaArns))
		for i, arn := range lambdaArns {
			// Check if it's a Terraform reference or direct ARN
			if strings.HasPrefix(arn, "aws_lambda_function.") {
				resources[i] = fmt.Sprintf("        \"${%s}\"", arn)
			} else {
				resources[i] = fmt.Sprintf("        \"%s\"", arn)
			}
		}
		lambdaResourcesJson = strings.Join(resources, ",\n")
	} else {
		// Fallback to wildcard if no Lambda functions found
		lambdaResourcesJson = "        \"arn:aws:lambda:*:*:function:*\""
	}

	return fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "bedrock:InvokeModel",
        "bedrock:InvokeModelWithResponseStream"
      ],
      "Resource": "arn:aws:bedrock:*::foundation-model/*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "bedrock:GetInferenceProfile",
        "bedrock:ListInferenceProfiles",
        "bedrock:UseInferenceProfile"
      ],
      "Resource": "arn:aws:bedrock:*:*:inference-profile/*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "lambda:InvokeFunction"
      ],
      "Resource": [
%s
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "bedrock:Retrieve",
        "bedrock:RetrieveAndGenerate"
      ],
      "Resource": "arn:aws:bedrock:*:*:knowledge-base/*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "arn:aws:logs:*:*:*"
    }
  ]
}`, lambdaResourcesJson)
}

// handleAgentExecutionRole determines whether to generate an IAM role or use an existing one
func (g *HCLGenerator) handleAgentExecutionRole(body *hclwrite.Body, agentName string, agent models.AgentSpec) error {
	// Check if user has provided IAM role configuration
	if agent.IAMRole != nil {
		// User has provided IAM role configuration
		if agent.IAMRole.RoleArn != "" {
			// User provided existing role ARN - no need to generate
			g.logger.WithField("agent", agentName).WithField("roleArn", agent.IAMRole.RoleArn).Info("Using existing IAM role ARN")
			return nil
		}

		if !agent.IAMRole.RoleName.IsEmpty() {
			// User provided reference to IAMRole resource - no need to generate
			g.logger.WithField("agent", agentName).WithField("roleName", agent.IAMRole.RoleName.String()).Info("Using referenced IAM role")
			return nil
		}

		if agent.IAMRole.AutoCreate != nil && !*agent.IAMRole.AutoCreate {
			// User explicitly disabled auto-creation
			g.logger.WithField("agent", agentName).Warn("IAM role auto-creation disabled but no existing role provided")
			return fmt.Errorf("IAM role auto-creation disabled but no existing role ARN or reference provided")
		}
	}

	// Default behavior: auto-generate IAM role
	g.logger.WithField("agent", agentName).Info("Auto-generating IAM role")
	return g.generateAgentExecutionRoleNative(body, agentName, agent)
}

// setAgentRoleReference sets the appropriate IAM role reference based on configuration
func (g *HCLGenerator) setAgentRoleReference(resourceBody *hclwrite.Body, agentName string, agent models.AgentSpec) error {
	if agent.IAMRole != nil {
		// User has provided IAM role configuration
		if agent.IAMRole.RoleArn != "" {
			// Direct ARN
			resourceBody.SetAttributeValue("agent_resource_role_arn", cty.StringVal(agent.IAMRole.RoleArn))
			return nil
		}

		if !agent.IAMRole.RoleName.IsEmpty() {
			// Reference to IAMRole resource
			roleResourceName := g.sanitizeResourceName(agent.IAMRole.RoleName.String())
			resourceBody.SetAttributeRaw("agent_resource_role_arn", hclwrite.Tokens{
				{Type: hclsyntax.TokenIdent, Bytes: []byte(fmt.Sprintf("aws_iam_role.%s.arn", roleResourceName))},
			})
			return nil
		}
	}

	// Default: reference auto-generated role
	agentRoleName := fmt.Sprintf("%s_execution_role", g.sanitizeResourceName(agentName))
	resourceBody.SetAttributeRaw("agent_resource_role_arn", hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(fmt.Sprintf("aws_iam_role.%s.arn", agentRoleName))},
	})
	return nil
}
