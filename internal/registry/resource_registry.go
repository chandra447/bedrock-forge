package registry

import (
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"

	"bedrock-forge/internal/models"
	"bedrock-forge/internal/parser"
)

type ResourceRegistry struct {
	logger    *logrus.Logger
	resources map[models.ResourceKind]map[string]*parser.ParsedResource
	mutex     sync.RWMutex
}

func NewResourceRegistry(logger *logrus.Logger) *ResourceRegistry {
	return &ResourceRegistry{
		logger:    logger,
		resources: make(map[models.ResourceKind]map[string]*parser.ParsedResource),
		mutex:     sync.RWMutex{},
	}
}

func (r *ResourceRegistry) AddResource(resource *parser.ParsedResource) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.resources[resource.Kind] == nil {
		r.resources[resource.Kind] = make(map[string]*parser.ParsedResource)
	}

	name := resource.Metadata.Name
	if _, exists := r.resources[resource.Kind][name]; exists {
		return fmt.Errorf("resource %s of kind %s already exists", name, resource.Kind)
	}

	r.resources[resource.Kind][name] = resource

	r.logger.WithFields(logrus.Fields{
		"kind": resource.Kind,
		"name": name,
		"file": resource.FilePath,
	}).Debug("Added resource to registry")

	return nil
}

func (r *ResourceRegistry) GetResource(kind models.ResourceKind, name string) (*parser.ParsedResource, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if resources, exists := r.resources[kind]; exists {
		if resource, exists := resources[name]; exists {
			return resource, true
		}
	}

	return nil, false
}

func (r *ResourceRegistry) GetResourcesByKind(kind models.ResourceKind) map[string]*parser.ParsedResource {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	result := make(map[string]*parser.ParsedResource)
	if resources, exists := r.resources[kind]; exists {
		for name, resource := range resources {
			result[name] = resource
		}
	}

	return result
}

func (r *ResourceRegistry) GetAllResources() map[models.ResourceKind]map[string]*parser.ParsedResource {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	result := make(map[models.ResourceKind]map[string]*parser.ParsedResource)
	for kind, resources := range r.resources {
		result[kind] = make(map[string]*parser.ParsedResource)
		for name, resource := range resources {
			result[kind][name] = resource
		}
	}

	return result
}

func (r *ResourceRegistry) ListResourceNames(kind models.ResourceKind) []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var names []string
	if resources, exists := r.resources[kind]; exists {
		for name := range resources {
			names = append(names, name)
		}
	}

	return names
}

func (r *ResourceRegistry) GetResourceCount(kind models.ResourceKind) int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if resources, exists := r.resources[kind]; exists {
		return len(resources)
	}

	return 0
}

func (r *ResourceRegistry) GetTotalResourceCount() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	total := 0
	for _, resources := range r.resources {
		total += len(resources)
	}

	return total
}

