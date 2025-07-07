package generator

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"

	"bedrock-forge/internal/models"
)

// generateCustomModuleModule creates a module call for a CustomModule resource
func (g *HCLGenerator) generateCustomModuleModule(body *hclwrite.Body, resource models.BaseResource) error {
	customModule, ok := resource.Spec.(models.CustomModuleSpec)
	if !ok {
		// Try to parse as map and convert to CustomModuleSpec
		specMap, mapOk := resource.Spec.(map[string]interface{})
		if !mapOk {
			return fmt.Errorf("invalid custom module spec format")
		}

		// Convert map to CustomModuleSpec
		specJSON, err := json.Marshal(specMap)
		if err != nil {
			return fmt.Errorf("failed to marshal custom module spec: %w", err)
		}

		if err := json.Unmarshal(specJSON, &customModule); err != nil {
			return fmt.Errorf("failed to unmarshal custom module spec: %w", err)
		}
	}

	resourceName := g.sanitizeResourceName(resource.Metadata.Name)

	g.logger.WithField("custom_module", resource.Metadata.Name).Debug("Generating custom module")

	// Create module block
	moduleBlock := body.AppendNewBlock("module", []string{resourceName})
	moduleBody := moduleBlock.Body()

	// Set module source
	moduleSource := customModule.Source
	if customModule.Version != "" {
		// Add version reference for git or registry modules
		if isGitSource(customModule.Source) {
			moduleSource += fmt.Sprintf("?ref=%s", customModule.Version)
		} else if isRegistrySource(customModule.Source) {
			// For registry modules, version is handled differently
			moduleBody.SetAttributeValue("version", cty.StringVal(customModule.Version))
		}
	}
	moduleBody.SetAttributeValue("source", cty.StringVal(moduleSource))

	// Set input variables
	if len(customModule.Variables) > 0 {
		for varName, varValue := range customModule.Variables {
			ctyValue, err := convertToCtyValue(varValue)
			if err != nil {
				g.logger.WithField("variable", varName).WithError(err).Warn("Failed to convert variable value")
				continue
			}
			moduleBody.SetAttributeValue(varName, ctyValue)
		}
	}

	// Add dependency references if specified
	if len(customModule.DependsOn) > 0 {
		dependsList := make([]cty.Value, 0, len(customModule.DependsOn))
		for _, dep := range customModule.DependsOn {
			// Check if dependency exists in registry and create proper reference
			if g.isValidDependency(dep) {
				depName := g.sanitizeResourceName(dep)
				dependsList = append(dependsList, cty.StringVal(fmt.Sprintf("module.%s", depName)))
			} else {
				g.logger.WithField("dependency", dep).Warn("Invalid dependency reference in custom module")
			}
		}
		if len(dependsList) > 0 {
			moduleBody.SetAttributeValue("depends_on", cty.ListVal(dependsList))
		}
	}

	body.AppendNewline()

	g.logger.WithField("custom_module", resource.Metadata.Name).Info("Generated custom module")
	return nil
}

// convertToCtyValue converts Go interface{} values to cty.Value
func convertToCtyValue(value interface{}) (cty.Value, error) {
	switch v := value.(type) {
	case string:
		return cty.StringVal(v), nil
	case int:
		return cty.NumberIntVal(int64(v)), nil
	case int64:
		return cty.NumberIntVal(v), nil
	case float64:
		return cty.NumberFloatVal(v), nil
	case bool:
		return cty.BoolVal(v), nil
	case []interface{}:
		var values []cty.Value
		for _, item := range v {
			ctyItem, err := convertToCtyValue(item)
			if err != nil {
				return cty.NilVal, err
			}
			values = append(values, ctyItem)
		}
		return cty.ListVal(values), nil
	case map[string]interface{}:
		values := make(map[string]cty.Value)
		for key, val := range v {
			ctyVal, err := convertToCtyValue(val)
			if err != nil {
				return cty.NilVal, err
			}
			values[key] = ctyVal
		}
		return cty.ObjectVal(values), nil
	default:
		// Try to convert to string as fallback
		return cty.StringVal(fmt.Sprintf("%v", v)), nil
	}
}

// isGitSource checks if the source is a git repository
func isGitSource(source string) bool {
	return len(source) > 4 && (source[:4] == "git:" ||
		source[:8] == "https://" && (source[8:] == "github.com" || source[8:] == "gitlab.com"))
}

// isRegistrySource checks if the source is a Terraform registry module
func isRegistrySource(source string) bool {
	// Registry modules typically have format: namespace/name/provider
	parts := len(source)
	slashCount := 0
	for i := 0; i < parts; i++ {
		if source[i] == '/' {
			slashCount++
		}
	}
	return slashCount == 2 && !isGitSource(source) && source[0] != '.' && source[0] != '/'
}

// isValidDependency checks if a dependency reference is valid
func (g *HCLGenerator) isValidDependency(dep string) bool {
	// Check if dependency exists in any of the supported resource types
	resourceTypes := []models.ResourceKind{
		models.AgentKind,
		models.LambdaKind,
		models.ActionGroupKind,
		models.KnowledgeBaseKind,
		models.GuardrailKind,
		models.PromptKind,
		models.IAMRoleKind,
		models.CustomModuleKind,
	}

	for _, resourceType := range resourceTypes {
		if g.registry.HasResource(resourceType, dep) {
			return true
		}
	}
	return false
}
