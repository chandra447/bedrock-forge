package validation

import (
	"fmt"
	"regexp"
	"strings"

	"bedrock-forge/internal/models"
)

// TaggingPolicyConfig defines mandatory and optional tagging requirements
type TaggingPolicyConfig struct {
	// Global tagging requirements applied to all resources
	Global *TaggingRequirements `yaml:"global,omitempty"`

	// Resource-specific tagging requirements
	Resources map[string]*TaggingRequirements `yaml:"resources,omitempty"`

	// Team-specific tagging requirements
	Teams map[string]*TaggingRequirements `yaml:"teams,omitempty"`

	// Environment-specific tagging requirements
	Environments map[string]*TaggingRequirements `yaml:"environments,omitempty"`

	// Tag value validation rules
	TagValidation map[string]*TagValidationRule `yaml:"tagValidation,omitempty"`
}

// TaggingRequirements defines what tags are required
type TaggingRequirements struct {
	// Required tags that must be present
	RequiredTags []string `yaml:"requiredTags"`

	// Optional tags that are recommended
	OptionalTags []string `yaml:"optionalTags,omitempty"`

	// Forbidden tags that must not be present
	ForbiddenTags []string `yaml:"forbiddenTags,omitempty"`

	// Whether to inherit tags from higher-level configurations
	InheritTags bool `yaml:"inheritTags,omitempty"`

	// Default tag values to apply if not specified
	DefaultTags map[string]string `yaml:"defaultTags,omitempty"`

	// Custom validation message
	ValidationMessage string `yaml:"validationMessage,omitempty"`
}

// TagValidationRule defines validation rules for specific tag values
type TagValidationRule struct {
	// Regex pattern the tag value must match
	Pattern string `yaml:"pattern,omitempty"`

	// Compiled regex pattern (internal use)
	CompiledPattern *regexp.Regexp `yaml:"-"`

	// Allowed values (whitelist)
	AllowedValues []string `yaml:"allowedValues,omitempty"`

	// Forbidden values (blacklist)
	ForbiddenValues []string `yaml:"forbiddenValues,omitempty"`

	// Minimum length for tag value
	MinLength int `yaml:"minLength,omitempty"`

	// Maximum length for tag value
	MaxLength int `yaml:"maxLength,omitempty"`

	// Whether the tag value is case sensitive
	CaseSensitive bool `yaml:"caseSensitive,omitempty"`

	// Custom validation message
	ValidationMessage string `yaml:"validationMessage,omitempty"`
}

// TaggingValidator validates resource tags against tagging policies
type TaggingValidator struct {
	config *TaggingPolicyConfig
}

// NewTaggingValidator creates a new tagging validator
func NewTaggingValidator(config *TaggingPolicyConfig) (*TaggingValidator, error) {
	validator := &TaggingValidator{
		config: config,
	}

	// Compile all regex patterns
	if err := validator.compilePatterns(); err != nil {
		return nil, fmt.Errorf("failed to compile tagging patterns: %w", err)
	}

	return validator, nil
}

// compilePatterns compiles all regex patterns in tag validation rules
func (v *TaggingValidator) compilePatterns() error {
	if v.config.TagValidation == nil {
		return nil
	}

	for tagName, rule := range v.config.TagValidation {
		if rule.Pattern != "" {
			compiled, err := regexp.Compile(rule.Pattern)
			if err != nil {
				return fmt.Errorf("invalid regex pattern for tag '%s': %w", tagName, err)
			}
			rule.CompiledPattern = compiled
		}
	}

	return nil
}

