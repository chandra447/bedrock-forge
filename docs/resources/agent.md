# Agent Resource

AWS Bedrock agents with guardrails, action groups, and knowledge bases.

## Overview

The Agent resource creates AWS Bedrock agents with comprehensive IAM permissions automatically generated. Agents can include guardrails, action groups, knowledge bases, and prompt overrides.

## Basic Example

```yaml
kind: Agent
metadata:
  name: "customer-support"
  description: "Customer support agent with order management capabilities"
spec:
  foundationModel: "anthropic.claude-3-sonnet-20240229-v1:0"
  instruction: "You are a helpful customer support agent"
  # IAM role is automatically generated!
```

## Complete Example

```yaml
kind: Agent
metadata:
  name: "customer-support"
  description: "Advanced customer support agent"
spec:
  foundationModel: "anthropic.claude-3-sonnet-20240229-v1:0"
  instruction: "You are a helpful customer support agent"
  idleSessionTtlInSeconds: 3600
  
  # Guardrail integration
  guardrail:
    name: "content-safety-guardrail"
    version: "1"
    mode: "pre"
  
  # Inline action groups
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

## Specification

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `foundationModel` | string | AWS Bedrock foundation model ARN |
| `instruction` | string | Agent's system instruction |

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `idleSessionTtlInSeconds` | number | Session timeout in seconds (default: 3600) |
| `guardrail` | object | Guardrail configuration |
| `actionGroups` | array | Inline action group definitions |
| `promptOverrides` | array | Custom prompt configurations |
| `memoryConfiguration` | object | Memory settings |

### Guardrail Configuration

```yaml
guardrail:
  name: "guardrail-name"    # Reference to Guardrail resource
  version: "1"              # Guardrail version
  mode: "pre"               # "pre" or "post"
```

### Action Groups

```yaml
actionGroups:
  - name: "group-name"
    description: "Group description"
    actionGroupExecutor:
      lambda: "lambda-name"           # Reference to Lambda resource
      # OR
      lambdaArn: "arn:aws:lambda:..." # External Lambda ARN
    functionSchema:
      functions:
        - name: "function_name"
          description: "Function description"
          parameters:
            param_name:
              type: "string"
              required: true
```

### Prompt Overrides

```yaml
promptOverrides:
  - promptType: "ORCHESTRATION"  # "ORCHESTRATION" or "KNOWLEDGE_BASE_RESPONSE_GENERATION"
    prompt: "prompt-name"        # Reference to Prompt resource
    variant: "production"        # Prompt variant
```

### Memory Configuration

```yaml
memoryConfiguration:
  enabledMemoryTypes: ["SESSION_SUMMARY"]  # Memory types to enable
  storageDays: 30                          # Days to store memory
```

## Auto-Generated IAM Permissions

When you create an Agent, Bedrock Forge automatically generates an IAM role with these permissions:

### Foundation Model Access
- `bedrock:InvokeModel`
- `bedrock:InvokeModelWithResponseStream`

### Lambda Invocation (if action groups are present)
- `lambda:InvokeFunction` for referenced Lambda functions

### Knowledge Base Access (if knowledge bases are associated)
- `bedrock:Retrieve`
- `bedrock:RetrieveAndGenerate`

### CloudWatch Logging
- `logs:CreateLogGroup`
- `logs:CreateLogStream`
- `logs:PutLogEvents`

## Custom IAM Roles

For enterprise scenarios requiring specific permissions, you can define custom IAM roles. See [iam-role.md](iam-role.md) for details.

## Dependencies

- **Lambda Functions**: Referenced in action groups must exist
- **Guardrails**: Referenced guardrails must exist
- **Prompts**: Referenced prompts must exist
- **Knowledge Bases**: Must be associated via separate AgentKnowledgeBaseAssociation resource

## Generated Resources

- AWS Bedrock Agent
- IAM Role (auto-generated)
- IAM Policy (auto-generated)

## Common Patterns

### Simple Agent
```yaml
kind: Agent
metadata:
  name: "simple-agent"
