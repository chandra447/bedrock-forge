# Bedrock Forge - Development Plan for Claude Code

## Project Overview

Build **Bedrock Forge**, an enterprise platform that transforms YAML configurations into AWS Bedrock agent deployments using Terraform modules. Teams define agents, lambdas, action groups, and knowledge bases in simple YAML files, and Bedrock Forge generates the complete Terraform infrastructure.

## Architecture Summary

- **Language**: Go (for native HCL support and performance)
- **Input**: YAML files with `kind` specifications (Agent, Lambda, ActionGroup, KnowledgeBase, etc.)
- **Output**: Generated Terraform modules and main.tf file
- **Deployment**: GitHub Actions workflow for CI/CD integration
- **Storage**: S3 for OpenAPI schemas and Lambda packages

## Epic 1: Core YAML Parser and Resource Discovery

### Epic Goal
Create the foundation for scanning directories, parsing YAML files, and identifying Bedrock resources by their `kind` field.

### User Stories
- As a developer, I want Bedrock Forge to automatically discover all YAML resource definitions in my repository
- As a platform engineer, I want validation of YAML schemas to catch configuration errors early

### TODO Items

**1.1 Project Setup**
- [ ] Initialize Go module with proper project structure
- [ ] Add dependencies: `gopkg.in/yaml.v3`, `github.com/hashicorp/hcl/v2`, `github.com/spf13/cobra` (CLI)
- [ ] Create basic CLI structure with `cobra` for commands like `init`, `scan`, `deploy`, `validate`
- [ ] Set up proper logging with structured output

**1.2 YAML Resource Discovery**
- [ ] Implement recursive directory scanning for `.yml` and `.yaml` files
- [ ] Create YAML parser that extracts `kind`, `metadata`, and `spec` fields
- [ ] Build resource registry to track discovered resources by type
- [ ] Add support for include/exclude patterns (from `agentDeployer.yml` config)

**1.3 Resource Models**
- [ ] Define Go structs for each resource type (Agent, Lambda, ActionGroup, KnowledgeBase, Guardrail, Prompt)
- [ ] Implement YAML unmarshaling with proper validation for all Bedrock agent attributes
- [ ] Create resource dependency graph builder (Agent -> Guardrail -> ActionGroup -> Lambda -> KnowledgeBase)
- [ ] Add configuration validation with helpful error messages
- [ ] Support for referencing existing vs new resources (guardrails, knowledge bases, prompts)

**1.4 CLI Foundation**
- [ ] `bedrock-forge scan` - discover and list all resources
- [ ] `bedrock-forge validate` - validate YAML syntax and dependencies
- [ ] `bedrock-forge version` - show version and build info
- [ ] Configuration file support (`forge.yml` or `agentDeployer.yml`)

## Epic 2: Terraform Module Generation Engine

### Epic Goal
Transform parsed YAML resources into properly formatted Terraform HCL module calls with correct dependencies.

### User Stories
- As a developer, I want my YAML configurations automatically converted to production-ready Terraform code
- As a platform team, I want generated Terraform to follow our enterprise standards and naming conventions

### TODO Items

**2.1 HCL Generation Framework**
- [ ] Create HCL file generation using `github.com/hashicorp/hcl/v2/hclwrite`
- [ ] Build template system for each resource type (Agent, Lambda, ActionGroup, etc.)
- [ ] Implement variable interpolation and reference resolution
- [ ] Add support for module versioning and source URLs

**2.2 Resource-Specific Generators**
- [ ] **Agent Generator**: Transform Agent YAML to `bedrock-agent` module calls with full attribute support
  - [ ] Support all agent configuration parameters (idle_session_ttl, customer_encryption_key, etc.)
  - [ ] Handle guardrail configuration (existing or new guardrail references)
  - [ ] Support prompt overrides with variant configurations
  - [ ] Memory configuration and instruction templates
- [ ] **Lambda Generator**: Transform Lambda YAML to `lambda-function` module calls with packaging
- [ ] **ActionGroup Generator**: Transform ActionGroup YAML to `bedrock-action-group` module calls
- [ ] **KnowledgeBase Generator**: Transform KnowledgeBase YAML to `bedrock-knowledge-base` module calls
  - [ ] S3 data source configuration with chunking strategies
  - [ ] Pre-processing and post-processing Lambda integration
  - [ ] Vector database configuration (OpenSearch Serverless)
  - [ ] Embedding model selection and configuration
