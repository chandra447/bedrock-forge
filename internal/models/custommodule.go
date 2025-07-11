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
	DependsOn []Reference `yaml:"dependsOn,omitempty"` // References to other resources

	// Description of what these resources provide
	Description string `yaml:"description,omitempty"`

	// Variables to pass to the Terraform configuration
	Variables map[string]interface{} `yaml:"variables,omitempty"`
}
