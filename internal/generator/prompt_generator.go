package generator

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"

	"bedrock-forge/internal/models"
)

// generatePromptModule creates a module call for a Prompt resource
func (g *HCLGenerator) generatePromptModule(body *hclwrite.Body, resource models.BaseResource) error {
	prompt, ok := resource.Spec.(models.PromptSpec)
	if !ok {
		// Try to parse as map and convert to PromptSpec
		specMap, mapOk := resource.Spec.(map[string]interface{})
		if !mapOk {
			return fmt.Errorf("invalid prompt spec format")
		}

		// Convert map to PromptSpec
		specJSON, err := json.Marshal(specMap)
		if err != nil {
			return fmt.Errorf("failed to marshal prompt spec: %w", err)
		}

		if err := json.Unmarshal(specJSON, &prompt); err != nil {
			return fmt.Errorf("failed to unmarshal prompt spec: %w", err)
		}
	}

	resourceName := g.sanitizeResourceName(resource.Metadata.Name)

	// Create module block
	moduleBlock := body.AppendNewBlock("module", []string{resourceName})
	moduleBody := moduleBlock.Body()

	// Set module source
	moduleSource := fmt.Sprintf("%s//modules/bedrock-prompt", g.config.ModuleRegistry)
	if g.config.ModuleVersion != "" {
		moduleSource += fmt.Sprintf("?ref=%s", g.config.ModuleVersion)
	}
	moduleBody.SetAttributeValue("source", cty.StringVal(moduleSource))

	// Set basic attributes
	moduleBody.SetAttributeValue("prompt_name", cty.StringVal(resource.Metadata.Name))

	// Optional description
	if prompt.Description != "" {
		moduleBody.SetAttributeValue("description", cty.StringVal(prompt.Description))
	}

	// Customer encryption key
	if prompt.CustomerEncryptionKeyArn != "" {
		moduleBody.SetAttributeValue("customer_encryption_key_arn", cty.StringVal(prompt.CustomerEncryptionKeyArn))
	}

	// Default variant
	if prompt.DefaultVariant != "" {
		moduleBody.SetAttributeValue("default_variant", cty.StringVal(prompt.DefaultVariant))
	}

	// Input variables at prompt level
	if len(prompt.InputVariables) > 0 {
		inputVarsList := make([]cty.Value, 0, len(prompt.InputVariables))
		for _, inputVar := range prompt.InputVariables {
			inputVarsList = append(inputVarsList, cty.ObjectVal(map[string]cty.Value{
				"name": cty.StringVal(inputVar.Name),
			}))
		}
		moduleBody.SetAttributeValue("input_variables", cty.ListVal(inputVarsList))
	}

	// Variants configuration
	if len(prompt.Variants) > 0 {
		variantsList := make([]cty.Value, 0, len(prompt.Variants))

		for _, variant := range prompt.Variants {
			variantValues := make(map[string]cty.Value)
			variantValues["name"] = cty.StringVal(variant.Name)
			variantValues["model_id"] = cty.StringVal(variant.ModelId)
			variantValues["template_type"] = cty.StringVal(variant.TemplateType)

			// Template configuration based on type
			if variant.TemplateConfiguration != nil {
				templateConfig, err := g.generateTemplateConfiguration(variant.TemplateConfiguration, variant.TemplateType)
				if err != nil {
					return fmt.Errorf("failed to generate template configuration: %w", err)
				}
				variantValues["template_configuration"] = templateConfig
			}

			// Inference configuration
			if variant.InferenceConfiguration != nil {
				inferenceConfig, err := g.generateInferenceConfiguration(variant.InferenceConfiguration)
				if err != nil {
					return fmt.Errorf("failed to generate inference configuration: %w", err)
				}
				variantValues["inference_configuration"] = inferenceConfig
			}

			// Gen AI Resource configuration
			if variant.GenAiResource != nil {
				genAiConfig, err := g.generateGenAiResourceConfiguration(variant.GenAiResource)
				if err != nil {
					return fmt.Errorf("failed to generate gen AI resource configuration: %w", err)
				}
				variantValues["gen_ai_resource"] = genAiConfig
			}

			variantsList = append(variantsList, cty.ObjectVal(variantValues))
		}

		moduleBody.SetAttributeValue("variants", cty.ListVal(variantsList))
	}

	// Tags
	if len(prompt.Tags) > 0 {
		tagValues := make(map[string]cty.Value)
		for key, value := range prompt.Tags {
			tagValues[key] = cty.StringVal(value)
		}
		moduleBody.SetAttributeValue("tags", cty.ObjectVal(tagValues))
	}

	body.AppendNewline()

	g.logger.WithField("prompt", resource.Metadata.Name).Info("Generated prompt module")
	return nil
}

