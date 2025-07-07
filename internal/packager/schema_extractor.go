package packager

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"

	"bedrock-forge/internal/models"
	"bedrock-forge/internal/registry"
)

// SchemaExtractor handles OpenAPI schema extraction and uploading
type SchemaExtractor struct {
	logger   *logrus.Logger
	registry *registry.ResourceRegistry
	s3Client S3Client
	config   *PackagerConfig
}

// SchemaPackage represents an OpenAPI schema package
type SchemaPackage struct {
	Name        string
	ActionGroup string
	Content     []byte
	S3Bucket    string
	S3Key       string
	S3URI       string
	Source      string // "manual", "fastapi", "chalice"
}

// NewSchemaExtractor creates a new schema extractor
func NewSchemaExtractor(logger *logrus.Logger, registry *registry.ResourceRegistry, s3Client S3Client, config *PackagerConfig) *SchemaExtractor {
	return &SchemaExtractor{
		logger:   logger,
		registry: registry,
		s3Client: s3Client,
		config:   config,
	}
}

// ExtractAllSchemas discovers and processes all OpenAPI schemas
func (e *SchemaExtractor) ExtractAllSchemas(baseDir string) (map[string]*SchemaPackage, error) {
	e.logger.Info("Starting OpenAPI schema extraction...")

	packages := make(map[string]*SchemaPackage)

	// Get all ActionGroup resources from registry
	actionGroups := e.registry.GetResourcesByType(models.ActionGroupKind)

	for _, actionGroup := range actionGroups {
		actionGroupSpec, ok := actionGroup.Spec.(models.ActionGroupSpec)
		if !ok {
			e.logger.WithField("action_group", actionGroup.Metadata.Name).Warn("Invalid ActionGroup spec, skipping")
			continue
		}

		// Check if action group has API schema configuration
		if actionGroupSpec.APISchema == nil {
			e.logger.WithField("action_group", actionGroup.Metadata.Name).Debug("ActionGroup has no API schema, skipping")
			continue
		}

		// Find action group directory
		actionGroupDir, err := e.findActionGroupDirectory(baseDir, actionGroup.Metadata.Name)
		if err != nil {
			e.logger.WithError(err).WithField("action_group", actionGroup.Metadata.Name).Error("Failed to find ActionGroup directory")
			continue
		}

		// Extract schema
		pkg, err := e.extractSchema(actionGroup.Metadata.Name, actionGroupDir)
		if err != nil {
			e.logger.WithError(err).WithField("action_group", actionGroup.Metadata.Name).Error("Failed to extract schema")
			continue
		}

		packages[actionGroup.Metadata.Name] = pkg
		e.logger.WithFields(logrus.Fields{
			"action_group": actionGroup.Metadata.Name,
			"source":       pkg.Source,
			"s3_uri":       pkg.S3URI,
		}).Info("Successfully extracted schema")
	}

	e.logger.WithField("count", len(packages)).Info("Schema extraction completed")
	return packages, nil
}

// findActionGroupDirectory locates the directory containing the ActionGroup
func (e *SchemaExtractor) findActionGroupDirectory(baseDir, actionGroupName string) (string, error) {
	var actionGroupDir string

	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Look for action-group.yml files
		if !info.IsDir() && (filepath.Base(path) == "action-group.yml" || filepath.Base(path) == "action-group.yaml") {
			// Check if this is for our target ActionGroup
			if e.isTargetActionGroup(path, actionGroupName) {
				actionGroupDir = filepath.Dir(path)
				return filepath.SkipDir // Found it, stop searching
			}
		}

		return nil
	})

	if err != nil {
		return "", fmt.Errorf("error walking directory: %w", err)
	}

	if actionGroupDir == "" {
		return "", fmt.Errorf("ActionGroup directory not found for %s", actionGroupName)
	}

	return actionGroupDir, nil
}

// isTargetActionGroup checks if an action-group.yml file corresponds to the target ActionGroup
func (e *SchemaExtractor) isTargetActionGroup(yamlPath, targetName string) bool {
	// This is a simplified check - in a real implementation,
	// we'd parse the YAML and check the metadata.name field
	dir := filepath.Dir(yamlPath)
	dirName := filepath.Base(dir)

	// Check if directory name matches ActionGroup name
	return strings.EqualFold(dirName, targetName) || strings.EqualFold(dirName, strings.ReplaceAll(targetName, "_", "-"))
}

// extractSchema extracts OpenAPI schema from manual files only
func (e *SchemaExtractor) extractSchema(actionGroupName, actionGroupDir string) (*SchemaPackage, error) {
	e.logger.WithFields(logrus.Fields{
		"action_group": actionGroupName,
		"dir":          actionGroupDir,
	}).Debug("Extracting OpenAPI schema")

	// Only support manual OpenAPI schema files
	if schema, err := e.extractManualSchema(actionGroupDir); err == nil {
		return e.packageSchema(actionGroupName, schema, "manual")
	}

	return nil, fmt.Errorf("no manual OpenAPI schema found for ActionGroup %s", actionGroupName)
}

// extractManualSchema reads manually created OpenAPI schema files
func (e *SchemaExtractor) extractManualSchema(dir string) ([]byte, error) {
	// Look for common OpenAPI schema file names
	schemaFiles := []string{
		"openapi.json", "openapi.yaml", "openapi.yml",
		"schema.json", "schema.yaml", "schema.yml",
		"api.json", "api.yaml", "api.yml",
	}

	for _, fileName := range schemaFiles {
		filePath := filepath.Join(dir, fileName)
		if _, err := os.Stat(filePath); err == nil {
			content, err := os.ReadFile(filePath)
			if err != nil {
				return nil, fmt.Errorf("failed to read schema file %s: %w", filePath, err)
			}

			// If it's YAML, we should convert to JSON for consistency
			// For now, we'll just return the content as-is
			return content, nil
		}
	}

	return nil, fmt.Errorf("no manual schema file found")
}

// packageSchema packages and uploads a schema to S3
func (e *SchemaExtractor) packageSchema(actionGroupName string, schema []byte, source string) (*SchemaPackage, error) {
	// Generate S3 key
	s3Key := fmt.Sprintf("%s/schemas/%s/openapi.json", e.config.S3KeyPrefix, actionGroupName)

	// Upload to S3
	s3URI, err := e.s3Client.UploadContent(e.config.S3Bucket, s3Key, schema, "application/json")
	if err != nil {
		return nil, fmt.Errorf("failed to upload schema to S3: %w", err)
	}

	return &SchemaPackage{
		Name:        fmt.Sprintf("%s-schema", actionGroupName),
		ActionGroup: actionGroupName,
		Content:     schema,
		S3Bucket:    e.config.S3Bucket,
		S3Key:       s3Key,
		S3URI:       s3URI,
		Source:      source,
	}, nil
}