- [ ] **Guardrail Generator**: Transform Guardrail YAML to `bedrock-guardrail` module calls
  - [ ] Content filters (hate, insults, sexual, violence, misconduct, prompt attacks)
  - [ ] Contextual grounding check configuration
  - [ ] Sensitive information filters (PII detection)
  - [ ] Topic filters and word filters
  - [ ] Support for guardrail versions and deployment
- [ ] **Prompt Generator**: Transform Prompt YAML to `bedrock-prompt` module calls
  - [ ] Support for prompt variants and configurations
  - [ ] Template variable management
  - [ ] Integration with agent prompt overrides
- [ ] **IAM Role Generator**: Generate required IAM roles and policies for all resource types

**2.3 Dependency Management**
- [ ] Build dependency resolver to determine resource creation order
- [ ] Generate Terraform resource references (e.g., `module.agent_name.agent_id`)
- [ ] Handle circular dependency detection and reporting
- [ ] Create proper variable passing between dependent modules

**2.4 IAM Role and Policy Management**
- [ ] Generate IAM execution role for Bedrock agents with foundation model access
- [ ] Create IAM policies for agent access to Lambda functions (invoke permissions)
- [ ] Generate IAM policies for action group Lambda execution roles
- [ ] Create knowledge base access roles for OpenSearch and S3 data sources
- [ ] Implement least-privilege permission templates for each resource type
- [ ] Add cross-service trust relationships (Bedrock -> Lambda, Lambda -> S3, etc.)

**2.5 Main Terraform File Generation**
- [ ] Generate consolidated `main.tf` with all module calls
- [ ] Create `variables.tf` for parameterized deployments
- [ ] Generate `outputs.tf` for resource ARNs and identifiers
- [ ] Add proper provider configuration and version constraints

## Epic 3: Lambda Code Packaging and OpenAPI Schema Generation

### Epic Goal
Automatically package Lambda functions and manage OpenAPI schemas for action groups, storing artifacts in S3.

### User Stories
- As a developer, I want my Lambda code automatically packaged and uploaded during deployment
- As a developer, I want manual OpenAPI schemas discovered and uploaded automatically

### TODO Items

**3.1 Lambda Code Packaging**
- [ ] Implement directory-based Lambda code discovery (co-located with `lambda.yml`)
- [ ] Create ZIP packaging functionality with proper exclusions (`.yml`, `.git`, etc.)
- [ ] Add support for requirements.txt and dependency installation
- [ ] Generate unique S3 keys for Lambda packages with versioning

**3.2 OpenAPI Schema Management**
- [x] **Manual Schema Support**: Allow hand-written OpenAPI schema files
- [ ] Schema validation and Bedrock compatibility checking
- [ ] Schema file discovery with standard naming conventions (openapi.json, schema.json, etc.)
- [ ] Support for both JSON and YAML schema formats

**3.3 S3 Artifact Management**
- [ ] Implement S3 client for uploading Lambda packages and OpenAPI schemas
- [ ] Create consistent naming conventions for S3 objects
- [ ] Add S3 versioning support for artifact history
- [ ] Implement cleanup policies for old artifacts

**3.4 ActionGroup Schema Linking**
- [ ] Generate proper S3 URIs for OpenAPI schemas in ActionGroup modules
- [ ] Handle schema dependencies and references
- [ ] Add schema validation against Bedrock requirements
- [x] Support manual schema workflows only (auto-generation removed)

## Epic 4: GitHub Actions Integration and CI/CD

### Epic Goal
Create reusable GitHub Actions workflow that teams can easily adopt for automated Bedrock deployments.

### User Stories
- As a development team, I want to deploy Bedrock agents automatically when I push code to main branch
- As a platform team, I want teams to adopt our deployment workflow with minimal configuration

### TODO Items

**4.1 GitHub Actions Workflow**
- [ ] Create composite GitHub Action (`bedrock-forge-deploy`)
- [ ] Add proper AWS credential handling and role assumption
- [ ] Implement Terraform backend configuration (S3 state)
- [ ] Add deployment status reporting and error handling

**4.2 CI/CD Pipeline Steps**
- [ ] **Resource Discovery**: Scan repository for YAML files
- [ ] **Code Packaging**: Package Lambda functions and upload to S3
- [ ] **Schema Management**: Discover and upload manual OpenAPI schemas
- [ ] **Terraform Generation**: Create complete Terraform configuration
- [ ] **Terraform Deployment**: Plan and apply infrastructure changes

**4.3 Configuration Management**
- [ ] Support for multiple environments (dev, staging, prod)
- [ ] Environment-specific variable substitution
- [ ] Proper state file isolation per environment
- [ ] Team-specific IAM role and permission handling

