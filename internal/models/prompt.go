package models

type Prompt struct {
	Kind     ResourceKind `yaml:"kind"`
	Metadata Metadata     `yaml:"metadata"`
	Spec     PromptSpec   `yaml:"spec"`
}

type PromptSpec struct {
	Description     string          `yaml:"description,omitempty"`
	DefaultVariant  string          `yaml:"defaultVariant,omitempty"`
	Variants        []PromptVariant `yaml:"variants"`
	Tags            map[string]string `yaml:"tags,omitempty"`
}

type PromptVariant struct {
	Name                    string                   `yaml:"name"`
	ModelId                 string                   `yaml:"modelId"`
	TemplateType            string                   `yaml:"templateType"`
	TemplateConfiguration   *TemplateConfiguration   `yaml:"templateConfiguration,omitempty"`
	InferenceConfiguration  *InferenceConfiguration  `yaml:"inferenceConfiguration,omitempty"`
}

type TemplateConfiguration struct {
	Text string `yaml:"text"`
}

type InferenceConfiguration struct {
	Text *TextInferenceConfiguration `yaml:"text,omitempty"`
}

type TextInferenceConfiguration struct {
	Temperature   float64  `yaml:"temperature,omitempty"`
	TopP          float64  `yaml:"topP,omitempty"`
	MaxTokens     int      `yaml:"maxTokens,omitempty"`
	StopSequences []string `yaml:"stopSequences,omitempty"`
}