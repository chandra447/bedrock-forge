package validation

import (
	"fmt"
	"regexp"
	"strings"

	"bedrock-forge/internal/models"
)

// NamingConventionConfig defines the naming rules for resources
type NamingConventionConfig struct {
	// Global naming rules applied to all resources
	Global *NamingRules `yaml:"global,omitempty"`

	// Resource-specific naming rules
	Resources map[string]*NamingRules `yaml:"resources,omitempty"`

	// Team-specific naming rules
	Teams map[string]*NamingRules `yaml:"teams,omitempty"`

	// Environment-specific naming rules
	Environments map[string]*NamingRules `yaml:"environments,omitempty"`
}

// NamingRules defines the naming convention rules
type NamingRules struct {
	// Required prefix for resource names
	Prefix string `yaml:"prefix,omitempty"`

	// Required suffix for resource names
	Suffix string `yaml:"suffix,omitempty"`

	// Regex pattern the name must match
	Pattern string `yaml:"pattern,omitempty"`

	// Compiled regex pattern (internal use)
	CompiledPattern *regexp.Regexp `yaml:"-"`

	// Minimum length for resource names
	MinLength int `yaml:"minLength,omitempty"`

	// Maximum length for resource names
	MaxLength int `yaml:"maxLength,omitempty"`

	// Allowed characters (regex character class)
	AllowedChars string `yaml:"allowedChars,omitempty"`

	// Forbidden characters (regex character class)
	ForbiddenChars string `yaml:"forbiddenChars,omitempty"`

	// Whether to enforce lowercase names
	ForceLowercase bool `yaml:"forceLowercase,omitempty"`

	// Whether to enforce uppercase names
	ForceUppercase bool `yaml:"forceUppercase,omitempty"`

	// Custom validation messages
	ValidationMessage string `yaml:"validationMessage,omitempty"`
}

// NamingValidator validates resource names against naming conventions
type NamingValidator struct {
	config *NamingConventionConfig
}

// NewNamingValidator creates a new naming validator
func NewNamingValidator(config *NamingConventionConfig) (*NamingValidator, error) {
	validator := &NamingValidator{
		config: config,
	}

	// Compile all regex patterns
	if err := validator.compilePatterns(); err != nil {
		return nil, fmt.Errorf("failed to compile naming patterns: %w", err)
	}

	return validator, nil
}

// compilePatterns compiles all regex patterns in the configuration
func (v *NamingValidator) compilePatterns() error {
	patterns := []*NamingRules{}

	// Collect all naming rules
	if v.config.Global != nil {
		patterns = append(patterns, v.config.Global)
	}

	for _, rules := range v.config.Resources {
		patterns = append(patterns, rules)
	}

	for _, rules := range v.config.Teams {
		patterns = append(patterns, rules)
	}

	for _, rules := range v.config.Environments {
		patterns = append(patterns, rules)
	}

	// Compile each pattern
	for _, rules := range patterns {
		if rules.Pattern != "" {
			compiled, err := regexp.Compile(rules.Pattern)
			if err != nil {
				return fmt.Errorf("invalid regex pattern '%s': %w", rules.Pattern, err)
			}
			rules.CompiledPattern = compiled
		}
	}

	return nil
}

// ValidateResourceName validates a resource name against naming conventions
func (v *NamingValidator) ValidateResourceName(resource interface{}, context *ValidationContext) []ValidationError {
	errors := []ValidationError{}

	// Extract resource metadata
	var metadata models.Metadata
	var resourceType string

	switch r := resource.(type) {
	case *models.Agent:
		metadata = r.Metadata
		resourceType = "Agent"
	case *models.Lambda:
		metadata = r.Metadata
		resourceType = "Lambda"
	case *models.ActionGroup:
		metadata = r.Metadata
		resourceType = "ActionGroup"
	case *models.KnowledgeBase:
		metadata = r.Metadata
		resourceType = "KnowledgeBase"
	case *models.Guardrail:
		metadata = r.Metadata
		resourceType = "Guardrail"
	case *models.Prompt:
		metadata = r.Metadata
		resourceType = "Prompt"
	case *models.IAMRole:
		metadata = r.Metadata
		resourceType = "IAMRole"
	default:
		// Skip unknown resource types
		return errors
	}

	// Get applicable naming rules
	rules := v.getApplicableRules(resourceType, context)

	// Validate against each rule
	for _, rule := range rules {
		if err := v.validateNameAgainstRule(metadata.Name, rule, resourceType, context); err != nil {
			errors = append(errors, *err)
		}
	}

	return errors
}

