package generator

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"

	"bedrock-forge/internal/models"
)

// generateCustomResourcesModule copies user's .tf files and handles CustomResources
func (g *HCLGenerator) generateCustomResourcesModule(body *hclwrite.Body, resource models.BaseResource) error {
	customResources, ok := resource.Spec.(models.CustomResourcesSpec)
	if !ok {
		// Try to parse as map and convert to CustomResourcesSpec
		specMap, mapOk := resource.Spec.(map[string]interface{})
		if !mapOk {
			return fmt.Errorf("invalid custom resources spec format")
		}

		// Convert map to CustomResourcesSpec
		specJSON, err := json.Marshal(specMap)
		if err != nil {
			return fmt.Errorf("failed to marshal custom resources spec: %w", err)
		}

		if err := json.Unmarshal(specJSON, &customResources); err != nil {
			return fmt.Errorf("failed to unmarshal custom resources spec: %w", err)
		}
	}

	resourceName := g.sanitizeResourceName(resource.Metadata.Name)
	g.logger.WithField("custom_resources", resource.Metadata.Name).Debug("Processing custom resources")

	// Copy user's .tf files to output directory
	if err := g.copyUserTerraformFiles(customResources); err != nil {
		return fmt.Errorf("failed to copy user terraform files: %w", err)
	}

	// Generate variables.tf file for user's custom resources if variables are provided
	if len(customResources.Variables) > 0 {
		if err := g.generateCustomResourcesVariables(customResources, resourceName); err != nil {
			return fmt.Errorf("failed to generate variables for custom resources: %w", err)
		}
	}

	g.logger.WithField("custom_resources", resource.Metadata.Name).Info("Generated custom resources")
	return nil
}

// generateCustomModuleModule creates a module call for a CustomModule resource (Deprecated)
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
		for _, depRef := range customModule.DependsOn {
			if !depRef.IsEmpty() {
				// Check if dependency exists in registry and create proper reference
				if g.isValidDependency(depRef.String()) {
					depName := g.sanitizeResourceName(depRef.String())
					dependsList = append(dependsList, cty.StringVal(fmt.Sprintf("module.%s", depName)))
				} else {
					g.logger.WithField("dependency", depRef.String()).Warn("Invalid dependency reference in custom module")
				}
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

// copyUserTerraformFiles copies user's .tf files to the output directory
func (g *HCLGenerator) copyUserTerraformFiles(spec models.CustomResourcesSpec) error {
	if spec.Path != "" {
		// Handle path-based approach
		return g.copyTerraformPath(spec.Path)
	}

	if len(spec.Files) > 0 {
		// Handle files list approach
		return g.copyTerraformFiles(spec.Files)
	}

	return fmt.Errorf("either 'path' or 'files' must be specified for CustomResources")
}

// copyTerraformPath copies all .tf files from a directory or a single .tf file
func (g *HCLGenerator) copyTerraformPath(path string) error {
	// Check if path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", path)
	}

	// Check if it's a file or directory
	fileInfo, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat path %s: %w", path, err)
	}

	if fileInfo.IsDir() {
		// Copy all .tf files from directory
		return g.copyTerraformFromDirectory(path)
	} else {
		// Copy single file
		if !strings.HasSuffix(path, ".tf") {
			return fmt.Errorf("file must have .tf extension: %s", path)
		}
		return g.copyTerraformFile(path)
	}
}

// copyTerraformFiles copies specific .tf files
func (g *HCLGenerator) copyTerraformFiles(files []string) error {
	for _, file := range files {
		if !strings.HasSuffix(file, ".tf") {
			return fmt.Errorf("file must have .tf extension: %s", file)
		}
		if err := g.copyTerraformFile(file); err != nil {
			return fmt.Errorf("failed to copy file %s: %w", file, err)
		}
	}
	return nil
}

// copyTerraformFromDirectory copies all .tf files from a directory
func (g *HCLGenerator) copyTerraformFromDirectory(dirPath string) error {
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-.tf files
		if info.IsDir() || !strings.HasSuffix(path, ".tf") {
			return nil
		}

		return g.copyTerraformFile(path)
	})
}

// copyTerraformFile copies a single .tf file to the output directory
func (g *HCLGenerator) copyTerraformFile(srcPath string) error {
	// Open source file
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", srcPath, err)
	}
	defer srcFile.Close()

	// Create destination file in output directory
	fileName := filepath.Base(srcPath)
	destPath := filepath.Join(g.config.OutputDir, fileName)

	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", destPath, err)
	}
	defer destFile.Close()

	// Copy file contents
	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy file contents from %s to %s: %w", srcPath, destPath, err)
	}

	g.logger.WithField("file", fileName).Debug("Copied user terraform file")
	return nil
}

// generateCustomResourcesVariables generates a variables.tf file for custom resources
func (g *HCLGenerator) generateCustomResourcesVariables(spec models.CustomResourcesSpec, resourceName string) error {
	variablesPath := filepath.Join(g.config.OutputDir, fmt.Sprintf("variables_%s.tf", resourceName))

	// Create new HCL file
	hclFile := hclwrite.NewEmptyFile()
	body := hclFile.Body()

	// Add comment
	body.AppendUnstructuredTokens(hclwrite.Tokens{
		{Type: hclsyntax.TokenComment, Bytes: []byte(fmt.Sprintf("# Variables for custom resources: %s", resourceName))},
		{Type: hclsyntax.TokenNewline, Bytes: []byte("\n")},
	})

	// Generate variable blocks
	for varName, varValue := range spec.Variables {
		varBlock := body.AppendNewBlock("variable", []string{varName})
		varBody := varBlock.Body()

		// Determine variable type and set default
		switch v := varValue.(type) {
		case string:
			varBody.SetAttributeValue("type", cty.StringVal("string"))
			varBody.SetAttributeValue("default", cty.StringVal(v))
		case bool:
			varBody.SetAttributeValue("type", cty.StringVal("bool"))
			varBody.SetAttributeValue("default", cty.BoolVal(v))
		case int, int64, float64:
			varBody.SetAttributeValue("type", cty.StringVal("number"))
			if num, ok := v.(int); ok {
				varBody.SetAttributeValue("default", cty.NumberIntVal(int64(num)))
			} else if num, ok := v.(int64); ok {
				varBody.SetAttributeValue("default", cty.NumberIntVal(num))
			} else if num, ok := v.(float64); ok {
				varBody.SetAttributeValue("default", cty.NumberFloatVal(num))
			}
		case []interface{}:
			varBody.SetAttributeValue("type", cty.StringVal("list"))
			ctyValue, _ := convertToCtyValue(v)
			varBody.SetAttributeValue("default", ctyValue)
		case map[string]interface{}:
			varBody.SetAttributeValue("type", cty.StringVal("map"))
			ctyValue, _ := convertToCtyValue(v)
			varBody.SetAttributeValue("default", ctyValue)
		default:
			// Fallback to string
			varBody.SetAttributeValue("type", cty.StringVal("string"))
			varBody.SetAttributeValue("default", cty.StringVal(fmt.Sprintf("%v", v)))
		}

		body.AppendNewline()
	}

	// Write to file
	file, err := os.Create(variablesPath)
	if err != nil {
		return fmt.Errorf("failed to create variables file %s: %w", variablesPath, err)
	}
	defer file.Close()

	_, err = file.Write(hclFile.Bytes())
	if err != nil {
		return fmt.Errorf("failed to write variables file %s: %w", variablesPath, err)
	}

	g.logger.WithField("file", fmt.Sprintf("variables_%s.tf", resourceName)).Debug("Generated variables file for custom resources")
	return nil
}
