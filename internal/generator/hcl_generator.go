package generator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"

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
	SourceDir      string
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

	// Ensure output directory exists
	if err := os.MkdirAll(g.config.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory %s: %w", g.config.OutputDir, err)
	}

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
	// Build dependency graph based on actual references
	dependencyGraph := g.buildDependencyGraph()

	// Perform topological sort to determine order
	orderedKinds, err := g.topologicalSort(dependencyGraph)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve dependencies: %w", err)
	}

	return orderedKinds, nil
}

// buildDependencyGraph analyzes all resources and builds a dependency graph
func (g *HCLGenerator) buildDependencyGraph() map[models.ResourceKind][]models.ResourceKind {
	dependencies := make(map[models.ResourceKind][]models.ResourceKind)

	// Initialize all resource kinds
	allKinds := []models.ResourceKind{
		models.IAMRoleKind,
		models.CustomResourcesKind,
		models.GuardrailKind,
		models.PromptKind,
		models.LambdaKind,
		models.OpenSearchServerlessKind,
		models.KnowledgeBaseKind,
		models.ActionGroupKind,
		models.AgentKnowledgeBaseAssociationKind,
		models.AgentKind,
	}

	for _, kind := range allKinds {
		dependencies[kind] = []models.ResourceKind{}
	}

	// Analyze dependencies for each resource kind
	for _, kind := range allKinds {
		resources := g.registry.GetResourcesByType(kind)
		for _, resource := range resources {
			resourceDeps := g.extractResourceDependencies(resource)
			for _, dep := range resourceDeps {
				if !g.containsKind(dependencies[kind], dep) {
					dependencies[kind] = append(dependencies[kind], dep)
				}
			}
		}
	}

	return dependencies
}

// extractResourceDependencies analyzes a resource and extracts its dependencies
func (g *HCLGenerator) extractResourceDependencies(resource models.BaseResource) []models.ResourceKind {
	var dependencies []models.ResourceKind

	switch resource.Kind {
	case models.AgentKind:
		// Agent depends on guardrails, prompts, and lambdas
		if agent, ok := resource.Spec.(models.AgentSpec); ok {
			if agent.Guardrail != nil && !agent.Guardrail.Name.IsEmpty() {
				dependencies = append(dependencies, models.GuardrailKind)
			}

			for _, promptOverride := range agent.PromptOverrides {
				if !promptOverride.Prompt.IsEmpty() {
					dependencies = append(dependencies, models.PromptKind)
				}
			}

			for _, ag := range agent.ActionGroups {
				if ag.ActionGroupExecutor != nil && !ag.ActionGroupExecutor.Lambda.IsEmpty() {
					dependencies = append(dependencies, models.LambdaKind)
				}
			}
		}

	case models.ActionGroupKind:
		// ActionGroup depends on agent and lambda
		if actionGroup, ok := resource.Spec.(models.ActionGroupSpec); ok {
			if !actionGroup.AgentId.IsEmpty() {
				dependencies = append(dependencies, models.AgentKind)
			}

			if actionGroup.ActionGroupExecutor != nil && !actionGroup.ActionGroupExecutor.Lambda.IsEmpty() {
				dependencies = append(dependencies, models.LambdaKind)
			}
		}

	case models.KnowledgeBaseKind:
		// KnowledgeBase depends on OpenSearch Serverless and optionally Lambda
		if knowledgeBase, ok := resource.Spec.(models.KnowledgeBaseSpec); ok {
			if knowledgeBase.StorageConfiguration != nil && knowledgeBase.StorageConfiguration.OpenSearchServerless != nil {
				if knowledgeBase.StorageConfiguration.OpenSearchServerless.CollectionName != nil && !knowledgeBase.StorageConfiguration.OpenSearchServerless.CollectionName.IsEmpty() {
					dependencies = append(dependencies, models.OpenSearchServerlessKind)
				}
			}

			for _, dataSource := range knowledgeBase.DataSources {
				if dataSource.CustomTransformation != nil && dataSource.CustomTransformation.TransformationLambda != nil {
					if !dataSource.CustomTransformation.TransformationLambda.Lambda.IsEmpty() {
						dependencies = append(dependencies, models.LambdaKind)
					}
				}
			}
		}

	case models.AgentKnowledgeBaseAssociationKind:
		// Association depends on agent and knowledge base
		if association, ok := resource.Spec.(models.AgentKnowledgeBaseAssociationSpec); ok {
			if !association.AgentName.IsEmpty() {
				dependencies = append(dependencies, models.AgentKind)
			}

			if !association.KnowledgeBaseName.IsEmpty() {
				dependencies = append(dependencies, models.KnowledgeBaseKind)
			}
		}

	case models.CustomResourcesKind:
		// Custom resources depend on their dependencies
		if customResources, ok := resource.Spec.(models.CustomResourcesSpec); ok {
			for _, depRef := range customResources.DependsOn {
				if !depRef.IsEmpty() {
					// Determine the kind of the dependency
					if depKind := g.getResourceKindByName(depRef.String()); depKind != "" {
						dependencies = append(dependencies, depKind)
					}
				}
			}
		}

	}

	return dependencies
}

// getResourceKindByName finds the resource kind for a given resource name
func (g *HCLGenerator) getResourceKindByName(resourceName string) models.ResourceKind {
	allKinds := []models.ResourceKind{
		models.IAMRoleKind,
		models.CustomResourcesKind,
		models.GuardrailKind,
		models.PromptKind,
		models.LambdaKind,
		models.OpenSearchServerlessKind,
		models.KnowledgeBaseKind,
		models.ActionGroupKind,
		models.AgentKnowledgeBaseAssociationKind,
		models.AgentKind,
	}

	for _, kind := range allKinds {
		if g.registry.HasResource(kind, resourceName) {
			return kind
		}
	}

	return ""
}

