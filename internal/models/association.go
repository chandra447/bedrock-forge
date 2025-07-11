package models

// AgentKnowledgeBaseAssociation represents an association between an agent and a knowledge base
type AgentKnowledgeBaseAssociation struct {
	Kind     ResourceKind                      `yaml:"kind"`
	Metadata Metadata                          `yaml:"metadata"`
	Spec     AgentKnowledgeBaseAssociationSpec `yaml:"spec"`
}

// AgentKnowledgeBaseAssociationSpec defines the specification for agent-knowledge base associations
type AgentKnowledgeBaseAssociationSpec struct {
	AgentId           Reference `yaml:"agentId"`                     // Reference to Agent resource
	AgentName         Reference `yaml:"agentName,omitempty"`         // Reference to Agent resource
	KnowledgeBaseId   Reference `yaml:"knowledgeBaseId"`             // Reference to KnowledgeBase resource
	KnowledgeBaseName Reference `yaml:"knowledgeBaseName,omitempty"` // Reference to KnowledgeBase resource
	Description       string    `yaml:"description,omitempty"`
	State             string    `yaml:"state,omitempty"`
}
