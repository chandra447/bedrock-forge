package parser

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

type Scanner struct {
	logger *logrus.Logger
}

func NewScanner(logger *logrus.Logger) *Scanner {
	return &Scanner{
		logger: logger,
	}
}

type ScanResult struct {
	Files  []string
	Errors []error
}

func (s *Scanner) ScanDirectory(rootPath string, includePatterns []string, excludePatterns []string) (*ScanResult, error) {
	result := &ScanResult{
		Files:  make([]string, 0),
		Errors: make([]error, 0),
	}

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			s.logger.WithError(err).WithField("path", path).Warn("Error accessing path")
			result.Errors = append(result.Errors, err)
			return nil
		}

		if info.IsDir() {
			if s.shouldExcludeDirectory(path, excludePatterns) {
				s.logger.WithField("path", path).Debug("Skipping excluded directory")
				return filepath.SkipDir
			}
			return nil
		}

		if s.isYAMLFile(path) && !s.shouldExcludeFile(path, excludePatterns) {
			s.logger.WithField("path", path).Debug("Found YAML file")
			result.Files = append(result.Files, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	s.logger.WithField("count", len(result.Files)).Info("Completed directory scan")
	return result, nil
}

func (s *Scanner) isYAMLFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".yml" || ext == ".yaml"
}

func (s *Scanner) shouldExcludeDirectory(path string, excludePatterns []string) bool {
	for _, pattern := range excludePatterns {
		if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
			return true
		}
		if strings.Contains(pattern, "**") {
			cleanPattern := strings.ReplaceAll(pattern, "**", "*")
			if matched, _ := filepath.Match(cleanPattern, path); matched {
				return true
			}
		}
	}
	return false
}

func (s *Scanner) shouldExcludeFile(path string, excludePatterns []string) bool {
	for _, pattern := range excludePatterns {
		if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
			return true
		}
		if strings.Contains(pattern, "**") {
			cleanPattern := strings.ReplaceAll(pattern, "**", "*")
			if matched, _ := filepath.Match(cleanPattern, path); matched {
				return true
			}
		}
	}
	return false
}
