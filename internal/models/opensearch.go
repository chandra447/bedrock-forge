package models

// OpenSearchServerless represents an OpenSearch Serverless collection with required security policies
type OpenSearchServerless struct {
	Kind     ResourceKind             `yaml:"kind"`
	Metadata Metadata                 `yaml:"metadata"`
	Spec     OpenSearchServerlessSpec `yaml:"spec"`
}

type OpenSearchServerlessSpec struct {
	// Collection configuration
	CollectionName string `yaml:"collectionName"`
	Description    string `yaml:"description,omitempty"`
	Type           string `yaml:"type,omitempty"` // Default: "VECTORSEARCH"

	// Security policies
	EncryptionPolicy *EncryptionPolicy `yaml:"encryptionPolicy,omitempty"`
	NetworkPolicy    *NetworkPolicy    `yaml:"networkPolicy,omitempty"`
	AccessPolicy     *AccessPolicy     `yaml:"accessPolicy,omitempty"`

	// Vector index configuration for Bedrock
	VectorIndex *VectorIndexConfig `yaml:"vectorIndex,omitempty"`

	// Tags
	Tags map[string]string `yaml:"tags,omitempty"`
}

type EncryptionPolicy struct {
	Name        string `yaml:"name,omitempty"`
	Description string `yaml:"description,omitempty"`
	Type        string `yaml:"type,omitempty"`     // Default: "encryption"
	KmsKeyId    string `yaml:"kmsKeyId,omitempty"` // Optional, uses AWS managed key if not provided
}

type NetworkPolicy struct {
	Name        string          `yaml:"name,omitempty"`
	Description string          `yaml:"description,omitempty"`
	Type        string          `yaml:"type,omitempty"` // Default: "network"
	Access      []NetworkAccess `yaml:"access,omitempty"`
}

type NetworkAccess struct {
	SourceVPCEs []string `yaml:"sourceVPCEs,omitempty"`
	SourceType  string   `yaml:"sourceType,omitempty"` // Default: "public"
}

type AccessPolicy struct {
	Name        string `yaml:"name,omitempty"`
	Description string `yaml:"description,omitempty"`
	Type        string `yaml:"type,omitempty"` // Default: "data"

	// Principals that will have access (IAM roles/users)
	Principals []string `yaml:"principals,omitempty"`

	// Permissions
	Permissions []string `yaml:"permissions,omitempty"`

	// Auto-configure for Bedrock (adds necessary permissions)
	AutoConfigureForBedrock bool `yaml:"autoConfigureForBedrock,omitempty"`
}

type VectorIndexConfig struct {
	Name         string             `yaml:"name"`
	FieldMapping VectorFieldMapping `yaml:"fieldMapping"`
}

type VectorFieldMapping struct {
	VectorField   string `yaml:"vectorField"`   // Default: "vector"
	TextField     string `yaml:"textField"`     // Default: "text"
	MetadataField string `yaml:"metadataField"` // Default: "metadata"
}

// OpenSearchServerlessReference represents a reference to an existing OpenSearch Serverless collection
type OpenSearchServerlessReference struct {
	// For existing collections
	CollectionArn *string `yaml:"collectionArn,omitempty"`
	CollectionId  *string `yaml:"collectionId,omitempty"`

	// For auto-created collections (reference by name)
	CollectionName *string `yaml:"collectionName,omitempty"`

	// Vector index configuration
	VectorIndexName string       `yaml:"vectorIndexName"`
	FieldMapping    FieldMapping `yaml:"fieldMapping"`
}
