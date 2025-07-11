package models

type Lambda struct {
	Kind     ResourceKind `yaml:"kind"`
	Metadata Metadata     `yaml:"metadata"`
	Spec     LambdaSpec   `yaml:"spec"`
}

type LambdaSpec struct {
	Runtime             string                `yaml:"runtime"`
	Handler             string                `yaml:"handler"`
	Code                CodeConfiguration     `yaml:"code"`
	Environment         map[string]string     `yaml:"environment,omitempty"`
	Timeout             int                   `yaml:"timeout,omitempty"`
	MemorySize          int                   `yaml:"memorySize,omitempty"`
	ReservedConcurrency int                   `yaml:"reservedConcurrency,omitempty"`
	Tags                map[string]string     `yaml:"tags,omitempty"`
	VpcConfig           *VpcConfig            `yaml:"vpcConfig,omitempty"`
	ResourcePolicy      *LambdaResourcePolicy `yaml:"resourcePolicy,omitempty"`

	// Missing critical Terraform attributes
	Role                           Reference         `yaml:"role,omitempty"`                 // Reference to IAM role or ARN
	RoleArn                        string            `yaml:"roleArn,omitempty"`              // Direct IAM role ARN
	Architectures                  []string          `yaml:"architectures,omitempty"`        // x86_64, arm64
	CodeSigningConfigArn           string            `yaml:"codeSigningConfigArn,omitempty"` // Code signing config ARN
	DeadLetterConfig               *DeadLetterConfig `yaml:"deadLetterConfig,omitempty"`     // DLQ configuration
	EphemeralStorage               *EphemeralStorage `yaml:"ephemeralStorage,omitempty"`     // /tmp storage size
	FileSystemConfig               *FileSystemConfig `yaml:"fileSystemConfig,omitempty"`     // EFS config
	ImageConfig                    *ImageConfig      `yaml:"imageConfig,omitempty"`          // Container image config
	KmsKeyArn                      string            `yaml:"kmsKeyArn,omitempty"`            // KMS key for encryption
	Layers                         []string          `yaml:"layers,omitempty"`               // Lambda layer ARNs
	PackageType                    string            `yaml:"packageType,omitempty"`          // Zip or Image
	Publish                        *bool             `yaml:"publish,omitempty"`              // Create version on update
	ReplaceSecurityGroupsOnDestroy *bool             `yaml:"replaceSecurityGroupsOnDestroy,omitempty"`
	ReplacementSecurityGroupIds    []string          `yaml:"replacementSecurityGroupIds,omitempty"`
	SkipDestroy                    *bool             `yaml:"skipDestroy,omitempty"`    // Skip destroy
	SnapStart                      *SnapStart        `yaml:"snapStart,omitempty"`      // SnapStart config
	SourceCodeHash                 string            `yaml:"sourceCodeHash,omitempty"` // Source code hash
	Timeouts                       *LambdaTimeouts   `yaml:"timeouts,omitempty"`       // Terraform timeouts
	TracingConfig                  *TracingConfig    `yaml:"tracingConfig,omitempty"`  // X-Ray tracing
}

type LambdaResourcePolicy struct {
	AllowBedrockAgents bool                       `yaml:"allowBedrockAgents,omitempty"`
	Statements         []LambdaResourcePolicyStmt `yaml:"statements,omitempty"`
}

type LambdaResourcePolicyStmt struct {
	Sid       string                 `yaml:"sid"`
	Effect    string                 `yaml:"effect"`
	Principal map[string]interface{} `yaml:"principal"`
	Action    interface{}            `yaml:"action"` // string or []string
	Resource  string                 `yaml:"resource,omitempty"`
	Condition map[string]interface{} `yaml:"condition,omitempty"`
}

type CodeConfiguration struct {
	Source          string `yaml:"source"`
	ZipFile         string `yaml:"zipFile,omitempty"`
	S3Bucket        string `yaml:"s3Bucket,omitempty"`
	S3Key           string `yaml:"s3Key,omitempty"`
	S3ObjectVersion string `yaml:"s3ObjectVersion,omitempty"`
}

type VpcConfig struct {
	SecurityGroupIds []string `yaml:"securityGroupIds"`
	SubnetIds        []string `yaml:"subnetIds"`
}

// New supporting types for additional Lambda attributes
type DeadLetterConfig struct {
	TargetArn string `yaml:"targetArn"`
}

type EphemeralStorage struct {
	Size int `yaml:"size"` // Size in MB between 512 and 10240
}

type FileSystemConfig struct {
	Arn            string `yaml:"arn"`
	LocalMountPath string `yaml:"localMountPath"`
}

type ImageConfig struct {
	Command          []string `yaml:"command,omitempty"`
	EntryPoint       []string `yaml:"entryPoint,omitempty"`
	WorkingDirectory string   `yaml:"workingDirectory,omitempty"`
}

type SnapStart struct {
	ApplyOn string `yaml:"applyOn"` // PublishedVersions or None
}

type LambdaTimeouts struct {
	Create string `yaml:"create,omitempty"`
	Update string `yaml:"update,omitempty"`
	Delete string `yaml:"delete,omitempty"`
}

type TracingConfig struct {
	Mode string `yaml:"mode"` // Active or PassThrough
}
