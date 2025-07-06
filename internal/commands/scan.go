package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"

	"bedrock-forge/internal/models"
	"bedrock-forge/internal/parser"
	"bedrock-forge/internal/registry"
)

type ScanCommand struct {
	logger   *logrus.Logger
	scanner  *parser.Scanner
	yamlParser *parser.YAMLParser
	registry *registry.ResourceRegistry
}

func NewScanCommand(logger *logrus.Logger) *ScanCommand {
	return &ScanCommand{
		logger:     logger,
		scanner:    parser.NewScanner(logger),
		yamlParser: parser.NewYAMLParser(logger),
		registry:   registry.NewResourceRegistry(logger),
	}
}

func (s *ScanCommand) Execute(rootPath string) error {
	if rootPath == "" {
		var err error
		rootPath, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}
	}

	s.logger.WithField("path", rootPath).Info("Starting resource scan")

	excludePatterns := []string{
		"**/node_modules/**",
		"**/.git/**",
		"**/.terraform/**",
		"**/vendor/**",
		"**/.vscode/**",
		"**/.idea/**",
	}

	scanResult, err := s.scanner.ScanDirectory(rootPath, nil, excludePatterns)
	if err != nil {
		return fmt.Errorf("failed to scan directory: %w", err)
	}

	s.logger.WithField("files", len(scanResult.Files)).Info("Found YAML files")

	for _, filePath := range scanResult.Files {
		err := s.processFile(filePath)
		if err != nil {
			s.logger.WithError(err).WithField("file", filePath).Warn("Failed to process file")
		}
	}

	s.printScanResults()

	return nil
}

func (s *ScanCommand) processFile(filePath string) error {
	resources, err := s.yamlParser.ParseFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to parse file %s: %w", filePath, err)
	}

	for _, resource := range resources {
		err := s.registry.AddResource(resource)
		if err != nil {
			s.logger.WithError(err).WithFields(logrus.Fields{
				"file": filePath,
				"kind": resource.Kind,
				"name": resource.Metadata.Name,
			}).Warn("Failed to add resource to registry")
		}
	}

	return nil
}

func (s *ScanCommand) printScanResults() {
	fmt.Printf("\n=== Bedrock Forge Resource Scan Results ===\n\n")

	allResources := s.registry.GetAllResources()
	totalCount := s.registry.GetTotalResourceCount()

	if totalCount == 0 {
		fmt.Printf("No Bedrock resources found.\n")
		return
	}

	fmt.Printf("Total Resources Found: %d\n\n", totalCount)

	resourceKinds := []models.ResourceKind{
		models.AgentKind,
		models.LambdaKind,
		models.ActionGroupKind,
		models.KnowledgeBaseKind,
		models.GuardrailKind,
		models.PromptKind,
	}

	for _, kind := range resourceKinds {
		resources := allResources[kind]
		if len(resources) == 0 {
			continue
		}

		fmt.Printf("📦 %s (%d)\n", kind, len(resources))
		fmt.Printf("└─ Resources:\n")

		for name, resource := range resources {
			relPath := s.getRelativePath(resource.FilePath)
			fmt.Printf("   ├─ %s (%s)\n", name, relPath)
			
			if resource.Metadata.Description != "" {
				fmt.Printf("   │  └─ %s\n", resource.Metadata.Description)
			}
		}
		fmt.Printf("\n")
	}
}

func (s *ScanCommand) getRelativePath(filePath string) string {
	cwd, err := os.Getwd()
	if err != nil {
		return filePath
	}

	relPath, err := filepath.Rel(cwd, filePath)
	if err != nil {
		return filePath
	}

	return relPath
}

func (s *ScanCommand) GetRegistry() *registry.ResourceRegistry {
	return s.registry
}