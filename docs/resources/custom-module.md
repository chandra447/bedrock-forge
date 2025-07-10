# Custom Module Resource (Deprecated)

⚠️ **DEPRECATED**: This resource type has been replaced with `CustomResources`. See [custom-resources.md](./custom-resources.md) for the new approach.

Integration with existing Terraform modules alongside Bedrock resources for comprehensive infrastructure deployment.

## Overview

Custom Module resources allow you to integrate your existing Terraform infrastructure with Bedrock resources. This enables you to deploy supporting infrastructure (VPC, storage, monitoring) alongside your Bedrock agents in a unified workflow.

## Basic Example

```yaml
kind: CustomModule
metadata:
  name: "s3-storage"
  description: "S3 bucket for storing agent artifacts"
spec:
  source: "./modules/s3-bucket"  # Local module
  
  variables:
    bucket_name: "bedrock-artifacts-${var.environment}"
    versioning_enabled: true
    tags:
      Purpose: "BedrockStorage"
      Environment: "${var.environment}"
```

## Complete Example

```yaml
kind: CustomModule
metadata:
  name: "vpc-infrastructure"
  description: "VPC infrastructure for Bedrock agents"
spec:
  # Module source (local, registry, or git)
  source: "terraform-aws-modules/vpc/aws"
  version: "5.0.0"  # For registry/git modules
  
  # Input variables
  variables:
    name: "bedrock-agent-vpc"
    cidr: "10.0.0.0/16"
    azs: ["us-east-1a", "us-east-1b", "us-east-1c"]
    private_subnets: ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
    public_subnets: ["10.0.101.0/24", "10.0.102.0/24", "10.0.103.0/24"]
    
    enable_nat_gateway: true
    enable_vpn_gateway: false
    enable_dns_hostnames: true
    enable_dns_support: true
    
    # Cross-module references
    additional_tags:
      Project: "${var.project_name}"
      Environment: "${var.environment}"
      ManagedBy: "bedrock-forge"
  
  # Dependencies (optional)
  dependsOn:
    - "base-networking"  # Wait for base networking first
  
  # Output values can be referenced by other resources
  outputs:
    vpc_id: "${module.vpc_infrastructure.vpc_id}"
    private_subnet_ids: "${module.vpc_infrastructure.private_subnets}"
    public_subnet_ids: "${module.vpc_infrastructure.public_subnets}"
  
  # Resource tags
  tags:
    Component: "networking"
    Team: "platform"
```

## Specification

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `source` | string | Module source (local path, registry, or git) |

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `version` | string | Module version (for registry/git sources) |
| `variables` | object | Input variables for the module |
| `dependsOn` | array | List of dependencies |
| `outputs` | object | Output values to expose |
| `tags` | object | Resource tags |

### Module Sources

#### Local Modules

```yaml
spec:
  source: "./modules/my-module"
  # Local path relative to project root
```

#### Terraform Registry

```yaml
spec:
  source: "terraform-aws-modules/vpc/aws"
  version: "5.0.0"  # Semantic version
```

#### Git Repositories

```yaml
spec:
  source: "git::https://github.com/org/terraform-modules.git//modules/vpc"
  version: "v1.2.0"  # Git tag or branch
  
  # With SSH
  # source: "git::ssh://git@github.com/org/terraform-modules.git//modules/vpc"
  
  # With specific subdirectory
  # source: "git::https://github.com/org/terraform-modules.git//infrastructure/vpc"
```

### Variable Types

#### Primitive Types

```yaml
variables:
  # String
  bucket_name: "my-bucket"
  
  # Number
  instance_count: 3
  memory_size: 512
  
  # Boolean
  enable_logging: true
  create_vpc: false
```

#### Complex Types

```yaml
variables:
  # Lists
  availability_zones: ["us-east-1a", "us-east-1b"]
  allowed_cidr_blocks: ["10.0.0.0/8", "172.16.0.0/12"]
  
  # Objects
  vpc_config:
    cidr: "10.0.0.0/16"
    enable_dns: true
    name: "bedrock-vpc"
  
  # Maps
  tags:
    Environment: "production"
    Team: "platform"
    Project: "bedrock-agents"
```

#### Variable References

```yaml
variables:
  # Reference global variables
  environment: "${var.environment}"
  project_name: "${var.project_name}"
  
  # Reference other module outputs
  vpc_id: "${module.vpc_infrastructure.vpc_id}"
  subnet_ids: "${module.vpc_infrastructure.private_subnets}"
  
  # Combine references
  bucket_name: "${var.project_name}-${var.environment}-artifacts"
```

