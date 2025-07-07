package validation

import (
	"fmt"
	"regexp"
	"strings"

	"bedrock-forge/internal/models"
)

// SecurityPolicyConfig defines security validation requirements
type SecurityPolicyConfig struct {
	// IAM policy validation rules
	IAMPolicies *IAMPolicyValidation `yaml:"iamPolicies,omitempty"`

	// Lambda security requirements
	LambdaSecurity *LambdaSecurityValidation `yaml:"lambdaSecurity,omitempty"`

	// Agent security requirements
	AgentSecurity *AgentSecurityValidation `yaml:"agentSecurity,omitempty"`

	// Knowledge base security requirements
	KnowledgeBaseSecurity *KnowledgeBaseSecurityValidation `yaml:"knowledgeBaseSecurity,omitempty"`

	// Encryption requirements
	EncryptionRequirements *EncryptionValidation `yaml:"encryptionRequirements,omitempty"`

	// Network security requirements
	NetworkSecurity *NetworkSecurityValidation `yaml:"networkSecurity,omitempty"`
}

// IAMPolicyValidation defines IAM policy validation rules
type IAMPolicyValidation struct {
	// Forbidden actions that should never be allowed
	ForbiddenActions []string `yaml:"forbiddenActions,omitempty"`

	// Required actions for specific resource types
	RequiredActions map[string][]string `yaml:"requiredActions,omitempty"`

	// Forbidden resources patterns
	ForbiddenResources []string `yaml:"forbiddenResources,omitempty"`

	// Maximum policy size
	MaxPolicySize int `yaml:"maxPolicySize,omitempty"`

	// Whether wildcard resources are allowed
	AllowWildcardResources bool `yaml:"allowWildcardResources,omitempty"`

	// Whether admin permissions are allowed
	AllowAdminPermissions bool `yaml:"allowAdminPermissions,omitempty"`

	// Require MFA for sensitive actions
	RequireMFAForSensitiveActions bool `yaml:"requireMFAForSensitiveActions,omitempty"`

	// Sensitive actions that require additional controls
	SensitiveActions []string `yaml:"sensitiveActions,omitempty"`
}

// LambdaSecurityValidation defines Lambda security requirements
type LambdaSecurityValidation struct {
	// Require VPC configuration for Lambda functions
	RequireVPC bool `yaml:"requireVPC,omitempty"`

	// Forbidden environment variable patterns
	ForbiddenEnvPatterns []string `yaml:"forbiddenEnvPatterns,omitempty"`

	// Required security headers for HTTP-triggered functions
	RequiredSecurityHeaders []string `yaml:"requiredSecurityHeaders,omitempty"`

	// Maximum execution timeout
	MaxTimeout int `yaml:"maxTimeout,omitempty"`

	// Maximum memory allocation
	MaxMemorySize int `yaml:"maxMemorySize,omitempty"`

	// Allowed runtimes
	AllowedRuntimes []string `yaml:"allowedRuntimes,omitempty"`

	// Require encryption for environment variables
	RequireEnvEncryption bool `yaml:"requireEnvEncryption,omitempty"`
}

// AgentSecurityValidation defines Bedrock agent security requirements
type AgentSecurityValidation struct {
	// Require guardrails for all agents
	RequireGuardrails bool `yaml:"requireGuardrails,omitempty"`

	// Required guardrail configurations
	RequiredGuardrailTypes []string `yaml:"requiredGuardrailTypes,omitempty"`

	// Maximum idle session timeout
	MaxIdleSessionTTL int `yaml:"maxIdleSessionTTL,omitempty"`

	// Require customer encryption key
	RequireCustomerEncryption bool `yaml:"requireCustomerEncryption,omitempty"`

	// Forbidden foundation models
	ForbiddenModels []string `yaml:"forbiddenModels,omitempty"`

	// Required memory configuration
	RequireMemoryConfiguration bool `yaml:"requireMemoryConfiguration,omitempty"`
}

// KnowledgeBaseSecurityValidation defines knowledge base security requirements
type KnowledgeBaseSecurityValidation struct {
	// Require encryption for data sources
	RequireDataSourceEncryption bool `yaml:"requireDataSourceEncryption,omitempty"`

	// Allowed data source types
	AllowedDataSourceTypes []string `yaml:"allowedDataSourceTypes,omitempty"`

	// Require VPC endpoints for data access
	RequireVPCEndpoints bool `yaml:"requireVPCEndpoints,omitempty"`

	// Maximum document retention period
	MaxRetentionDays int `yaml:"maxRetentionDays,omitempty"`

	// Require data source access logging
	RequireAccessLogging bool `yaml:"requireAccessLogging,omitempty"`
}