// generateTemplateConfiguration generates template configuration based on type
func (g *HCLGenerator) generateTemplateConfiguration(templateConfig *models.TemplateConfiguration, templateType string) (cty.Value, error) {
	templateValues := make(map[string]cty.Value)

	switch templateType {
	case "TEXT":
		if templateConfig.Text != nil {
			textValues := make(map[string]cty.Value)
			textValues["text"] = cty.StringVal(templateConfig.Text.Text)

			// Input variables for text template
			if len(templateConfig.Text.InputVariables) > 0 {
				inputVarsList := make([]cty.Value, 0, len(templateConfig.Text.InputVariables))
				for _, inputVar := range templateConfig.Text.InputVariables {
					inputVarsList = append(inputVarsList, cty.ObjectVal(map[string]cty.Value{
						"name": cty.StringVal(inputVar.Name),
					}))
				}
				textValues["input_variables"] = cty.ListVal(inputVarsList)
			}

			templateValues["text"] = cty.ObjectVal(textValues)
		}

	case "CHAT":
		if templateConfig.Chat != nil {
			chatValues := make(map[string]cty.Value)

			// Messages
			if len(templateConfig.Chat.Messages) > 0 {
				messagesList := make([]cty.Value, 0, len(templateConfig.Chat.Messages))
				for _, message := range templateConfig.Chat.Messages {
					messageValues := make(map[string]cty.Value)
					messageValues["role"] = cty.StringVal(message.Role)

					// Content
					if len(message.Content) > 0 {
						contentList := make([]cty.Value, 0, len(message.Content))
						for _, content := range message.Content {
							contentList = append(contentList, cty.ObjectVal(map[string]cty.Value{
								"text": cty.StringVal(content.Text),
							}))
						}
						messageValues["content"] = cty.ListVal(contentList)
					}

					messagesList = append(messagesList, cty.ObjectVal(messageValues))
				}
				chatValues["messages"] = cty.ListVal(messagesList)
			}

			// System messages
			if len(templateConfig.Chat.System) > 0 {
				systemList := make([]cty.Value, 0, len(templateConfig.Chat.System))
				for _, system := range templateConfig.Chat.System {
					systemList = append(systemList, cty.ObjectVal(map[string]cty.Value{
						"text": cty.StringVal(system.Text),
					}))
				}
				chatValues["system"] = cty.ListVal(systemList)
			}

			// Tool configuration
			if templateConfig.Chat.ToolConfiguration != nil {
				toolConfig, err := g.generateToolConfiguration(templateConfig.Chat.ToolConfiguration)
				if err != nil {
					return cty.NilVal, fmt.Errorf("failed to generate tool configuration: %w", err)
				}
				chatValues["tool_configuration"] = toolConfig
			}

			// Input variables for chat template
			if len(templateConfig.Chat.InputVariables) > 0 {
				inputVarsList := make([]cty.Value, 0, len(templateConfig.Chat.InputVariables))
				for _, inputVar := range templateConfig.Chat.InputVariables {
					inputVarsList = append(inputVarsList, cty.ObjectVal(map[string]cty.Value{
						"name": cty.StringVal(inputVar.Name),
					}))
				}
				chatValues["input_variables"] = cty.ListVal(inputVarsList)
			}

			templateValues["chat"] = cty.ObjectVal(chatValues)
		}

	default:
		return cty.NilVal, fmt.Errorf("unsupported template type: %s", templateType)
	}

	return cty.ObjectVal(templateValues), nil
}

