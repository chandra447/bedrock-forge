package models

type ActionGroup struct {
	Kind     ResourceKind    `yaml:"kind"`
	Metadata Metadata        `yaml:"metadata"`
	Spec     ActionGroupSpec `yaml:"spec"`
}

type ActionGroupSpec struct {
	AgentId                    string               `yaml:"agentId"`                // Required: agent_id is required per AWS docs
	AgentVersion               string               `yaml:"agentVersion,omitempty"` // Optional: agent version (defaults to DRAFT)
	Description                string               `yaml:"description,omitempty"`
	ParentActionGroupSignature string               `yaml:"parentActionGroupSignature,omitempty"`
	ActionGroupExecutor        *ActionGroupExecutor `yaml:"actionGroupExecutor"` // Required: action_group_executor is required
	ActionGroupState           string               `yaml:"actionGroupState,omitempty"`
	APISchema                  *APISchema           `yaml:"apiSchema,omitempty"`
	FunctionSchema             *FunctionSchema      `yaml:"functionSchema,omitempty"`
	SkipResourceInUseCheck     bool                 `yaml:"skipResourceInUseCheck,omitempty"` // Optional: skip_resource_in_use_check
	Tags                       map[string]string    `yaml:"tags,omitempty"`
}

type ActionGroupExecutor struct {
	Lambda        string `yaml:"lambda,omitempty"`
	LambdaArn     string `yaml:"lambdaArn,omitempty"`
	CustomControl string `yaml:"customControl,omitempty"`
}

type APISchema struct {
	S3      *S3APISchema `yaml:"s3,omitempty"`
	Payload string       `yaml:"payload,omitempty"`
}

type S3APISchema struct {
	S3BucketName string `yaml:"s3BucketName"`
	S3ObjectKey  string `yaml:"s3ObjectKey"`
}

type FunctionSchema struct {
	Functions []Function `yaml:"functions"`
}

type Function struct {
	Name        string               `yaml:"name"`
	Description string               `yaml:"description,omitempty"`
	Parameters  map[string]Parameter `yaml:"parameters,omitempty"`
}

type Parameter struct {
	Description string `yaml:"description,omitempty"`
	Required    bool   `yaml:"required,omitempty"`
	Type        string `yaml:"type,omitempty"`
}
