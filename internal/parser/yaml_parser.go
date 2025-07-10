package parser

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"bedrock-forge/internal/models"
)

type YAMLParser struct {
	logger *logrus.Logger
}

func NewYAMLParser(logger *logrus.Logger) *YAMLParser {
	return &YAMLParser{
		logger: logger,
	}
}

type ParsedResource struct {
	Kind       models.ResourceKind
	Metadata   models.Metadata
	Resource   interface{}
	FilePath   string
	RawContent []byte
}

func (p *YAMLParser) ParseFile(filePath string) ([]*ParsedResource, error) {
	p.logger.WithField("file", filePath).Debug("Parsing YAML file")

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	return p.ParseContent(content, filePath)
}

func (p *YAMLParser) ParseContent(content []byte, filePath string) ([]*ParsedResource, error) {
	resources := make([]*ParsedResource, 0)

	documents := strings.Split(string(content), "---")
	for i, doc := range documents {
		doc = strings.TrimSpace(doc)
		if doc == "" {
			continue
		}

		resource, err := p.parseDocument([]byte(doc), filePath, i)
		if err != nil {
			p.logger.WithError(err).WithFields(logrus.Fields{
				"file":     filePath,
				"document": i,
			}).Warn("Failed to parse document")
			continue
		}

		if resource != nil {
			resources = append(resources, resource)
		}
	}

	p.logger.WithFields(logrus.Fields{
		"file":  filePath,
		"count": len(resources),
	}).Debug("Parsed resources from file")

	return resources, nil
}

func (p *YAMLParser) parseDocument(content []byte, filePath string, docIndex int) (*ParsedResource, error) {
	var base models.BaseResource
	if err := yaml.Unmarshal(content, &base); err != nil {
		return nil, fmt.Errorf("failed to unmarshal base resource: %w", err)
	}

	if base.Kind == "" {
		return nil, fmt.Errorf("resource kind is required")
	}

	parsedResource := &ParsedResource{
		Kind:       base.Kind,
		Metadata:   base.Metadata,
		FilePath:   filePath,
		RawContent: content,
	}

	switch base.Kind {
	case models.AgentKind:
		var agent models.Agent
		if err := yaml.Unmarshal(content, &agent); err != nil {
			return nil, fmt.Errorf("failed to unmarshal Agent: %w", err)
		}
		parsedResource.Resource = &agent

	case models.LambdaKind:
		var lambda models.Lambda
		if err := yaml.Unmarshal(content, &lambda); err != nil {
			return nil, fmt.Errorf("failed to unmarshal Lambda: %w", err)
		}
		parsedResource.Resource = &lambda

	case models.ActionGroupKind:
		var actionGroup models.ActionGroup
		if err := yaml.Unmarshal(content, &actionGroup); err != nil {
			return nil, fmt.Errorf("failed to unmarshal ActionGroup: %w", err)
		}
		parsedResource.Resource = &actionGroup

	case models.KnowledgeBaseKind:
		var knowledgeBase models.KnowledgeBase
		if err := yaml.Unmarshal(content, &knowledgeBase); err != nil {
			return nil, fmt.Errorf("failed to unmarshal KnowledgeBase: %w", err)
		}
		parsedResource.Resource = &knowledgeBase

	case models.GuardrailKind:
		var guardrail models.Guardrail
		if err := yaml.Unmarshal(content, &guardrail); err != nil {
			return nil, fmt.Errorf("failed to unmarshal Guardrail: %w", err)
		}
		parsedResource.Resource = &guardrail

	case models.PromptKind:
		var prompt models.Prompt
		if err := yaml.Unmarshal(content, &prompt); err != nil {
			return nil, fmt.Errorf("failed to unmarshal Prompt: %w", err)
		}
		parsedResource.Resource = &prompt

	case models.IAMRoleKind:
		var iamRole models.IAMRole
		if err := yaml.Unmarshal(content, &iamRole); err != nil {
			return nil, fmt.Errorf("failed to unmarshal IAMRole: %w", err)
		}
		parsedResource.Resource = &iamRole

	case models.CustomResourcesKind:
		var customResources models.CustomResources
		if err := yaml.Unmarshal(content, &customResources); err != nil {
			return nil, fmt.Errorf("failed to unmarshal CustomResources: %w", err)
		}
		parsedResource.Resource = &customResources

	case models.CustomModuleKind:
		var customModule models.CustomModule
		if err := yaml.Unmarshal(content, &customModule); err != nil {
			return nil, fmt.Errorf("failed to unmarshal CustomModule: %w", err)
		}
		parsedResource.Resource = &customModule

	case models.OpenSearchServerlessKind:
		var opensearchServerless models.OpenSearchServerless
		if err := yaml.Unmarshal(content, &opensearchServerless); err != nil {
			return nil, fmt.Errorf("failed to unmarshal OpenSearchServerless: %w", err)
		}
		parsedResource.Resource = &opensearchServerless

	default:
		return nil, fmt.Errorf("unsupported resource kind: %s", base.Kind)
	}

	return parsedResource, nil
}