spec:
  foundationModel: "anthropic.claude-3-sonnet-20240229-v1:0"
  instruction: "You are a helpful assistant"
```

### Agent with Lambda Functions
```yaml
kind: Agent
metadata:
  name: "tool-agent"
spec:
  foundationModel: "anthropic.claude-3-sonnet-20240229-v1:0"
  instruction: "You are an assistant with tools"
  actionGroups:
    - name: "tools"
      description: "Useful tools"
      actionGroupExecutor:
        lambda: "my-lambda-function"
      functionSchema:
        functions:
          - name: "get_weather"
            description: "Get weather"
            parameters:
              location:
                type: "string"
                required: true
```

### Enterprise Agent with Guardrails
```yaml
kind: Agent
metadata:
  name: "enterprise-agent"
spec:
  foundationModel: "anthropic.claude-3-sonnet-20240229-v1:0"
  instruction: "You are a compliant enterprise assistant"
  guardrail:
    name: "enterprise-guardrail"
    version: "1"
    mode: "pre"
  idleSessionTtlInSeconds: 1800
  memoryConfiguration:
    enabledMemoryTypes: ["SESSION_SUMMARY"]
    storageDays: 7
```

## Agent Aliases

Agent aliases enable deployment strategies like staging, production, and canary deployments by creating named references to specific agent versions with traffic routing capabilities.

### Basic Alias Configuration

```yaml
kind: Agent
metadata:
  name: "customer-support"
spec:
  foundationModel: "anthropic.claude-3-sonnet-20240229-v1:0"
  instruction: "You are a helpful customer support agent"
  
  # Agent aliases for different environments
  aliases:
    - name: "dev"
      description: "Development version for testing"
      routingConfiguration:
        - agentVersion: "DRAFT"
    
    - name: "prod"
      description: "Production version"
      routingConfiguration:
        - agentVersion: "1"
```

### Advanced Alias with Traffic Splitting

```yaml
aliases:
  - name: "prod"
    description: "Production with blue-green deployment"
    routingConfiguration:
      # 90% traffic to stable version
      - agentVersion: "2"
        provisionedThroughput: "90"
      # 10% traffic to new version
      - agentVersion: "3"
        provisionedThroughput: "10"
    tags:
      Environment: "prod"
      DeploymentStrategy: "blue-green"
  
  - name: "canary"
    description: "Canary testing with minimal traffic"
    routingConfiguration:
      - agentVersion: "DRAFT"
        provisionedThroughput: "5"
    tags:
      Purpose: "canary-testing"
```

### Alias Configuration Options

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Alias name (required) |
| `description` | string | Alias description |
| `routingConfiguration` | array | Traffic routing rules |
| `tags` | object | Alias-specific tags |

### Routing Configuration

| Field | Type | Description |
|-------|------|-------------|
| `agentVersion` | string | Agent version ("DRAFT", "1", "2", etc.) |
| `provisionedThroughput` | string | Traffic percentage (0-100) |

### Deployment Benefits

- **Environment Separation**: Dev, staging, and production aliases
- **Traffic Splitting**: Gradual rollouts with version-specific traffic
- **Canary Deployments**: Test new versions with minimal traffic
- **Blue-Green Deployments**: Instant traffic switching between versions
- **Version Management**: Named references to specific agent versions

## Best Practices

1. **Use descriptive names** for agents and action groups
2. **Set appropriate session timeouts** based on use case
3. **Include guardrails** for production deployments
4. **Use memory configuration** for conversational agents
5. **Test with different foundation models** to find the best fit
6. **Keep instructions clear and specific** for better agent behavior

## See Also

- [Action Group Resource](action-group.md)
- [Lambda Resource](lambda.md)
- [Guardrail Resource](guardrail.md)
- [IAM Management](../iam-management.md)