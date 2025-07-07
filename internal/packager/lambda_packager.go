package packager

import (
	"archive/zip"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"bedrock-forge/internal/models"
	"bedrock-forge/internal/registry"
)

// LambdaPackager handles packaging Lambda functions and uploading to S3
type LambdaPackager struct {
	logger   *logrus.Logger
	registry *registry.ResourceRegistry
	s3Client S3Client
	config   *PackagerConfig
}

// PackagerConfig holds configuration for the packager
type PackagerConfig struct {
	S3Bucket        string
	S3KeyPrefix     string
	TempDir         string
	ExcludePatterns []string
}

// S3Client interface for uploading artifacts
type S3Client interface {
	UploadFile(bucket, key string, filePath string) (string, error)
	UploadContent(bucket, key string, content []byte, contentType string) (string, error)
}

// LambdaPackage represents a packaged Lambda function
type LambdaPackage struct {
	Name         string
	FilePath     string
	S3Bucket     string
	S3Key        string
	S3URI        string
	Hash         string
	Size         int64
	Dependencies []string
}

// NewLambdaPackager creates a new Lambda packager
func NewLambdaPackager(logger *logrus.Logger, registry *registry.ResourceRegistry, s3Client S3Client, config *PackagerConfig) *LambdaPackager {
	if config.ExcludePatterns == nil {
		config.ExcludePatterns = []string{
			"*.yml", "*.yaml", "*.md", "*.txt",
			".git", ".gitignore", ".DS_Store",
			"__pycache__", "*.pyc", "*.pyo",
			".pytest_cache", ".coverage",
			"node_modules", ".npm",
			".env", ".env.*",
		}
	}

	if config.TempDir == "" {
		config.TempDir = "/tmp/bedrock-forge"
	}

	return &LambdaPackager{
		logger:   logger,
		registry: registry,
		s3Client: s3Client,
		config:   config,
	}
}

// PackageAllLambdas discovers and packages all Lambda functions
func (p *LambdaPackager) PackageAllLambdas(baseDir string) (map[string]*LambdaPackage, error) {
	p.logger.Info("Starting Lambda packaging process...")

	packages := make(map[string]*LambdaPackage)

	// Get all Lambda resources from registry
	lambdas := p.registry.GetResourcesByType(models.LambdaKind)

	for _, lambda := range lambdas {
		lambdaSpec, ok := lambda.Spec.(models.LambdaSpec)
		if !ok {
			p.logger.WithField("lambda", lambda.Metadata.Name).Warn("Invalid Lambda spec, skipping")
			continue
		}

		// Only package directory-based Lambdas
		if lambdaSpec.Code.Source != "directory" {
			p.logger.WithField("lambda", lambda.Metadata.Name).Debug("Lambda uses non-directory source, skipping packaging")
			continue
		}

		// Find Lambda directory
		lambdaDir, err := p.findLambdaDirectory(baseDir, lambda.Metadata.Name)
		if err != nil {
			p.logger.WithError(err).WithField("lambda", lambda.Metadata.Name).Error("Failed to find Lambda directory")
			continue
		}

		// Package the Lambda
		pkg, err := p.packageLambda(lambda.Metadata.Name, lambdaDir)
		if err != nil {
			p.logger.WithError(err).WithField("lambda", lambda.Metadata.Name).Error("Failed to package Lambda")
			continue
		}

		packages[lambda.Metadata.Name] = pkg
		p.logger.WithFields(logrus.Fields{
			"lambda": lambda.Metadata.Name,
			"size":   pkg.Size,
			"s3_uri": pkg.S3URI,
		}).Info("Successfully packaged Lambda")
	}

	p.logger.WithField("count", len(packages)).Info("Lambda packaging completed")
	return packages, nil
}

// findLambdaDirectory locates the directory containing the Lambda code
func (p *LambdaPackager) findLambdaDirectory(baseDir, lambdaName string) (string, error) {
	var lambdaDir string

	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Look for lambda.yml files
		if !info.IsDir() && (filepath.Base(path) == "lambda.yml" || filepath.Base(path) == "lambda.yaml") {
			// Check if this lambda.yml is for our target Lambda
			if p.isTargetLambda(path, lambdaName) {
				lambdaDir = filepath.Dir(path)
				return filepath.SkipDir // Found it, stop searching
			}
		}

		return nil
	})

	if err != nil {
		return "", fmt.Errorf("error walking directory: %w", err)
	}

	if lambdaDir == "" {
		return "", fmt.Errorf("lambda directory not found for %s", lambdaName)
	}

	return lambdaDir, nil
}

