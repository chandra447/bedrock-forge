package models

// AgentKnowledgeBaseAssociation represents an association between an agent and a knowledge base
type AgentKnowledgeBaseAssociation struct {
	Kind     ResourceKind                          `yaml:"kind"`
	Metadata Metadata                              `yaml:"metadata"`
	Spec     AgentKnowledgeBaseAssociationSpec     `yaml:"spec"`
}

// AgentKnowledgeBaseAssociationSpec defines the specification for agent-knowledge base associations
type AgentKnowledgeBaseAssociationSpec struct {
	AgentId         string `yaml:"agentId"`
	AgentName       string `yaml:"agentName,omitempty"`
	KnowledgeBaseId string `yaml:"knowledgeBaseId"`
	KnowledgeBaseName string `yaml:"knowledgeBaseName,omitempty"`
	Description     string `yaml:"description,omitempty"`
	State           string `yaml:"state,omitempty"`
}