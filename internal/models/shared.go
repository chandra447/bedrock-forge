package models

type ResourceKind string

const (
	AgentKind                           ResourceKind = "Agent"
	LambdaKind                          ResourceKind = "Lambda"
	ActionGroupKind                     ResourceKind = "ActionGroup"
	KnowledgeBaseKind                   ResourceKind = "KnowledgeBase"
	GuardrailKind                       ResourceKind = "Guardrail"
	PromptKind                          ResourceKind = "Prompt"
	IAMRoleKind                         ResourceKind = "IAMRole"
	AgentKnowledgeBaseAssociationKind   ResourceKind = "AgentKnowledgeBaseAssociation"
	CustomModuleKind                    ResourceKind = "CustomModule"
)

type BaseResource struct {
	Kind       ResourceKind `yaml:"kind"`
	APIVersion string       `yaml:"apiVersion,omitempty"`
	Metadata   Metadata     `yaml:"metadata"`
	Spec       interface{}  `yaml:"spec"`
}

type Metadata struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}