### Dependencies

```yaml
dependsOn:
  - "base-networking"     # Custom module name
  - "security-groups"     # Another custom module
  - "iam-roles"          # Wait for IAM setup
```

## Common Patterns

### VPC Infrastructure

```yaml
kind: CustomModule
metadata:
  name: "vpc-setup"
spec:
  source: "terraform-aws-modules/vpc/aws"
  version: "5.0.0"
  
  variables:
    name: "bedrock-vpc"
    cidr: "10.0.0.0/16"
    azs: ["us-east-1a", "us-east-1b"]
    private_subnets: ["10.0.1.0/24", "10.0.2.0/24"]
    public_subnets: ["10.0.101.0/24", "10.0.102.0/24"]
    
    enable_nat_gateway: true
    single_nat_gateway: false
    enable_vpn_gateway: false
    
    tags:
      Purpose: "BedrockAgents"
```

### S3 Storage

```yaml
kind: CustomModule
metadata:
  name: "s3-storage"
spec:
  source: "terraform-aws-modules/s3-bucket/aws"
  version: "3.0.0"
  
  variables:
    bucket: "bedrock-artifacts-${var.environment}"
    
    versioning:
      enabled: true
    
    server_side_encryption_configuration:
      rule:
        apply_server_side_encryption_by_default:
          sse_algorithm: "AES256"
    
    public_access_block:
      block_public_acls: true
      block_public_policy: true
      ignore_public_acls: true
      restrict_public_buckets: true
```

### OpenSearch Serverless

```yaml
kind: CustomModule
metadata:
  name: "opensearch-serverless"
spec:
  source: "./modules/opensearch-serverless"
  
  variables:
    collection_name: "bedrock-knowledge-base"
    
    network_policy:
      type: "encryption"
      rules:
        - resource_type: "collection"
          resource: ["collection/bedrock-knowledge-base"]
    
    encryption_policy:
      type: "encryption"
      rules:
        - resource_type: "collection"
          resource: ["collection/bedrock-knowledge-base"]
    
    access_policy:
      type: "data"
      rules:
        - resource_type: "collection"
          resource: ["collection/bedrock-knowledge-base"]
          permissions: ["aoss:*"]
  
  dependsOn:
    - "vpc-setup"
```

### Monitoring Setup

```yaml
kind: CustomModule
metadata:
  name: "monitoring"
spec:
  source: "./modules/cloudwatch-dashboard"
  
  variables:
    dashboard_name: "BedrockAgentMonitoring"
    
    widgets:
      - type: "metric"
        properties:
          metrics:
            - ["AWS/Lambda", "Duration", "FunctionName", "${module.customer_lambda.function_name}"]
            - ["AWS/Lambda", "Errors", "FunctionName", "${module.customer_lambda.function_name}"]
          period: 300
          stat: "Average"
          region: "us-east-1"
          title: "Lambda Performance"
      
      - type: "log"
        properties:
          query: |
            SOURCE '/aws/bedrock/agents/${module.customer_agent.agent_name}'
            | fields @timestamp, @message
            | sort @timestamp desc
            | limit 100
          title: "Agent Logs"
  
  dependsOn:
    - "customer-agent"    # Wait for agent to be created
    - "customer-lambda"   # Wait for Lambda to be created
```

### Database Setup

```yaml
kind: CustomModule
metadata:
  name: "rds-database"
spec:
  source: "terraform-aws-modules/rds/aws"
  version: "6.0.0"
  
  variables:
    identifier: "bedrock-agent-db"
    
    engine: "postgres"
    engine_version: "15.4"
    family: "postgres15"
    major_engine_version: "15"
    instance_class: "db.t3.micro"
    
    allocated_storage: 20
    max_allocated_storage: 100
    
    db_name: "bedrock_agents"
    username: "bedrock_user"
    manage_master_user_password: true
    
    # VPC configuration
    db_subnet_group_name: "${module.vpc_setup.database_subnet_group}"
    vpc_security_group_ids: ["${module.security_groups.database_sg_id}"]
    
    # Backup configuration
    backup_retention_period: 7
    backup_window: "03:00-04:00"
    maintenance_window: "Mon:04:00-Mon:05:00"
    
    deletion_protection: true
    
    tags:
      Purpose: "BedrockAgentData"
  
  dependsOn:
    - "vpc-setup"
    - "security-groups"
```

