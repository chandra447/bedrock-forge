# Custom Resources

Include your own Terraform files alongside bedrock-forge generated resources for comprehensive infrastructure deployment.

## Overview

Custom Resources allow you to write standard Terraform `.tf` files for AWS resources not directly supported by bedrock-forge and include them in your deployment. This enables you to:

- Use any AWS resource with standard Terraform syntax
- Cross-reference these resources in your agent YAML files  
- Deploy everything together as a single Terraform deployment
- Maintain full flexibility while benefiting from bedrock-forge's simplified YAML syntax

## Basic Example

```yaml
kind: CustomResources
metadata:
  name: "infrastructure"
  description: "SNS and EventBridge for agent notifications"
spec:
  path: "./terraform/"  # Directory containing .tf files
  
  variables:
    environment: "${var.environment}"
    sns_topic_name: "agent-notifications"
```

## File-Specific Example

```yaml
kind: CustomResources
metadata:
  name: "vpc-security"
  description: "VPC and security group resources"
spec:
  files:  # List specific files instead of directory
    - "vpc.tf"
    - "security-groups.tf"
    - "variables.tf"
  
  variables:
    vpc_cidr: "10.0.0.0/16"
    availability_zones: ["us-east-1a", "us-east-1b"]
```

## Cross-Reference in Agents

```yaml
kind: Agent
metadata:
  name: notification-agent
spec:
  foundationModel: "anthropic.claude-3-haiku-20240307-v1:0"
  instruction: "You are a helpful agent that sends notifications."
  
  # Reference resources from your .tf files
  environment:
    SNS_TOPIC_ARN: "${aws_sns_topic.agent_notifications.arn}"
    VPC_ID: "${aws_vpc.main.id}"
  
  # Ensure custom resources are created first
  dependsOn:
    - infrastructure
```

## Configuration Reference

### Required Fields

| Field | Description | Example |
|-------|-------------|---------|
| `path` OR `files` | Directory path or list of .tf files | `"./terraform/"` or `["vpc.tf", "sns.tf"]` |

### Optional Fields

| Field | Description | Default | Example |
|-------|-------------|---------|---------|
| `variables` | Variables to pass to Terraform | `{}` | `{"environment": "dev"}` |
| `dependsOn` | Resource dependencies | `[]` | `["vpc-module", "agent-name"]` |
| `description` | Description of resources | `""` | `"Infrastructure for notifications"` |

## How It Works

1. **File Copying**: bedrock-forge copies your `.tf` files into the generated Terraform output directory
2. **Variable Merging**: Your variables are merged with bedrock-forge generated variables
3. **Dependency Management**: Use `dependsOn` to ensure proper creation order
4. **Single Deployment**: Everything becomes part of one Terraform configuration

## Directory Structure

```
project/
├── infrastructure.yml           # CustomResources YAML
├── agent.yml                   # Agent that uses custom resources
└── terraform/                  # Your .tf files
    ├── sns.tf                  # SNS topic
    ├── eventbridge.tf          # EventBridge rules
    └── variables.tf            # Variables
```

## Example Terraform Files

### terraform/sns.tf
```hcl
resource "aws_sns_topic" "agent_notifications" {
  name = var.sns_topic_name
  
  tags = {
    Environment = var.environment
    ManagedBy   = "bedrock-forge"
  }
}

output "sns_topic_arn" {
  value = aws_sns_topic.agent_notifications.arn
}
```

### terraform/variables.tf
```hcl
variable "environment" {
  description = "Environment name"
  type        = string
  default     = "dev"
}

variable "sns_topic_name" {
  description = "SNS topic name"
  type        = string
  default     = "agent-notifications"
}
```

## Best Practices

### 1. Organize Files Logically
```yaml
spec:
  files:
    - "networking.tf"      # VPC, subnets, security groups
    - "storage.tf"         # S3, RDS, DynamoDB
    - "monitoring.tf"      # CloudWatch, SNS
    - "variables.tf"       # All variables
```