// EncryptionValidation defines encryption requirements
type EncryptionValidation struct {
	// Require encryption at rest
	RequireEncryptionAtRest bool `yaml:"requireEncryptionAtRest,omitempty"`

	// Require encryption in transit
	RequireEncryptionInTransit bool `yaml:"requireEncryptionInTransit,omitempty"`

	// Allowed KMS key patterns
	AllowedKMSKeyPatterns []string `yaml:"allowedKMSKeyPatterns,omitempty"`

	// Require customer-managed keys
	RequireCustomerManagedKeys bool `yaml:"requireCustomerManagedKeys,omitempty"`
}

// NetworkSecurityValidation defines network security requirements
type NetworkSecurityValidation struct {
	// Require private subnets for sensitive resources
	RequirePrivateSubnets bool `yaml:"requirePrivateSubnets,omitempty"`

	// Allowed security group patterns
	AllowedSecurityGroups []string `yaml:"allowedSecurityGroups,omitempty"`

	// Forbidden port ranges
	ForbiddenPorts []string `yaml:"forbiddenPorts,omitempty"`

	// Require VPC flow logs
	RequireVPCFlowLogs bool `yaml:"requireVPCFlowLogs,omitempty"`
}

// SecurityValidator validates resources against security policies
type SecurityValidator struct {
	config *SecurityPolicyConfig
}

// NewSecurityValidator creates a new security validator
func NewSecurityValidator(config *SecurityPolicyConfig) (*SecurityValidator, error) {
	return &SecurityValidator{
		config: config,
	}, nil
}

// ValidateResourceSecurity validates a resource against security policies
func (v *SecurityValidator) ValidateResourceSecurity(resource interface{}, context *ValidationContext) []ValidationError {
	errors := []ValidationError{}

	switch r := resource.(type) {
	case *models.Agent:
		errors = append(errors, v.validateAgentSecurity(r)...)
	case *models.Lambda:
		errors = append(errors, v.validateLambdaSecurity(r)...)
	case *models.KnowledgeBase:
		errors = append(errors, v.validateKnowledgeBaseSecurity(r)...)
	case *models.IAMRole:
		errors = append(errors, v.validateIAMRoleSecurity(r)...)
	}

	return errors
}

// validateAgentSecurity validates Bedrock agent security requirements
func (v *SecurityValidator) validateAgentSecurity(agent *models.Agent) []ValidationError {
	errors := []ValidationError{}

	if v.config.AgentSecurity == nil {
		return errors
	}

	config := v.config.AgentSecurity
	resourceName := fmt.Sprintf("Agent/%s", agent.Metadata.Name)

	// Check if guardrails are required
	if config.RequireGuardrails && agent.Spec.Guardrail == nil {
		errors = append(errors, ValidationError{
			Type:     "security_policy",
			Message:  "Bedrock agents must have guardrails configured for security compliance",
			Resource: resourceName,
			Field:    "spec.guardrail",
			Severity: "error",
		})
	}

	// Check idle session timeout
	if config.MaxIdleSessionTTL > 0 && agent.Spec.IdleSessionTTL > config.MaxIdleSessionTTL {
		errors = append(errors, ValidationError{
			Type:     "security_policy",
			Message:  fmt.Sprintf("Idle session timeout (%d) exceeds maximum allowed (%d)", agent.Spec.IdleSessionTTL, config.MaxIdleSessionTTL),
			Resource: resourceName,
			Field:    "spec.idleSessionTtl",
			Severity: "error",
		})
	}

	// Check customer encryption requirement
	if config.RequireCustomerEncryption && agent.Spec.CustomerEncryptionKey == "" {
		errors = append(errors, ValidationError{
			Type:     "security_policy",
			Message:  "Customer-managed encryption key is required for this agent",
			Resource: resourceName,
			Field:    "spec.customerEncryptionKey",
			Severity: "error",
		})
	}

	// Check forbidden models
	for _, forbiddenModel := range config.ForbiddenModels {
		if strings.Contains(agent.Spec.FoundationModel, forbiddenModel) {
			errors = append(errors, ValidationError{
				Type:     "security_policy",
				Message:  fmt.Sprintf("Foundation model '%s' contains forbidden pattern '%s'", agent.Spec.FoundationModel, forbiddenModel),
				Resource: resourceName,
				Field:    "spec.foundationModel",
				Severity: "error",
			})
		}
	}

	// Check memory configuration requirement
	if config.RequireMemoryConfiguration && agent.Spec.MemoryConfiguration == nil {
		errors = append(errors, ValidationError{
			Type:     "security_policy",
			Message:  "Memory configuration is required for security compliance",
			Resource: resourceName,
			Field:    "spec.memoryConfiguration",
			Severity: "error",
		})
	}

	return errors
}

