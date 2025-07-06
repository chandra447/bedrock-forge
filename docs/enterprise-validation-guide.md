# Enterprise Validation Guide

## Overview

Bedrock Forge provides comprehensive enterprise-grade validation capabilities that enforce organizational standards for resource naming, tagging, and security compliance. This validation framework helps ensure that all Bedrock agent deployments follow enterprise policies while providing clear feedback to development teams.

## Features

### ✅ Resource Naming Convention Enforcement
- **Flexible Patterns**: Support for regex patterns, prefixes, suffixes, and character restrictions
- **Team-Specific Rules**: Different naming conventions for different teams
- **Environment-Aware**: Naming rules that change based on deployment environment
- **Comprehensive Coverage**: All resource types (Agents, Lambdas, Action Groups, etc.)

### ✅ Mandatory Tagging for Cost Allocation and Compliance
- **Required Tags**: Enforce mandatory tags for cost tracking and compliance
- **Tag Validation**: Validate tag values against allowed lists, patterns, and formats
- **Resource-Specific Requirements**: Different tagging requirements for different resource types
- **Enterprise Standards**: Built-in support for common enterprise tagging patterns

### ✅ Security Policy Validation
- **IAM Policy Scanning**: Detect forbidden actions, overly permissive policies, and security violations
- **Lambda Security**: Enforce VPC requirements, timeout limits, and runtime restrictions
- **Agent Security**: Require guardrails, encryption, and memory configuration
- **Network Security**: Validate VPC configurations and security group requirements

## Validation Profiles

### Default Profile
Suitable for development environments with flexible requirements:
- Basic naming conventions (prefixes/suffixes)
- Essential tags (Environment, Project, Owner)
- Relaxed security policies

### Enterprise Profile
Strict validation for production environments:
- Comprehensive naming patterns (team-env-name-type format)
- Complete tagging requirements (12+ required tags)
- Strict security policies (guardrails required, encryption mandatory)

### Custom Profile
Load validation rules from a local `validation.yml` file for complete customization.

## Configuration

### Built-in Profiles

```bash
# Use default validation (development-friendly)
./bedrock-forge validate

# Use enterprise validation (production-ready)
./bedrock-forge validate --profile enterprise

# Use custom validation configuration
./bedrock-forge validate --config path/to/validation.yml
```

### Local Configuration

Place a `validation.yml` file in your project root for automatic custom validation:

```yaml
enabledValidators:
  - naming
  - tagging
  - security

namingConventions:
  global:
    minLength: 5
    maxLength: 50
    allowedChars: "a-z0-9-"
    forceLowercase: true
    pattern: "^[a-z][a-z0-9-]*[a-z0-9]$"

  resources:
    Agent:
      pattern: "^[a-z]+-(dev|staging|prod)-[a-z0-9-]+-agent$"
      validationMessage: "Agent names must follow pattern: <team>-<env>-<name>-agent"

taggingPolicies:
  global:
    requiredTags:
      - Environment
      - Project
      - Owner
      - CostCenter
      - Team

  tagValidation:
    Environment:
      allowedValues: ["dev", "staging", "prod"]
      caseSensitive: false

securityPolicies:
  agentSecurity:
    requireGuardrails: true
    maxIdleSessionTTL: 1800
    requireCustomerEncryption: true
```

## Validation Results

### Success Output
```
✅ All resources are valid!
   └─ 5 resources passed validation

⚠️  2 warnings:
   1. Optional tag 'CostCenter' is missing (recommended for compliance)
   2. Optional tag 'MonitoringLevel' is missing (recommended for compliance)
```

### Failure Output
```
❌ Validation failed with 12 errors:

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

## Enterprise Standards

### Naming Conventions

#### Pattern: `{team}-{environment}-{name}-{type}`

Examples:
- `data-dev-customer-support-agent`
- `engineering-prod-order-lookup-lambda`
- `security-staging-content-filter-guardrail`

#### Benefits:
- **Searchability**: Easy to find resources by team or environment
- **Cost Attribution**: Clear ownership for billing and cost tracking
- **Security**: Consistent patterns make security policies easier to enforce
- **Automation**: Predictable names enable automated tools and scripts

### Required Tags

#### Core Enterprise Tags
```yaml
tags:
  Environment: "prod"                    # dev, staging, prod
  Project: "customer-support-platform"  # Project identifier
  Owner: "john.doe@company.com"         # Resource owner email
  CostCenter: "CC-123456"               # 6-digit cost center
  Team: "data"                          # Organizational team
  BusinessUnit: "engineering"           # Business unit
  DataClassification: "internal"        # Data sensitivity level
  ComplianceLevel: "pci"                # Applicable compliance frameworks