// ValidateResourceTags validates resource tags against tagging policies
func (v *TaggingValidator) ValidateResourceTags(resource interface{}, context *ValidationContext) []ValidationError {
	errors := []ValidationError{}

	// Extract resource tags and metadata
	var tags map[string]string
	var metadata models.Metadata
	var resourceType string

	switch r := resource.(type) {
	case *models.Agent:
		tags = r.Spec.Tags
		metadata = r.Metadata
		resourceType = "Agent"
	case *models.Lambda:
		tags = r.Spec.Tags
		metadata = r.Metadata
		resourceType = "Lambda"
	case *models.ActionGroup:
		tags = make(map[string]string) // ActionGroup doesn't have tags in spec
		metadata = r.Metadata
		resourceType = "ActionGroup"
	case *models.KnowledgeBase:
		tags = r.Spec.Tags
		metadata = r.Metadata
		resourceType = "KnowledgeBase"
	case *models.Guardrail:
		tags = r.Spec.Tags
		metadata = r.Metadata
		resourceType = "Guardrail"
	case *models.Prompt:
		tags = r.Spec.Tags
		metadata = r.Metadata
		resourceType = "Prompt"
	case *models.IAMRole:
		tags = r.Spec.Tags
		metadata = r.Metadata
		resourceType = "IAMRole"
	default:
		// Skip unknown resource types
		return errors
	}

	// If no tags map exists, create an empty one
	if tags == nil {
		tags = make(map[string]string)
	}

	// Get applicable tagging requirements
	requirements := v.getApplicableRequirements(resourceType, context)

	// Validate against each requirement
	for _, req := range requirements {
		validationErrors := v.validateTagsAgainstRequirement(tags, req, resourceType, metadata.Name, context)
		errors = append(errors, validationErrors...)
	}

	// Validate individual tag values
	for tagName, tagValue := range tags {
		if rule, exists := v.config.TagValidation[tagName]; exists {
			if err := v.validateTagValue(tagName, tagValue, rule, resourceType, metadata.Name); err != nil {
				errors = append(errors, *err)
			}
		}
	}

	return errors
}

// getApplicableRequirements returns the tagging requirements that apply to a resource
func (v *TaggingValidator) getApplicableRequirements(resourceType string, context *ValidationContext) []*TaggingRequirements {
	requirements := []*TaggingRequirements{}

	// Add global requirements
	if v.config.Global != nil {
		requirements = append(requirements, v.config.Global)
	}

	// Add resource-specific requirements
	if resourceReqs, exists := v.config.Resources[resourceType]; exists {
		requirements = append(requirements, resourceReqs)
	}

	// Add team-specific requirements
	if context != nil && context.Team != "" {
		if teamReqs, exists := v.config.Teams[context.Team]; exists {
			requirements = append(requirements, teamReqs)
		}
	}

	// Add environment-specific requirements
	if context != nil && context.Environment != "" {
		if envReqs, exists := v.config.Environments[context.Environment]; exists {
			requirements = append(requirements, envReqs)
		}
	}

	return requirements
}

// validateTagsAgainstRequirement validates tags against a specific requirement
func (v *TaggingValidator) validateTagsAgainstRequirement(tags map[string]string, req *TaggingRequirements, resourceType, resourceName string, context *ValidationContext) []ValidationError {
	errors := []ValidationError{}

	// Check required tags
	for _, requiredTag := range req.RequiredTags {
		if _, exists := tags[requiredTag]; !exists {
			message := fmt.Sprintf("Required tag '%s' is missing", requiredTag)
			if req.ValidationMessage != "" {
				message = req.ValidationMessage
			}

			errors = append(errors, ValidationError{
				Type:     "tagging_policy",
				Message:  message,
				Resource: fmt.Sprintf("%s/%s", resourceType, resourceName),
				Field:    fmt.Sprintf("spec.tags.%s", requiredTag),
				Severity: "error",
			})
		}
	}

	// Check forbidden tags
	for _, forbiddenTag := range req.ForbiddenTags {
		if _, exists := tags[forbiddenTag]; exists {
			errors = append(errors, ValidationError{
				Type:     "tagging_policy",
				Message:  fmt.Sprintf("Forbidden tag '%s' is present", forbiddenTag),
				Resource: fmt.Sprintf("%s/%s", resourceType, resourceName),
				Field:    fmt.Sprintf("spec.tags.%s", forbiddenTag),
				Severity: "error",
			})
		}
	}

	// Warn about missing optional tags
	for _, optionalTag := range req.OptionalTags {
		if _, exists := tags[optionalTag]; !exists {
			errors = append(errors, ValidationError{
				Type:     "tagging_policy",
				Message:  fmt.Sprintf("Optional tag '%s' is missing (recommended for compliance)", optionalTag),
				Resource: fmt.Sprintf("%s/%s", resourceType, resourceName),
				Field:    fmt.Sprintf("spec.tags.%s", optionalTag),
				Severity: "warning",
			})
		}
	}

	return errors
}