// validateLambdaSecurity validates Lambda function security requirements
func (v *SecurityValidator) validateLambdaSecurity(lambda *models.Lambda) []ValidationError {
	errors := []ValidationError{}

	if v.config.LambdaSecurity == nil {
		return errors
	}

	config := v.config.LambdaSecurity
	resourceName := fmt.Sprintf("Lambda/%s", lambda.Metadata.Name)

	// Check VPC requirement
	if config.RequireVPC && lambda.Spec.VpcConfig == nil {
		errors = append(errors, ValidationError{
			Type:     "security_policy",
			Message:  "Lambda functions must be deployed in a VPC for security compliance",
			Resource: resourceName,
			Field:    "spec.vpcConfig",
			Severity: "error",
		})
	}

	// Check timeout limits
	if config.MaxTimeout > 0 && lambda.Spec.Timeout > config.MaxTimeout {
		errors = append(errors, ValidationError{
			Type:     "security_policy",
			Message:  fmt.Sprintf("Lambda timeout (%d) exceeds maximum allowed (%d)", lambda.Spec.Timeout, config.MaxTimeout),
			Resource: resourceName,
			Field:    "spec.timeout",
			Severity: "error",
		})
	}

	// Check memory limits
	if config.MaxMemorySize > 0 && lambda.Spec.MemorySize > config.MaxMemorySize {
		errors = append(errors, ValidationError{
			Type:     "security_policy",
			Message:  fmt.Sprintf("Lambda memory size (%d) exceeds maximum allowed (%d)", lambda.Spec.MemorySize, config.MaxMemorySize),
			Resource: resourceName,
			Field:    "spec.memorySize",
			Severity: "error",
		})
	}

	// Check allowed runtimes
	if len(config.AllowedRuntimes) > 0 {
		runtimeAllowed := false
		for _, allowedRuntime := range config.AllowedRuntimes {
			if lambda.Spec.Runtime == allowedRuntime {
				runtimeAllowed = true
				break
			}
		}
		if !runtimeAllowed {
			errors = append(errors, ValidationError{
				Type:     "security_policy",
				Message:  fmt.Sprintf("Runtime '%s' is not in the allowed list: %v", lambda.Spec.Runtime, config.AllowedRuntimes),
				Resource: resourceName,
				Field:    "spec.runtime",
				Severity: "error",
			})
		}
	}

	// Check environment variable patterns
	for envName, envValue := range lambda.Spec.Environment {
		for _, forbiddenPattern := range config.ForbiddenEnvPatterns {
			if matched, _ := regexp.MatchString(forbiddenPattern, envName); matched {
				errors = append(errors, ValidationError{
					Type:     "security_policy",
					Message:  fmt.Sprintf("Environment variable '%s' matches forbidden pattern '%s'", envName, forbiddenPattern),
					Resource: resourceName,
					Field:    fmt.Sprintf("spec.environment.%s", envName),
					Severity: "error",
				})
			}
			if matched, _ := regexp.MatchString(forbiddenPattern, envValue); matched {
				errors = append(errors, ValidationError{
					Type:     "security_policy",
					Message:  fmt.Sprintf("Environment variable value for '%s' matches forbidden pattern '%s'", envName, forbiddenPattern),
					Resource: resourceName,
					Field:    fmt.Sprintf("spec.environment.%s", envName),
					Severity: "error",
				})
			}
		}
	}

	return errors
}

// validateKnowledgeBaseSecurity validates knowledge base security requirements
func (v *SecurityValidator) validateKnowledgeBaseSecurity(kb *models.KnowledgeBase) []ValidationError {
	errors := []ValidationError{}

	if v.config.KnowledgeBaseSecurity == nil {
		return errors
	}

	config := v.config.KnowledgeBaseSecurity
	resourceName := fmt.Sprintf("KnowledgeBase/%s", kb.Metadata.Name)

	// Check allowed data source types
	if len(config.AllowedDataSourceTypes) > 0 {
		for _, dataSource := range kb.Spec.DataSources {
			typeAllowed := false
			for _, allowedType := range config.AllowedDataSourceTypes {
				if dataSource.Type == allowedType {
					typeAllowed = true
					break
				}
			}
			if !typeAllowed {
				errors = append(errors, ValidationError{
					Type:     "security_policy",
					Message:  fmt.Sprintf("Data source type '%s' is not in the allowed list: %v", dataSource.Type, config.AllowedDataSourceTypes),
					Resource: resourceName,
					Field:    "spec.dataSources[].type",
					Severity: "error",
				})
			}
		}
	}

	return errors
}

