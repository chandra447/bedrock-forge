package generator

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"

	"bedrock-forge/internal/models"
)

// generateLambdaNative creates a native AWS Terraform resource for a Lambda function
func (g *HCLGenerator) generateLambdaNative(body *hclwrite.Body, resource models.BaseResource) error {
	lambda, ok := resource.Spec.(models.LambdaSpec)
	if !ok {
		// Try to parse as map and convert to LambdaSpec
		specMap, mapOk := resource.Spec.(map[string]interface{})
		if !mapOk {
			return fmt.Errorf("invalid lambda spec format")
		}

		// Convert map to LambdaSpec
		specJSON, err := json.Marshal(specMap)
		if err != nil {
			return fmt.Errorf("failed to marshal lambda spec: %w", err)
		}

		if err := json.Unmarshal(specJSON, &lambda); err != nil {
			return fmt.Errorf("failed to unmarshal lambda spec: %w", err)
		}
	}

	resourceName := g.sanitizeResourceName(resource.Metadata.Name)

	// Generate IAM role for Lambda execution first
	if err := g.generateLambdaExecutionRole(body, resourceName, lambda); err != nil {
		return fmt.Errorf("failed to generate Lambda execution role: %w", err)
	}

	// Create native AWS Lambda function resource
	resourceBlock := body.AppendNewBlock("resource", []string{"aws_lambda_function", resourceName})
	resourceBody := resourceBlock.Body()

	// Set basic attributes according to AWS provider schema
	resourceBody.SetAttributeValue("function_name", cty.StringVal(resource.Metadata.Name))
	resourceBody.SetAttributeValue("runtime", cty.StringVal(lambda.Runtime))
	resourceBody.SetAttributeValue("handler", cty.StringVal(lambda.Handler))

	// Role reference
	if lambda.RoleArn != "" {
		resourceBody.SetAttributeValue("role", cty.StringVal(lambda.RoleArn))
	} else if !lambda.Role.IsEmpty() {
		// Handle reference to IAM role
		roleResourceName := g.sanitizeResourceName(lambda.Role.String())
		resourceBody.SetAttributeRaw("role", hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(fmt.Sprintf("aws_iam_role.%s.arn", roleResourceName))},
		})
	} else {
		// Use auto-generated role
		roleResourceName := fmt.Sprintf("%s_execution_role", resourceName)
		resourceBody.SetAttributeRaw("role", hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(fmt.Sprintf("aws_iam_role.%s.arn", roleResourceName))},
		})
	}

	// Optional description
	if resource.Metadata.Description != "" {
		resourceBody.SetAttributeValue("description", cty.StringVal(resource.Metadata.Description))
	}

	// Code configuration
	if lambda.Code.ZipFile != "" {
		// Inline code
		resourceBody.SetAttributeValue("filename", cty.StringVal("lambda_function.zip"))
		resourceBody.SetAttributeValue("source_code_hash", cty.StringVal("${filebase64sha256(\"lambda_function.zip\")}"))
	} else if lambda.Code.S3Bucket != "" {
		// S3 source
		resourceBody.SetAttributeValue("s3_bucket", cty.StringVal(lambda.Code.S3Bucket))
		resourceBody.SetAttributeValue("s3_key", cty.StringVal(lambda.Code.S3Key))
		if lambda.Code.S3ObjectVersion != "" {
			resourceBody.SetAttributeValue("s3_object_version", cty.StringVal(lambda.Code.S3ObjectVersion))
		}
	} else if lambda.Code.Source != "" {
		// Local source directory - need to create zip
		resourceBody.SetAttributeValue("filename", cty.StringVal(fmt.Sprintf("%s.zip", resourceName)))
		resourceBody.SetAttributeRaw("source_code_hash", hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(fmt.Sprintf("data.archive_file.%s.output_base64sha256", resourceName))},
		})

		// Generate archive data source
		g.generateArchiveDataSource(body, resourceName, lambda.Code.Source)
	}

	// Environment variables
	if len(lambda.Environment) > 0 {
		envBlock := resourceBody.AppendNewBlock("environment", nil)
		envBody := envBlock.Body()

		envVarMap := make(map[string]string)
		for key, value := range lambda.Environment {
			envVarMap[key] = value
		}

		// Build the variables block content
		var tokens hclwrite.Tokens
		tokens = append(tokens, &hclwrite.Token{Type: hclsyntax.TokenOBrace, Bytes: []byte("{\n")})
		for key, value := range envVarMap {
			tokens = append(tokens, &hclwrite.Token{Type: hclsyntax.TokenIdent, Bytes: []byte("    " + key)})
			tokens = append(tokens, &hclwrite.Token{Type: hclsyntax.TokenEqual, Bytes: []byte(" = ")})

			// Check if this is a Terraform reference
			if strings.HasPrefix(value, "${") && strings.HasSuffix(value, "}") {
				// Extract the reference without the ${} wrapper
				refContent := value[2 : len(value)-1]
				tokens = append(tokens, &hclwrite.Token{Type: hclsyntax.TokenIdent, Bytes: []byte(refContent)})
			} else {
				tokens = append(tokens, &hclwrite.Token{Type: hclsyntax.TokenOQuote, Bytes: []byte("\"")})
				tokens = append(tokens, &hclwrite.Token{Type: hclsyntax.TokenQuotedLit, Bytes: []byte(value)})
				tokens = append(tokens, &hclwrite.Token{Type: hclsyntax.TokenCQuote, Bytes: []byte("\"")})
			}
			tokens = append(tokens, &hclwrite.Token{Type: hclsyntax.TokenNewline, Bytes: []byte("\n")})
		}
		tokens = append(tokens, &hclwrite.Token{Type: hclsyntax.TokenCBrace, Bytes: []byte("  }")})

		envBody.SetAttributeRaw("variables", tokens)
	}

	// Configuration settings
	if lambda.Timeout > 0 {
		resourceBody.SetAttributeValue("timeout", cty.NumberIntVal(int64(lambda.Timeout)))
	}

	if lambda.MemorySize > 0 {
		resourceBody.SetAttributeValue("memory_size", cty.NumberIntVal(int64(lambda.MemorySize)))
	}

	if lambda.ReservedConcurrency > 0 {
		resourceBody.SetAttributeValue("reserved_concurrent_executions", cty.NumberIntVal(int64(lambda.ReservedConcurrency)))
	}

	// Tags
	if len(lambda.Tags) > 0 {
		tagValues := make(map[string]cty.Value)
		for key, value := range lambda.Tags {
			tagValues[key] = cty.StringVal(value)
		}
		resourceBody.SetAttributeValue("tags", cty.ObjectVal(tagValues))
	}

	// VPC configuration
	if lambda.VpcConfig != nil {
		vpcConfigBlock := resourceBody.AppendNewBlock("vpc_config", nil)
		vpcConfigBody := vpcConfigBlock.Body()

		if len(lambda.VpcConfig.SecurityGroupIds) > 0 {
			sgIds := make([]cty.Value, 0, len(lambda.VpcConfig.SecurityGroupIds))
			for _, sgId := range lambda.VpcConfig.SecurityGroupIds {
				sgIds = append(sgIds, cty.StringVal(sgId))
			}
			vpcConfigBody.SetAttributeValue("security_group_ids", cty.ListVal(sgIds))
		}

		if len(lambda.VpcConfig.SubnetIds) > 0 {
			subnetIds := make([]cty.Value, 0, len(lambda.VpcConfig.SubnetIds))
			for _, subnetId := range lambda.VpcConfig.SubnetIds {
				subnetIds = append(subnetIds, cty.StringVal(subnetId))
			}
			vpcConfigBody.SetAttributeValue("subnet_ids", cty.ListVal(subnetIds))
		}
	}

	// Advanced attributes
	g.setLambdaNativeAdvancedAttributes(resourceBody, lambda)

	body.AppendNewline()

	// Generate resource-based policies for Bedrock agent access
	if err := g.generateLambdaResourcePermissions(body, resourceName, resource.Metadata.Name, lambda); err != nil {
		return fmt.Errorf("failed to generate Lambda resource permissions: %w", err)
	}

	g.logger.WithField("lambda", resource.Metadata.Name).Info("Generated native Lambda resource")
	return nil
}

