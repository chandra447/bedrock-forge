package generator

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"

	"bedrock-forge/internal/models"
)

// generateActionGroupModule creates a module call for an ActionGroup resource
func (g *HCLGenerator) generateActionGroupModule(body *hclwrite.Body, resource models.BaseResource) error {
	actionGroup, ok := resource.Spec.(models.ActionGroupSpec)
	if !ok {
		// Try to parse as map and convert to ActionGroupSpec
		specMap, mapOk := resource.Spec.(map[string]interface{})
		if !mapOk {
			return fmt.Errorf("invalid action group spec format")
		}

		// Convert map to ActionGroupSpec
		specJSON, err := json.Marshal(specMap)
		if err != nil {
			return fmt.Errorf("failed to marshal action group spec: %w", err)
		}

		if err := json.Unmarshal(specJSON, &actionGroup); err != nil {
			return fmt.Errorf("failed to unmarshal action group spec: %w", err)
		}
	}

	resourceName := g.sanitizeResourceName(resource.Metadata.Name)

	// Create module block
	moduleBlock := body.AppendNewBlock("module", []string{resourceName})
	moduleBody := moduleBlock.Body()

	// Set module source
	moduleSource := fmt.Sprintf("%s//modules/bedrock-action-group", g.config.ModuleRegistry)
	if g.config.ModuleVersion != "" {
		moduleSource += fmt.Sprintf("?ref=%s", g.config.ModuleVersion)
	}
	moduleBody.SetAttributeValue("source", cty.StringVal(moduleSource))

	// Set basic attributes
	moduleBody.SetAttributeValue("action_group_name", cty.StringVal(resource.Metadata.Name))

	// Set required agent_id
	if actionGroup.AgentId.IsEmpty() {
		return fmt.Errorf("action group %s must specify agentId", resource.Metadata.Name)
	}

	// Resolve agent reference to get agent ID
	if agentId, err := g.resolveReferenceToOutput(actionGroup.AgentId, models.AgentKind, "agent_id"); err == nil {
		moduleBody.SetAttributeValue("agent_id", cty.StringVal(agentId))
	} else {
		// Fallback to direct string value for backward compatibility
		moduleBody.SetAttributeValue("agent_id", cty.StringVal(actionGroup.AgentId.String()))
		g.logger.WithError(err).WithField("agent", actionGroup.AgentId.String()).Warn("Failed to resolve agent reference, using direct value")
	}

	// Set agent_version (defaults to DRAFT if not specified)
	agentVersion := actionGroup.AgentVersion
	if agentVersion == "" {
		agentVersion = "DRAFT"
	}
	moduleBody.SetAttributeValue("agent_version", cty.StringVal(agentVersion))

	// Optional description
	if actionGroup.Description != "" {
		moduleBody.SetAttributeValue("description", cty.StringVal(actionGroup.Description))
	}

	// Parent action group signature
	if actionGroup.ParentActionGroupSignature != "" {
		moduleBody.SetAttributeValue("parent_action_group_signature", cty.StringVal(actionGroup.ParentActionGroupSignature))
	}

	// Action group state
	if actionGroup.ActionGroupState != "" {
		moduleBody.SetAttributeValue("action_group_state", cty.StringVal(actionGroup.ActionGroupState))
	}

	// Skip resource in use check
	if actionGroup.SkipResourceInUseCheck {
		moduleBody.SetAttributeValue("skip_resource_in_use_check", cty.BoolVal(true))
	}

	// Action group executor (required)
	if actionGroup.ActionGroupExecutor == nil {
		return fmt.Errorf("action group %s must specify actionGroupExecutor", resource.Metadata.Name)
	}
	{
		executorValues := make(map[string]cty.Value)

		// Handle Lambda reference (either local resource or existing ARN)
		if actionGroup.ActionGroupExecutor.LambdaArn != "" {
			// Direct ARN reference to existing Lambda function
			executorValues["lambda"] = cty.StringVal(actionGroup.ActionGroupExecutor.LambdaArn)
			g.logger.WithFields(logrus.Fields{
				"action_group": resource.Metadata.Name,
				"lambda_arn":   actionGroup.ActionGroupExecutor.LambdaArn,
			}).Debug("Using existing Lambda ARN for action group executor")
		} else if !actionGroup.ActionGroupExecutor.Lambda.IsEmpty() {
			// Reference to a Lambda module defined in the same project
			if lambdaArn, err := g.resolveReferenceToOutput(actionGroup.ActionGroupExecutor.Lambda, models.LambdaKind, "lambda_function_arn"); err == nil {
				executorValues["lambda"] = cty.StringVal(lambdaArn)
				g.logger.WithFields(logrus.Fields{
					"action_group":  resource.Metadata.Name,
					"lambda_module": actionGroup.ActionGroupExecutor.Lambda.String(),
				}).Debug("Using Lambda module reference for action group executor")
			} else {
				// Treat as direct ARN reference for backward compatibility
				executorValues["lambda"] = cty.StringVal(actionGroup.ActionGroupExecutor.Lambda.String())
				g.logger.WithFields(logrus.Fields{
					"action_group": resource.Metadata.Name,
					"lambda_value": actionGroup.ActionGroupExecutor.Lambda.String(),
				}).Debug("Using direct Lambda value for action group executor")
			}
		}

		if actionGroup.ActionGroupExecutor.CustomControl != "" {
			executorValues["custom_control"] = cty.StringVal(actionGroup.ActionGroupExecutor.CustomControl)
		}

		if len(executorValues) > 0 {
			moduleBody.SetAttributeValue("action_group_executor", cty.ObjectVal(executorValues))
		}
	}

	// API Schema configuration
	if actionGroup.APISchema != nil {
		apiSchemaValues := make(map[string]cty.Value)

		if actionGroup.APISchema.S3 != nil {
			s3Values := make(map[string]cty.Value)

			// Check if we have packaged schema with updated S3 location
			if bucket, key := g.context.GetSchemaS3Location(resource.Metadata.Name); bucket != "" && key != "" {
				s3Values["s3_bucket_name"] = cty.StringVal(bucket)
				s3Values["s3_object_key"] = cty.StringVal(key)
				g.logger.WithFields(logrus.Fields{
					"action_group": resource.Metadata.Name,
					"bucket":       bucket,
					"key":          key,
				}).Debug("Using packaged schema S3 location")
			} else {
				// Use original S3 configuration from YAML
				s3Values["s3_bucket_name"] = cty.StringVal(actionGroup.APISchema.S3.S3BucketName)
				s3Values["s3_object_key"] = cty.StringVal(actionGroup.APISchema.S3.S3ObjectKey)
				g.logger.WithField("action_group", resource.Metadata.Name).Debug("Using original schema S3 location from YAML")
			}

			apiSchemaValues["s3"] = cty.ObjectVal(s3Values)
		}

		if actionGroup.APISchema.Payload != "" {
			apiSchemaValues["payload"] = cty.StringVal(actionGroup.APISchema.Payload)
		}

		if len(apiSchemaValues) > 0 {
			moduleBody.SetAttributeValue("api_schema", cty.ObjectVal(apiSchemaValues))
		}
	}

	// Function Schema configuration
	if actionGroup.FunctionSchema != nil {
		functionList := make([]cty.Value, 0, len(actionGroup.FunctionSchema.Functions))

		for _, function := range actionGroup.FunctionSchema.Functions {
			functionValues := make(map[string]cty.Value)
			functionValues["name"] = cty.StringVal(function.Name)

			// Always include description field for consistent structure
			if function.Description != "" {
				functionValues["description"] = cty.StringVal(function.Description)
			} else {
				functionValues["description"] = cty.NullVal(cty.String)
			}

			// Convert parameters to JSON string for consistent structure
			if len(function.Parameters) > 0 {
				// Convert to a map structure that can be JSON encoded
				paramMap := make(map[string]interface{})
				for paramName, param := range function.Parameters {
					paramInfo := map[string]interface{}{
						"description": param.Description,
						"type":        param.Type,
						"required":    param.Required,
					}
					paramMap[paramName] = paramInfo
				}

				// Convert to JSON string for consistent type
				paramJSON, err := json.Marshal(paramMap)
				if err != nil {
					g.logger.WithError(err).Warn("Failed to marshal parameters")
					functionValues["parameters"] = cty.StringVal("{}")
				} else {
					functionValues["parameters"] = cty.StringVal(string(paramJSON))
				}
			} else {
				// Include empty parameters JSON for consistency
				functionValues["parameters"] = cty.StringVal("{}")
			}

			functionList = append(functionList, cty.ObjectVal(functionValues))
		}

		moduleBody.SetAttributeValue("function_schema", cty.ObjectVal(map[string]cty.Value{
			"functions": cty.ListVal(functionList),
		}))
	}

	// Tags
	if len(actionGroup.Tags) > 0 {
		tagValues := make(map[string]cty.Value)
		for key, value := range actionGroup.Tags {
			tagValues[key] = cty.StringVal(value)
		}
		moduleBody.SetAttributeValue("tags", cty.ObjectVal(tagValues))
	}

	// New Terraform-specific attributes
	if actionGroup.PrepareAgent != nil {
		moduleBody.SetAttributeValue("prepare_agent", cty.BoolVal(*actionGroup.PrepareAgent))
	}

	// Timeouts configuration
	if actionGroup.Timeouts != nil {
		timeoutValues := make(map[string]cty.Value)
		if actionGroup.Timeouts.Create != "" {
			timeoutValues["create"] = cty.StringVal(actionGroup.Timeouts.Create)
		}
		if actionGroup.Timeouts.Update != "" {
			timeoutValues["update"] = cty.StringVal(actionGroup.Timeouts.Update)
		}
		if actionGroup.Timeouts.Delete != "" {
			timeoutValues["delete"] = cty.StringVal(actionGroup.Timeouts.Delete)
		}
		if len(timeoutValues) > 0 {
			moduleBody.SetAttributeValue("timeouts", cty.ObjectVal(timeoutValues))
		}
	}

	body.AppendNewline()

	g.logger.WithField("action_group", resource.Metadata.Name).Info("Generated action group module")
	return nil
}