func (p *YAMLParser) ValidateResource(resource *ParsedResource) error {
	if resource.Kind == "" {
		return fmt.Errorf("resource kind is required")
	}

	if resource.Metadata.Name == "" {
		return fmt.Errorf("resource metadata.name is required")
	}

	switch resource.Kind {
	case models.AgentKind:
		return p.validateAgent(resource.Resource.(*models.Agent))
	case models.LambdaKind:
		return p.validateLambda(resource.Resource.(*models.Lambda))
	case models.ActionGroupKind:
		return p.validateActionGroup(resource.Resource.(*models.ActionGroup))
	case models.KnowledgeBaseKind:
		return p.validateKnowledgeBase(resource.Resource.(*models.KnowledgeBase))
	case models.GuardrailKind:
		return p.validateGuardrail(resource.Resource.(*models.Guardrail))
	case models.PromptKind:
		return p.validatePrompt(resource.Resource.(*models.Prompt))
	case models.IAMRoleKind:
		return p.validateIAMRole(resource.Resource.(*models.IAMRole))
	case models.CustomResourcesKind:
		return p.validateCustomResources(resource.Resource.(*models.CustomResources))
	case models.CustomModuleKind:
		return p.validateCustomModule(resource.Resource.(*models.CustomModule))
	case models.OpenSearchServerlessKind:
		return p.validateOpenSearchServerless(resource.Resource.(*models.OpenSearchServerless))
	}

	return nil
}

func (p *YAMLParser) validateAgent(agent *models.Agent) error {
	if agent.Spec.FoundationModel == "" {
		return fmt.Errorf("agent foundationModel is required")
	}
	if agent.Spec.Instruction == "" {
		return fmt.Errorf("agent instruction is required")
	}
	return nil
}

func (p *YAMLParser) validateLambda(lambda *models.Lambda) error {
	if lambda.Spec.Runtime == "" {
		return fmt.Errorf("lambda runtime is required")
	}
	if lambda.Spec.Handler == "" {
		return fmt.Errorf("lambda handler is required")
	}
	if lambda.Spec.Code.Source == "" {
		return fmt.Errorf("lambda code.source is required")
	}
	return nil
}

func (p *YAMLParser) validateActionGroup(actionGroup *models.ActionGroup) error {
	if actionGroup.Spec.ActionGroupExecutor == nil {
		return fmt.Errorf("actionGroup executor is required")
	}
	return nil
}

func (p *YAMLParser) validateKnowledgeBase(kb *models.KnowledgeBase) error {
	if kb.Spec.KnowledgeBaseConfiguration == nil {
		return fmt.Errorf("knowledgeBase configuration is required")
	}
	if kb.Spec.StorageConfiguration == nil {
		return fmt.Errorf("knowledgeBase storage configuration is required")
	}
	return nil
}

func (p *YAMLParser) validateGuardrail(guardrail *models.Guardrail) error {
	hasPolicy := guardrail.Spec.ContentPolicyConfig != nil ||
		guardrail.Spec.SensitiveInformationPolicyConfig != nil ||
		guardrail.Spec.ContextualGroundingPolicyConfig != nil ||
		guardrail.Spec.TopicPolicyConfig != nil ||
		guardrail.Spec.WordPolicyConfig != nil

	if !hasPolicy {
		return fmt.Errorf("guardrail must have at least one policy configuration")
	}
	return nil
}

func (p *YAMLParser) validatePrompt(prompt *models.Prompt) error {
	if len(prompt.Spec.Variants) == 0 {
		return fmt.Errorf("prompt must have at least one variant")
	}
	for _, variant := range prompt.Spec.Variants {
		if variant.Name == "" {
			return fmt.Errorf("prompt variant name is required")
		}
		if variant.ModelId == "" {
			return fmt.Errorf("prompt variant modelId is required")
		}
	}
	return nil
}

func (p *YAMLParser) validateIAMRole(iamRole *models.IAMRole) error {
	if iamRole.Spec.AssumeRolePolicy == nil {
		return fmt.Errorf("IAM role assumeRolePolicy is required")
	}
	if iamRole.Spec.AssumeRolePolicy.Version == "" {
		return fmt.Errorf("IAM role assumeRolePolicy version is required")
	}
	if len(iamRole.Spec.AssumeRolePolicy.Statement) == 0 {
		return fmt.Errorf("IAM role assumeRolePolicy must have at least one statement")
	}
	return nil
}

func (p *YAMLParser) validateCustomResources(customResources *models.CustomResources) error {
	if customResources.Spec.Path == "" && len(customResources.Spec.Files) == 0 {
		return fmt.Errorf("custom resources must specify either 'path' or 'files'")
	}

	if customResources.Spec.Path != "" && len(customResources.Spec.Files) > 0 {
		return fmt.Errorf("custom resources cannot specify both 'path' and 'files' - use one or the other")
	}

	return nil
}

func (p *YAMLParser) validateCustomModule(customModule *models.CustomModule) error {
	if customModule.Spec.Source == "" {
		return fmt.Errorf("custom module source is required")
	}
	return nil
}

func (p *YAMLParser) validateOpenSearchServerless(opensearchServerless *models.OpenSearchServerless) error {
	if opensearchServerless.Spec.CollectionName == "" {
		return fmt.Errorf("OpenSearch Serverless collectionName is required")
	}
	return nil
}
