# Internal Architecture Guide - Bedrock Forge

This document provides a comprehensive walkthrough of the `internal/` folder structure and the types of logic each package contains in Bedrock Forge.

## 📁 `/internal/` Folder Structure Overview

### 🔧 `/internal/models/` - **Data Structures & Types**
**Purpose:** Contains all the Go structs that represent Bedrock resources.

**Contains:**
- `types.go` - All resource type definitions
  - `Agent`, `Lambda`, `ActionGroup`, `KnowledgeBase`, `Guardrail`, `Prompt`, `IAMRole` structs
  - Complete AWS Bedrock API mappings with YAML tags
  - Complex nested structures (e.g., `VectorKnowledgeBaseConfiguration`, `ContentPolicyConfig`)
  - 430+ lines of comprehensive type definitions

**Logic Type:** Pure data models - no business logic, just struct definitions with YAML/JSON serialization tags

**Key Structs:**
```go
type Agent struct {
    Kind     ResourceKind `yaml:"kind"`
    Metadata Metadata     `yaml:"metadata"`
    Spec     AgentSpec    `yaml:"spec"`
}

type AgentSpec struct {
    FoundationModel         string
    Instruction             string
    Guardrail               *GuardrailConfig
    KnowledgeBases          []KnowledgeBaseReference
    IAMRole                 *IAMRoleConfig
    // ... many more fields
}
```

---

### 🖥️ `/internal/commands/` - **CLI Command Implementations**
**Purpose:** Contains the actual CLI command logic and user interface.

**Contains:**
- `scan.go` - Discovers and catalogs YAML resources in directories
- `validate.go` - Runs comprehensive validation (naming, tagging, security)
- `generate.go` - Orchestrates Terraform generation from YAML resources

**Logic Type:** Command orchestration - each file implements a CLI subcommand, coordinates other packages, handles user input/output

**Example Flow:**
```go
func (s *ScanCommand) Execute(rootPath string) error {
    // 1. Coordinate scanner to find files
    files := s.scanner.ScanDirectory(rootPath)
    
    // 2. Parse each file using YAML parser
    for _, file := range files {
        resources := s.yamlParser.ParseFile(file)
        
        // 3. Register resources for dependency tracking
        s.registry.AddResource(resource)
    }
    
    // 4. Present results to user
    s.printResults()
}
```

---

### 🔍 `/internal/parser/` - **YAML Processing & File Discovery**
**Purpose:** Finds and parses YAML files into Go structs.

**Contains:**
- `scanner.go` - Recursive directory scanning for `.yml`/`.yaml` files
  - Configurable include/exclude patterns
  - Ignores common directories (`node_modules`, `.git`, `.terraform`)
  - Concurrent file processing for performance
- `yaml_parser.go` - Converts YAML content into typed Go structs
  - Handles multi-document YAML files (documents separated by `---`)
  - Type-safe parsing with validation
  - Error reporting with file/line context
  - Resource kind detection and routing

**Logic Type:** Input processing - file I/O, YAML unmarshaling, content validation

**Key Features:**
- **Multi-document Support**: Single YAML file can contain multiple resources
- **Type Safety**: Validates YAML structure against Go structs
- **Error Context**: Provides file paths and line numbers for debugging

---

### 📊 `/internal/registry/` - **Resource Management & Dependencies**
**Purpose:** Central store for discovered resources with dependency tracking.

**Contains:**
- `resource_registry.go` - In-memory registry of all discovered resources
  - Thread-safe resource storage with mutexes
  - Dependency validation (e.g., Agent → Guardrail references)
  - Resource lookup by kind and name
  - Duplicate detection and conflict resolution

**Logic Type:** Data management - resource storage, relationship tracking, dependency resolution

**Key Capabilities:**
```go
// Store resources by kind and name
resources map[ResourceKind]map[string]*ParsedResource

// Validate dependencies
func (r *Registry) ValidateDependencies() []error {
    // Check that referenced resources exist
    // e.g., Agent references existing Guardrail
}
```