// validateIAMRoleSecurity validates IAM role security requirements
func (v *SecurityValidator) validateIAMRoleSecurity(role *models.IAMRole) []ValidationError {
	errors := []ValidationError{}

	if v.config.IAMPolicies == nil {
		return errors
	}

	resourceName := fmt.Sprintf("IAMRole/%s", role.Metadata.Name)

	// Validate inline policies
	for _, inlinePolicy := range role.Spec.InlinePolicies {
		policyErrors := v.validateIAMPolicyDocument(&inlinePolicy.Policy, resourceName, fmt.Sprintf("spec.inlinePolicies[%s]", inlinePolicy.Name))
		errors = append(errors, policyErrors...)
	}

	return errors
}

// validateIAMPolicyDocument validates an IAM policy document
func (v *SecurityValidator) validateIAMPolicyDocument(policy *models.IAMPolicyDocument, resourceName, fieldPath string) []ValidationError {
	errors := []ValidationError{}
	config := v.config.IAMPolicies

	for i, statement := range policy.Statement {
		statementPath := fmt.Sprintf("%s.statement[%d]", fieldPath, i)

		// Check for forbidden actions
		actions := v.normalizeActions(statement.Action)
		for _, action := range actions {
			for _, forbidden := range config.ForbiddenActions {
				if matched, _ := regexp.MatchString(forbidden, action); matched {
					errors = append(errors, ValidationError{
						Type:     "security_policy",
						Message:  fmt.Sprintf("IAM policy contains forbidden action '%s'", action),
						Resource: resourceName,
						Field:    fmt.Sprintf("%s.action", statementPath),
						Severity: "error",
					})
				}
			}
		}

		// Check for admin permissions
		if !config.AllowAdminPermissions {
			for _, action := range actions {
				if action == "*" || strings.HasSuffix(action, ":*") {
					errors = append(errors, ValidationError{
						Type:     "security_policy",
						Message:  fmt.Sprintf("IAM policy contains admin permissions '%s' which are not allowed", action),
						Resource: resourceName,
						Field:    fmt.Sprintf("%s.action", statementPath),
						Severity: "error",
					})
				}
			}
		}

		// Check for wildcard resources
		if !config.AllowWildcardResources {
			resources := v.normalizeResources(statement.Resource)
			for _, resource := range resources {
				if resource == "*" {
					errors = append(errors, ValidationError{
						Type:     "security_policy",
						Message:  "IAM policy contains wildcard resource '*' which is not allowed",
						Resource: resourceName,
						Field:    fmt.Sprintf("%s.resource", statementPath),
						Severity: "error",
					})
				}
			}
		}

		// Check for sensitive actions requiring MFA
		if config.RequireMFAForSensitiveActions && statement.Effect == "Allow" {
			for _, action := range actions {
				for _, sensitiveAction := range config.SensitiveActions {
					if matched, _ := regexp.MatchString(sensitiveAction, action); matched {
						// Check if MFA condition is present
						if !v.hasMFACondition(statement.Condition) {
							errors = append(errors, ValidationError{
								Type:     "security_policy",
								Message:  fmt.Sprintf("Sensitive action '%s' requires MFA condition", action),
								Resource: resourceName,
								Field:    fmt.Sprintf("%s.condition", statementPath),
								Severity: "error",
							})
						}
					}
				}
			}
		}
	}

	return errors
}

// normalizeActions converts action field to string slice
func (v *SecurityValidator) normalizeActions(action interface{}) []string {
	switch a := action.(type) {
	case string:
		return []string{a}
	case []string:
		return a
	case []interface{}:
		var actions []string
		for _, item := range a {
			if str, ok := item.(string); ok {
				actions = append(actions, str)
			}
		}
		return actions
	default:
		return []string{}
	}
}

// normalizeResources converts resource field to string slice
func (v *SecurityValidator) normalizeResources(resource interface{}) []string {
	switch r := resource.(type) {
	case string:
		return []string{r}
	case []string:
		return r
	case []interface{}:
		var resources []string
		for _, item := range r {
			if str, ok := item.(string); ok {
				resources = append(resources, str)
			}
		}
		return resources
	default:
		return []string{}
	}
}

