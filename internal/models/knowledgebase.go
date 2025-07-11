package models

type KnowledgeBase struct {
	Kind     ResourceKind      `yaml:"kind"`
	Metadata Metadata          `yaml:"metadata"`
	Spec     KnowledgeBaseSpec `yaml:"spec"`
}

type KnowledgeBaseSpec struct {
	Description                string                      `yaml:"description,omitempty"`
	KnowledgeBaseConfiguration *KnowledgeBaseConfiguration `yaml:"knowledgeBaseConfiguration,omitempty"`
	StorageConfiguration       *StorageConfiguration       `yaml:"storageConfiguration,omitempty"`
	DataSources                []DataSource                `yaml:"dataSources,omitempty"`
	Tags                       map[string]string           `yaml:"tags,omitempty"`
}

type KnowledgeBaseConfiguration struct {
	Type                             string                            `yaml:"type"`
	VectorKnowledgeBaseConfiguration *VectorKnowledgeBaseConfiguration `yaml:"vectorKnowledgeBaseConfiguration,omitempty"`
}

type VectorKnowledgeBaseConfiguration struct {
	EmbeddingModelArn           string                       `yaml:"embeddingModelArn"`
	EmbeddingModelConfiguration *EmbeddingModelConfiguration `yaml:"embeddingModelConfiguration,omitempty"`
}

type EmbeddingModelConfiguration struct {
	BedrockEmbeddingModelConfiguration *BedrockEmbeddingModelConfiguration `yaml:"bedrockEmbeddingModelConfiguration,omitempty"`
}

type BedrockEmbeddingModelConfiguration struct {
	Dimensions int `yaml:"dimensions,omitempty"`
}

type StorageConfiguration struct {
	Type                              string                             `yaml:"type"`
	OpensearchServerlessConfiguration *OpensearchServerlessConfiguration `yaml:"opensearchServerlessConfiguration,omitempty"`

	// Enhanced OpenSearch Serverless configuration with auto-creation support
	OpenSearchServerless *OpenSearchServerlessReference `yaml:"openSearchServerless,omitempty"`
}

type OpensearchServerlessConfiguration struct {
	CollectionArn   string       `yaml:"collectionArn"`
	VectorIndexName string       `yaml:"vectorIndexName"`
	FieldMapping    FieldMapping `yaml:"fieldMapping"`
}

type FieldMapping struct {
	VectorField   string `yaml:"vectorField"`
	TextField     string `yaml:"textField"`
	MetadataField string `yaml:"metadataField"`
}

type DataSource struct {
	Name                         string                        `yaml:"name"`
	Type                         string                        `yaml:"type"`
	S3Configuration              *S3Configuration              `yaml:"s3Configuration,omitempty"`
	ChunkingConfiguration        *ChunkingConfiguration        `yaml:"chunkingConfiguration,omitempty"`
	VectorIngestionConfiguration *VectorIngestionConfiguration `yaml:"vectorIngestionConfiguration,omitempty"`
	CustomTransformation         *CustomTransformation         `yaml:"customTransformation,omitempty"`
}

type S3Configuration struct {
	BucketArn         string   `yaml:"bucketArn"`
	InclusionPrefixes []string `yaml:"inclusionPrefixes,omitempty"`
	ExclusionPrefixes []string `yaml:"exclusionPrefixes,omitempty"`
}

type ChunkingConfiguration struct {
	ChunkingStrategy               string                          `yaml:"chunkingStrategy"`
	FixedSizeChunkingConfiguration *FixedSizeChunkingConfiguration `yaml:"fixedSizeChunkingConfiguration,omitempty"`
	SemanticChunkingConfiguration  *SemanticChunkingConfiguration  `yaml:"semanticChunkingConfiguration,omitempty"`
}

type FixedSizeChunkingConfiguration struct {
	MaxTokens         int `yaml:"maxTokens"`
	OverlapPercentage int `yaml:"overlapPercentage"`
}

type SemanticChunkingConfiguration struct {
	MaxTokens                     int `yaml:"maxTokens"`
	BufferSize                    int `yaml:"bufferSize"`
	BreakpointPercentileThreshold int `yaml:"breakpointPercentileThreshold"`
}

type VectorIngestionConfiguration struct {
	ChunkingConfiguration *ChunkingConfiguration `yaml:"chunkingConfiguration,omitempty"`
}

type CustomTransformation struct {
	TransformationLambda *TransformationLambda `yaml:"transformationLambda,omitempty"`
	IntermediateStorage  *IntermediateStorage  `yaml:"intermediateStorage,omitempty"`
}

type TransformationLambda struct {
	LambdaArn string    `yaml:"lambdaArn,omitempty"` // External Lambda ARN
	Lambda    Reference `yaml:"lambda,omitempty"`    // Reference to Lambda resource
}

type IntermediateStorage struct {
	S3Location *S3Location `yaml:"s3Location,omitempty"`
}

type S3Location struct {
	URI string `yaml:"uri"`
}