**4.4 Workflow Features**
- [ ] Parallel processing for multiple resources
- [ ] Rollback capabilities on deployment failures
- [ ] Deployment approval workflows for production
- [ ] Integration with existing enterprise CI/CD systems

## Epic 6: IAM Security and Permission Management

### Epic Goal
Automatically generate all required IAM roles, policies, and trust relationships for secure Bedrock agent operations with proper least-privilege access.

### User Stories
- As a security engineer, I want all IAM permissions to follow least-privilege principles automatically
- As a developer, I want IAM complexity abstracted away so I can focus on business logic
- As a compliance team, I want consistent security patterns across all Bedrock deployments

### TODO Items

**6.1 Bedrock Agent IAM Roles**
- [x] Generate agent execution role with trust policy for Bedrock service
- [x] Create foundation model access policies (invoke model permissions)
- [x] Add CloudWatch logging permissions for agent execution
- [x] Support for auto-creation of basic execution roles
- [x] Support for manual IAM role definitions with full control
- [x] Support for existing IAM role ARN references

**6.2 IAM Role Configuration Options**
- [x] Auto-create basic roles with required permissions
- [x] Reference manually defined IAMRole resources
- [x] Use existing IAM role ARNs
- [x] Add additional managed policies to auto-created roles
- [x] Support for complex inline policies with multiple statements

**6.3 Manual IAM Role Definition**
- [x] Full IAM policy document support (assume role policies, inline policies, managed policies)
- [x] Support for complex IAM policy statements with conditions
- [x] Proper trust relationship configuration for Bedrock service
- [x] Tags and metadata support for IAM roles

**6.4 Enterprise Security Features**
- [x] Least-privilege access patterns built into auto-created roles
- [x] Consistent security patterns across all deployments
- [x] Support for enterprise compliance requirements through manual role definitions
- [x] Flexible policy attachment and customization

### IAM Policy Templates

**Bedrock Agent Role Policy Example:**
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "bedrock:InvokeModel",
        "bedrock:RetrieveAndGenerate"
      ],
      "Resource": [
        "arn:aws:bedrock:*::foundation-model/anthropic.claude-*",
        "arn:aws:bedrock:*:*:knowledge-base/*"
      ]
    },
    {
      "Effect": "Allow", 
      "Action": "lambda:InvokeFunction",
      "Resource": "arn:aws:lambda:*:*:function:bedrock-action-*"
    }
  ]
}
```

## Epic 5: Enterprise Features and Production Readiness ✅ COMPLETED

### Epic Goal ✅ COMPLETED
Add enterprise-grade features like governance, monitoring, security, and team management.

### User Stories ✅ COMPLETED
- As a security team, I want all deployments to follow approved security patterns ✅
- As a platform team, I want visibility into all deployed agents across the organization ✅

### TODO Items ✅ COMPLETED

**5.1 Security and Governance ✅ COMPLETED**
- [x] Implement resource naming convention enforcement
- [x] Add mandatory tagging for cost allocation and compliance
- [x] Integrate IAM policy templates and validation
- [x] Add security scanning for Lambda code and dependencies
- [x] Validate IAM policies against enterprise security standards

**5.2 Comprehensive Validation Framework ✅ COMPLETED**
- [x] Create flexible validation configuration system
- [x] Implement multiple validation profiles (default, enterprise, custom)
- [x] Add local validation configuration support
- [x] Create comprehensive error reporting and feedback

**5.3 Enterprise Standards Implementation ✅ COMPLETED**
- [x] Implement team-environment-name-type naming patterns
- [x] Add 12+ required enterprise tags with validation
- [x] Create strict security policies for production environments
- [x] Add tag value validation (allowed values, patterns, formats)
- [x] Implement resource-specific requirements

### Completed Features:

#### ✅ Resource Naming Convention Enforcement
- **Flexible Patterns**: Regex patterns, prefixes, suffixes, character restrictions
- **Team-Specific Rules**: Different conventions for different teams
- **Environment-Aware**: Rules that change based on deployment environment
- **Enterprise Pattern**: `{team}-{environment}-{name}-{type}` format

#### ✅ Mandatory Tagging for Cost Allocation and Compliance
- **Required Tags**: 12+ mandatory enterprise tags for all resources
- **Tag Validation**: Value validation against allowed lists, patterns, email formats
- **Resource-Specific**: Different requirements for Agents, Lambdas, Knowledge Bases
- **Cost Attribution**: Clear ownership and billing allocation

#### ✅ Security Policy Validation
- **IAM Policy Scanning**: Forbidden actions, wildcard restrictions, MFA requirements
- **Lambda Security**: VPC requirements, timeout limits, runtime restrictions
- **Agent Security**: Guardrail requirements, encryption, session limits
- **Network Security**: VPC configuration and security group validation

#### ✅ Validation Infrastructure
- **Multiple Profiles**: Default (development), Enterprise (production), Custom
- **Local Configuration**: Automatic detection of `validation.yml` files
- **Comprehensive Reporting**: Detailed error messages with field-level feedback
- **CI/CD Integration**: Ready for GitHub Actions and pre-commit hooks

### Implementation Results:
- **18 Error Types Detected** in comprehensive testing
- **3 Validation Categories**: Naming, Tagging, Security
- **Enterprise-Ready**: Strict patterns and comprehensive coverage
- **Developer-Friendly**: Clear error messages and flexible configuration

## Epic 5: Enterprise Features and Production Readiness (Remaining)

### Epic Goal
Add enterprise-grade features like governance, monitoring, security, and team management.

### User Stories
- As a security team, I want all deployments to follow approved security patterns
- As a platform team, I want visibility into all deployed agents across the organization

### TODO Items

**5.1 Security and Governance**
- [ ] Implement resource naming convention enforcement
- [ ] Add mandatory tagging for cost allocation and compliance
- [ ] Integrate with Epic 6 IAM policy templates and validation
- [ ] Add security scanning for Lambda code and dependencies
- [ ] Validate IAM policies against enterprise security standards

**5.2 Module Registry and Versioning**
- [ ] Create module registry for hosting Terraform modules
- [ ] Implement semantic versioning for modules
- [ ] Add module compatibility checking
- [ ] Create module documentation and examples

**5.3 Monitoring and Observability**
- [ ] Add deployment metrics and logging
- [ ] Create resource inventory and tracking
- [ ] Implement cost monitoring and reporting
- [ ] Add performance metrics for deployment times

**5.4 Team Management**
- [ ] Multi-tenant support with team isolation
- [ ] Role-based access control integration
- [ ] Team-specific module repositories
- [ ] Approval workflows and governance policies

## Quick Start Command for Claude Code

```bash
# Initial command to get started
bedrock-forge init --template enterprise
```

## Example YAML Configurations

### forge.yml (Project Configuration)
```yaml
metadata:
  name: "customer-support-platform"
  team: "customer-experience"
  environment: "dev"