// getApplicableRules returns the naming rules that apply to a resource
func (v *NamingValidator) getApplicableRules(resourceType string, context *ValidationContext) []*NamingRules {
	rules := []*NamingRules{}

	// Add global rules
	if v.config.Global != nil {
		rules = append(rules, v.config.Global)
	}

	// Add resource-specific rules
	if resourceRules, exists := v.config.Resources[resourceType]; exists {
		rules = append(rules, resourceRules)
	}

	// Add team-specific rules
	if context != nil && context.Team != "" {
		if teamRules, exists := v.config.Teams[context.Team]; exists {
			rules = append(rules, teamRules)
		}
	}

	// Add environment-specific rules
	if context != nil && context.Environment != "" {
		if envRules, exists := v.config.Environments[context.Environment]; exists {
			rules = append(rules, envRules)
		}
	}

	return rules
}

// validateNameAgainstRule validates a name against a specific naming rule
func (v *NamingValidator) validateNameAgainstRule(name string, rule *NamingRules, resourceType string, context *ValidationContext) *ValidationError {
	// Check prefix
	if rule.Prefix != "" && !strings.HasPrefix(name, rule.Prefix) {
		return &ValidationError{
			Type:     "naming_convention",
			Message:  v.getValidationMessage(rule, fmt.Sprintf("Resource name '%s' must start with prefix '%s'", name, rule.Prefix)),
			Resource: fmt.Sprintf("%s/%s", resourceType, name),
			Field:    "metadata.name",
		}
	}

	// Check suffix
	if rule.Suffix != "" && !strings.HasSuffix(name, rule.Suffix) {
		return &ValidationError{
			Type:     "naming_convention",
			Message:  v.getValidationMessage(rule, fmt.Sprintf("Resource name '%s' must end with suffix '%s'", name, rule.Suffix)),
			Resource: fmt.Sprintf("%s/%s", resourceType, name),
			Field:    "metadata.name",
		}
	}

	// Check regex pattern
	if rule.CompiledPattern != nil && !rule.CompiledPattern.MatchString(name) {
		return &ValidationError{
			Type:     "naming_convention",
			Message:  v.getValidationMessage(rule, fmt.Sprintf("Resource name '%s' does not match required pattern '%s'", name, rule.Pattern)),
			Resource: fmt.Sprintf("%s/%s", resourceType, name),
			Field:    "metadata.name",
		}
	}

	// Check length constraints
	if rule.MinLength > 0 && len(name) < rule.MinLength {
		return &ValidationError{
			Type:     "naming_convention",
			Message:  v.getValidationMessage(rule, fmt.Sprintf("Resource name '%s' must be at least %d characters long", name, rule.MinLength)),
			Resource: fmt.Sprintf("%s/%s", resourceType, name),
			Field:    "metadata.name",
		}
	}

	if rule.MaxLength > 0 && len(name) > rule.MaxLength {
		return &ValidationError{
			Type:     "naming_convention",
			Message:  v.getValidationMessage(rule, fmt.Sprintf("Resource name '%s' must be at most %d characters long", name, rule.MaxLength)),
			Resource: fmt.Sprintf("%s/%s", resourceType, name),
			Field:    "metadata.name",
		}
	}

	// Check allowed characters
	if rule.AllowedChars != "" {
		allowedPattern := fmt.Sprintf("^[%s]+$", rule.AllowedChars)
		if matched, _ := regexp.MatchString(allowedPattern, name); !matched {
			return &ValidationError{
				Type:     "naming_convention",
				Message:  v.getValidationMessage(rule, fmt.Sprintf("Resource name '%s' contains invalid characters. Allowed: %s", name, rule.AllowedChars)),
				Resource: fmt.Sprintf("%s/%s", resourceType, name),
				Field:    "metadata.name",
			}
		}
	}

	// Check forbidden characters
	if rule.ForbiddenChars != "" {
		forbiddenPattern := fmt.Sprintf("[%s]", rule.ForbiddenChars)
		if matched, _ := regexp.MatchString(forbiddenPattern, name); matched {
			return &ValidationError{
				Type:     "naming_convention",
				Message:  v.getValidationMessage(rule, fmt.Sprintf("Resource name '%s' contains forbidden characters: %s", name, rule.ForbiddenChars)),
				Resource: fmt.Sprintf("%s/%s", resourceType, name),
				Field:    "metadata.name",
			}
		}
	}

	// Check case enforcement
	if rule.ForceLowercase && name != strings.ToLower(name) {
		return &ValidationError{
			Type:     "naming_convention",
			Message:  v.getValidationMessage(rule, fmt.Sprintf("Resource name '%s' must be lowercase", name)),
			Resource: fmt.Sprintf("%s/%s", resourceType, name),
			Field:    "metadata.name",
		}
	}

	if rule.ForceUppercase && name != strings.ToUpper(name) {
		return &ValidationError{
			Type:     "naming_convention",
			Message:  v.getValidationMessage(rule, fmt.Sprintf("Resource name '%s' must be uppercase", name)),
			Resource: fmt.Sprintf("%s/%s", resourceType, name),
			Field:    "metadata.name",
		}
	}

	return nil
}

