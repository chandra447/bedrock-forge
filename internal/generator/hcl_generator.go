package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
	"github.com/sirupsen/logrus"

	"bedrock-forge/internal/models"
	"bedrock-forge/internal/registry"
)

// HCLGenerator handles the transformation of YAML resources to HCL Terraform modules
type HCLGenerator struct {
	logger   *logrus.Logger
	registry *registry.ResourceRegistry
	config   *GeneratorConfig
	context  *GenerationContext
}

// GeneratorConfig holds configuration for HCL generation
type GeneratorConfig struct {
	ModuleRegistry string
	ModuleVersion  string
	OutputDir      string
	ProjectName    string
	Environment    string
}

// NewHCLGenerator creates a new HCL generator instance
func NewHCLGenerator(logger *logrus.Logger, registry *registry.ResourceRegistry, config *GeneratorConfig) *HCLGenerator {
	return &HCLGenerator{
		logger:   logger,
		registry: registry,
		config:   config,
		context:  NewGenerationContext(),
	}
}

// SetGenerationContext sets the generation context with packaging results
func (g *HCLGenerator) SetGenerationContext(context *GenerationContext) {
	g.context = context
}

// Generate creates Terraform configuration from the resource registry
func (g *HCLGenerator) Generate() error {
	g.logger.Info("Starting HCL generation...")

	// Build dependency graph
	dependencyOrder, err := g.buildDependencyOrder()
	if err != nil {
		return fmt.Errorf("failed to build dependency order: %w", err)
	}

	// Generate main.tf file
	mainFile := hclwrite.NewEmptyFile()
	body := mainFile.Body()

	// Add terraform block
	g.addTerraformBlock(body)

	// Add provider block
	g.addProviderBlock(body)

	// Add variables block
	g.addVariablesBlock(body)

	// First pass: Generate auto-IAM roles for agents that need them
	g.generateAutoIAMRoles(body)

	// Generate module calls for each resource type in dependency order
	for _, resourceType := range dependencyOrder {
		resources := g.registry.GetResourcesByType(resourceType)
		for _, resource := range resources {
			if err := g.generateModuleCall(body, resource); err != nil {
				return fmt.Errorf("failed to generate module call for %s: %w", resource.Metadata.Name, err)
			}
		}
	}

	// Add outputs block
	g.addOutputsBlock(body)

	// Write the file
	outputPath := filepath.Join(g.config.OutputDir, "main.tf")
	if err := g.writeHCLFile(outputPath, mainFile); err != nil {
		return fmt.Errorf("failed to write main.tf: %w", err)
	}

	g.logger.WithField("output", outputPath).Info("Generated main.tf successfully")
	return nil
}

// buildDependencyOrder determines the order in which resources should be created
func (g *HCLGenerator) buildDependencyOrder() ([]models.ResourceKind, error) {
	// Define the dependency order based on Bedrock resource relationships
	// Guardrails and Prompts can be created first (no dependencies)
	// KnowledgeBases and Lambdas can be created next
	// ActionGroups depend on Lambdas
	// Agents depend on everything else
	
	order := []models.ResourceKind{
		models.IAMRoleKind,         // IAM roles must be created first
		models.CustomModuleKind,    // Custom modules can be created early
		models.GuardrailKind,
		models.PromptKind,
		models.LambdaKind,
		models.OpenSearchServerlessKind, // OpenSearch Serverless must be created before KnowledgeBase
		models.KnowledgeBaseKind,
		models.ActionGroupKind,
		models.AgentKind,
	}

	return order, nil
}

// generateModuleCall creates a module call for a specific resource
func (g *HCLGenerator) generateModuleCall(body *hclwrite.Body, resource models.BaseResource) error {
	switch resource.Kind {
	case models.AgentKind:
		return g.generateAgentModule(body, resource)
	case models.LambdaKind:
		return g.generateLambdaModule(body, resource)
	case models.ActionGroupKind:
		return g.generateActionGroupModule(body, resource)
	case models.KnowledgeBaseKind:
		return g.generateKnowledgeBaseModule(body, resource)
	case models.GuardrailKind:
		return g.generateGuardrailModule(body, resource)
	case models.PromptKind:
		return g.generatePromptModule(body, resource)
	case models.IAMRoleKind:
		return g.generateIAMRoleModule(body, resource)
	case models.CustomModuleKind:
		return g.generateCustomModuleModule(body, resource)
	case models.OpenSearchServerlessKind:
		return g.generateOpenSearchServerlessModule(body, resource)
	default:
		return fmt.Errorf("unsupported resource kind: %s", resource.Kind)
	}
}

// addTerraformBlock adds the terraform configuration block
func (g *HCLGenerator) addTerraformBlock(body *hclwrite.Body) {
	terraformBlock := body.AppendNewBlock("terraform", nil)
	terraformBody := terraformBlock.Body()
	
	// Add required providers
	reqProvidersBlock := terraformBody.AppendNewBlock("required_providers", nil)
	reqProvidersBody := reqProvidersBlock.Body()
	
	reqProvidersBody.SetAttributeValue("aws", cty.ObjectVal(map[string]cty.Value{
		"source":  cty.StringVal("hashicorp/aws"),
		"version": cty.StringVal("~> 5.0"),
	}))
	
	// Add required version
	terraformBody.SetAttributeValue("required_version", cty.StringVal(">= 1.0"))
	
	body.AppendNewline()
}

