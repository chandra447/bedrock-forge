<div align="center">
  <img src="golang-bedrock.png" alt="Bedrock Forge Logo" width="200"/>
  
  # Bedrock Forge

  **Enterprise platform for transforming YAML configurations into AWS Bedrock agent deployments using Terraform modules**
</div>

Bedrock Forge simplifies AWS Bedrock agent deployment by allowing teams to define agents, Lambda functions, action groups, knowledge bases, and IAM roles in simple YAML files, then automatically generating production-ready Terraform infrastructure.

## 🚀 Features

### Core Capabilities
- **YAML-to-Terraform Generation**: Transform declarative YAML configurations into Terraform modules
- **Complete Resource Support**: Agents, Lambda functions, Action Groups, Knowledge Bases, Guardrails, Prompts, IAM Roles, and Custom Modules
- **Custom Module Integration**: Include your own Terraform modules alongside Bedrock resources
- **Dependency Management**: Automatic resource ordering and cross-module references
- **Artifact Packaging**: Automatic Lambda code packaging and S3 upload
- **Schema Management**: OpenAPI schema discovery and validation
- **IAM Security**: Automatic IAM role generation with comprehensive permissions

### Enterprise Features
- **GitHub Actions Integration**: Complete CI/CD pipeline with automated deployment
- **Multi-Environment Support**: Development, staging, and production deployments
- **Security Best Practices**: Least-privilege IAM roles and enterprise compliance patterns
- **Scalable Architecture**: Support for complex enterprise deployments
- **Team Collaboration**: Git-based workflow with approval processes

## 📋 Table of Contents