// getValidationMessage returns the appropriate validation message
func (v *NamingValidator) getValidationMessage(rule *NamingRules, defaultMessage string) string {
	if rule.ValidationMessage != "" {
		return rule.ValidationMessage
	}
	return defaultMessage
}

// ValidationContext provides context for validation
type ValidationContext struct {
	Team        string
	Environment string
	Project     string
	Region      string
}

// ValidationError represents a naming convention validation error
type ValidationError struct {
	Type     string
	Message  string
	Resource string
	Field    string
	Severity string
}

// DefaultNamingConventions returns a set of enterprise-friendly default naming conventions
func DefaultNamingConventions() *NamingConventionConfig {
	return &NamingConventionConfig{
		Global: &NamingRules{
			MinLength:      3,
			MaxLength:      64,
			AllowedChars:   "a-zA-Z0-9-_",
			ForbiddenChars: " ",
			ForceLowercase: false,
			Pattern:        "^[a-zA-Z][a-zA-Z0-9-_]*$", // Must start with letter
		},
		Resources: map[string]*NamingRules{
			"Agent": {
				Suffix:  "-agent",
				Pattern: "^[a-z][a-z0-9-]*-agent$",
			},
			"Lambda": {
				Suffix:  "-lambda",
				Pattern: "^[a-z][a-z0-9-]*-lambda$",
			},
			"ActionGroup": {
				Suffix:  "-action-group",
				Pattern: "^[a-z][a-z0-9-]*-action-group$",
			},
			"KnowledgeBase": {
				Suffix:  "-kb",
				Pattern: "^[a-z][a-z0-9-]*-kb$",
			},
			"Guardrail": {
				Suffix:  "-guardrail",
				Pattern: "^[a-z][a-z0-9-]*-guardrail$",
			},
			"Prompt": {
				Suffix:  "-prompt",
				Pattern: "^[a-z][a-z0-9-]*-prompt$",
			},
			"IAMRole": {
				Suffix:  "-role",
				Pattern: "^[a-zA-Z][a-zA-Z0-9-]*-role$",
			},
		},
		Teams: map[string]*NamingRules{
			"engineering": {
				Prefix: "eng-",
			},
			"data": {
				Prefix: "data-",
			},
			"security": {
				Prefix: "sec-",
			},
		},
		Environments: map[string]*NamingRules{
			"dev": {
				Prefix: "dev-",
			},
			"staging": {
				Prefix: "staging-",
			},
			"prod": {
				Prefix: "prod-",
			},
		},
	}
}

// EnterpriseNamingConventions returns a stricter set of naming conventions for enterprise use
func EnterpriseNamingConventions() *NamingConventionConfig {
	return &NamingConventionConfig{
		Global: &NamingRules{
			MinLength:      5,
			MaxLength:      50,
			AllowedChars:   "a-z0-9-",
			ForceLowercase: true,
			Pattern:        "^[a-z][a-z0-9-]*[a-z0-9]$", // Must start with letter, end with letter or number
		},
		Resources: map[string]*NamingRules{
			"Agent": {
				Pattern:           "^[a-z]+-(dev|staging|prod)-[a-z0-9-]+-agent$",
				ValidationMessage: "Agent names must follow pattern: <team>-<env>-<name>-agent",
			},
			"Lambda": {
				Pattern:           "^[a-z]+-(dev|staging|prod)-[a-z0-9-]+-lambda$",
				ValidationMessage: "Lambda names must follow pattern: <team>-<env>-<name>-lambda",
			},
			"ActionGroup": {
				Pattern:           "^[a-z]+-(dev|staging|prod)-[a-z0-9-]+-action-group$",
				ValidationMessage: "ActionGroup names must follow pattern: <team>-<env>-<name>-action-group",
			},
			"KnowledgeBase": {
				Pattern:           "^[a-z]+-(dev|staging|prod)-[a-z0-9-]+-kb$",
				ValidationMessage: "KnowledgeBase names must follow pattern: <team>-<env>-<name>-kb",
			},
			"Guardrail": {
				Pattern:           "^[a-z]+-(dev|staging|prod)-[a-z0-9-]+-guardrail$",
				ValidationMessage: "Guardrail names must follow pattern: <team>-<env>-<name>-guardrail",
			},
			"Prompt": {
				Pattern:           "^[a-z]+-(dev|staging|prod)-[a-z0-9-]+-prompt$",
				ValidationMessage: "Prompt names must follow pattern: <team>-<env>-<name>-prompt",
			},
			"IAMRole": {
				Pattern:           "^[a-z]+-(dev|staging|prod)-[a-z0-9-]+-role$",
				ValidationMessage: "IAMRole names must follow pattern: <team>-<env>-<name>-role",
			},
		},
	}
}