// generateLambdaExecutionRole creates an IAM role for Lambda execution
func (g *HCLGenerator) generateLambdaExecutionRole(body *hclwrite.Body, lambdaResourceName string, lambda models.LambdaSpec) error {
	roleResourceName := fmt.Sprintf("%s_execution_role", lambdaResourceName)

	// Create IAM role
	roleBlock := body.AppendNewBlock("resource", []string{"aws_iam_role", roleResourceName})
	roleBody := roleBlock.Body()

	roleBody.SetAttributeValue("name", cty.StringVal(fmt.Sprintf("%s-execution-role", lambdaResourceName)))
	roleBody.SetAttributeValue("assume_role_policy", cty.StringVal(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Effect": "Allow",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      }
    }
  ]
}`))

	// Attach basic execution role policy
	policyAttachmentBlock := body.AppendNewBlock("resource", []string{"aws_iam_role_policy_attachment", fmt.Sprintf("%s_basic", roleResourceName)})
	policyAttachmentBody := policyAttachmentBlock.Body()

	policyAttachmentBody.SetAttributeRaw("role", hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(fmt.Sprintf("aws_iam_role.%s.name", roleResourceName))},
	})
	policyAttachmentBody.SetAttributeValue("policy_arn", cty.StringVal("arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"))

	// If VPC config is specified, attach VPC execution role
	if lambda.VpcConfig != nil {
		vpcPolicyAttachmentBlock := body.AppendNewBlock("resource", []string{"aws_iam_role_policy_attachment", fmt.Sprintf("%s_vpc", roleResourceName)})
		vpcPolicyAttachmentBody := vpcPolicyAttachmentBlock.Body()

		vpcPolicyAttachmentBody.SetAttributeRaw("role", hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(fmt.Sprintf("aws_iam_role.%s.name", roleResourceName))},
		})
		vpcPolicyAttachmentBody.SetAttributeValue("policy_arn", cty.StringVal("arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"))
	}

	// Add S3 permissions if environment variables reference S3 buckets
	if g.needsS3Permissions(lambda) {
		s3PolicyBlock := body.AppendNewBlock("resource", []string{"aws_iam_role_policy", fmt.Sprintf("%s_s3_policy", roleResourceName)})
		s3PolicyBody := s3PolicyBlock.Body()

		s3PolicyBody.SetAttributeValue("name", cty.StringVal("S3AccessPolicy"))
		s3PolicyBody.SetAttributeRaw("role", hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(fmt.Sprintf("aws_iam_role.%s.id", roleResourceName))},
		})
		s3PolicyBody.SetAttributeValue("policy", cty.StringVal(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetObject",
        "s3:PutObject",
        "s3:DeleteObject"
      ],
      "Resource": "arn:aws:s3:::*/*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "s3:ListBucket",
        "s3:GetBucketLocation"
      ],
      "Resource": "arn:aws:s3:::*"
    }
  ]
}`))
	}

	body.AppendNewline()
	return nil
}

