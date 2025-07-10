# Custom Resources Examples

This directory demonstrates how to use `CustomResources` to include your own Terraform files alongside bedrock-forge generated resources.

## Overview

`CustomResources` allows you to:
- Write standard Terraform `.tf` files for AWS resources not directly supported by bedrock-forge
- Include them in your bedrock-forge deployment
- Cross-reference these resources in your agent YAML files
- Deploy everything together as a single Terraform deployment

## Examples

### 1. Directory-based approach (`sns-eventbridge.yml`)

```yaml
kind: CustomResources
metadata:
  name: infrastructure
spec:
  path: "./terraform/"  # Points to directory with .tf files
  variables:
    environment: "${var.environment}"
    sns_topic_name: "agent-notifications"
```

### 2. File-specific approach (`specific-files.yml`)

```yaml
kind: CustomResources  
metadata:
  name: vpc-and-security
spec:
  files:              # Lists specific .tf files
    - "vpc.tf"
    - "security-groups.tf"
  variables:
    vpc_cidr: "10.0.0.0/16"
```

### 3. Cross-referencing in agents (`agent-with-custom-resources.yml`)

```yaml
kind: Agent
metadata:
  name: notification-agent
spec:
  environment:
    SNS_TOPIC_ARN: "${aws_sns_topic.agent_notifications.arn}"
  dependsOn:
    - infrastructure  # Ensures custom resources are created first
```

## Directory Structure

```
project/
├── custom-resources.yml       # CustomResources YAML
├── agent.yml                  # Agent that uses custom resources
└── terraform/                 # Your .tf files
    ├── sns.tf
    ├── eventbridge.tf
    └── variables.tf
```

## How it works

1. **Creation**: bedrock-forge copies your `.tf` files into the generated Terraform output
2. **Variables**: Any variables you specify get merged with bedrock-forge variables
3. **Dependencies**: Use `dependsOn` to ensure proper creation order
4. **Cross-references**: Use standard Terraform syntax to reference your resources

## Terraform Files Included

- `terraform/sns.tf` - SNS topic for agent notifications
- `terraform/eventbridge.tf` - EventBridge rules for capturing agent events  
- `terraform/variables.tf` - Variables for the custom infrastructure

These files create:
- SNS topic for notifications with proper permissions
- EventBridge rule to capture Bedrock agent execution events
- Outputs that can be referenced in agent configurations

## Deployment

When you run `bedrock-forge generate`, it will:
1. Copy your `.tf` files to the output directory
2. Generate bedrock-forge resources (agents, lambdas, etc.)
3. Create a single Terraform configuration that includes everything
4. You can then run `terraform plan` and `terraform apply` as usual

This approach gives you the flexibility to use any AWS resource while still benefiting from bedrock-forge's simplified YAML syntax for Bedrock-specific resources.