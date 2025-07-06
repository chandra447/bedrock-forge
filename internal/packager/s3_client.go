package packager

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// MockS3Client is a mock implementation for testing
type MockS3Client struct {
	logger    *logrus.Logger
	localDir  string
	uploads   map[string]string // key -> local file path
}

// RealS3Client would be the actual AWS S3 implementation
type RealS3Client struct {
	logger *logrus.Logger
	// AWS SDK client would go here
}

// NewMockS3Client creates a mock S3 client that stores files locally
func NewMockS3Client(logger *logrus.Logger, localDir string) *MockS3Client {
	return &MockS3Client{
		logger:   logger,
		localDir: localDir,
		uploads:  make(map[string]string),
	}
}

// UploadFile uploads a file to S3 (mock implementation saves to local directory)
func (c *MockS3Client) UploadFile(bucket, key string, filePath string) (string, error) {
	c.logger.WithFields(logrus.Fields{
		"bucket":   bucket,
		"key":      key,
		"file":     filePath,
	}).Debug("Mock S3 upload file")
	
	// Create destination directory
	destPath := filepath.Join(c.localDir, bucket, key)
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create destination directory: %w", err)
	}
	
	// Copy file
	if err := c.copyFile(filePath, destPath); err != nil {
		return "", fmt.Errorf("failed to copy file: %w", err)
	}
	
	// Store upload record
	s3URI := fmt.Sprintf("s3://%s/%s", bucket, key)
	c.uploads[key] = destPath
	
	c.logger.WithFields(logrus.Fields{
		"bucket": bucket,
		"key":    key,
		"uri":    s3URI,
	}).Info("Mock S3 file uploaded")
	
	return s3URI, nil
}

// UploadContent uploads content to S3 (mock implementation saves to local directory)
func (c *MockS3Client) UploadContent(bucket, key string, content []byte, contentType string) (string, error) {
	c.logger.WithFields(logrus.Fields{
		"bucket":       bucket,
		"key":          key,
		"content_type": contentType,
		"size":         len(content),
	}).Debug("Mock S3 upload content")
	
	// Create destination directory
	destPath := filepath.Join(c.localDir, bucket, key)
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create destination directory: %w", err)
	}
	
	// Write content to file
	if err := os.WriteFile(destPath, content, 0644); err != nil {
		return "", fmt.Errorf("failed to write content: %w", err)
	}
	
	// Store upload record
	s3URI := fmt.Sprintf("s3://%s/%s", bucket, key)
	c.uploads[key] = destPath
	
	c.logger.WithFields(logrus.Fields{
		"bucket": bucket,
		"key":    key,
		"uri":    s3URI,
	}).Info("Mock S3 content uploaded")
	
	return s3URI, nil
}

// GetUploads returns the map of uploaded files (for testing)
func (c *MockS3Client) GetUploads() map[string]string {
	return c.uploads
}

// copyFile copies a file from src to dst
func (c *MockS3Client) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()
	
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()
	
	_, err = io.Copy(destFile, sourceFile)
	return err
}

// NewRealS3Client would create a real AWS S3 client
func NewRealS3Client(logger *logrus.Logger) *RealS3Client {
	return &RealS3Client{
		logger: logger,
	}
}

// UploadFile uploads a file to real AWS S3
func (c *RealS3Client) UploadFile(bucket, key string, filePath string) (string, error) {
	// Real AWS S3 implementation would go here
	// For now, return an error indicating it's not implemented
	return "", fmt.Errorf("real S3 client not implemented yet")
}

// UploadContent uploads content to real AWS S3
func (c *RealS3Client) UploadContent(bucket, key string, content []byte, contentType string) (string, error) {
	// Real AWS S3 implementation would go here
	// For now, return an error indicating it's not implemented
	return "", fmt.Errorf("real S3 client not implemented yet")
}