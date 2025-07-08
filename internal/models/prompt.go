package models

type Prompt struct {
	Kind     ResourceKind `yaml:"kind"`
	Metadata Metadata     `yaml:"metadata"`
	Spec     PromptSpec   `yaml:"spec"`
}

type PromptSpec struct {
	Description              string                `yaml:"description,omitempty"`
	DefaultVariant           string                `yaml:"defaultVariant,omitempty"`
	CustomerEncryptionKeyArn string                `yaml:"customerEncryptionKeyArn,omitempty"`
	InputVariables           []PromptInputVariable `yaml:"inputVariables,omitempty"`
	Variants                 []PromptVariant       `yaml:"variants"`
	Tags                     map[string]string     `yaml:"tags,omitempty"`
}

type PromptInputVariable struct {
	Name string `yaml:"name"`
}

type PromptVariant struct {
	Name                   string                  `yaml:"name"`
	ModelId                string                  `yaml:"modelId"`
	TemplateType           string                  `yaml:"templateType"` // "TEXT" or "CHAT"
	TemplateConfiguration  *TemplateConfiguration  `yaml:"templateConfiguration,omitempty"`
	InferenceConfiguration *InferenceConfiguration `yaml:"inferenceConfiguration,omitempty"`
	GenAiResource          *GenAiResourceConfig    `yaml:"genAiResource,omitempty"`
}

type GenAiResourceConfig struct {
	Agent *AgentResourceConfig `yaml:"agent,omitempty"`
}

type AgentResourceConfig struct {
	// Reference to an agent YAML config in the same project
	AgentName string `yaml:"agentName,omitempty"`

	// Direct ARN reference to an existing deployed agent
	AgentArn string `yaml:"agentArn,omitempty"`
}

type TemplateConfiguration struct {
	// For TEXT template type
	Text *TextTemplateConfiguration `yaml:"text,omitempty"`

	// For CHAT template type
	Chat *ChatTemplateConfiguration `yaml:"chat,omitempty"`
}

type TextTemplateConfiguration struct {
	Text           string                  `yaml:"text"`
	InputVariables []TemplateInputVariable `yaml:"inputVariables,omitempty"`
}

type ChatTemplateConfiguration struct {
	Messages          []ChatMessage           `yaml:"messages,omitempty"`
	System            []SystemMessage         `yaml:"system,omitempty"`
	ToolConfiguration *ToolConfiguration      `yaml:"toolConfiguration,omitempty"`
	InputVariables    []TemplateInputVariable `yaml:"inputVariables,omitempty"`
}

type TemplateInputVariable struct {
	Name string `yaml:"name"`
}

type ChatMessage struct {
	Role    string           `yaml:"role"` // "user", "assistant", "system"
	Content []MessageContent `yaml:"content"`
}

type MessageContent struct {
	Text string `yaml:"text,omitempty"`
}

type SystemMessage struct {
	Text string `yaml:"text"`
}

type ToolConfiguration struct {
	Tools      []Tool      `yaml:"tools,omitempty"`
	ToolChoice *ToolChoice `yaml:"toolChoice,omitempty"`
}

type Tool struct {
	ToolSpec *ToolSpec `yaml:"toolSpec,omitempty"`
}

type ToolSpec struct {
	Name        string           `yaml:"name"`
	Description string           `yaml:"description"`
	InputSchema *ToolInputSchema `yaml:"inputSchema"`
}

type ToolInputSchema struct {
	Json map[string]any `yaml:"json,omitempty"`
}

type ToolChoice struct {
	Auto *ToolChoiceAuto `yaml:"auto,omitempty"`
	Any  *ToolChoiceAny  `yaml:"any,omitempty"`
	Tool *ToolChoiceTool `yaml:"tool,omitempty"`
}

type ToolChoiceAuto struct{}

type ToolChoiceAny struct{}

type ToolChoiceTool struct {
	Name string `yaml:"name"`
}

type InferenceConfiguration struct {
	Text *TextInferenceConfiguration `yaml:"text,omitempty"`
}

type TextInferenceConfiguration struct {
	Temperature   *float64 `yaml:"temperature,omitempty"`
	TopP          *float64 `yaml:"topP,omitempty"`
	TopK          *int     `yaml:"topK,omitempty"`
	MaxTokens     *int     `yaml:"maxTokens,omitempty"`
	StopSequences []string `yaml:"stopSequences,omitempty"`
}