// isTargetLambda checks if a lambda.yml file corresponds to the target Lambda
func (p *LambdaPackager) isTargetLambda(yamlPath, targetName string) bool {
	// This is a simplified check - in a real implementation,
	// we'd parse the YAML and check the metadata.name field
	dir := filepath.Dir(yamlPath)
	dirName := filepath.Base(dir)

	// Check if directory name matches Lambda name
	return strings.EqualFold(dirName, targetName) || strings.EqualFold(dirName, strings.ReplaceAll(targetName, "_", "-"))
}

// packageLambda creates a ZIP package of the Lambda function
func (p *LambdaPackager) packageLambda(lambdaName, lambdaDir string) (*LambdaPackage, error) {
	p.logger.WithFields(logrus.Fields{
		"lambda": lambdaName,
		"dir":    lambdaDir,
	}).Debug("Packaging Lambda function")

	// Create temp directory for packaging
	tempDir := filepath.Join(p.config.TempDir, lambdaName)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Create ZIP file
	zipPath := filepath.Join(tempDir, fmt.Sprintf("%s.zip", lambdaName))
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create ZIP file: %w", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Add files to ZIP
	err = p.addDirectoryToZip(zipWriter, lambdaDir, "")
	if err != nil {
		return nil, fmt.Errorf("failed to add files to ZIP: %w", err)
	}

	// Close ZIP writer to flush contents
	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close ZIP writer: %w", err)
	}

	// Get file info
	zipInfo, err := zipFile.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get ZIP file info: %w", err)
	}

	// Calculate hash
	hash, err := p.calculateFileHash(zipPath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate file hash: %w", err)
	}

	// Generate S3 key
	s3Key := p.generateS3Key(lambdaName, hash)

	// Upload to S3
	s3URI, err := p.s3Client.UploadFile(p.config.S3Bucket, s3Key, zipPath)
	if err != nil {
		return nil, fmt.Errorf("failed to upload to S3: %w", err)
	}

	return &LambdaPackage{
		Name:     lambdaName,
		FilePath: zipPath,
		S3Bucket: p.config.S3Bucket,
		S3Key:    s3Key,
		S3URI:    s3URI,
		Hash:     hash,
		Size:     zipInfo.Size(),
	}, nil
}

// addDirectoryToZip recursively adds directory contents to ZIP
func (p *LambdaPackager) addDirectoryToZip(zipWriter *zip.Writer, sourceDir, basePath string) error {
	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		// Skip excluded files
		if p.shouldExcludeFile(relPath, info) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip directories (they're created automatically)
		if info.IsDir() {
			return nil
		}

		// Create ZIP entry
		zipPath := filepath.Join(basePath, relPath)
		zipPath = filepath.ToSlash(zipPath) // Ensure forward slashes in ZIP

		zipEntry, err := zipWriter.Create(zipPath)
		if err != nil {
			return err
		}

		// Copy file contents
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(zipEntry, file)
		return err
	})
}

// shouldExcludeFile checks if a file should be excluded from packaging
func (p *LambdaPackager) shouldExcludeFile(relPath string, info os.FileInfo) bool {
	fileName := info.Name()

	for _, pattern := range p.config.ExcludePatterns {
		// Simple pattern matching (could be enhanced with glob patterns)
		if strings.HasSuffix(pattern, "*") {
			prefix := strings.TrimSuffix(pattern, "*")
			if strings.HasPrefix(fileName, prefix) {
				return true
			}
		} else if pattern == fileName || pattern == relPath {
			return true
		}
	}

	return false
}

// calculateFileHash calculates SHA256 hash of a file
func (p *LambdaPackager) calculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

// generateS3Key creates a unique S3 key for the Lambda package
func (p *LambdaPackager) generateS3Key(lambdaName, hash string) string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%s/lambdas/%s/%d-%s.zip",
		p.config.S3KeyPrefix, lambdaName, timestamp, hash[:8])
}