// validateTagValue validates a tag value against its validation rule
func (v *TaggingValidator) validateTagValue(tagName, tagValue string, rule *TagValidationRule, resourceType, resourceName string) *ValidationError {
	// Check regex pattern
	if rule.CompiledPattern != nil && !rule.CompiledPattern.MatchString(tagValue) {
		return &ValidationError{
			Type:     "tag_validation",
			Message:  v.getTagValidationMessage(rule, fmt.Sprintf("Tag '%s' value '%s' does not match required pattern '%s'", tagName, tagValue, rule.Pattern)),
			Resource: fmt.Sprintf("%s/%s", resourceType, resourceName),
			Field:    fmt.Sprintf("spec.tags.%s", tagName),
			Severity: "error",
		}
	}

	// Check allowed values
	if len(rule.AllowedValues) > 0 {
		allowed := false
		compareValue := tagValue
		if !rule.CaseSensitive {
			compareValue = strings.ToLower(tagValue)
		}

		for _, allowedValue := range rule.AllowedValues {
			checkValue := allowedValue
			if !rule.CaseSensitive {
				checkValue = strings.ToLower(allowedValue)
			}
			if compareValue == checkValue {
				allowed = true
				break
			}
		}

		if !allowed {
			return &ValidationError{
				Type:     "tag_validation",
				Message:  v.getTagValidationMessage(rule, fmt.Sprintf("Tag '%s' value '%s' is not in allowed values: %v", tagName, tagValue, rule.AllowedValues)),
				Resource: fmt.Sprintf("%s/%s", resourceType, resourceName),
				Field:    fmt.Sprintf("spec.tags.%s", tagName),
				Severity: "error",
			}
		}
	}

	// Check forbidden values
	if len(rule.ForbiddenValues) > 0 {
		compareValue := tagValue
		if !rule.CaseSensitive {
			compareValue = strings.ToLower(tagValue)
		}

		for _, forbiddenValue := range rule.ForbiddenValues {
			checkValue := forbiddenValue
			if !rule.CaseSensitive {
				checkValue = strings.ToLower(forbiddenValue)
			}
			if compareValue == checkValue {
				return &ValidationError{
					Type:     "tag_validation",
					Message:  v.getTagValidationMessage(rule, fmt.Sprintf("Tag '%s' value '%s' is forbidden", tagName, tagValue)),
					Resource: fmt.Sprintf("%s/%s", resourceType, resourceName),
					Field:    fmt.Sprintf("spec.tags.%s", tagName),
					Severity: "error",
				}
			}
		}
	}

	// Check length constraints
	if rule.MinLength > 0 && len(tagValue) < rule.MinLength {
		return &ValidationError{
			Type:     "tag_validation",
			Message:  v.getTagValidationMessage(rule, fmt.Sprintf("Tag '%s' value '%s' must be at least %d characters long", tagName, tagValue, rule.MinLength)),
			Resource: fmt.Sprintf("%s/%s", resourceType, resourceName),
			Field:    fmt.Sprintf("spec.tags.%s", tagName),
			Severity: "error",
		}
	}

	if rule.MaxLength > 0 && len(tagValue) > rule.MaxLength {
		return &ValidationError{
			Type:     "tag_validation",
			Message:  v.getTagValidationMessage(rule, fmt.Sprintf("Tag '%s' value '%s' must be at most %d characters long", tagName, tagValue, rule.MaxLength)),
			Resource: fmt.Sprintf("%s/%s", resourceType, resourceName),
			Field:    fmt.Sprintf("spec.tags.%s", tagName),
			Severity: "error",
		}
	}

	return nil
}