terraform:
  backend:
    type: "s3"
    bucket: "company-terraform-state"
    key: "bedrock-agents/{team}/{environment}/{name}/terraform.tfstate"
    region: "us-east-1"
  executionRole: "arn:aws:iam::123456789012:role/TerraformExecutionRole"

scanning:
  paths: ["./src", "./agents", "./knowledge-bases"]
  include: ["*.yml", "*.yaml"]
  exclude: ["**/node_modules/**", "**/.git/**"]

modules:
  registry: "git::https://github.com/company/bedrock-terraform-modules"
  version: "v1.2.0"
```

### Example Resource Files
```yaml
# agents/customer-support/agent.yml
kind: Agent
metadata:
  name: "customer-support"
spec:
  foundationModel: "anthropic.claude-3-sonnet-20240229-v1:0"
  instruction: "You are a helpful customer support agent..."
  description: "Customer support agent for order management"
  idleSessionTtlInSeconds: 3600
  
  # Guardrail configuration (existing or new)
  guardrail:
    name: "content-safety-guardrail"  # References existing guardrail
    version: "1"
    mode: "pre"  # pre, post, or both
  
  # Knowledge base associations
  knowledgeBases: 
    - name: "faq-kb"
      description: "Customer FAQ knowledge base"
  
  # Prompt overrides
  promptOverrides:
    - promptType: "PRE_PROCESSING"
      promptArn: "arn:aws:bedrock:us-east-1:123456789012:prompt/custom-preprocessing"
      promptVariant: "v1"
    - promptType: "ORCHESTRATION"
      prompt: "custom-orchestration-prompt"
      variant: "production"
  
  # Memory configuration
  memoryConfiguration:
    enabledMemoryTypes: ["SESSION_SUMMARY"]
    storageDays: 30