// generateArchiveDataSource creates a data source for archiving Lambda source code
func (g *HCLGenerator) generateArchiveDataSource(body *hclwrite.Body, resourceName, sourcePath string) {
	dataBlock := body.AppendNewBlock("data", []string{"archive_file", resourceName})
	dataBody := dataBlock.Body()

	dataBody.SetAttributeValue("type", cty.StringVal("zip"))
	dataBody.SetAttributeValue("source_dir", cty.StringVal(sourcePath))
	dataBody.SetAttributeValue("output_path", cty.StringVal(fmt.Sprintf("%s.zip", resourceName)))

	body.AppendNewline()
}

// generateLambdaResourcePermissions creates aws_lambda_permission resources for Bedrock agent access
func (g *HCLGenerator) generateLambdaResourcePermissions(body *hclwrite.Body, lambdaResourceName, lambdaName string, lambda models.LambdaSpec) error {
	// Find all agents that reference this Lambda function
	referencingAgents := g.findAgentsReferencingLambda(lambdaName)

	if len(referencingAgents) > 0 {
		// Create agent-specific permissions
		for _, agentName := range referencingAgents {
			agentResourceName := g.sanitizeResourceName(agentName)
			permissionResourceName := fmt.Sprintf("%s_allow_%s", lambdaResourceName, agentResourceName)

			permissionBlock := body.AppendNewBlock("resource", []string{"aws_lambda_permission", permissionResourceName})
			permissionBody := permissionBlock.Body()

			permissionBody.SetAttributeValue("statement_id", cty.StringVal(fmt.Sprintf("AllowBedrockAgent_%s", agentResourceName)))
			permissionBody.SetAttributeValue("action", cty.StringVal("lambda:InvokeFunction"))
			permissionBody.SetAttributeRaw("function_name", hclwrite.Tokens{
				{Type: hclsyntax.TokenIdent, Bytes: []byte(fmt.Sprintf("aws_lambda_function.%s.function_name", lambdaResourceName))},
			})
			permissionBody.SetAttributeValue("principal", cty.StringVal("bedrock.amazonaws.com"))
			permissionBody.SetAttributeRaw("source_arn", hclwrite.Tokens{
				{Type: hclsyntax.TokenIdent, Bytes: []byte(fmt.Sprintf("aws_bedrockagent_agent.%s.agent_arn", agentResourceName))},
			})

			body.AppendNewline()

			g.logger.WithField("lambda", lambdaName).WithField("agent", agentName).Debug("Generated agent-specific Lambda permission")
		}
	} else {
		// If no agents reference this Lambda, add general Bedrock permission (unless explicitly disabled)
		if lambda.ResourcePolicy == nil || lambda.ResourcePolicy.AllowBedrockAgents {
			permissionResourceName := fmt.Sprintf("%s_allow_bedrock", lambdaResourceName)

			permissionBlock := body.AppendNewBlock("resource", []string{"aws_lambda_permission", permissionResourceName})
			permissionBody := permissionBlock.Body()

			permissionBody.SetAttributeValue("statement_id", cty.StringVal("AllowBedrockAgentInvoke"))
			permissionBody.SetAttributeValue("action", cty.StringVal("lambda:InvokeFunction"))
			permissionBody.SetAttributeRaw("function_name", hclwrite.Tokens{
				{Type: hclsyntax.TokenIdent, Bytes: []byte(fmt.Sprintf("aws_lambda_function.%s.function_name", lambdaResourceName))},
			})
			permissionBody.SetAttributeValue("principal", cty.StringVal("bedrock.amazonaws.com"))

			body.AppendNewline()
		}
	}

	return nil
}

