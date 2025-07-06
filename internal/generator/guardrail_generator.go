package generator

import (
	"fmt"
	"encoding/json"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"

	"bedrock-forge/internal/models"
)

// generateGuardrailModule creates a module call for a Guardrail resource
func (g *HCLGenerator) generateGuardrailModule(body *hclwrite.Body, resource models.BaseResource) error {
	guardrail, ok := resource.Spec.(models.GuardrailSpec)
	if !ok {
		// Try to parse as map and convert to GuardrailSpec
		specMap, mapOk := resource.Spec.(map[string]interface{})
		if !mapOk {
			return fmt.Errorf("invalid guardrail spec format")
		}
		
		// Convert map to GuardrailSpec
		specJSON, err := json.Marshal(specMap)
		if err != nil {
			return fmt.Errorf("failed to marshal guardrail spec: %w", err)
		}
		
		if err := json.Unmarshal(specJSON, &guardrail); err != nil {
			return fmt.Errorf("failed to unmarshal guardrail spec: %w", err)
		}
	}

	resourceName := g.sanitizeResourceName(resource.Metadata.Name)
	
	// Create module block
	moduleBlock := body.AppendNewBlock("module", []string{resourceName})
	moduleBody := moduleBlock.Body()
	
	// Set module source
	moduleSource := fmt.Sprintf("%s//modules/bedrock-guardrail", g.config.ModuleRegistry)
	if g.config.ModuleVersion != "" {
		moduleSource += fmt.Sprintf("?ref=%s", g.config.ModuleVersion)
	}
	moduleBody.SetAttributeValue("source", cty.StringVal(moduleSource))
	
	// Set basic attributes
	moduleBody.SetAttributeValue("guardrail_name", cty.StringVal(resource.Metadata.Name))
	
	// Optional description
	if guardrail.Description != "" {
		moduleBody.SetAttributeValue("description", cty.StringVal(guardrail.Description))
	}
	
	// Content policy configuration
	if guardrail.ContentPolicyConfig != nil {
		contentPolicyValues := make(map[string]cty.Value)
		
		if len(guardrail.ContentPolicyConfig.FiltersConfig) > 0 {
			filtersList := make([]cty.Value, 0, len(guardrail.ContentPolicyConfig.FiltersConfig))
			
			for _, filter := range guardrail.ContentPolicyConfig.FiltersConfig {
				filterValues := make(map[string]cty.Value)
				filterValues["type"] = cty.StringVal(filter.Type)
				filterValues["input_strength"] = cty.StringVal(filter.InputStrength)
				filterValues["output_strength"] = cty.StringVal(filter.OutputStrength)
				
				filtersList = append(filtersList, cty.ObjectVal(filterValues))
			}
			
			contentPolicyValues["filters_config"] = cty.ListVal(filtersList)
		}
		
		moduleBody.SetAttributeValue("content_policy_config", cty.ObjectVal(contentPolicyValues))
	}
	
	// Sensitive information policy configuration
	if guardrail.SensitiveInformationPolicyConfig != nil {
		sensitiveInfoValues := make(map[string]cty.Value)
		
		if len(guardrail.SensitiveInformationPolicyConfig.PiiEntitiesConfig) > 0 {
			piiEntitiesList := make([]cty.Value, 0, len(guardrail.SensitiveInformationPolicyConfig.PiiEntitiesConfig))
			
			for _, piiEntity := range guardrail.SensitiveInformationPolicyConfig.PiiEntitiesConfig {
				piiValues := make(map[string]cty.Value)
				piiValues["type"] = cty.StringVal(piiEntity.Type)
				piiValues["action"] = cty.StringVal(piiEntity.Action)
				
				piiEntitiesList = append(piiEntitiesList, cty.ObjectVal(piiValues))
			}
			
			sensitiveInfoValues["pii_entities_config"] = cty.ListVal(piiEntitiesList)
		}
		
		moduleBody.SetAttributeValue("sensitive_information_policy_config", cty.ObjectVal(sensitiveInfoValues))
	}
	
	// Contextual grounding policy configuration
	if guardrail.ContextualGroundingPolicyConfig != nil {
		contextualGroundingValues := make(map[string]cty.Value)
		
		if len(guardrail.ContextualGroundingPolicyConfig.FiltersConfig) > 0 {
			filtersList := make([]cty.Value, 0, len(guardrail.ContextualGroundingPolicyConfig.FiltersConfig))
			
			for _, filter := range guardrail.ContextualGroundingPolicyConfig.FiltersConfig {
				filterValues := make(map[string]cty.Value)
				filterValues["type"] = cty.StringVal(filter.Type)
				filterValues["threshold"] = cty.NumberFloatVal(filter.Threshold)
				
				filtersList = append(filtersList, cty.ObjectVal(filterValues))
			}
			
			contextualGroundingValues["filters_config"] = cty.ListVal(filtersList)
		}
		
		moduleBody.SetAttributeValue("contextual_grounding_policy_config", cty.ObjectVal(contextualGroundingValues))
	}
	
	// Topic policy configuration
	if guardrail.TopicPolicyConfig != nil {
		topicPolicyValues := make(map[string]cty.Value)
		
		if len(guardrail.TopicPolicyConfig.TopicsConfig) > 0 {
			topicsList := make([]cty.Value, 0, len(guardrail.TopicPolicyConfig.TopicsConfig))
			
			for _, topic := range guardrail.TopicPolicyConfig.TopicsConfig {
				topicValues := make(map[string]cty.Value)
				topicValues["name"] = cty.StringVal(topic.Name)
				topicValues["definition"] = cty.StringVal(topic.Definition)
				topicValues["type"] = cty.StringVal(topic.Type)
				
				if len(topic.Examples) > 0 {
					examplesList := make([]cty.Value, 0, len(topic.Examples))
					for _, example := range topic.Examples {
						examplesList = append(examplesList, cty.StringVal(example))
					}
					topicValues["examples"] = cty.ListVal(examplesList)
				}
				
				topicsList = append(topicsList, cty.ObjectVal(topicValues))
			}
			
			topicPolicyValues["topics_config"] = cty.ListVal(topicsList)
		}
		
		moduleBody.SetAttributeValue("topic_policy_config", cty.ObjectVal(topicPolicyValues))
	}
	
	// Word policy configuration
	if guardrail.WordPolicyConfig != nil {
		wordPolicyValues := make(map[string]cty.Value)
		
		if len(guardrail.WordPolicyConfig.WordsConfig) > 0 {
			wordsList := make([]cty.Value, 0, len(guardrail.WordPolicyConfig.WordsConfig))
			
			for _, word := range guardrail.WordPolicyConfig.WordsConfig {
				wordValues := make(map[string]cty.Value)
				wordValues["text"] = cty.StringVal(word.Text)
				
				wordsList = append(wordsList, cty.ObjectVal(wordValues))
			}
			
			wordPolicyValues["words_config"] = cty.ListVal(wordsList)
		}
		
		if len(guardrail.WordPolicyConfig.ManagedWordListsConfig) > 0 {
			managedWordsList := make([]cty.Value, 0, len(guardrail.WordPolicyConfig.ManagedWordListsConfig))
			
			for _, managedWordList := range guardrail.WordPolicyConfig.ManagedWordListsConfig {
				managedWordValues := make(map[string]cty.Value)
				managedWordValues["type"] = cty.StringVal(managedWordList.Type)
				
				managedWordsList = append(managedWordsList, cty.ObjectVal(managedWordValues))
			}
			
			wordPolicyValues["managed_word_lists_config"] = cty.ListVal(managedWordsList)
		}
		
		if len(wordPolicyValues) > 0 {
			moduleBody.SetAttributeValue("word_policy_config", cty.ObjectVal(wordPolicyValues))
		}
	}
	
	// Tags
	if len(guardrail.Tags) > 0 {
		tagValues := make(map[string]cty.Value)
		for key, value := range guardrail.Tags {
			tagValues[key] = cty.StringVal(value)
		}
		moduleBody.SetAttributeValue("tags", cty.ObjectVal(tagValues))
	}
	
	body.AppendNewline()
	
	g.logger.WithField("guardrail", resource.Metadata.Name).Info("Generated guardrail module")
	return nil
}