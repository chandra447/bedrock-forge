package generator

import (
	"fmt"
	"encoding/json"

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
	
	// Default variant
	if prompt.DefaultVariant != "" {
		moduleBody.SetAttributeValue("default_variant", cty.StringVal(prompt.DefaultVariant))
	}
	
	// Variants configuration
	if len(prompt.Variants) > 0 {
		variantsList := make([]cty.Value, 0, len(prompt.Variants))
		
		for _, variant := range prompt.Variants {
			variantValues := make(map[string]cty.Value)
			variantValues["name"] = cty.StringVal(variant.Name)
			variantValues["model_id"] = cty.StringVal(variant.ModelId)
			variantValues["template_type"] = cty.StringVal(variant.TemplateType)
			
			// Template configuration - always include
			templateValues := make(map[string]cty.Value)
			if variant.TemplateConfiguration != nil {
				templateValues["text"] = cty.StringVal(variant.TemplateConfiguration.Text)
			} else {
				templateValues["text"] = cty.StringVal("")
			}
			variantValues["template_configuration"] = cty.ObjectVal(templateValues)
			
			// Inference configuration - always include with all fields
			inferenceValues := make(map[string]cty.Value)
			textInferenceValues := make(map[string]cty.Value)
			
			if variant.InferenceConfiguration != nil && variant.InferenceConfiguration.Text != nil {
				textConfig := variant.InferenceConfiguration.Text
				
				textInferenceValues["temperature"] = cty.NumberFloatVal(textConfig.Temperature)
				textInferenceValues["top_p"] = cty.NumberFloatVal(textConfig.TopP)
				textInferenceValues["max_tokens"] = cty.NumberIntVal(int64(textConfig.MaxTokens))
				
				if len(textConfig.StopSequences) > 0 {
					stopSequences := make([]cty.Value, 0, len(textConfig.StopSequences))
					for _, stopSequence := range textConfig.StopSequences {
						stopSequences = append(stopSequences, cty.StringVal(stopSequence))
					}
					textInferenceValues["stop_sequences"] = cty.ListVal(stopSequences)
				} else {
					textInferenceValues["stop_sequences"] = cty.NullVal(cty.List(cty.String))
				}
			} else {
				// Default values to ensure consistent structure
				textInferenceValues["temperature"] = cty.NumberFloatVal(0)
				textInferenceValues["top_p"] = cty.NumberFloatVal(0)
				textInferenceValues["max_tokens"] = cty.NumberIntVal(0)
				textInferenceValues["stop_sequences"] = cty.NullVal(cty.List(cty.String))
			}
			
			inferenceValues["text"] = cty.ObjectVal(textInferenceValues)
			variantValues["inference_configuration"] = cty.ObjectVal(inferenceValues)
			
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