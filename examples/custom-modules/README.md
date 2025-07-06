# Custom Module Examples

This directory contains examples of how to use the `CustomModule` resource kind to include your own Terraform modules alongside Bedrock resources.

## Overview

Custom modules allow you to:
- Include existing Terraform modules from various sources
- Pass variables and configuration to modules
- Define dependencies between modules and Bedrock resources
- Integrate infrastructure components like VPC, storage, monitoring, etc.

## Supported Module Sources

### 1. Local Modules
```yaml
source: "./modules/my-module"          # Relative path
source: "/absolute/path/to/module"     # Absolute path
```

### 2. Terraform Registry Modules
```yaml
source: "terraform-aws-modules/vpc/aws"
version: "5.0.0"
```

### 3. Git Repository Modules
```yaml
source: "git::https://github.com/org/repo.git"
version: "v1.2.0"  # Git tag or branch
```

## Examples

### S3 Bucket (`s3-bucket.yml`)
- **Purpose**: Storage for Lambda code and API schemas
- **Type**: Local module
- **Features**: Versioning, lifecycle rules, tagging

### VPC Infrastructure (`vpc-module.yml`)
- **Purpose**: Network infrastructure for Lambda functions
- **Type**: Terraform registry module
- **Features**: Public/private subnets, NAT gateway, DNS support

### OpenSearch Cluster (`opensearch-cluster.yml`)
- **Purpose**: Vector storage for knowledge bases
- **Type**: Git repository module
- **Features**: Encryption, access policies, VPC integration
- **Dependencies**: Requires VPC module to be created first

### Monitoring Dashboard (`monitoring-dashboard.yml`)
- **Purpose**: CloudWatch monitoring for Bedrock agents
- **Type**: Local shared module
- **Features**: Custom dashboards, alerts, metric aggregation
- **Dependencies**: Requires agents and Lambda functions to exist

## Variable Types Supported

Custom modules support all Terraform variable types:

- **Strings**: `bucket_name: "my-bucket"`
- **Numbers**: `instance_count: 3`
- **Booleans**: `enable_encryption: true`
- **Lists**: `subnets: ["subnet-1", "subnet-2"]`
- **Objects**: Complex nested configurations
- **References**: `"${module.other.output_value}"`

## Dependency Management

Use the `dependsOn` field to ensure proper resource ordering:

```yaml
spec:
  dependsOn:
    - "vpc-module"      # CustomModule dependency
    - "my-agent"        # Agent dependency
    - "my-lambda"       # Lambda dependency
```

## Best Practices

1. **Use descriptive names** for modules that indicate their purpose
2. **Define dependencies** explicitly to ensure correct deployment order
3. **Use variables** instead of hardcoded values for flexibility
4. **Add meaningful tags** for resource management
5. **Include descriptions** to document module purpose
6. **Version your modules** when using external sources

## Integration with Bedrock Resources

Custom modules can:
- Provide infrastructure that Bedrock resources depend on (VPC, storage)
- Consume outputs from Bedrock resources (agent ARNs, Lambda ARNs)
- Add monitoring and observability for Bedrock deployments
- Implement security and compliance requirements

## Generated Terraform

Each custom module becomes a `module` block in the generated `main.tf`:

```hcl
module "bedrock_vpc" {
  source = "terraform-aws-modules/vpc/aws"
  version = "5.0.0"
  
  name = "bedrock-agent-vpc"
  cidr = "10.0.0.0/16"
  # ... other variables
}

module "bedrock_opensearch" {
  source = "git::https://github.com/org/repo.git?ref=v1.2.0"
  
  vpc_id = module.bedrock_vpc.vpc_id
  # ... other variables
  
  depends_on = [module.bedrock_vpc]
}
```