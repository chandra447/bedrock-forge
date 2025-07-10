package models

// CustomResources represents user-defined Terraform files to be included in the deployment
type CustomResources struct {
	Kind     ResourceKind        `yaml:"kind"`
	Metadata Metadata            `yaml:"metadata"`
	Spec     CustomResourcesSpec `yaml:"spec"`
}

// CustomResourcesSpec defines the specification for custom Terraform files
type CustomResourcesSpec struct {
	// Path to directory containing .tf files OR path to main.tf file
	Path string `yaml:"path"`

	// List of specific .tf files to include (alternative to Path)
	Files []string `yaml:"files,omitempty"`

	// Dependencies on other resources (for ordering)
	DependsOn []string `yaml:"dependsOn,omitempty"`

	// Description of what these resources provide
	Description string `yaml:"description,omitempty"`

	// Variables to pass to the Terraform configuration
	Variables map[string]interface{} `yaml:"variables,omitempty"`
}

// CustomModule represents a user-defined Terraform module to be included in the deployment
// Deprecated: Use CustomResources instead
type CustomModule struct {
	Kind     ResourceKind     `yaml:"kind"`
	Metadata Metadata         `yaml:"metadata"`
	Spec     CustomModuleSpec `yaml:"spec"`
}

// CustomModuleSpec defines the specification for custom Terraform modules
// Deprecated: Use CustomResourcesSpec instead
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
