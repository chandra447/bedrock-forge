package models

type ActionGroup struct {
	Kind     ResourceKind    `yaml:"kind"`
	Metadata Metadata        `yaml:"metadata"`
	Spec     ActionGroupSpec `yaml:"spec"`
}

type ActionGroupSpec struct {
	AgentId                    string               `yaml:"agentId,omitempty"`
	AgentName                  string               `yaml:"agentName,omitempty"`
	Description                string               `yaml:"description,omitempty"`
	ParentActionGroupSignature string               `yaml:"parentActionGroupSignature,omitempty"`
	ActionGroupExecutor        *ActionGroupExecutor `yaml:"actionGroupExecutor,omitempty"`
	ActionGroupState           string               `yaml:"actionGroupState,omitempty"`
	APISchema                  *APISchema           `yaml:"apiSchema,omitempty"`
	FunctionSchema             *FunctionSchema      `yaml:"functionSchema,omitempty"`
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
