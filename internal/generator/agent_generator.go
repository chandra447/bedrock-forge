package generator

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"

	"bedrock-forge/internal/models"
)

// generateAgentModule creates a native AWS Terraform resource for an Agent
func (g *HCLGenerator) generateAgentModule(body *hclwrite.Body, resource models.BaseResource) error {
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

	// Create native AWS resource block
	resourceBlock := body.AppendNewBlock("resource", []string{"aws_bedrockagent_agent", resourceName})
	resourceBody := resourceBlock.Body()

	// Set basic attributes according to AWS provider schema
	resourceBody.SetAttributeValue("agent_name", cty.StringVal(resource.Metadata.Name))
	resourceBody.SetAttributeValue("foundation_model", cty.StringVal(agent.FoundationModel))
	resourceBody.SetAttributeValue("instruction", cty.StringVal(agent.Instruction))

	// IAM role is generated separately and referenced via resource output
	agentRoleName := fmt.Sprintf("%s_execution_role", g.sanitizeResourceName(resource.Metadata.Name))
	resourceBody.SetAttributeValue("agent_resource_role_arn", cty.StringVal(fmt.Sprintf("${aws_iam_role.%s.arn}", agentRoleName)))

	// Optional attributes according to AWS provider schema
	if agent.Description != "" {
		resourceBody.SetAttributeValue("agent_description", cty.StringVal(agent.Description))
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

	// New Terraform-specific attributes
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

	// Guardrail configuration
	if agent.Guardrail != nil && !agent.Guardrail.Name.IsEmpty() {
		guardrailValues := make(map[string]cty.Value)
		guardrailValues["name"] = cty.StringVal(agent.Guardrail.Name.String())

		if agent.Guardrail.Version != "" {
			guardrailValues["version"] = cty.StringVal(agent.Guardrail.Version)
		}

		if agent.Guardrail.Mode != "" {
			guardrailValues["mode"] = cty.StringVal(agent.Guardrail.Mode)
		}

		// Resolve guardrail reference
		if guardrailId, err := g.resolveReferenceToOutput(agent.Guardrail.Name, models.GuardrailKind, "guardrail_id"); err == nil {
			guardrailValues["guardrail_id"] = cty.StringVal(guardrailId)
			if guardrailVersion, err := g.resolveReferenceToOutput(agent.Guardrail.Name, models.GuardrailKind, "guardrail_version"); err == nil {
				guardrailValues["guardrail_version"] = cty.StringVal(guardrailVersion)
			}
		} else {
			g.logger.WithError(err).WithField("guardrail", agent.Guardrail.Name.String()).Warn("Failed to resolve guardrail reference")
		}

		resourceBody.SetAttributeValue("guardrail", cty.ObjectVal(guardrailValues))
	}

	// Knowledge bases are handled through separate association resources

	// Inline action groups
	if len(agent.ActionGroups) > 0 {
		agList := make([]cty.Value, 0, len(agent.ActionGroups))
		for _, ag := range agent.ActionGroups {
			agValues := make(map[string]cty.Value)
			agValues["name"] = cty.StringVal(ag.Name)

			if ag.Description != "" {
				agValues["description"] = cty.StringVal(ag.Description)
			}

			if ag.ParentActionGroupSignature != "" {
				agValues["parent_action_group_signature"] = cty.StringVal(ag.ParentActionGroupSignature)
			}

			if ag.ActionGroupState != "" {
				agValues["action_group_state"] = cty.StringVal(ag.ActionGroupState)
			}

			if ag.SkipResourceInUseCheck {
				agValues["skip_resource_in_use_check"] = cty.BoolVal(true)
			}

			// Action group executor configuration
			if ag.ActionGroupExecutor != nil {
				executorValues := make(map[string]cty.Value)

				if !ag.ActionGroupExecutor.Lambda.IsEmpty() {
					// Reference to a Lambda resource
					if lambdaArn, err := g.resolveReferenceToOutput(ag.ActionGroupExecutor.Lambda, models.LambdaKind, "lambda_function_arn"); err == nil {
						executorValues["lambda"] = cty.StringVal(lambdaArn)
					} else {
						g.logger.WithError(err).WithField("lambda", ag.ActionGroupExecutor.Lambda.String()).Warn("Failed to resolve lambda reference")
					}
				} else if ag.ActionGroupExecutor.LambdaArn != "" {
					// Direct Lambda ARN
					executorValues["lambda"] = cty.StringVal(ag.ActionGroupExecutor.LambdaArn)
				} else if ag.ActionGroupExecutor.CustomControl != "" {
					executorValues["custom_control"] = cty.StringVal(ag.ActionGroupExecutor.CustomControl)
				}

				agValues["action_group_executor"] = cty.ObjectVal(executorValues)
			}

			// API Schema configuration
			if ag.APISchema != nil {
				schemaValues := make(map[string]cty.Value)

				if ag.APISchema.S3 != nil {
					s3Values := make(map[string]cty.Value)
					s3Values["s3_bucket_name"] = cty.StringVal(ag.APISchema.S3.S3BucketName)
					s3Values["s3_object_key"] = cty.StringVal(ag.APISchema.S3.S3ObjectKey)
					schemaValues["s3"] = cty.ObjectVal(s3Values)
				} else if ag.APISchema.Payload != "" {
					schemaValues["payload"] = cty.StringVal(ag.APISchema.Payload)
				}

				agValues["api_schema"] = cty.ObjectVal(schemaValues)
			}

			// Function Schema configuration
			if ag.FunctionSchema != nil {
				functionsList := make([]cty.Value, 0, len(ag.FunctionSchema.Functions))

				for _, fn := range ag.FunctionSchema.Functions {
					fnValues := make(map[string]cty.Value)
					fnValues["name"] = cty.StringVal(fn.Name)

					if fn.Description != "" {
						fnValues["description"] = cty.StringVal(fn.Description)
					}

					if len(fn.Parameters) > 0 {
						paramValues := make(map[string]cty.Value)
						for paramName, param := range fn.Parameters {
							paramObj := make(map[string]cty.Value)
							paramObj["type"] = cty.StringVal(param.Type)
							paramObj["required"] = cty.BoolVal(param.Required)
							if param.Description != "" {
								paramObj["description"] = cty.StringVal(param.Description)
							}
							paramValues[paramName] = cty.ObjectVal(paramObj)
						}
						fnValues["parameters"] = cty.ObjectVal(paramValues)
					}

					functionsList = append(functionsList, cty.ObjectVal(fnValues))
				}

				functionSchemaValues := make(map[string]cty.Value)
				functionSchemaValues["functions"] = cty.ListVal(functionsList)
				agValues["function_schema"] = cty.ObjectVal(functionSchemaValues)
			}

			agList = append(agList, cty.ObjectVal(agValues))
		}
		resourceBody.SetAttributeValue("action_groups", cty.ListVal(agList))
	}

	// Prompt overrides
	if len(agent.PromptOverrides) > 0 {
		poList := make([]cty.Value, 0, len(agent.PromptOverrides))
		for _, po := range agent.PromptOverrides {
			poValues := make(map[string]cty.Value)
			poValues["prompt_type"] = cty.StringVal(po.PromptType)

			// Always include these fields to ensure consistent structure
			if po.PromptArn != "" {
				poValues["prompt_arn"] = cty.StringVal(po.PromptArn)
			} else if !po.Prompt.IsEmpty() {
				// Reference to a prompt module
				if promptArn, err := g.resolveReferenceToOutput(po.Prompt, models.PromptKind, "prompt_arn"); err == nil {
					poValues["prompt_arn"] = cty.StringVal(promptArn)
				} else {
					g.logger.WithError(err).WithField("prompt", po.Prompt.String()).Warn("Failed to resolve prompt reference")
					poValues["prompt_arn"] = cty.StringVal("")
				}
			} else {
				poValues["prompt_arn"] = cty.StringVal("")
			}

			// Use unified variant field
			variant := ""
			if po.PromptVariant != "" {
				variant = po.PromptVariant
			} else if po.Variant != "" {
				variant = po.Variant
			}
			poValues["variant"] = cty.StringVal(variant)

			poList = append(poList, cty.ObjectVal(poValues))
		}
		resourceBody.SetAttributeValue("prompt_overrides", cty.ListVal(poList))
	}

	// Memory configuration
	if agent.MemoryConfiguration != nil {
		memoryValues := make(map[string]cty.Value)

		if len(agent.MemoryConfiguration.EnabledMemoryTypes) > 0 {
			memoryTypes := make([]cty.Value, 0, len(agent.MemoryConfiguration.EnabledMemoryTypes))
			for _, memType := range agent.MemoryConfiguration.EnabledMemoryTypes {
				memoryTypes = append(memoryTypes, cty.StringVal(memType))
			}
			memoryValues["enabled_memory_types"] = cty.ListVal(memoryTypes)
		}

		if agent.MemoryConfiguration.StorageDays > 0 {
			memoryValues["storage_days"] = cty.NumberIntVal(int64(agent.MemoryConfiguration.StorageDays))
		}

		resourceBody.SetAttributeValue("memory_configuration", cty.ObjectVal(memoryValues))
	}

	body.AppendNewline()

	// Generate agent aliases if specified
	if len(agent.Aliases) > 0 {
		if err := g.generateAgentAliases(body, resource.Metadata.Name, agent.Aliases); err != nil {
			return fmt.Errorf("failed to generate agent aliases: %w", err)
		}
	}

	g.logger.WithField("agent", resource.Metadata.Name).Info("Generated agent module")
	return nil
}
