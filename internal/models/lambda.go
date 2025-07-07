package models

type Lambda struct {
	Kind     ResourceKind `yaml:"kind"`
	Metadata Metadata     `yaml:"metadata"`
	Spec     LambdaSpec   `yaml:"spec"`
}

type LambdaSpec struct {
	Runtime             string            `yaml:"runtime"`
	Handler             string            `yaml:"handler"`
	Code                CodeConfiguration `yaml:"code"`
	Environment         map[string]string `yaml:"environment,omitempty"`
	Timeout             int               `yaml:"timeout,omitempty"`
	MemorySize          int               `yaml:"memorySize,omitempty"`
	ReservedConcurrency int               `yaml:"reservedConcurrency,omitempty"`
	Tags                map[string]string `yaml:"tags,omitempty"`
	VpcConfig           *VpcConfig        `yaml:"vpcConfig,omitempty"`
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