```

#### Resource-Specific Tags
```yaml
# For Agents
AgentType: "conversational"           # Type of agent
BusinessFunction: "customer-service"  # Business purpose
DataProcessing: "conversational-ai"   # Data processing type
SecurityLevel: "high"                 # Security classification

# For Lambdas  
Runtime: "python3.11"                # Lambda runtime
FunctionType: "api"                   # Function category
ExecutionRole: "lambda-execution"     # IAM role type

# For Knowledge Bases
DataSource: "s3"                      # Source of knowledge
ContentType: "faq"                    # Type of content
DataSensitivity: "internal"           # Data sensitivity
RetentionPeriod: "7years"             # Data retention policy
```

## Security Policies

### IAM Security
- **Forbidden Actions**: Prevents overly permissive IAM policies
- **Wildcard Restrictions**: Blocks `*` resources in production
- **MFA Requirements**: Enforces MFA for sensitive operations
- **Admin Prevention**: Blocks admin-level permissions

### Lambda Security
- **VPC Requirements**: Forces Lambda functions into VPCs
- **Runtime Restrictions**: Only allows approved runtime versions
- **Timeout Limits**: Prevents excessive execution times
- **Environment Scanning**: Detects secrets in environment variables

### Agent Security
- **Guardrail Requirements**: Mandates content safety guardrails
- **Encryption Requirements**: Enforces customer-managed encryption
- **Session Limits**: Restricts idle session timeouts
- **Model Restrictions**: Blocks non-approved foundation models

## Integration with CI/CD

### GitHub Actions Integration

```yaml
name: Validate Bedrock Resources

on:
  pull_request:
    paths:
      - 'agents/**'
      - 'lambdas/**'
      - 'action-groups/**'

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Download Bedrock Forge
        run: |
          curl -L https://github.com/org/bedrock-forge/releases/latest/download/bedrock-forge-linux-amd64 -o bedrock-forge
          chmod +x bedrock-forge
      
      - name: Validate Resources (Enterprise)
        run: |
          ./bedrock-forge validate --profile enterprise
```

### Pre-commit Hooks

```yaml
# .pre-commit-config.yaml
repos:
  - repo: local
    hooks:
      - id: bedrock-forge-validate
        name: Validate Bedrock Resources
        entry: ./bedrock-forge validate --profile enterprise
        language: system
        pass_filenames: false
        files: \.(yml|yaml)$
```

## Benefits

### For Development Teams
- **Clear Standards**: Automated enforcement of naming and tagging conventions
- **Early Feedback**: Catch policy violations before deployment
- **Reduced Errors**: Prevent common configuration mistakes
- **Faster Reviews**: Automated validation reduces manual review overhead

### For Platform Teams
- **Consistent Standards**: Ensure all deployments follow organizational policies
- **Cost Visibility**: Mandatory tagging enables accurate cost allocation
- **Security Compliance**: Automated security policy enforcement
- **Audit Trail**: Complete validation logs for compliance reporting

### For Security Teams
- **Policy Enforcement**: Automated security policy validation
- **Risk Reduction**: Prevent insecure configurations from reaching production
- **Compliance**: Built-in support for regulatory requirements
- **Visibility**: Clear reporting on security policy violations

## Troubleshooting

### Common Validation Errors

#### Naming Convention Violations
```
[naming_convention] Resource name 'my-agent' must follow pattern: <team>-<env>-<name>-agent
```
**Solution**: Rename resource to follow the enterprise pattern, e.g., `data-dev-my-agent`

#### Missing Required Tags
```
[tagging_policy] Required tag 'Environment' is missing
```
**Solution**: Add all required tags to the resource specification

#### Security Policy Violations
```
[security_policy] Bedrock agents must have guardrails configured
```
**Solution**: Add guardrail configuration to the agent specification

### Debugging Validation

Enable debug logging for detailed validation information:
```bash
./bedrock-forge validate --log-level debug
```

This provides detailed information about:
- Which validation rules are being applied
- Why specific validations are failing
- Complete validation context and settings

---

**Bedrock Forge Enterprise Validation** - Ensuring enterprise-grade compliance and security for AWS Bedrock deployments.