### 2. Use Descriptive Names
```yaml
metadata:
  name: "vpc-and-security"     # Clear, descriptive name
  description: "VPC infrastructure and security groups for Lambda functions"
```

### 3. Define Dependencies
```yaml
spec:
  dependsOn:
    - "vpc-infrastructure"     # Other CustomResources
    - "my-lambda"             # Lambda functions
    - "my-agent"              # Agents
```

### 4. Use Variables for Flexibility
```yaml
spec:
  variables:
    environment: "${var.environment}"        # Reference bedrock-forge variables
    vpc_cidr: "10.0.0.0/16"                 # Custom values
    instance_count: 2                        # Numbers
    enable_monitoring: true                  # Booleans
    subnets: ["subnet-1", "subnet-2"]       # Lists
```

### 5. Add Outputs for Cross-References
```hcl
# In your .tf files
output "vpc_id" {
  description = "VPC ID for use in other resources"
  value       = aws_vpc.main.id
}

output "security_group_id" {
  description = "Security group ID"
  value       = aws_security_group.lambda.id
}
```

## Advanced Examples

### Multi-File Infrastructure
```yaml
kind: CustomResources
metadata:
  name: "complete-infrastructure"
spec:
  files:
    - "vpc.tf"
    - "security-groups.tf"
    - "rds.tf"
    - "elasticache.tf"
    - "monitoring.tf"
    - "variables.tf"
    - "outputs.tf"
  
  variables:
    environment: "${var.environment}"
    db_instance_class: "db.t3.micro"
    redis_node_type: "cache.t3.micro"
    enable_multi_az: false
```

### Cross-Resource References
```yaml
# Agent using custom infrastructure
kind: Agent
metadata:
  name: "data-processing-agent"
spec:
  environment:
    DATABASE_URL: "${aws_db_instance.main.endpoint}"
    REDIS_URL: "${aws_elasticache_cluster.main.cache_nodes.0.address}"
    VPC_ID: "${aws_vpc.main.id}"
    SECURITY_GROUP: "${aws_security_group.lambda.id}"
  
  dependsOn:
    - complete-infrastructure
```

## Common Use Cases

### 1. VPC and Networking
- VPC with public/private subnets
- Security groups for Lambda functions
- NAT gateways for private resources

### 2. Data Storage
- RDS databases for persistent data
- S3 buckets for file storage
- DynamoDB tables for fast access

### 3. Monitoring and Alerting
- CloudWatch dashboards
- SNS topics for notifications
- EventBridge rules for event processing

### 4. Security and Compliance
- KMS keys for encryption
- IAM policies for fine-grained access
- WAF rules for web applications

## Troubleshooting

### Common Issues

#### 1. File Not Found
```
Error: path does not exist: ./terraform/
```
**Solution**: Ensure the path or files exist relative to where you run bedrock-forge

#### 2. Variable Conflicts
```
Error: variable "environment" already defined
```
**Solution**: Avoid redefining variables that bedrock-forge already provides

#### 3. Cross-Reference Errors
```
Error: resource "aws_sns_topic.notifications" not found
```
**Solution**: Ensure your `.tf` files are included and the resource name matches

### Validation

bedrock-forge validates that:
- Either `path` or `files` is specified (not both)
- Specified paths and files exist
- File extensions are `.tf`
- Dependencies reference valid resources

## Migration from CustomModule

If you were using the deprecated `CustomModule` approach:

```yaml
# Old (deprecated)
kind: CustomModule
metadata:
  name: vpc
spec:
  source: "terraform-aws-modules/vpc/aws"
  version: "5.0.0"
  variables: {...}

# New (recommended)
kind: CustomResources  
metadata:
  name: vpc
spec:
  path: "./terraform/"
  variables: {...}
```

1. Create your own `.tf` files with the resources you need
2. Change `kind: CustomModule` to `kind: CustomResources`
3. Change `source:` to `path:` pointing to your `.tf` files
4. Remove `version:` field (not needed for local files)

This approach gives you complete control over your infrastructure while maintaining the simplicity of bedrock-forge for Bedrock-specific resources.