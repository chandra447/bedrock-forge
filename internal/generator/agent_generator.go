package generator

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"

	"bedrock-forge/internal/models"
)

// generateAgentModule creates a module call for an Agent resource
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

	// Create module block
	moduleBlock := body.AppendNewBlock("module", []string{resourceName})
	moduleBody := moduleBlock.Body()

	// Set module source
	moduleSource := fmt.Sprintf("%s//modules/bedrock-agent", g.config.ModuleRegistry)
	if g.config.ModuleVersion != "" {
		moduleSource += fmt.Sprintf("?ref=%s", g.config.ModuleVersion)
	}
	moduleBody.SetAttributeValue("source", cty.StringVal(moduleSource))

	// Set basic attributes
	moduleBody.SetAttributeValue("name", cty.StringVal(resource.Metadata.Name))
	moduleBody.SetAttributeValue("foundation_model", cty.StringVal(agent.FoundationModel))
	moduleBody.SetAttributeValue("instruction", cty.StringVal(agent.Instruction))

	// IAM role is generated separately and referenced via module output
	agentRoleName := fmt.Sprintf("%s_execution_role", g.sanitizeResourceName(resource.Metadata.Name))
	moduleBody.SetAttributeValue("agent_resource_role_arn", cty.StringVal(fmt.Sprintf("${module.%s.role_arn}", agentRoleName)))

	// Optional attributes
	if agent.Description != "" {
		moduleBody.SetAttributeValue("description", cty.StringVal(agent.Description))
	}

	if agent.IdleSessionTTL > 0 {
		moduleBody.SetAttributeValue("idle_session_ttl", cty.NumberIntVal(int64(agent.IdleSessionTTL)))
	}

	if agent.CustomerEncryptionKey != "" {
		moduleBody.SetAttributeValue("customer_encryption_key", cty.StringVal(agent.CustomerEncryptionKey))
	}

	// Tags
	if len(agent.Tags) > 0 {
		tagValues := make(map[string]cty.Value)
		for key, value := range agent.Tags {
			tagValues[key] = cty.StringVal(value)
		}
		moduleBody.SetAttributeValue("tags", cty.ObjectVal(tagValues))
	}

	// Guardrail configuration
	if agent.Guardrail != nil {
		guardrailValues := make(map[string]cty.Value)
		guardrailValues["name"] = cty.StringVal(agent.Guardrail.Name)

		if agent.Guardrail.Version != "" {
			guardrailValues["version"] = cty.StringVal(agent.Guardrail.Version)
		}

		if agent.Guardrail.Mode != "" {
			guardrailValues["mode"] = cty.StringVal(agent.Guardrail.Mode)
		}

		// Check if this is a reference to an existing guardrail or a new one
		if g.registry.HasResource(models.GuardrailKind, agent.Guardrail.Name) {
			// Reference to module output
			guardrailName := g.sanitizeResourceName(agent.Guardrail.Name)
			guardrailValues["guardrail_id"] = cty.StringVal(fmt.Sprintf("${module.%s.guardrail_id}", guardrailName))
			guardrailValues["guardrail_version"] = cty.StringVal(fmt.Sprintf("${module.%s.guardrail_version}", guardrailName))
		}

		moduleBody.SetAttributeValue("guardrail", cty.ObjectVal(guardrailValues))
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

				if ag.ActionGroupExecutor.Lambda != "" {
					// Reference to a Lambda resource
					if g.registry.HasResource(models.LambdaKind, ag.ActionGroupExecutor.Lambda) {
						lambdaName := g.sanitizeResourceName(ag.ActionGroupExecutor.Lambda)
						executorValues["lambda"] = cty.StringVal(fmt.Sprintf("${module.%s.lambda_function_arn}", lambdaName))
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
		moduleBody.SetAttributeValue("action_groups", cty.ListVal(agList))
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
			} else if po.Prompt != "" {
				// Reference to a prompt module
				if g.registry.HasResource(models.PromptKind, po.Prompt) {
					promptName := g.sanitizeResourceName(po.Prompt)
					poValues["prompt_arn"] = cty.StringVal(fmt.Sprintf("${module.%s.prompt_arn}", promptName))
				} else {
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
		moduleBody.SetAttributeValue("prompt_overrides", cty.ListVal(poList))
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

		moduleBody.SetAttributeValue("memory_configuration", cty.ObjectVal(memoryValues))
	}

	body.AppendNewline()

	g.logger.WithField("agent", resource.Metadata.Name).Info("Generated agent module")
	return nil
}
