package models

type Guardrail struct {
	Kind     ResourceKind  `yaml:"kind"`
	Metadata Metadata      `yaml:"metadata"`
	Spec     GuardrailSpec `yaml:"spec"`
}

type GuardrailSpec struct {
	Description                      string                            `yaml:"description,omitempty"`
	ContentPolicyConfig              *ContentPolicyConfig              `yaml:"contentPolicyConfig,omitempty"`
	SensitiveInformationPolicyConfig *SensitiveInformationPolicyConfig `yaml:"sensitiveInformationPolicyConfig,omitempty"`
	ContextualGroundingPolicyConfig  *ContextualGroundingPolicyConfig  `yaml:"contextualGroundingPolicyConfig,omitempty"`
	TopicPolicyConfig                *TopicPolicyConfig                `yaml:"topicPolicyConfig,omitempty"`
	WordPolicyConfig                 *WordPolicyConfig                 `yaml:"wordPolicyConfig,omitempty"`
	Tags                             map[string]string                 `yaml:"tags,omitempty"`
}

type ContentPolicyConfig struct {
	FiltersConfig []ContentFilter `yaml:"filtersConfig"`
}

type ContentFilter struct {
	Type           string `yaml:"type"`
	InputStrength  string `yaml:"inputStrength"`
	OutputStrength string `yaml:"outputStrength"`
}

type SensitiveInformationPolicyConfig struct {
	PiiEntitiesConfig []PiiEntity `yaml:"piiEntitiesConfig"`
}

type PiiEntity struct {
	Type   string `yaml:"type"`
	Action string `yaml:"action"`
}

type ContextualGroundingPolicyConfig struct {
	FiltersConfig []ContextualGroundingFilter `yaml:"filtersConfig"`
}

type ContextualGroundingFilter struct {
	Type      string  `yaml:"type"`
	Threshold float64 `yaml:"threshold"`
}

type TopicPolicyConfig struct {
	TopicsConfig []Topic `yaml:"topicsConfig"`
}

type Topic struct {
	Name       string   `yaml:"name"`
	Definition string   `yaml:"definition"`
	Examples   []string `yaml:"examples"`
	Type       string   `yaml:"type"`
}

type WordPolicyConfig struct {
	WordsConfig            []Word            `yaml:"wordsConfig,omitempty"`
	ManagedWordListsConfig []ManagedWordList `yaml:"managedWordListsConfig,omitempty"`
}

type Word struct {
	Text string `yaml:"text"`
}

type ManagedWordList struct {
	Type string `yaml:"type"`
}
