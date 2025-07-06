package validation

import (
	"fmt"
	"path/filepath"

	"bedrock-forge/internal/parser"
	"bedrock-forge/internal/registry"
	"github.com/sirupsen/logrus"
)

// ValidationConfig holds all validation configuration
type ValidationConfig struct {
	NamingConventions *NamingConventionConfig `yaml:"namingConventions,omitempty"`
	TaggingPolicies   *TaggingPolicyConfig    `yaml:"taggingPolicies,omitempty"`
	SecurityPolicies  *SecurityPolicyConfig   `yaml:"securityPolicies,omitempty"`
	EnabledValidators []string                `yaml:"enabledValidators,omitempty"`
}

// Validator coordinates all validation activities
type Validator struct {
	logger            *logrus.Logger
	config            *ValidationConfig
	namingValidator   *NamingValidator
	taggingValidator  *TaggingValidator
	securityValidator *SecurityValidator
}

// NewValidator creates a new validator with the given configuration
func NewValidator(logger *logrus.Logger, config *ValidationConfig) (*Validator, error) {
	if config == nil {
		config = DefaultValidationConfig()
	}

	validator := &Validator{
		logger: logger,
		config: config,
	}

	// Initialize naming validator
	if config.NamingConventions != nil {
		namingValidator, err := NewNamingValidator(config.NamingConventions)
		if err != nil {
			return nil, fmt.Errorf("failed to create naming validator: %w", err)
		}
		validator.namingValidator = namingValidator
	}

	// Initialize tagging validator
	if config.TaggingPolicies != nil {
		taggingValidator, err := NewTaggingValidator(config.TaggingPolicies)
		if err != nil {
			return nil, fmt.Errorf("failed to create tagging validator: %w", err)
		}
		validator.taggingValidator = taggingValidator
	}

	// Initialize security validator
	if config.SecurityPolicies != nil {
		securityValidator, err := NewSecurityValidator(config.SecurityPolicies)
		if err != nil {
			return nil, fmt.Errorf("failed to create security validator: %w", err)
		}
		validator.securityValidator = securityValidator
	}

	return validator, nil
}

// ValidateRegistry validates all resources in a registry
func (v *Validator) ValidateRegistry(reg *registry.ResourceRegistry, context *ValidationContext) *ValidationResult {
	result := &ValidationResult{
		TotalResources: reg.GetTotalResourceCount(),
		Errors:         []ValidationError{},
		Warnings:       []ValidationError{},
	}

	allResources := reg.GetAllResources()
	for _, resources := range allResources {
		for _, resource := range resources {
			resourceErrors := v.ValidateResource(resource, context)
			for _, err := range resourceErrors {
				if err.Severity == "error" {
					result.Errors = append(result.Errors, err)
				} else {
					result.Warnings = append(result.Warnings, err)
				}
			}
		}
	}

	// Validate dependencies
	dependencyErrors := reg.ValidateDependencies()
	for _, err := range dependencyErrors {
		result.Errors = append(result.Errors, ValidationError{
			Type:     "dependency",
			Message:  err.Error(),
			Resource: "registry",
			Field:    "",
			Severity: "error",
		})
	}

	result.ValidResources = result.TotalResources - len(result.Errors)
	result.Success = len(result.Errors) == 0

	return result
}

// ValidateResource validates a single resource
func (v *Validator) ValidateResource(resource *parser.ParsedResource, context *ValidationContext) []ValidationError {
	errors := []ValidationError{}

	// Basic YAML structure validation (already done by parser)
	
	// Naming convention validation
	if v.namingValidator != nil && v.isValidatorEnabled("naming") {
		namingErrors := v.namingValidator.ValidateResourceName(resource.Resource, context)
		errors = append(errors, namingErrors...)
	}

	// Tagging policy validation
	if v.taggingValidator != nil && v.isValidatorEnabled("tagging") {
		taggingErrors := v.taggingValidator.ValidateResourceTags(resource.Resource, context)
		errors = append(errors, taggingErrors...)
	}

	// Security policy validation
	if v.securityValidator != nil && v.isValidatorEnabled("security") {
		securityErrors := v.securityValidator.ValidateResourceSecurity(resource.Resource, context)
		errors = append(errors, securityErrors...)
	}

	// Add file path context to errors
	for i := range errors {
		if errors[i].Resource == "" {
			errors[i].Resource = filepath.Base(resource.FilePath)
		}
	}

	return errors
}

// isValidatorEnabled checks if a validator is enabled
func (v *Validator) isValidatorEnabled(validatorType string) bool {
	if len(v.config.EnabledValidators) == 0 {
		return true // All enabled by default
	}

	for _, enabled := range v.config.EnabledValidators {
		if enabled == validatorType || enabled == "all" {
			return true
		}
	}

	return false
}

// ValidationResult holds the results of validation
type ValidationResult struct {
	TotalResources  int
	ValidResources  int
	Errors          []ValidationError
	Warnings        []ValidationError
	Success         bool
}

// PrintSummary prints a summary of validation results
func (r *ValidationResult) PrintSummary() {
	if r.Success {
		fmt.Printf("✅ All resources are valid!\n")
		fmt.Printf("   └─ %d resources passed validation\n\n", r.ValidResources)
		
		if len(r.Warnings) > 0 {
			fmt.Printf("⚠️  %d warnings:\n", len(r.Warnings))
			for i, warning := range r.Warnings {
				fmt.Printf("   %d. %s\n", i+1, warning.Message)
			}
			fmt.Printf("\n")
		}
		return
	}

	fmt.Printf("❌ Validation failed with %d errors:\n\n", len(r.Errors))

	for i, err := range r.Errors {
		fmt.Printf("   %d. [%s] %s\n", i+1, err.Type, err.Message)
		if err.Resource != "" {
			fmt.Printf("      Resource: %s\n", err.Resource)
		}
		if err.Field != "" {
			fmt.Printf("      Field: %s\n", err.Field)
		}
		fmt.Printf("\n")
	}

	if r.ValidResources > 0 {
		fmt.Printf("✅ %d resources passed validation\n", r.ValidResources)
	}
	fmt.Printf("❌ %d validation errors found\n", len(r.Errors))
	
	if len(r.Warnings) > 0 {
		fmt.Printf("⚠️  %d warnings found\n", len(r.Warnings))
	}
	
	fmt.Printf("\n")
}

// DefaultValidationConfig returns a default validation configuration
func DefaultValidationConfig() *ValidationConfig {
	return &ValidationConfig{
		NamingConventions: DefaultNamingConventions(),
		TaggingPolicies:   DefaultTaggingPolicies(),
		SecurityPolicies:  DefaultSecurityPolicies(),
		EnabledValidators: []string{"naming", "tagging", "security"},
	}
}

// EnterpriseValidationConfig returns an enterprise-grade validation configuration
func EnterpriseValidationConfig() *ValidationConfig {
	return &ValidationConfig{
		NamingConventions: EnterpriseNamingConventions(),
		TaggingPolicies:   EnterpriseTaggingPolicies(),
		SecurityPolicies:  EnterpriseSecurityPolicies(),
		EnabledValidators: []string{"naming", "tagging", "security"},
	}
}