- [Quick Start](#quick-start)
- [Resource Types](#resource-types)
- [Custom Modules](#custom-modules)
- [IAM Role Management](#iam-role-management)
- [GitHub Actions Workflow](#github-actions-workflow)
- [Configuration](#configuration)
- [Examples](#examples)
- [CLI Reference](#cli-reference)
- [Enterprise Setup](#enterprise-setup)
- [Troubleshooting](#troubleshooting)

## 🏃 Quick Start

### Prerequisites
- Go 1.21+
- AWS CLI configured
- Terraform 1.0+
- Git repository

### Installation

```bash
# Clone the repository
git clone https://github.com/your-org/bedrock-forge
cd bedrock-forge

# Build the binary
go build -o bedrock-forge ./cmd/bedrock-forge

# Verify installation
./bedrock-forge version
```

### Basic Usage

1. **Create your first agent**:
```yaml
# agents/my-agent.yml
kind: Agent
metadata:
  name: "my-agent"
  description: "My first Bedrock agent"
spec:
  foundationModel: "anthropic.claude-3-sonnet-20240229-v1:0"
  instruction: "You are a helpful assistant"
  # IAM role is automatically generated!
```

2. **Generate Terraform**:
```bash
./bedrock-forge generate . ./terraform
```

3. **Deploy**:
```bash
cd terraform
terraform init
terraform plan
terraform apply
```

## 📦 Resource Types

### Agent
Complete Bedrock agent configuration with guardrails, knowledge bases, and action groups.

```yaml
kind: Agent
metadata:
  name: "customer-support"
spec:
  foundationModel: "anthropic.claude-3-sonnet-20240229-v1:0"
  instruction: "You are a helpful customer support agent"
  idleSessionTtlInSeconds: 3600
  
  # Guardrail integration
  guardrail:
    name: "content-safety-guardrail"
    version: "1"
    mode: "pre"
  
  # Inline action groups (new structure!)
  actionGroups:
    - name: "order-management"
      description: "Handle order lookups and updates"
      actionGroupExecutor:
        lambda: "order-lookup-lambda"
      functionSchema:
        functions:
          - name: "lookup_order"
            description: "Look up order by ID"
            parameters:
              order_id:
                type: "string"
                required: true
  
  # Prompt overrides
  promptOverrides:
    - promptType: "ORCHESTRATION"
      prompt: "custom-orchestration-prompt"
      variant: "production"
  
  # Memory configuration
  memoryConfiguration:
    enabledMemoryTypes: ["SESSION_SUMMARY"]
    storageDays: 30
```

### Lambda Function
Lambda function with automatic code packaging and S3 upload.

```yaml
kind: Lambda
metadata:
  name: "order-lookup"
spec:
  runtime: "python3.11"
  handler: "app.handler"
  description: "Lambda function to look up customer orders"
  timeout: 30
  memorySize: 256
  
  environmentVariables:
    LOG_LEVEL: "INFO"
    API_URL: "https://api.company.com"
  
  # VPC configuration (optional)
  vpcConfig:
    securityGroupIds: ["sg-12345"]
    subnetIds: ["subnet-12345", "subnet-67890"]
```

### Action Group
Action group with agent association and automatic Lambda integration.

```yaml
kind: ActionGroup
metadata:
  name: "order-management"
spec:
  agentName: "customer-support"  # Links to agent
  description: "Provides order lookup and management capabilities"
  
  # Lambda executor (can reference local Lambda or external ARN)
  actionGroupExecutor:
    lambda: "order-lookup"          # Local Lambda reference
    # lambdaArn: "arn:aws:lambda:..."  # Or external Lambda ARN
  
  # Function schema (optional - can use OpenAPI instead)
  functionSchema:
    functions:
      - name: "lookup_order"
        description: "Look up order details by order ID"
        parameters:
          order_id:
            type: "string"
            description: "The unique order identifier"
            required: true
```

### Knowledge Base
Vector knowledge base with S3 data sources and chunking strategies.

```yaml
kind: KnowledgeBase
metadata:
  name: "faq-kb"
spec:
  description: "Customer FAQ knowledge base"
  
  knowledgeBaseConfiguration:
    type: "VECTOR"
    vectorKnowledgeBaseConfiguration:
      embeddingModelArn: "arn:aws:bedrock:us-east-1::foundation-model/amazon.titan-embed-text-v1"
  
  storageConfiguration:
    type: "OPENSEARCH_SERVERLESS"
    opensearchServerlessConfiguration:
      collectionArn: "arn:aws:aoss:us-east-1:123456789012:collection/bedrock-kb"
      vectorIndexName: "bedrock-knowledge-base-index"
  
  dataSources:
    - name: "faq-documents"
      type: "S3"
      s3Configuration:
        bucketArn: "arn:aws:s3:::company-kb-documents"
        inclusionPrefixes: ["faq/"]
      
      chunkingConfiguration:
        chunkingStrategy: "FIXED_SIZE"
        fixedSizeChunkingConfiguration:
          maxTokens: 512
          overlapPercentage: 20
```

### Guardrail
Content safety and compliance guardrails.

```yaml
kind: Guardrail
metadata:
  name: "content-safety-guardrail"
spec:
  description: "Enterprise content safety guardrail"
  
  contentPolicyConfig:
    filtersConfig:
      - type: "SEXUAL"
        inputStrength: "HIGH"
        outputStrength: "HIGH"
      - type: "VIOLENCE"
        inputStrength: "MEDIUM"
        outputStrength: "HIGH"
  
  sensitiveInformationPolicyConfig:
    piiEntitiesConfig:
      - type: "EMAIL"
        action: "BLOCK"
      - type: "PHONE"
        action: "ANONYMIZE"
  
  topicPolicyConfig:
    topicsConfig:
      - name: "Investment Advice"
        definition: "Financial investment discussions"
        type: "DENY"
```

### Prompt
Custom prompts with multiple variants.

```yaml
kind: Prompt
metadata:
  name: "custom-orchestration-prompt"
spec:
  description: "Custom orchestration prompt for customer support"
  defaultVariant: "production"
  
  variants:
    - name: "production"
      modelId: "anthropic.claude-3-sonnet-20240229-v1:0"
      templateType: "TEXT"
      templateConfiguration:
        text: |
          You are a customer support agent. Always be helpful and professional.
          
          Customer Query: {{query}}
          Context: {{context}}
      
      inferenceConfiguration:
        text:
          temperature: 0.1
          topP: 0.9
          maxTokens: 2048
```

### Custom Module
Include your own Terraform modules alongside Bedrock resources.

```yaml
kind: CustomModule
metadata:
  name: "s3-storage"
  description: "S3 bucket for storing agent artifacts"
spec:
  # Module source (local, registry, or git)
  source: "./modules/s3-bucket"
  # source: "terraform-aws-modules/s3-bucket/aws"  # Registry module
  # version: "3.0.0"  # For registry/git modules
  
  # Input variables
  variables:
    bucket_name: "bedrock-artifacts-${var.environment}"
    versioning_enabled: true
    tags:
      Purpose: "BedrockStorage"
      Environment: "${var.environment}"
  
  # Dependencies (optional)
  dependsOn:
    - "vpc-module"  # Wait for VPC first
  
  description: "S3 bucket for Lambda code and schemas"
```

**Supported Sources:**
- **Local modules**: `"./modules/my-module"`
- **Terraform Registry**: `"terraform-aws-modules/vpc/aws"` + `version`
- **Git repositories**: `"git::https://github.com/org/repo.git"` + `version`

**Variable Types:**
- Strings, numbers, booleans
- Lists and objects (complex configurations)
- Module references: `"${module.vpc.vpc_id}"`

## Custom Modules

Custom modules allow you to integrate your existing Terraform infrastructure with Bedrock resources. This enables you to:

- **Add supporting infrastructure**: VPC, storage, monitoring
- **Integrate with existing systems**: Databases, APIs, security tools
- **Implement enterprise patterns**: Compliance, governance, cost management
- **Reuse proven modules**: Leverage your organization's Terraform library

### Example: Complete Infrastructure Stack

```yaml
# 1. VPC Infrastructure
kind: CustomModule
metadata:
  name: "bedrock-vpc"
spec:
  source: "terraform-aws-modules/vpc/aws"
  version: "5.0.0"
  variables:
    name: "bedrock-agent-vpc"
    cidr: "10.0.0.0/16"
    azs: ["us-east-1a", "us-east-1b"]
    private_subnets: ["10.0.1.0/24", "10.0.2.0/24"]
    public_subnets: ["10.0.101.0/24", "10.0.102.0/24"]
    enable_nat_gateway: true

---
# 2. OpenSearch for Knowledge Base
kind: CustomModule
metadata:
  name: "bedrock-opensearch"
spec:
  source: "git::https://github.com/your-org/terraform-opensearch.git"
  version: "v1.2.0"
  dependsOn: ["bedrock-vpc"]
  variables:
    cluster_name: "bedrock-kb"
    vpc_id: "${module.bedrock_vpc.vpc_id}"
    subnet_ids: "${module.bedrock_vpc.private_subnets}"

---
# 3. Bedrock Agent
kind: Agent
metadata:
  name: "customer-support"
spec:
  foundationModel: "anthropic.claude-3-sonnet-20240229-v1:0"
  instruction: "You are a helpful support agent"
  # IAM role auto-generated with all required permissions!

---
# 4. Monitoring Dashboard  
kind: CustomModule
metadata:
  name: "monitoring"
spec:
  source: "./modules/cloudwatch-dashboard"
  dependsOn: ["customer-support"]
  variables:
    agent_name: "customer-support"
    dashboard_name: "BedrockAgentMonitoring"
```

### Integration Benefits

1. **Unified Deployment**: Deploy infrastructure and Bedrock resources together
2. **Dependency Management**: Proper ordering ensures VPC → OpenSearch → Agent → Monitoring
3. **Cross-Module References**: Use `${module.vpc.vpc_id}` to link resources
4. **Version Control**: Pin module versions for reproducible deployments
5. **Enterprise Compliance**: Include required security and governance modules

## 🔐 IAM Role Management

**🎉 IAM roles are now automatically generated for all agents!** No configuration needed.

### Automatic IAM Role Generation

Every agent automatically gets an IAM role with comprehensive permissions:

```yaml
kind: Agent
metadata:
  name: "my-agent"
spec:
  foundationModel: "anthropic.claude-3-sonnet-20240229-v1:0"
  instruction: "You are a helpful assistant"
  # No IAM configuration needed - role auto-generated!
```

**Auto-generated permissions include:**
- ✅ **Foundation model access**: `bedrock:InvokeModel`, `bedrock:InvokeModelWithResponseStream`
- ✅ **Lambda invocation**: `lambda:InvokeFunction` for action groups
- ✅ **Knowledge base access**: `bedrock:Retrieve`, `bedrock:RetrieveAndGenerate`
- ✅ **CloudWatch logging**: `logs:CreateLogGroup`, `logs:CreateLogStream`, `logs:PutLogEvents`

### Legacy IAM Configuration (Optional)

For advanced use cases, you can still define custom IAM roles:

### Manual IAM Role Definition
For enterprise scenarios requiring specific permissions:

```yaml
kind: IAMRole
metadata:
  name: "custom-agent-execution-role"
spec:
  description: "Custom execution role with specific permissions"
  
  assumeRolePolicy:
    version: "2012-10-17"
    statement:
      - effect: "Allow"
        principal:
          service: "bedrock.amazonaws.com"
        action: "sts:AssumeRole"
  
  policies:
    - policyArn: "arn:aws:iam::aws:policy/service-role/AmazonBedrockAgentResourcePolicy"
  
  inlinePolicies:
    - name: "CustomPermissions"
      policy:
        version: "2012-10-17"
        statement:
          - effect: "Allow"
            action: ["lambda:InvokeFunction"]
            resource: "arn:aws:lambda:*:*:function:my-functions-*"

---
kind: Agent
metadata:
  name: "my-agent"
spec:
  foundationModel: "anthropic.claude-3-sonnet-20240229-v1:0"
  instruction: "You are a helpful assistant"
  # IAM role auto-generated with all required permissions!
```

> **Note**: Manual IAM role configuration is optional. The system automatically generates roles with all necessary permissions for most use cases.

## 🔄 GitHub Actions Workflow

### Setup

1. **Copy the workflow to your repository**:
```bash
mkdir -p .github/workflows
cp .github/workflows/bedrock-forge-deploy.yml .github/workflows/
```

2. **Configure AWS OIDC** (recommended for security):
```bash
# Use the provided setup script
chmod +x .github/workflows/setup-aws-oidc.sh
./github/workflows/setup-aws-oidc.sh
```

3. **Set repository variables**:
- `AWS_DEPLOYMENT_ROLE`: `arn:aws:iam::123456789012:role/BedrockForgeDeploymentRole`
- `AWS_REGION`: `us-east-1`
- `TF_STATE_BUCKET`: `your-terraform-state-bucket`
- `TF_STATE_KEY_PREFIX`: `bedrock-forge`
- `TF_STATE_LOCK_TABLE`: `terraform-state-lock`

### Workflow Features

#### Multi-Stage Pipeline
1. **Validate**: YAML configuration validation and resource scanning
2. **Package**: Lambda code packaging and schema extraction
3. **Deploy**: Terraform generation and infrastructure deployment
4. **Cleanup**: Status reporting and artifact management

#### Trigger Conditions
- **Push to main/master**: Automatic deployment to development
- **Pull requests**: Validation and planning only
- **Manual dispatch**: Deploy to any environment with approval

#### Environment Support
```yaml
# Manual deployment with environment selection
on:
  workflow_dispatch:
    inputs:
      environment:
        description: 'Environment to deploy'
        required: true
        default: 'dev'
        type: choice
        options:
          - dev
          - staging
          - prod
```

#### Security Features
- AWS OIDC authentication (no long-lived credentials)
- Environment-specific approvals
- Terraform state locking
- Deployment status reporting

### Example Workflow Usage

```yaml
# .github/workflows/deploy.yml
name: Deploy Bedrock Agents

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]
  workflow_dispatch:
    inputs:
      environment:
        description: 'Environment to deploy'
        required: true
        default: 'dev'
        type: choice
        options: [dev, staging, prod]

jobs:
  deploy:
    uses: your-org/bedrock-forge/.github/workflows/bedrock-forge-deploy.yml@main
    with:
      environment: ${{ github.event.inputs.environment || 'dev' }}
    secrets: inherit
```

## ⚙️ Configuration

### Project Configuration (forge.yml)
```yaml
metadata:
  name: "customer-support-platform"
  team: "customer-experience"
  environment: "{{ .Environment }}"

terraform:
  backend:
    type: "s3"
    bucket: "{{ .TFStateBucket }}"
    key: "{{ .TFStateKeyPrefix }}/{{ .Environment }}/terraform.tfstate"
    region: "{{ .AWSRegion }}"
    encrypt: true

scanning:
  paths: ["./agents", "./lambdas", "./action-groups"]
  include: ["*.yml", "*.yaml"]
  exclude: ["**/node_modules/**", "**/.git/**"]

modules:
  registry: "git::https://github.com/company/bedrock-terraform-modules"
  version: "v1.2.0"

# Environment-specific overrides
environments:
  dev:
    modules:
      version: "v1.2.0-dev"
  prod:
    modules:
      version: "v1.2.0"
```

### Lambda Packaging Configuration
```yaml
lambda:
  packaging:
    exclude_patterns:
      - "*.yml"
      - ".git/**"
      - "__pycache__/**"
      - "tests/**"
    
    python:
      runtime: "python3.11"
      install_requirements: true
    
    nodejs:
      runtime: "nodejs18.x"
      install_dependencies: true
```

### Schema Management
```yaml
schemas:
  discovery:
    file_patterns:
      - "openapi.json"
      - "openapi.yaml"
      - "schema.json"
  
  validation:
    enabled: true
    bedrock_compatibility: true
```

## 📚 CLI Reference

### Commands

#### `bedrock-forge scan [path]`
Discover and list all resources in the specified directory.

```bash
./bedrock-forge scan .
./bedrock-forge scan ./examples
```

#### `bedrock-forge validate [path]`
Validate YAML syntax and dependencies.

```bash
./bedrock-forge validate .
./bedrock-forge validate ./agents
```

#### `bedrock-forge generate [input-path] [output-path]`
Generate Terraform configuration from YAML resources.

```bash
./bedrock-forge generate . ./terraform
./bedrock-forge generate ./examples ./output
```

#### `bedrock-forge version`
Show version information.

```bash
./bedrock-forge version
```

### Flags

- `--log-level`: Set logging level (debug, info, warn, error)
- `--config`: Specify configuration file path
- `--dry-run`: Preview changes without generating files

## 🏢 Enterprise Setup

### Module Registry Setup
```yaml
modules:
  registry: "git::https://github.com/your-org/bedrock-terraform-modules"
  version: "v1.2.0"
```

### Team-Specific Configuration
```yaml
metadata:
  team: "customer-experience"
  cost_center: "engineering"
  environment: "production"

terraform:
  executionRole: "arn:aws:iam::123456789012:role/TeamTerraformRole"
  
scanning:
  paths: ["./src/agents", "./src/lambdas"]
  
tags:
  default:
    Team: "{{ .metadata.team }}"
    CostCenter: "{{ .metadata.cost_center }}"
    Environment: "{{ .metadata.environment }}"
```

### Multi-Environment Strategy
```yaml
environments:
  dev:
    terraform:
      backend:
        key: "dev/terraform.tfstate"
    variables:
      log_level: "DEBUG"
  
  staging:
    terraform:
      backend:
        key: "staging/terraform.tfstate"
    variables:
      log_level: "INFO"
  
  prod:
    terraform:
      backend:
        key: "prod/terraform.tfstate"
    variables:
      log_level: "WARN"
      enable_monitoring: true
```

## 🔧 Troubleshooting

### Common Issues

#### 1. Missing Dependencies
```bash
# Error: Agent references non-existent guardrail
Error: agent my-agent references non-existent guardrail my-guardrail

# Solution: Create the guardrail resource or fix the reference
```

#### 2. IAM Permission Errors
```bash
# Error: Access denied when deploying
Error: AccessDenied: User is not authorized to perform: bedrock:CreateAgent

# Solution: Ensure deployment role has required permissions
```

#### 3. Terraform State Issues
```bash
# Error: Backend configuration changed
Error: Backend configuration changed

# Solution: Run terraform init -reconfigure
```

#### 4. Lambda Packaging Failures
```bash
# Error: Failed to package Lambda
Error: requirements.txt not found

# Solution: Ensure requirements.txt exists in Lambda directory
```

### Debug Mode
```bash
# Enable debug logging
./bedrock-forge --log-level debug generate . ./terraform

# Validate with verbose output
./bedrock-forge --log-level debug validate .
```

### Validation Checklist
- [ ] All YAML files have valid syntax
- [ ] Resource dependencies are satisfied
- [ ] IAM roles are properly configured
- [ ] Lambda functions have required files
- [ ] OpenAPI schemas are valid
- [ ] Terraform backend is configured
- [ ] AWS credentials are available

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Update documentation
6. Submit a pull request

## 📄 License

MIT License - see LICENSE file for details.

## 🆘 Support

- **Documentation**: [docs/](./docs/)
- **Issues**: [GitHub Issues](https://github.com/your-org/bedrock-forge/issues)
- **Discussions**: [GitHub Discussions](https://github.com/your-org/bedrock-forge/discussions)
- **Enterprise Support**: Contact your platform team

---

**Bedrock Forge** - Simplifying AWS Bedrock deployments for enterprise teams.