package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"bedrock-forge/internal/validation"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type ValidateCommand struct {
	logger       *logrus.Logger
	scanCommand  *ScanCommand
	validator    *validation.Validator
	configPath   string
	validationProfile string // "default", "enterprise", "custom"
}

func NewValidateCommand(logger *logrus.Logger) *ValidateCommand {
	return &ValidateCommand{
		logger:            logger,
		scanCommand:       NewScanCommand(logger),
		validationProfile: "default",
	}
}

// SetValidationProfile sets the validation profile to use
func (v *ValidateCommand) SetValidationProfile(profile string) {
	v.validationProfile = profile
}

// SetConfigPath sets the path to a custom validation configuration file
func (v *ValidateCommand) SetConfigPath(configPath string) {
	v.configPath = configPath
}

func (v *ValidateCommand) Execute(rootPath string) error {
	if rootPath == "" {
		var err error
		rootPath, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}
	}

	v.logger.WithField("path", rootPath).Info("Starting comprehensive resource validation")

	// Initialize validator with appropriate configuration
	err := v.initializeValidator(rootPath)
	if err != nil {
		return fmt.Errorf("failed to initialize validator: %w", err)
	}

	// Scan resources
	err = v.scanCommand.Execute(rootPath)
	if err != nil {
		return fmt.Errorf("failed to scan resources: %w", err)
	}

	registry := v.scanCommand.GetRegistry()
	
	fmt.Printf("\n=== Bedrock Forge Enterprise Resource Validation ===\n")
	fmt.Printf("Profile: %s\n", v.validationProfile)
	if v.configPath != "" {
		fmt.Printf("Config: %s\n", v.configPath)
	}
	fmt.Printf("\n")

	totalResources := registry.GetTotalResourceCount()
	if totalResources == 0 {
		fmt.Printf("No resources found to validate.\n")
		return nil
	}

	fmt.Printf("Validating %d resources...\n\n", totalResources)

	// Create validation context
	context := &validation.ValidationContext{
		Team:        v.extractTeamFromPath(rootPath),
		Environment: v.extractEnvironmentFromPath(rootPath),
		Project:     v.extractProjectFromPath(rootPath),
	}

	// Run comprehensive validation
	result := v.validator.ValidateRegistry(registry, context)

	// Print results
	result.PrintSummary()

	if !result.Success {
		return fmt.Errorf("validation failed with %d errors", len(result.Errors))
	}

	return nil
}

// initializeValidator creates a validator with the appropriate configuration
func (v *ValidateCommand) initializeValidator(rootPath string) error {
	var config *validation.ValidationConfig
	var err error

	if v.configPath != "" {
		// Load custom configuration
		config, err = v.loadCustomConfig(v.configPath)
		if err != nil {
			return fmt.Errorf("failed to load custom validation config: %w", err)
		}
		v.logger.WithField("config", v.configPath).Info("Using custom validation configuration")
	} else {
		// Try to find local validation configuration
		localConfigPath := filepath.Join(rootPath, "validation.yml")
		if _, err := os.Stat(localConfigPath); err == nil {
			config, err = v.loadCustomConfig(localConfigPath)
			if err != nil {
				v.logger.WithError(err).Warn("Failed to load local validation config, using default")
				config = v.getBuiltinConfig()
			} else {
				v.logger.WithField("config", localConfigPath).Info("Using local validation configuration")
			}
		} else {
			// Use built-in configuration based on profile
			config = v.getBuiltinConfig()
			v.logger.WithField("profile", v.validationProfile).Info("Using built-in validation configuration")
		}
	}

	// Create validator
	v.validator, err = validation.NewValidator(v.logger, config)
	if err != nil {
		return fmt.Errorf("failed to create validator: %w", err)
	}

	return nil
}

// loadCustomConfig loads a custom validation configuration from file
func (v *ValidateCommand) loadCustomConfig(configPath string) (*validation.ValidationConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config validation.ValidationConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// getBuiltinConfig returns the appropriate built-in configuration
func (v *ValidateCommand) getBuiltinConfig() *validation.ValidationConfig {
	switch v.validationProfile {
	case "enterprise":
		return validation.EnterpriseValidationConfig()
	case "default":
		fallthrough
	default:
		return validation.DefaultValidationConfig()
	}
}

// extractTeamFromPath attempts to extract team name from the directory path
func (v *ValidateCommand) extractTeamFromPath(rootPath string) string {
	// Look for team indicators in the path
	pathComponents := strings.Split(rootPath, string(os.PathSeparator))
	for _, component := range pathComponents {
		if strings.HasPrefix(component, "team-") {
			return strings.TrimPrefix(component, "team-")
		}
		// Common team directory patterns
		teamPatterns := []string{"engineering", "data", "security", "operations", "product", "finance"}
		for _, pattern := range teamPatterns {
			if strings.Contains(strings.ToLower(component), pattern) {
				return pattern
			}
		}
	}
	return ""
}

// extractEnvironmentFromPath attempts to extract environment from the directory path
func (v *ValidateCommand) extractEnvironmentFromPath(rootPath string) string {
	pathComponents := strings.Split(rootPath, string(os.PathSeparator))
	for _, component := range pathComponents {
		lowerComponent := strings.ToLower(component)
		if lowerComponent == "dev" || lowerComponent == "development" {
			return "dev"
		}
		if lowerComponent == "staging" || lowerComponent == "stage" {
			return "staging"
		}
		if lowerComponent == "prod" || lowerComponent == "production" {
			return "prod"
		}
	}
	return ""
}

// extractProjectFromPath attempts to extract project name from the directory path
func (v *ValidateCommand) extractProjectFromPath(rootPath string) string {
	// Use the base directory name as project name
	return filepath.Base(rootPath)
}