**Dependency Examples:**
- `Agent` → `Guardrail` (guardrail must exist)
- `Agent` → `KnowledgeBase` (knowledge base must exist)
- `Agent` → `Prompt` (prompt must exist for overrides)
- `ActionGroup` → `Lambda` (lambda must exist for executor)

---

### ⚙️ `/internal/generator/` - **Terraform Code Generation**
**Purpose:** Transforms parsed YAML resources into Terraform HCL modules.

**Contains:**
- `hcl_generator.go` - Main coordinator and HCL file writer
  - Dependency ordering (Guardrails → Prompts → Lambdas → Agents)
  - Provider configuration and version constraints
  - Variable definitions and output values
- `agent_generator.go` - Generates Terraform for Bedrock agents
- `lambda_generator.go` - Generates Terraform for Lambda functions  
- `action_group_generator.go` - Generates Terraform for action groups
- `knowledge_base_generator.go` - Generates Terraform for knowledge bases
- `guardrail_generator.go` - Generates Terraform for guardrails
- `prompt_generator.go` - Generates Terraform for prompts
- `iam_role_generator.go` - Generates Terraform for IAM roles
- `generation_context.go` - Shared context for artifact management

**Logic Type:** Code generation - template processing, HCL syntax generation, module references, dependency ordering

**Generated Output Example:**
```hcl
module "customer_support_agent" {
  source = "git::https://github.com/org/terraform-modules//bedrock-agent?ref=v1.2.0"
  
  name               = "customer-support"
  foundation_model   = "anthropic.claude-3-sonnet-20240229-v1:0"
  instruction        = "You are a helpful customer support agent"
  
  guardrail_id      = module.content_safety_guardrail.guardrail_id
  knowledge_base_id = module.faq_kb.knowledge_base_id
  
  tags = {
    Environment = "dev"
    Project     = "customer-support"
  }
}
```

---

### 📦 `/internal/packager/` - **Artifact Management**
**Purpose:** Handles Lambda code packaging and schema processing for deployment.

**Contains:**
- `lambda_packager.go` - ZIP packaging of Lambda source code
  - Directory-based discovery (co-located with `lambda.yml`)
  - File exclusion patterns (`.git`, `node_modules`, `*.yml`, etc.)
  - Unique versioned S3 keys (timestamp + hash)
  - Dependency installation for Python/Node.js
- `s3_client.go` - S3 upload client for artifacts
  - Configurable bucket and key prefixes
  - Retry logic and error handling
  - Object metadata and tagging
- `schema_extractor.go` - OpenAPI schema discovery and upload
  - Finds manual schema files (`openapi.json`, `schema.json`, `api.json`)
  - Validates schema format and structure
  - Uploads to S3 for ActionGroup integration

**Logic Type:** Artifact processing - file compression, cloud storage, content discovery

**Workflow:**
1. **Discover**: Find Lambda directories containing source code
2. **Package**: Create ZIP files with proper exclusions
3. **Upload**: Store artifacts in S3 with unique keys
4. **Reference**: Update Terraform generation with S3 URIs

---

### ✅ `/internal/validation/` - **Enterprise Governance & Compliance**
**Purpose:** Enforces organizational standards and security policies.

**Contains:**
- `validator.go` - Main validation coordinator with multiple profiles
  - Default profile (development-friendly)
  - Enterprise profile (production-ready)
  - Custom profile (local configuration files)
  - Comprehensive error reporting with field-level feedback
- `naming_conventions.go` - Resource naming pattern enforcement
  - Regex patterns, prefixes/suffixes, character restrictions
  - Team-specific and environment-specific rules
  - Enterprise patterns like `{team}-{environment}-{name}-{type}`
  - Length constraints and case enforcement