// addProviderBlock adds the AWS provider configuration
func (g *HCLGenerator) addProviderBlock(body *hclwrite.Body) {
	providerBlock := body.AppendNewBlock("provider", []string{"aws"})
	providerBody := providerBlock.Body()
	
	// Add default tags as a block
	defaultTagsBlock := providerBody.AppendNewBlock("default_tags", nil)
	defaultTagsBody := defaultTagsBlock.Body()
	
	defaultTagsBody.SetAttributeValue("tags", cty.ObjectVal(map[string]cty.Value{
		"Project":     cty.StringVal(g.config.ProjectName),
		"Environment": cty.StringVal(g.config.Environment),
		"ManagedBy":   cty.StringVal("bedrock-forge"),
	}))
	
	body.AppendNewline()
}

// addVariablesBlock adds common variables
func (g *HCLGenerator) addVariablesBlock(body *hclwrite.Body) {
	// Add project name variable
	projVarBlock := body.AppendNewBlock("variable", []string{"project_name"})
	projVarBody := projVarBlock.Body()
	projVarBody.SetAttributeValue("description", cty.StringVal("Name of the project"))
	projVarBody.SetAttributeRaw("type", hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte("string")},
	})
	projVarBody.SetAttributeValue("default", cty.StringVal(g.config.ProjectName))
	
	// Add environment variable
	envVarBlock := body.AppendNewBlock("variable", []string{"environment"})
	envVarBody := envVarBlock.Body()
	envVarBody.SetAttributeValue("description", cty.StringVal("Environment name"))
	envVarBody.SetAttributeRaw("type", hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte("string")},
	})
	envVarBody.SetAttributeValue("default", cty.StringVal(g.config.Environment))
	
	body.AppendNewline()
}

// addOutputsBlock adds outputs for created resources
func (g *HCLGenerator) addOutputsBlock(body *hclwrite.Body) {
	// Add outputs for each resource type
	agents := g.registry.GetResourcesByType(models.AgentKind)
	for _, agent := range agents {
		agentName := g.sanitizeResourceName(agent.Metadata.Name)
		
		// Agent ID output
		agentIdBlock := body.AppendNewBlock("output", []string{fmt.Sprintf("%s_agent_id", agentName)})
		agentIdBody := agentIdBlock.Body()
		agentIdBody.SetAttributeValue("description", cty.StringVal(fmt.Sprintf("ID of the %s agent", agent.Metadata.Name)))
		agentIdBody.SetAttributeTraversal("value", hcl.Traversal{
			hcl.TraverseRoot{Name: "module"},
			hcl.TraverseAttr{Name: agentName},
			hcl.TraverseAttr{Name: "agent_id"},
		})
		
		// Agent ARN output
		agentArnBlock := body.AppendNewBlock("output", []string{fmt.Sprintf("%s_agent_arn", agentName)})
		agentArnBody := agentArnBlock.Body()
		agentArnBody.SetAttributeValue("description", cty.StringVal(fmt.Sprintf("ARN of the %s agent", agent.Metadata.Name)))
		agentArnBody.SetAttributeTraversal("value", hcl.Traversal{
			hcl.TraverseRoot{Name: "module"},
			hcl.TraverseAttr{Name: agentName},
			hcl.TraverseAttr{Name: "agent_arn"},
		})
	}
	
	body.AppendNewline()
}

// sanitizeResourceName converts resource names to valid Terraform identifiers
func (g *HCLGenerator) sanitizeResourceName(name string) string {
	// Replace hyphens and spaces with underscores
	sanitized := strings.ReplaceAll(name, "-", "_")
	sanitized = strings.ReplaceAll(sanitized, " ", "_")
	
	// Convert to lowercase
	sanitized = strings.ToLower(sanitized)
	
	return sanitized
}

// writeHCLFile writes the HCL file to disk
func (g *HCLGenerator) writeHCLFile(path string, file *hclwrite.File) error {
	content := file.Bytes()
	
	// Create directory if it doesn't exist
	if err := g.ensureDir(filepath.Dir(path)); err != nil {
		return err
	}
	
	return g.writeFile(path, content)
}

// ensureDir creates a directory if it doesn't exist
func (g *HCLGenerator) ensureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// writeFile writes content to a file
func (g *HCLGenerator) writeFile(path string, content []byte) error {
	return os.WriteFile(path, content, 0644)
}

// generateAutoIAMRoles generates IAM roles for all agents automatically
func (g *HCLGenerator) generateAutoIAMRoles(body *hclwrite.Body) {
	agents := g.registry.GetResourcesByType(models.AgentKind)
	
	for _, agentResource := range agents {
		// Generate IAM role for every agent
		if err := g.generateAutoIAMRole(body, agentResource.Metadata.Name, nil); err != nil {
			g.logger.WithError(err).WithField("agent", agentResource.Metadata.Name).Warn("Failed to generate auto IAM role")
		}
	}
}

