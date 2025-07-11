package models

import (
	"fmt"
	"gopkg.in/yaml.v3"
)

type ResourceKind string

const (
	AgentKind                         ResourceKind = "Agent"
	LambdaKind                        ResourceKind = "Lambda"
	ActionGroupKind                   ResourceKind = "ActionGroup"
	KnowledgeBaseKind                 ResourceKind = "KnowledgeBase"
	GuardrailKind                     ResourceKind = "Guardrail"
	PromptKind                        ResourceKind = "Prompt"
	IAMRoleKind                       ResourceKind = "IAMRole"
	AgentKnowledgeBaseAssociationKind ResourceKind = "AgentKnowledgeBaseAssociation"
	CustomResourcesKind               ResourceKind = "CustomResources"
	OpenSearchServerlessKind          ResourceKind = "OpenSearchServerless"
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

// Reference represents a reference to another resource, supporting both:
// - Simple string reference: "resource-name"
// - Object reference: { ref: "resource-name" }
type Reference struct {
	Name string // The referenced resource name
}

// UnmarshalYAML implements custom YAML unmarshaling to support both syntaxes
func (r *Reference) UnmarshalYAML(node *yaml.Node) error {
	// Try to unmarshal as a simple string first
	var str string
	if err := node.Decode(&str); err == nil {
		r.Name = str
		return nil
	}

	// Try to unmarshal as an object with ref field
	var obj struct {
		Ref string `yaml:"ref"`
	}
	if err := node.Decode(&obj); err != nil {
		return fmt.Errorf("reference must be either a string or an object with 'ref' field")
	}

	if obj.Ref == "" {
		return fmt.Errorf("reference object must have non-empty 'ref' field")
	}

	r.Name = obj.Ref
	return nil
}

// MarshalYAML implements custom YAML marshaling to output as a string for simplicity
func (r Reference) MarshalYAML() (interface{}, error) {
	return r.Name, nil
}

// IsEmpty returns true if the reference is empty
func (r Reference) IsEmpty() bool {
	return r.Name == ""
}

// String returns the referenced resource name
func (r Reference) String() string {
	return r.Name
}
