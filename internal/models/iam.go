package models

// IAMRoleConfig provides flexible IAM role management for agents
type IAMRoleConfig struct {
	// For auto-generated roles (default: true)
	AutoCreate *bool `yaml:"autoCreate,omitempty"`

	// For existing role ARN
	RoleArn string `yaml:"roleArn,omitempty"`

	// For referencing a manually defined IAMRole resource
	RoleName Reference `yaml:"roleName,omitempty"` // Reference to IAMRole resource

	// Additional policies to attach to auto-generated roles
	AdditionalPolicies []IAMPolicyReference `yaml:"additionalPolicies,omitempty"`
}

type IAMRole struct {
	Kind     ResourceKind `yaml:"kind"`
	Metadata Metadata     `yaml:"metadata"`
	Spec     IAMRoleSpec  `yaml:"spec"`
}

type IAMRoleSpec struct {
	Description      string                `yaml:"description,omitempty"`
	AssumeRolePolicy *AssumeRolePolicy     `yaml:"assumeRolePolicy"`
	Policies         []IAMPolicyAttachment `yaml:"policies,omitempty"`
	InlinePolicies   []IAMInlinePolicy     `yaml:"inlinePolicies,omitempty"`
	Tags             map[string]string     `yaml:"tags,omitempty"`
}

type AssumeRolePolicy struct {
	Version   string                      `yaml:"version"`
	Statement []AssumeRolePolicyStatement `yaml:"statement"`
}

type AssumeRolePolicyStatement struct {
	Effect    string                 `yaml:"effect"`
	Principal map[string]interface{} `yaml:"principal"`
	Action    interface{}            `yaml:"action"`
	Condition map[string]interface{} `yaml:"condition,omitempty"`
}

type IAMPolicyAttachment struct {
	PolicyArn  string    `yaml:"policyArn"`            // External AWS policy ARN
	PolicyName Reference `yaml:"policyName,omitempty"` // Reference to policy resource
}

type IAMInlinePolicy struct {
	Name   string            `yaml:"name"`
	Policy IAMPolicyDocument `yaml:"policy"`
}

type IAMPolicyDocument struct {
	Version   string               `yaml:"version"`
	Statement []IAMPolicyStatement `yaml:"statement"`
}

type IAMPolicyStatement struct {
	Sid       string                 `yaml:"sid,omitempty"`
	Effect    string                 `yaml:"effect"`
	Action    interface{}            `yaml:"action"`
	Resource  interface{}            `yaml:"resource"`
	Condition map[string]interface{} `yaml:"condition,omitempty"`
}

type IAMPolicyReference struct {
	PolicyArn  string    `yaml:"policyArn,omitempty"`  // External AWS policy ARN
	PolicyName Reference `yaml:"policyName,omitempty"` // Reference to policy resource
}