// hasMFACondition checks if a condition map contains MFA requirement
func (v *SecurityValidator) hasMFACondition(condition map[string]interface{}) bool {
	if condition == nil {
		return false
	}

	// Check for common MFA condition patterns
	mfaPatterns := []string{
		"aws:MultiFactorAuthPresent",
		"aws:MultiFactorAuthAge",
	}

	for key := range condition {
		for _, pattern := range mfaPatterns {
			if strings.Contains(key, pattern) {
				return true
			}
		}
	}

	return false
}

// DefaultSecurityPolicies returns a set of default security policies
func DefaultSecurityPolicies() *SecurityPolicyConfig {
	return &SecurityPolicyConfig{
		IAMPolicies: &IAMPolicyValidation{
			ForbiddenActions: []string{
				"iam:CreateAccessKey",
				"iam:DeleteAccessKey",
				"sts:AssumeRole.*Root",
			},
			AllowWildcardResources: true,
			AllowAdminPermissions:  false,
			SensitiveActions: []string{
				"iam:.*",
				"sts:AssumeRole",
				"kms:.*",
			},
		},
		LambdaSecurity: &LambdaSecurityValidation{
			RequireVPC: false,
			ForbiddenEnvPatterns: []string{
				"(?i)(password|secret|key|token)",
			},
			MaxTimeout:    900, // 15 minutes
			MaxMemorySize: 3008,
			AllowedRuntimes: []string{
				"python3.11", "python3.10", "python3.9",
				"nodejs18.x", "nodejs16.x",
				"java17", "java11",
				"dotnet6",
			},
		},
		AgentSecurity: &AgentSecurityValidation{
			RequireGuardrails:          false,
			MaxIdleSessionTTL:          3600, // 1 hour
			RequireCustomerEncryption:  false,
			RequireMemoryConfiguration: false,
		},
		KnowledgeBaseSecurity: &KnowledgeBaseSecurityValidation{
			AllowedDataSourceTypes: []string{"S3", "Web", "Confluence", "SharePoint"},
			RequireAccessLogging:   false,
		},
	}
}

// EnterpriseSecurityPolicies returns a stricter set of security policies for enterprise use
func EnterpriseSecurityPolicies() *SecurityPolicyConfig {
	return &SecurityPolicyConfig{
		IAMPolicies: &IAMPolicyValidation{
			ForbiddenActions: []string{
				"iam:CreateAccessKey",
				"iam:DeleteAccessKey",
				"iam:CreateUser",
				"iam:DeleteUser",
				"sts:AssumeRole.*Root",
				".*:.*Admin.*",
				".*:.*Full.*",
			},
			AllowWildcardResources:        false,
			AllowAdminPermissions:         false,
			RequireMFAForSensitiveActions: true,
			SensitiveActions: []string{
				"iam:.*",
				"sts:AssumeRole",
				"kms:.*",
				"secretsmanager:.*",
				"bedrock:.*Agent.*",
			},
		},
		LambdaSecurity: &LambdaSecurityValidation{
			RequireVPC: true,
			ForbiddenEnvPatterns: []string{
				"(?i)(password|secret|key|token|api_key|auth)",
				"(?i)(prod|production).*(?i)(pass|secret)",
			},
			MaxTimeout:           300, // 5 minutes
			MaxMemorySize:        1024,
			RequireEnvEncryption: true,
			AllowedRuntimes: []string{
				"python3.11", "python3.10",
				"nodejs18.x",
				"java17",
			},
		},
		AgentSecurity: &AgentSecurityValidation{
			RequireGuardrails:          true,
			RequiredGuardrailTypes:     []string{"CONTENT", "SENSITIVE_INFORMATION"},
			MaxIdleSessionTTL:          1800, // 30 minutes
			RequireCustomerEncryption:  true,
			RequireMemoryConfiguration: true,
			ForbiddenModels: []string{
				"anthropic.claude-instant",
				"meta.llama2",
			},
		},
		KnowledgeBaseSecurity: &KnowledgeBaseSecurityValidation{
			RequireDataSourceEncryption: true,
			AllowedDataSourceTypes:      []string{"S3"},
			RequireVPCEndpoints:         true,
			MaxRetentionDays:            2555, // 7 years
			RequireAccessLogging:        true,
		},
		EncryptionRequirements: &EncryptionValidation{
			RequireEncryptionAtRest:    true,
			RequireEncryptionInTransit: true,
			RequireCustomerManagedKeys: true,
			AllowedKMSKeyPatterns: []string{
				"arn:aws:kms:.*:.*:key/.*",
			},
		},
		NetworkSecurity: &NetworkSecurityValidation{
			RequirePrivateSubnets: true,
			RequireVPCFlowLogs:    true,
			ForbiddenPorts: []string{
				"22", "3389", "1433", "3306", "5432",
			},
		},
	}
}