// containsKind checks if a kind is already in the slice
func (g *HCLGenerator) containsKind(kinds []models.ResourceKind, kind models.ResourceKind) bool {
	for _, k := range kinds {
		if k == kind {
			return true
		}
	}
	return false
}

// topologicalSort performs a topological sort on the dependency graph
func (g *HCLGenerator) topologicalSort(graph map[models.ResourceKind][]models.ResourceKind) ([]models.ResourceKind, error) {
	// Kahn's algorithm for topological sorting
	inDegree := make(map[models.ResourceKind]int)
	queue := []models.ResourceKind{}
	result := []models.ResourceKind{}

	// Initialize in-degree count
	for kind := range graph {
		inDegree[kind] = 0
	}

	// Calculate in-degrees
	for _, dependencies := range graph {
		for _, dep := range dependencies {
			inDegree[dep]++
		}
	}

	// Find all nodes with in-degree 0
	for kind, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, kind)
		}
	}

	// Process nodes
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		// Reduce in-degree for all dependent nodes
		for _, dep := range graph[current] {
			inDegree[dep]--
			if inDegree[dep] == 0 {
				queue = append(queue, dep)
			}
		}
	}

	// Check for cycles
	if len(result) != len(graph) {
		return nil, fmt.Errorf("circular dependency detected")
	}

	return result, nil
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
	case models.CustomResourcesKind:
		return g.generateCustomResourcesModule(body, resource)
	case models.OpenSearchServerlessKind:
		return g.generateOpenSearchServerlessModule(body, resource)
	case models.AgentKnowledgeBaseAssociationKind:
		return g.generateAgentKnowledgeBaseAssociationModule(body, resource)
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


// resolveReferenceToOutput resolves a Reference to a specific module output
func (g *HCLGenerator) resolveReferenceToOutput(ref models.Reference, expectedKind models.ResourceKind, outputName string) (string, error) {
	if ref.IsEmpty() {
		return "", fmt.Errorf("empty reference")
	}

	resourceName := ref.String()

	// Check if the resource exists in the registry
	if !g.registry.HasResource(expectedKind, resourceName) {
		return "", fmt.Errorf("resource %s of kind %s not found in registry", resourceName, expectedKind)
	}

	// Return the module output reference
	sanitizedName := g.sanitizeResourceName(resourceName)
	return fmt.Sprintf("${module.%s.%s}", sanitizedName, outputName), nil
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

// generateAgentKnowledgeBaseAssociationModule creates a module call for an AgentKnowledgeBaseAssociation resource
func (g *HCLGenerator) generateAgentKnowledgeBaseAssociationModule(body *hclwrite.Body, resource models.BaseResource) error {
	association, ok := resource.Spec.(models.AgentKnowledgeBaseAssociationSpec)
	if !ok {
		// Try to parse as map and convert to AgentKnowledgeBaseAssociationSpec
		specMap, mapOk := resource.Spec.(map[string]interface{})
		if !mapOk {
			return fmt.Errorf("invalid agent knowledge base association spec format")
		}

		// Convert map to AgentKnowledgeBaseAssociationSpec
		specJSON, err := json.Marshal(specMap)
		if err != nil {
			return fmt.Errorf("failed to marshal association spec: %w", err)
		}

		if err := json.Unmarshal(specJSON, &association); err != nil {
			return fmt.Errorf("failed to unmarshal association spec: %w", err)
		}
	}

	resourceName := g.sanitizeResourceName(resource.Metadata.Name)

	// Create module block
	moduleBlock := body.AppendNewBlock("module", []string{resourceName})
	moduleBody := moduleBlock.Body()

	// Set module source
	moduleSource := fmt.Sprintf("%s//modules/bedrock-agent-knowledge-base-association", g.config.ModuleRegistry)
	if g.config.ModuleVersion != "" {
		moduleSource += fmt.Sprintf("?ref=%s", g.config.ModuleVersion)
	}
	moduleBody.SetAttributeValue("source", cty.StringVal(moduleSource))

	// Set basic attributes
	moduleBody.SetAttributeValue("association_name", cty.StringVal(resource.Metadata.Name))

	// Resolve agent reference
	if agentId, err := g.resolveReferenceToOutput(association.AgentName, models.AgentKind, "agent_id"); err == nil {
		moduleBody.SetAttributeValue("agent_id", cty.StringVal(agentId))
	} else {
		return fmt.Errorf("failed to resolve agent reference: %w", err)
	}

	// Resolve knowledge base reference
	if kbId, err := g.resolveReferenceToOutput(association.KnowledgeBaseName, models.KnowledgeBaseKind, "knowledge_base_id"); err == nil {
		moduleBody.SetAttributeValue("knowledge_base_id", cty.StringVal(kbId))
	} else {
		return fmt.Errorf("failed to resolve knowledge base reference: %w", err)
	}

	// Optional description
	if association.Description != "" {
		moduleBody.SetAttributeValue("description", cty.StringVal(association.Description))
	}

	// Optional state
	if association.State != "" {
		moduleBody.SetAttributeValue("state", cty.StringVal(association.State))
	}

	body.AppendNewline()

	g.logger.WithField("association", resource.Metadata.Name).Info("Generated agent knowledge base association module")
	return nil
}