func (r *ResourceRegistry) ValidateDependencies() []error {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var errors []error

	agents := r.resources[models.AgentKind]
	for _, agentResource := range agents {
		agent := agentResource.Resource.(*models.Agent)

		if agent.Spec.Guardrail != nil && !agent.Spec.Guardrail.Name.IsEmpty() {
			guardrailName := agent.Spec.Guardrail.Name.String()
			if _, exists := r.resources[models.GuardrailKind][guardrailName]; !exists {
				errors = append(errors, fmt.Errorf("agent %s references non-existent guardrail %s", agent.Metadata.Name, guardrailName))
			}
		}

		// Knowledge bases are now handled through separate association resources

		// Action groups are now inline definitions within the agent
		for _, ag := range agent.Spec.ActionGroups {
			// Validate Lambda references for action group executors
			if ag.ActionGroupExecutor != nil {
				if !ag.ActionGroupExecutor.Lambda.IsEmpty() {
					lambdaName := ag.ActionGroupExecutor.Lambda.String()
					if _, exists := r.resources[models.LambdaKind][lambdaName]; !exists {
						errors = append(errors, fmt.Errorf("agent %s action group %s references non-existent lambda %s", agent.Metadata.Name, ag.Name, lambdaName))
					}
				}
				// LambdaArn references are external and don't need validation
			}
		}

		for _, promptOverride := range agent.Spec.PromptOverrides {
			if !promptOverride.Prompt.IsEmpty() {
				promptName := promptOverride.Prompt.String()
				if _, exists := r.resources[models.PromptKind][promptName]; !exists {
					errors = append(errors, fmt.Errorf("agent %s references non-existent prompt %s", agent.Metadata.Name, promptName))
				}
			}
		}
	}

	actionGroups := r.resources[models.ActionGroupKind]
	for _, agResource := range actionGroups {
		actionGroup := agResource.Resource.(*models.ActionGroup)

		if actionGroup.Spec.ActionGroupExecutor != nil {
			// If lambdaArn is specified, no dependency validation needed (external Lambda)
			if actionGroup.Spec.ActionGroupExecutor.LambdaArn != "" {
				r.logger.WithFields(logrus.Fields{
					"action_group": actionGroup.Metadata.Name,
					"lambda_arn":   actionGroup.Spec.ActionGroupExecutor.LambdaArn,
				}).Debug("Action group uses external Lambda ARN, skipping dependency validation")
				continue
			}

			// If lambda name is specified, validate it exists in the registry
			if !actionGroup.Spec.ActionGroupExecutor.Lambda.IsEmpty() {
				lambdaName := actionGroup.Spec.ActionGroupExecutor.Lambda.String()
				if _, exists := r.resources[models.LambdaKind][lambdaName]; !exists {
					errors = append(errors, fmt.Errorf("action group %s references non-existent lambda %s", actionGroup.Metadata.Name, lambdaName))
				}
			}
		}
	}


	return errors
}

func (r *ResourceRegistry) Clear() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.resources = make(map[models.ResourceKind]map[string]*parser.ParsedResource)
	r.logger.Debug("Cleared resource registry")
}

// HasResource checks if a resource exists in the registry
func (r *ResourceRegistry) HasResource(kind models.ResourceKind, name string) bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if resources, exists := r.resources[kind]; exists {
		_, exists := resources[name]
		return exists
	}
	return false
}

// GetResourcesByType returns all resources of a specific type
func (r *ResourceRegistry) GetResourcesByType(kind models.ResourceKind) []models.BaseResource {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var result []models.BaseResource
	if resources, exists := r.resources[kind]; exists {
		for _, resource := range resources {
			// Extract spec based on resource type
			var spec interface{}
			switch kind {
			case models.AgentKind:
				if agent, ok := resource.Resource.(*models.Agent); ok {
					spec = agent.Spec
				}
			case models.LambdaKind:
				if lambda, ok := resource.Resource.(*models.Lambda); ok {
					spec = lambda.Spec
				}
			case models.ActionGroupKind:
				if actionGroup, ok := resource.Resource.(*models.ActionGroup); ok {
					spec = actionGroup.Spec
				}
			case models.KnowledgeBaseKind:
				if knowledgeBase, ok := resource.Resource.(*models.KnowledgeBase); ok {
					spec = knowledgeBase.Spec
				}
			case models.GuardrailKind:
				if guardrail, ok := resource.Resource.(*models.Guardrail); ok {
					spec = guardrail.Spec
				}
			case models.PromptKind:
				if prompt, ok := resource.Resource.(*models.Prompt); ok {
					spec = prompt.Spec
				}
			case models.IAMRoleKind:
				if iamRole, ok := resource.Resource.(*models.IAMRole); ok {
					spec = iamRole.Spec
				}
		case models.OpenSearchServerlessKind:
			if opensearchServerless, ok := resource.Resource.(*models.OpenSearchServerless); ok {
				spec = opensearchServerless.Spec
			}
		case models.CustomResourcesKind:
			if customResources, ok := resource.Resource.(*models.CustomResources); ok {
				spec = customResources.Spec
			}
		case models.AgentKnowledgeBaseAssociationKind:
			if association, ok := resource.Resource.(*models.AgentKnowledgeBaseAssociation); ok {
				spec = association.Spec
			}
		}

			result = append(result, models.BaseResource{
				Kind:     resource.Kind,
				Metadata: resource.Metadata,
				Spec:     spec,
			})
		}
	}
	return result
}
