package models

// CustomModule represents a user-defined Terraform module to be included in the deployment
type CustomModule struct {
	Kind     ResourceKind     `yaml:"kind"`
	Metadata Metadata         `yaml:"metadata"`
	Spec     CustomModuleSpec `yaml:"spec"`
}

// CustomModuleSpec defines the specification for custom Terraform modules
type CustomModuleSpec struct {
	// Source of the Terraform module (can be local path, git repo, registry, etc.)
	Source string `yaml:"source"`

	// Version of the module (for registry modules or git tags)
	Version string `yaml:"version,omitempty"`

	// Input variables to pass to the module
	Variables map[string]interface{} `yaml:"variables,omitempty"`

	// Dependencies on other resources (for ordering)
	DependsOn []string `yaml:"dependsOn,omitempty"`

	// Description of what this module does
	Description string `yaml:"description,omitempty"`

	// Tags to apply to resources created by this module (if supported)
	Tags map[string]string `yaml:"tags,omitempty"`
}