// generateToolConfiguration generates tool configuration for chat templates
func (g *HCLGenerator) generateToolConfiguration(toolConfig *models.ToolConfiguration) (cty.Value, error) {
	toolConfigValues := make(map[string]cty.Value)

	// Tools
	if len(toolConfig.Tools) > 0 {
		toolsList := make([]cty.Value, 0, len(toolConfig.Tools))
		for _, tool := range toolConfig.Tools {
			toolValues := make(map[string]cty.Value)

			if tool.ToolSpec != nil {
				toolSpecValues := make(map[string]cty.Value)
				toolSpecValues["name"] = cty.StringVal(tool.ToolSpec.Name)
				toolSpecValues["description"] = cty.StringVal(tool.ToolSpec.Description)

				if tool.ToolSpec.InputSchema != nil && tool.ToolSpec.InputSchema.Json != nil {
					jsonBytes, err := json.Marshal(tool.ToolSpec.InputSchema.Json)
					if err != nil {
						return cty.NilVal, fmt.Errorf("failed to marshal tool input schema: %w", err)
					}
					toolSpecValues["input_schema"] = cty.ObjectVal(map[string]cty.Value{
						"json": cty.StringVal(string(jsonBytes)),
					})
				}

				toolValues["tool_spec"] = cty.ObjectVal(toolSpecValues)
			}

			toolsList = append(toolsList, cty.ObjectVal(toolValues))
		}
		toolConfigValues["tools"] = cty.ListVal(toolsList)
	}

	// Tool choice
	if toolConfig.ToolChoice != nil {
		toolChoiceValues := make(map[string]cty.Value)

		if toolConfig.ToolChoice.Auto != nil {
			toolChoiceValues["auto"] = cty.ObjectVal(map[string]cty.Value{})
		} else if toolConfig.ToolChoice.Any != nil {
			toolChoiceValues["any"] = cty.ObjectVal(map[string]cty.Value{})
		} else if toolConfig.ToolChoice.Tool != nil {
			toolChoiceValues["tool"] = cty.ObjectVal(map[string]cty.Value{
				"name": cty.StringVal(toolConfig.ToolChoice.Tool.Name),
			})
		}

		toolConfigValues["tool_choice"] = cty.ObjectVal(toolChoiceValues)
	}

	return cty.ObjectVal(toolConfigValues), nil
}

// generateInferenceConfiguration generates inference configuration
func (g *HCLGenerator) generateInferenceConfiguration(inferenceConfig *models.InferenceConfiguration) (cty.Value, error) {
	inferenceValues := make(map[string]cty.Value)

	if inferenceConfig.Text != nil {
		textInferenceValues := make(map[string]cty.Value)

		if inferenceConfig.Text.Temperature != nil {
			textInferenceValues["temperature"] = cty.NumberFloatVal(*inferenceConfig.Text.Temperature)
		}

		if inferenceConfig.Text.TopP != nil {
			textInferenceValues["top_p"] = cty.NumberFloatVal(*inferenceConfig.Text.TopP)
		}

		if inferenceConfig.Text.TopK != nil {
			textInferenceValues["top_k"] = cty.NumberIntVal(int64(*inferenceConfig.Text.TopK))
		}

		if inferenceConfig.Text.MaxTokens != nil {
			textInferenceValues["max_tokens"] = cty.NumberIntVal(int64(*inferenceConfig.Text.MaxTokens))
		}

		if len(inferenceConfig.Text.StopSequences) > 0 {
			stopSequences := make([]cty.Value, 0, len(inferenceConfig.Text.StopSequences))
			for _, stopSequence := range inferenceConfig.Text.StopSequences {
				stopSequences = append(stopSequences, cty.StringVal(stopSequence))
			}
			textInferenceValues["stop_sequences"] = cty.ListVal(stopSequences)
		}

		inferenceValues["text"] = cty.ObjectVal(textInferenceValues)
	}

	return cty.ObjectVal(inferenceValues), nil
}

// generateGenAiResourceConfiguration generates gen AI resource configuration for prompt variants
func (g *HCLGenerator) generateGenAiResourceConfiguration(genAiConfig *models.GenAiResourceConfig) (cty.Value, error) {
	genAiValues := make(map[string]cty.Value)

	if genAiConfig.Agent != nil {
		agentValues := make(map[string]cty.Value)

		if !genAiConfig.Agent.AgentName.IsEmpty() {
			// Reference to an agent YAML config in the same project
			if agentId, err := g.resolveReferenceToOutput(genAiConfig.Agent.AgentName, models.AgentKind, "agent_id"); err == nil {
				agentValues["agent_identifier"] = cty.StringVal(agentId)
				g.logger.WithField("prompt_agent", genAiConfig.Agent.AgentName.String()).Debug("Generated agent reference for prompt variant")
			} else {
				return cty.NilVal, fmt.Errorf("referenced agent '%s' not found in registry: %w", genAiConfig.Agent.AgentName.String(), err)
			}
		} else if genAiConfig.Agent.AgentArn != "" {
			// Direct ARN reference to an existing deployed agent
			agentValues["agent_identifier"] = cty.StringVal(genAiConfig.Agent.AgentArn)

			g.logger.WithField("prompt_agent_arn", genAiConfig.Agent.AgentArn).Debug("Generated agent ARN reference for prompt variant")
		} else {
			return cty.NilVal, fmt.Errorf("agent configuration must specify either agentName or agentArn")
		}

		genAiValues["agent"] = cty.ObjectVal(agentValues)
	}

	return cty.ObjectVal(genAiValues), nil
}