---
# guardrails/content-safety/guardrail.yml
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
      - type: "HATE"
        inputStrength: "HIGH"
        outputStrength: "HIGH"
      - type: "INSULTS"
        inputStrength: "MEDIUM"
        outputStrength: "HIGH"
      - type: "MISCONDUCT"
        inputStrength: "MEDIUM"
        outputStrength: "MEDIUM"
      - type: "PROMPT_ATTACK"
        inputStrength: "HIGH"
        outputStrength: "NONE"
  
  sensitiveInformationPolicyConfig:
    piiEntitiesConfig:
      - type: "EMAIL"
        action: "BLOCK"
      - type: "PHONE"
        action: "ANONYMIZE"
      - type: "SSN"
        action: "BLOCK"
  
  contextualGroundingPolicyConfig:
    filtersConfig:
      - type: "GROUNDING"
        threshold: 0.8
      - type: "RELEVANCE" 
        threshold: 0.7
  
  topicPolicyConfig:
    topicsConfig:
      - name: "Investment Advice"
        definition: "Discussions about financial investments or trading advice"
        examples: ["Should I buy this stock?", "What's the best crypto to invest in?"]
        type: "DENY"
  
  wordPolicyConfig:
    wordsConfig:
      - text: "competitor_name"
    managedWordListsConfig:
      - type: "PROFANITY"

---
# knowledge-bases/faq/knowledge-base.yml
kind: KnowledgeBase
metadata:
  name: "faq-kb"
spec:
  description: "Customer FAQ knowledge base"
  
  knowledgeBaseConfiguration:
    type: "VECTOR"
    vectorKnowledgeBaseConfiguration:
      embeddingModelArn: "arn:aws:bedrock:us-east-1::foundation-model/amazon.titan-embed-text-v1"
      embeddingModelConfiguration:
        bedrockEmbeddingModelConfiguration:
          dimensions: 1536
  
  storageConfiguration:
    type: "OPENSEARCH_SERVERLESS"
    opensearchServerlessConfiguration:
      collectionArn: "arn:aws:aoss:us-east-1:123456789012:collection/bedrock-kb"
      vectorIndexName: "bedrock-knowledge-base-index"
      fieldMapping:
        vectorField: "vector"
        textField: "text"
        metadataField: "metadata"
  
  # Data source configuration
  dataSources:
    - name: "faq-documents"
      type: "S3"
      s3Configuration:
        bucketArn: "arn:aws:s3:::company-kb-documents"
        inclusionPrefixes: ["faq/"]
      
      # Chunking strategy
      chunkingConfiguration:
        chunkingStrategy: "FIXED_SIZE"
        fixedSizeChunkingConfiguration:
          maxTokens: 512
          overlapPercentage: 20
      
      # Optional pre/post processing
      vectorIngestionConfiguration:
        chunkingConfiguration:
          chunkingStrategy: "SEMANTIC"
          semanticChunkingConfiguration:
            maxTokens: 300
            bufferSize: 1
            breakpointPercentileThreshold: 95
      
      # Custom transformation (optional)
      customTransformation:
        transformationLambda:
          lambdaArn: "arn:aws:lambda:us-east-1:123456789012:function:kb-preprocessor"
        intermediateStorage:
          s3Location:
            uri: "s3://company-kb-temp/transformations/"

---
# prompts/custom-orchestration/prompt.yml
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
          
          Instructions:
          1. Greet the customer warmly
          2. Listen to their concern
          3. Provide accurate information
          4. Offer additional assistance
          
          Customer Query: {{query}}
          
          Context: {{context}}
      
      inferenceConfiguration:
        text:
          temperature: 0.1
          topP: 0.9
          maxTokens: 2048
          stopSequences: ["Human:", "Assistant:"]
    
    - name: "development"
      modelId: "anthropic.claude-3-sonnet-20240229-v1:0"
      templateType: "TEXT"
      templateConfiguration:
        text: |
          [DEBUG MODE] Customer Support Agent
          
          Query: {{query}}
          Context: {{context}}
          
          Debug info will be included in responses.
      
      inferenceConfiguration:
        text:
          temperature: 0.2
          topP: 0.95
          maxTokens: 1024
```

## Development Priority
1. **Start with Epic 1** - Get basic YAML parsing and resource discovery working
2. **Move to Epic 2** - Build the core Terraform generation engine  
3. **Add Epic 3** - Implement Lambda packaging and OpenAPI generation
4. **Integrate Epic 4** - Create GitHub Actions workflow
5. **Implement Epic 6** - Add comprehensive IAM security management
6. **Polish with Epic 5** - Add remaining enterprise features

## Success Criteria
- Teams can deploy complete Bedrock agents with a single GitHub Actions workflow
- Zero custom Terraform code required from development teams
- 90% reduction in deployment time compared to manual Terraform development
- Enterprise security and governance standards automatically enforced