package models

type Agent struct {
	Kind     ResourceKind `yaml:"kind"`
	Metadata Metadata     `yaml:"metadata"`
	Spec     AgentSpec    `yaml:"spec"`
}

type AgentSpec struct {
	FoundationModel       string               `yaml:"foundationModel"`
	Instruction           string               `yaml:"instruction"`
	Description           string               `yaml:"description,omitempty"`
	IdleSessionTTL        int                  `yaml:"idleSessionTtl,omitempty"`
	CustomerEncryptionKey string               `yaml:"customerEncryptionKey,omitempty"`
	Tags                  map[string]string    `yaml:"tags,omitempty"`
	Guardrail             *GuardrailConfig     `yaml:"guardrail,omitempty"`
	ActionGroups          []InlineActionGroup  `yaml:"actionGroups,omitempty"`
	PromptOverrides       []PromptOverride     `yaml:"promptOverrides,omitempty"`
	MemoryConfiguration   *MemoryConfiguration `yaml:"memoryConfiguration,omitempty"`
	Aliases               []AgentAlias         `yaml:"aliases,omitempty"`
	
	// IAM Role configuration - allows users to specify existing roles or customize auto-generated ones
	IAMRole               *IAMRoleConfig       `yaml:"iamRole,omitempty"`
	
	// Missing Terraform attributes
	PrepareAgent              *bool             `yaml:"prepareAgent,omitempty"`              // Default: true
	SkipResourceInUseCheck    *bool             `yaml:"skipResourceInUseCheck,omitempty"`    // Default: false
	Timeouts                  *AgentTimeouts    `yaml:"timeouts,omitempty"`
}

type GuardrailConfig struct {
	Name    Reference `yaml:"name"`
	Version string    `yaml:"version,omitempty"`
	Mode    string    `yaml:"mode,omitempty"`
}

// InlineActionGroup represents an action group defined directly within an agent
type InlineActionGroup struct {
	Name                       string               `yaml:"name"`
	Description                string               `yaml:"description,omitempty"`
	ParentActionGroupSignature string               `yaml:"parentActionGroupSignature,omitempty"`
	ActionGroupExecutor        *ActionGroupExecutor `yaml:"actionGroupExecutor,omitempty"`
	ActionGroupState           string               `yaml:"actionGroupState,omitempty"`
	APISchema                  *APISchema           `yaml:"apiSchema,omitempty"`
	FunctionSchema             *FunctionSchema      `yaml:"functionSchema,omitempty"`
	SkipResourceInUseCheck     bool                 `yaml:"skipResourceInUseCheck,omitempty"`
}

type PromptOverride struct {
	PromptType    string    `yaml:"promptType"`
	PromptArn     string    `yaml:"promptArn,omitempty"` // External ARN
	Prompt        Reference `yaml:"prompt,omitempty"`    // Reference to Prompt resource
	PromptVariant string    `yaml:"promptVariant,omitempty"`
	Variant       string    `yaml:"variant,omitempty"`
}

type MemoryConfiguration struct {
	EnabledMemoryTypes []string `yaml:"enabledMemoryTypes"`
	StorageDays        int      `yaml:"storageDays,omitempty"`
}

type AgentAlias struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description,omitempty"`
	Tags        map[string]string `yaml:"tags,omitempty"`
}

// AgentTimeouts represents timeout configuration for agent operations
type AgentTimeouts struct {
	Create string `yaml:"create,omitempty"` // Default: 10m
	Update string `yaml:"update,omitempty"` // Default: 10m
	Delete string `yaml:"delete,omitempty"` // Default: 10m
}