- `tagging_policies.go` - Mandatory tagging requirements
  - Required tags for cost allocation and compliance
  - Resource-specific tagging requirements
  - Tag value validation (allowed values, email formats, patterns)
  - 12+ enterprise tags with comprehensive validation
- `security_policies.go` - Security and compliance validation
  - IAM policy scanning (forbidden actions, MFA requirements)
  - Lambda security (VPC requirements, timeout limits, runtime restrictions)
  - Agent requirements (guardrails, encryption, memory configuration)
  - Network security and encryption validation

**Logic Type:** Policy enforcement - rule validation, compliance checking, security scanning

**Validation Categories:**
1. **Naming Conventions**: Ensure consistent resource naming
2. **Tagging Policies**: Enforce cost allocation and compliance tags
3. **Security Policies**: Validate security configurations and requirements

**Example Validation Results:**
```
❌ Validation failed with 3 errors:

1. [naming_convention] Agent names must follow pattern: <team>-<env>-<name>-agent
   Resource: Agent/bad-agent
   Field: metadata.name

2. [tagging_policy] Required tag 'Environment' is missing
   Resource: Agent/bad-agent
   Field: spec.tags.Environment

3. [security_policy] Bedrock agents must have guardrails configured
   Resource: Agent/bad-agent
   Field: spec.guardrail
```

---

## 🏗️ **Architecture Flow**

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   CLI Commands  │───▶│     Parser      │───▶│    Registry     │
│                 │    │                 │    │                 │
│ • scan.go       │    │ • scanner.go    │    │ • resource_     │
│ • validate.go   │    │ • yaml_parser.go│    │   registry.go   │
│ • generate.go   │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                                              │
         │                                              ▼
         │              ┌─────────────────┐    ┌─────────────────┐
         │              │   Validation    │◀───│     Models      │
         │              │                 │    │                 │
         │              │ • validator.go  │    │ • types.go      │
         │              │ • naming_*.go   │    │                 │
         │              │ • tagging_*.go  │    │ (All structs)   │
         │              │ • security_*.go │    │                 │
         │              └─────────────────┘    └─────────────────┘
         │                                              │
         ▼                                              │
┌─────────────────┐    ┌─────────────────┐              │
│    Packager     │───▶│   Generator     │◀─────────────┘
│                 │    │                 │
│ • lambda_       │    │ • hcl_generator │
│   packager.go   │    │ • *_generator.go│
│ • s3_client.go  │    │ • generation_   │
│ • schema_       │    │   context.go    │
│   extractor.go  │    │                 │
└─────────────────┘    └─────────────────┘
```

**Data Flow:**
1. **Commands** orchestrate the overall process and handle user interaction
2. **Parser** discovers and reads YAML files using **Models** for type safety
3. **Registry** stores and manages resources with dependency tracking
4. **Validation** enforces enterprise policies and compliance using **Models**
5. **Packager** prepares Lambda code and schemas for deployment
6. **Generator** creates Terraform modules using **Models** and packaged artifacts

**Models** serves as the foundation - used by all other packages for type safety and consistent structure definitions.

## 🎯 **Design Principles**

### Clean Architecture
- **Separation of Concerns**: Each package has a single, well-defined responsibility
- **Dependency Inversion**: Core business logic doesn't depend on external frameworks
- **Interface Segregation**: Packages expose minimal, focused interfaces

### Enterprise Scalability
- **Thread Safety**: Registry uses mutexes for concurrent access
- **Error Handling**: Comprehensive error reporting with context
- **Extensibility**: Easy to add new resource types and validation rules
- **Configuration**: Multiple profiles for different environments

### Developer Experience
- **Type Safety**: Strong typing prevents runtime errors
- **Clear Feedback**: Detailed error messages with file/line context
- **Documentation**: Self-documenting code with clear interfaces
- **Testing**: Modular design enables comprehensive unit testing

This architecture enables Bedrock Forge to handle enterprise-scale deployments while maintaining simplicity for development teams.