// getTagValidationMessage returns the appropriate validation message
func (v *TaggingValidator) getTagValidationMessage(rule *TagValidationRule, defaultMessage string) string {
	if rule.ValidationMessage != "" {
		return rule.ValidationMessage
	}
	return defaultMessage
}

// DefaultTaggingPolicies returns a set of default tagging policies
func DefaultTaggingPolicies() *TaggingPolicyConfig {
	return &TaggingPolicyConfig{
		Global: &TaggingRequirements{
			RequiredTags: []string{"Environment", "Project", "Owner"},
			OptionalTags: []string{"CostCenter", "Team", "Contact"},
		},
		Resources: map[string]*TaggingRequirements{
			"Agent": {
				RequiredTags: []string{"AgentType", "BusinessFunction"},
				OptionalTags: []string{"DataClassification", "ComplianceLevel"},
			},
			"Lambda": {
				RequiredTags: []string{"Runtime", "FunctionType"},
				OptionalTags: []string{"ScheduleType", "TriggerType"},
			},
			"KnowledgeBase": {
				RequiredTags: []string{"DataSource", "ContentType"},
				OptionalTags: []string{"DataClassification", "RefreshSchedule"},
			},
		},
		TagValidation: map[string]*TagValidationRule{
			"Environment": {
				AllowedValues: []string{"dev", "staging", "prod", "test"},
				CaseSensitive: false,
			},
			"Owner": {
				Pattern:           `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
				ValidationMessage: "Owner must be a valid email address",
			},
			"CostCenter": {
				Pattern:           `^CC-\d{4}$`,
				ValidationMessage: "CostCenter must follow format: CC-XXXX (e.g., CC-1234)",
			},
		},
	}
}

// EnterpriseTaggingPolicies returns a stricter set of tagging policies for enterprise use
func EnterpriseTaggingPolicies() *TaggingPolicyConfig {
	return &TaggingPolicyConfig{
		Global: &TaggingRequirements{
			RequiredTags: []string{
				"Environment",
				"Project",
				"Owner",
				"CostCenter",
				"Team",
				"BusinessUnit",
				"DataClassification",
				"ComplianceLevel",
			},
			OptionalTags: []string{"BackupRequired", "MonitoringLevel", "SLA"},
		},
		Resources: map[string]*TaggingRequirements{
			"Agent": {
				RequiredTags: []string{
					"AgentType",
					"BusinessFunction",
					"DataProcessing",
					"SecurityLevel",
				},
				ValidationMessage: "Bedrock agents require comprehensive tagging for compliance and cost tracking",
			},
			"Lambda": {
				RequiredTags: []string{
					"Runtime",
					"FunctionType",
					"ExecutionRole",
					"SecurityLevel",
				},
			},
			"KnowledgeBase": {
				RequiredTags: []string{
					"DataSource",
					"ContentType",
					"DataSensitivity",
					"RetentionPeriod",
				},
			},
			"IAMRole": {
				RequiredTags: []string{
					"RoleType",
					"AccessLevel",
					"AuditRequired",
				},
			},
		},
		TagValidation: map[string]*TagValidationRule{
			"Environment": {
				AllowedValues: []string{"dev", "staging", "prod"},
				CaseSensitive: false,
			},
			"Owner": {
				Pattern:           `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
				ValidationMessage: "Owner must be a valid corporate email address",
			},
			"CostCenter": {
				Pattern:           `^CC-\d{6}$`,
				ValidationMessage: "CostCenter must follow corporate format: CC-XXXXXX",
			},
			"Team": {
				AllowedValues: []string{
					"engineering", "data", "security", "operations",
					"product", "compliance", "finance", "legal",
				},
				CaseSensitive: false,
			},
			"DataClassification": {
				AllowedValues: []string{"public", "internal", "confidential", "restricted"},
				CaseSensitive: false,
			},
			"ComplianceLevel": {
				AllowedValues: []string{"none", "pci", "hipaa", "sox", "gdpr"},
				CaseSensitive: false,
			},
			"SecurityLevel": {
				AllowedValues: []string{"low", "medium", "high", "critical"},
				CaseSensitive: false,
			},
		},
	}
}
