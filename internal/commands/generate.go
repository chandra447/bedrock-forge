package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"

	"bedrock-forge/internal/generator"
	"bedrock-forge/internal/models"
	"bedrock-forge/internal/packager"
	"bedrock-forge/internal/parser"
	"bedrock-forge/internal/registry"
)

type GenerateCommand struct {
	logger *logrus.Logger
}

func NewGenerateCommand(logger *logrus.Logger) *GenerateCommand {
	return &GenerateCommand{
		logger: logger,
	}
}

func (c *GenerateCommand) Execute(scanPath, outputDir string) error {
	c.logger.Info("Starting Terraform generation...")

	// Use current directory if scanPath is empty
	if scanPath == "" {
		var err error
		scanPath, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	// Use './terraform' as default output directory
	if outputDir == "" {
		outputDir = filepath.Join(scanPath, "terraform")
	}

	// Initialize registry and parser
	resourceRegistry := registry.NewResourceRegistry(c.logger)
	yamlParser := parser.NewYAMLParser(c.logger)

	// Scan and parse YAML files
	if err := c.scanAndParseFiles(scanPath, resourceRegistry, yamlParser); err != nil {
		return fmt.Errorf("failed to scan and parse files: %w", err)
	}

	// Validate dependencies
	if errors := resourceRegistry.ValidateDependencies(); len(errors) > 0 {
		c.logger.Error("Dependency validation failed:")
		for _, err := range errors {
			c.logger.WithError(err).Error("Dependency error")
		}
		return fmt.Errorf("found %d dependency validation errors", len(errors))
	}

	// Package Lambdas and extract schemas
	lambdaPackages, schemaPackages, err := c.packageArtifacts(scanPath, resourceRegistry)
	if err != nil {
		return fmt.Errorf("failed to package artifacts: %w", err)
	}

	// Generate Terraform configuration
	generatorConfig := &generator.GeneratorConfig{
		ModuleRegistry: "git::https://github.com/company/bedrock-terraform-modules",
		ModuleVersion:  "v1.0.0",
		OutputDir:      outputDir,
		ProjectName:    "bedrock-project",
		Environment:    "dev",
	}

	hclGenerator := generator.NewHCLGenerator(c.logger, resourceRegistry, generatorConfig)

	// Set generation context with packaging results
	generationContext := generator.NewGenerationContext()
	generationContext.LambdaPackages = lambdaPackages
	generationContext.SchemaPackages = schemaPackages
	hclGenerator.SetGenerationContext(generationContext)
	if err := hclGenerator.Generate(); err != nil {
		return fmt.Errorf("failed to generate HCL: %w", err)
	}

	// Print summary
	totalResources := resourceRegistry.GetTotalResourceCount()
	c.logger.WithFields(logrus.Fields{
		"total_resources": totalResources,
		"output_dir":      outputDir,
	}).Info("Terraform generation completed successfully")

	// Print resource breakdown
	for _, kind := range []string{"Agent", "Lambda", "ActionGroup", "KnowledgeBase", "Guardrail", "Prompt"} {
		count := resourceRegistry.GetResourceCount(models.ResourceKind(kind))
		if count > 0 {
			c.logger.WithFields(logrus.Fields{
				"kind":  kind,
				"count": count,
			}).Info("Generated modules")
		}
	}

	return nil
}

func (c *GenerateCommand) scanAndParseFiles(scanPath string, resourceRegistry *registry.ResourceRegistry, yamlParser *parser.YAMLParser) error {
	return filepath.Walk(scanPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if file is a YAML file
		if !isYAMLFile(path) {
			return nil
		}

		// Parse the file
		resources, err := yamlParser.ParseFile(path)
		if err != nil {
			c.logger.WithError(err).WithField("file", path).Warn("Failed to parse YAML file")
			return nil // Continue processing other files
		}

		// Add resources to registry
		for _, resource := range resources {
			if err := resourceRegistry.AddResource(resource); err != nil {
				c.logger.WithError(err).WithFields(logrus.Fields{
					"file": path,
					"kind": resource.Kind,
					"name": resource.Metadata.Name,
				}).Warn("Failed to add resource to registry")
			}
		}

		return nil
	})
}

func isYAMLFile(path string) bool {
	ext := filepath.Ext(path)
	return ext == ".yml" || ext == ".yaml"
}

func (c *GenerateCommand) packageArtifacts(scanPath string, resourceRegistry *registry.ResourceRegistry) (map[string]*packager.LambdaPackage, map[string]*packager.SchemaPackage, error) {
	c.logger.Info("Starting artifact packaging...")

	// Create S3 client (using mock for now)
	s3LocalDir := filepath.Join(scanPath, ".bedrock-forge", "s3-mock")
	s3Client := packager.NewMockS3Client(c.logger, s3LocalDir)

	// Package configuration
	packagerConfig := &packager.PackagerConfig{
		S3Bucket:    "bedrock-artifacts",
		S3KeyPrefix: "bedrock-forge",
		TempDir:     filepath.Join(scanPath, ".bedrock-forge", "temp"),
	}

	// Package Lambda functions
	lambdaPackager := packager.NewLambdaPackager(c.logger, resourceRegistry, s3Client, packagerConfig)
	lambdaPackages, err := lambdaPackager.PackageAllLambdas(scanPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to package Lambdas: %w", err)
	}

	// Extract OpenAPI schemas
	schemaExtractor := packager.NewSchemaExtractor(c.logger, resourceRegistry, s3Client, packagerConfig)
	schemaPackages, err := schemaExtractor.ExtractAllSchemas(scanPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to extract schemas: %w", err)
	}

	// Log summary
	c.logger.WithFields(logrus.Fields{
		"lambda_packages": len(lambdaPackages),
		"schema_packages": len(schemaPackages),
	}).Info("Artifact packaging completed")

	return lambdaPackages, schemaPackages, nil
}