## Integration with Bedrock Resources

### Complete Infrastructure Stack

```yaml
# 1. VPC Infrastructure
kind: CustomModule
metadata:
  name: "vpc"
spec:
  source: "terraform-aws-modules/vpc/aws"
  version: "5.0.0"
  variables:
    name: "bedrock-vpc"
    cidr: "10.0.0.0/16"
    azs: ["us-east-1a", "us-east-1b"]
    private_subnets: ["10.0.1.0/24", "10.0.2.0/24"]
    public_subnets: ["10.0.101.0/24", "10.0.102.0/24"]
    enable_nat_gateway: true

---
# 2. OpenSearch for Knowledge Base
kind: CustomModule
metadata:
  name: "opensearch"
spec:
  source: "./modules/opensearch-serverless"
  dependsOn: ["vpc"]
  variables:
    collection_name: "bedrock-kb"
    vpc_id: "${module.vpc.vpc_id}"
    subnet_ids: "${module.vpc.private_subnets}"

---
# 3. Lambda Function
kind: Lambda
metadata:
  name: "customer-tools"
spec:
  runtime: "python3.11"
  handler: "app.handler"
  vpcConfig:
    securityGroupIds: ["${module.vpc.default_security_group_id}"]
    subnetIds: "${module.vpc.private_subnets}"

---
# 4. Bedrock Agent
kind: Agent
metadata:
  name: "customer-agent"
spec:
  foundationModel: "anthropic.claude-3-sonnet-20240229-v1:0"
  instruction: "You are a customer support agent"
  actionGroups:
    - name: "tools"
      description: "Customer tools"
      actionGroupExecutor:
        lambda: "customer-tools"

---
# 5. Monitoring
kind: CustomModule
metadata:
  name: "monitoring"
spec:
  source: "./modules/monitoring"
  dependsOn: ["customer-agent", "customer-tools"]
  variables:
    agent_name: "customer-agent"
    lambda_name: "customer-tools"
```

## Deployment Order

Custom modules are deployed in dependency order:

1. **Infrastructure Modules** (VPC, networking)
2. **Storage Modules** (S3, databases, OpenSearch)
3. **Bedrock Resources** (agents, knowledge bases, Lambda functions)
4. **Monitoring Modules** (dashboards, alerts)

## Best Practices

### Module Organization
1. **Use semantic versioning** for module versions
2. **Pin module versions** for reproducible deployments
3. **Create reusable modules** for common patterns
4. **Document module interfaces** clearly
5. **Test modules** independently before integration

### Variable Management
1. **Use descriptive variable names** that indicate purpose
2. **Provide default values** where appropriate
3. **Validate variable inputs** in module definitions
4. **Use consistent naming conventions** across modules
5. **Document required vs optional** variables

### Dependency Management
1. **Define explicit dependencies** using `dependsOn`
2. **Avoid circular dependencies** between modules
3. **Use module outputs** for cross-module references
4. **Consider deployment order** when designing dependencies
5. **Test dependency resolution** during development

### Security
1. **Use least privilege** IAM policies in modules
2. **Encrypt sensitive data** at rest and in transit
3. **Use secure module sources** (private registries)
4. **Validate module inputs** for security implications
5. **Audit module permissions** regularly

## Dependencies

- **Module Sources**: Referenced modules must be accessible
- **Variable Inputs**: All required variables must be provided
- **Resource Dependencies**: Resources referenced via `dependsOn` must exist
- **Provider Requirements**: Modules may require specific Terraform providers

## Generated Resources

- Terraform module call
- Module variable assignments
- Module output references
- Provider configurations (if required)

## Common Issues

### Module Not Found
```
Error: Module not found at source
```
**Solution**: Verify module source path or URL is correct and accessible.

### Variable Type Mismatch
```
Error: Invalid value for variable
```
**Solution**: Check variable types match module expectations.

### Circular Dependencies
```
Error: Cycle detected in module dependencies
```
**Solution**: Review and restructure dependencies to eliminate cycles.

### Version Conflicts
```
Error: Module version constraint not satisfied
```
**Solution**: Update version constraints or use compatible module versions.

## See Also

- [Agent Resource](agent.md)
- [Lambda Resource](lambda.md)
- [Knowledge Base Resource](knowledge-base.md)
- [Terraform Module Documentation](https://developer.hashicorp.com/terraform/language/modules)