// setLambdaNativeAdvancedAttributes sets advanced Lambda attributes
func (g *HCLGenerator) setLambdaNativeAdvancedAttributes(resourceBody *hclwrite.Body, lambda models.LambdaSpec) {
	// Architectures
	if len(lambda.Architectures) > 0 {
		archVals := make([]cty.Value, 0, len(lambda.Architectures))
		for _, arch := range lambda.Architectures {
			archVals = append(archVals, cty.StringVal(arch))
		}
		resourceBody.SetAttributeValue("architectures", cty.ListVal(archVals))
	}

	// Code signing config
	if lambda.CodeSigningConfigArn != "" {
		resourceBody.SetAttributeValue("code_signing_config_arn", cty.StringVal(lambda.CodeSigningConfigArn))
	}

	// Dead letter config
	if lambda.DeadLetterConfig != nil {
		dlcBlock := resourceBody.AppendNewBlock("dead_letter_config", nil)
		dlcBody := dlcBlock.Body()
		dlcBody.SetAttributeValue("target_arn", cty.StringVal(lambda.DeadLetterConfig.TargetArn))
	}

	// Ephemeral storage
	if lambda.EphemeralStorage != nil {
		esBlock := resourceBody.AppendNewBlock("ephemeral_storage", nil)
		esBody := esBlock.Body()
		esBody.SetAttributeValue("size", cty.NumberIntVal(int64(lambda.EphemeralStorage.Size)))
	}

	// File system config
	if lambda.FileSystemConfig != nil {
		fscBlock := resourceBody.AppendNewBlock("file_system_config", nil)
		fscBody := fscBlock.Body()
		fscBody.SetAttributeValue("arn", cty.StringVal(lambda.FileSystemConfig.Arn))
		fscBody.SetAttributeValue("local_mount_path", cty.StringVal(lambda.FileSystemConfig.LocalMountPath))
	}

	// Image config
	if lambda.ImageConfig != nil {
		imgBlock := resourceBody.AppendNewBlock("image_config", nil)
		imgBody := imgBlock.Body()

		if len(lambda.ImageConfig.Command) > 0 {
			cmdVals := make([]cty.Value, 0, len(lambda.ImageConfig.Command))
			for _, cmd := range lambda.ImageConfig.Command {
				cmdVals = append(cmdVals, cty.StringVal(cmd))
			}
			imgBody.SetAttributeValue("command", cty.ListVal(cmdVals))
		}
		if len(lambda.ImageConfig.EntryPoint) > 0 {
			epVals := make([]cty.Value, 0, len(lambda.ImageConfig.EntryPoint))
			for _, ep := range lambda.ImageConfig.EntryPoint {
				epVals = append(epVals, cty.StringVal(ep))
			}
			imgBody.SetAttributeValue("entry_point", cty.ListVal(epVals))
		}
		if lambda.ImageConfig.WorkingDirectory != "" {
			imgBody.SetAttributeValue("working_directory", cty.StringVal(lambda.ImageConfig.WorkingDirectory))
		}
	}

	// KMS key
	if lambda.KmsKeyArn != "" {
		resourceBody.SetAttributeValue("kms_key_arn", cty.StringVal(lambda.KmsKeyArn))
	}

	// Layers
	if len(lambda.Layers) > 0 {
		layerVals := make([]cty.Value, 0, len(lambda.Layers))
		for _, layer := range lambda.Layers {
			layerVals = append(layerVals, cty.StringVal(layer))
		}
		resourceBody.SetAttributeValue("layers", cty.ListVal(layerVals))
	}

	// Package type
	if lambda.PackageType != "" {
		resourceBody.SetAttributeValue("package_type", cty.StringVal(lambda.PackageType))
	}

	// Publish
	if lambda.Publish != nil {
		resourceBody.SetAttributeValue("publish", cty.BoolVal(*lambda.Publish))
	}

	// Source code hash
	if lambda.SourceCodeHash != "" {
		resourceBody.SetAttributeValue("source_code_hash", cty.StringVal(lambda.SourceCodeHash))
	}

	// Timeouts
	if lambda.Timeouts != nil {
		timeoutBlock := resourceBody.AppendNewBlock("timeouts", nil)
		timeoutBody := timeoutBlock.Body()

		if lambda.Timeouts.Create != "" {
			timeoutBody.SetAttributeValue("create", cty.StringVal(lambda.Timeouts.Create))
		}
		if lambda.Timeouts.Update != "" {
			timeoutBody.SetAttributeValue("update", cty.StringVal(lambda.Timeouts.Update))
		}
		if lambda.Timeouts.Delete != "" {
			timeoutBody.SetAttributeValue("delete", cty.StringVal(lambda.Timeouts.Delete))
		}
	}

	// Tracing config
	if lambda.TracingConfig != nil {
		tracingBlock := resourceBody.AppendNewBlock("tracing_config", nil)
		tracingBody := tracingBlock.Body()
		tracingBody.SetAttributeValue("mode", cty.StringVal(lambda.TracingConfig.Mode))
	}
}

// needsS3Permissions checks if the Lambda function needs S3 permissions based on environment variables
func (g *HCLGenerator) needsS3Permissions(lambda models.LambdaSpec) bool {
	// Check if any environment variables reference S3 buckets
	for _, value := range lambda.Environment {
		if strings.Contains(value, "aws_s3_bucket.") || strings.Contains(value, "s3://") {
			return true
		}
	}
	return false
}
