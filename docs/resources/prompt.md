# Prompt Resource

Custom prompts with multiple variants, input variables, and support for both TEXT and CHAT template types.

## Overview

The Prompt resource creates AWS Bedrock prompts with comprehensive capabilities including input variable definitions, multiple template types (TEXT and CHAT), tool configurations for function calling, and automatic IAM role generation.

## Basic Examples

### TEXT Template

```yaml
kind: Prompt
metadata:
  name: "customer-support-prompt"
spec:
  description: "Customer support prompt with input variables"
  defaultVariant: "production"
  
  # Define input variables that can be used in templates
  inputVariables:
    - name: "query"
    - name: "context"
    - name: "customer_name"
  
  variants:
    - name: "production"
      modelId: "anthropic.claude-3-sonnet-20240229-v1:0"
      templateType: "TEXT"
      templateConfiguration:
        text:
          text: |
            You are a professional customer support agent.
            
            Customer: {{customer_name}}
            Query: {{query}}
            Context: {{context}}
            
            Please provide a helpful response.
          
          inputVariables:
            - name: "query"
            - name: "context"
            - name: "customer_name"
      
      inferenceConfiguration:
        text:
          temperature: 0.1
          topP: 0.9
          maxTokens: 2048
```

### CHAT Template with Tools

```yaml
kind: Prompt
metadata:
  name: "ai-assistant-chat"
spec:
  description: "Chat-based AI assistant with tool calling"
  defaultVariant: "production"
  
  inputVariables:
    - name: "user_query"
    - name: "user_name"
  
  variants:
    - name: "production"
      modelId: "anthropic.claude-3-sonnet-20240229-v1:0"
      templateType: "CHAT"
      templateConfiguration:
        chat:
          system:
            - text: "You are a helpful AI assistant with access to tools."
          
          messages:
            - role: "user"
              content:
                - text: "Hello, I'm {{user_name}}. {{user_query}}"
          
          toolConfiguration:
            tools:
              - toolSpec:
                  name: "search_knowledge_base"
                  description: "Search knowledge base for information"
                  inputSchema:
                    json:
                      type: "object"
                      properties:
                        query:
                          type: "string"
                          description: "Search query"
                      required: ["query"]
            
            toolChoice:
              auto: {}
          
          inputVariables:
            - name: "user_query"
            - name: "user_name"
```

## Complete Example

```yaml
kind: Prompt
metadata:
  name: "enterprise-assistant"
spec:
  description: "Enterprise AI assistant with encryption and comprehensive features"
  defaultVariant: "production"
  
  # Optional: Customer encryption key for sensitive prompts
  customerEncryptionKeyArn: "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012"
  
  # Global input variables
  inputVariables:
    - name: "user_query"
    - name: "context"
    - name: "user_name"
    - name: "department"
  
  variants:
    - name: "production"
      modelId: "anthropic.claude-3-sonnet-20240229-v1:0"
      templateType: "CHAT"
      templateConfiguration:
        chat:
          system:
            - text: |
                You are an enterprise AI assistant with access to company tools and information.
                Maintain professional communication and follow company policies.
          
          messages:
            - role: "user"
              content:
                - text: "I'm {{user_name}} from {{department}}. {{user_query}}"
          
          toolConfiguration:
            tools:
              - toolSpec:
                  name: "search_company_docs"
                  description: "Search company documentation and policies"
                  inputSchema:
                    json:
                      type: "object"
                      properties:
                        query:
                          type: "string"
                          description: "Search query"
                        category:
                          type: "string"
                          enum: ["policies", "procedures", "faq"]
                      required: ["query"]
              
              - toolSpec:
                  name: "get_employee_info"
                  description: "Get employee information (with proper authorization)"
                  inputSchema:
                    json:
                      type: "object"
                      properties:
                        employee_id:
                          type: "string"
                          description: "Employee ID"
                      required: ["employee_id"]
            
            toolChoice:
              auto: {}
          
          inputVariables:
            - name: "user_query"
            - name: "user_name"
            - name: "department"
      
      inferenceConfiguration:
        text:
          temperature: 0.2
          topP: 0.9
          topK: 50
          maxTokens: 4096
          stopSequences: ["Human:", "User:"]
    
    - name: "development"
      modelId: "anthropic.claude-3-sonnet-20240229-v1:0"
      templateType: "TEXT"
      templateConfiguration:
        text:
          text: |
            [DEBUG] Enterprise Assistant
            User: {{user_name}} | Dept: {{department}}
            Query: {{user_query}}
            Context: {{context}}
          
          inputVariables:
            - name: "user_query"
            - name: "user_name"
            - name: "department"
            - name: "context"
      
      inferenceConfiguration:
        text:
          temperature: 0.3
          topP: 0.95
          maxTokens: 2048
  
  tags:
    Environment: "production"
    Classification: "internal"
    Team: "ai-platform"
```

## Specification

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `variants` | array | List of prompt variants |

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `description` | string | Prompt description |
| `defaultVariant` | string | Name of the default variant |
| `customerEncryptionKeyArn` | string | KMS key ARN for encryption |
| `inputVariables` | array | Global input variables |
| `tags` | object | Resource tags |

### Input Variables

```yaml
inputVariables:
  - name: "variable_name"
```

### Variant Configuration

```yaml
variants:
  - name: "variant-name"
    modelId: "model-arn"
    templateType: "TEXT" | "CHAT"
    templateConfiguration:
      # Configuration based on template type
    inferenceConfiguration:
      text:
        # Inference parameters
```

### Template Types

#### TEXT Template

```yaml
templateType: "TEXT"
templateConfiguration:
  text:
    text: |
      Your prompt template with {{variables}}
    inputVariables:
      - name: "variable_name"
```

#### CHAT Template

```yaml
templateType: "CHAT"
templateConfiguration:
  chat:
    system:
      - text: "System message"
    
    messages:
      - role: "user" | "assistant" | "system"
        content:
          - text: "Message content with {{variables}}"
    
    toolConfiguration:
      tools:
        - toolSpec:
            name: "tool_name"
            description: "Tool description"
            inputSchema:
              json:
                type: "object"
                properties: {}
      
      toolChoice:
        auto: {}    # Let model decide
        # any: {}   # Must use at least one tool
        # tool:     # Force specific tool
        #   name: "tool_name"
    
    inputVariables:
      - name: "variable_name"
```

### Chat Messages

```yaml
messages:
  - role: "user"
    content:
      - text: "User message with {{variables}}"
  
  - role: "assistant"
    content:
      - text: "Assistant response"
  
  - role: "system"
    content:
      - text: "System instruction"
```

### Tool Configuration

```yaml
toolConfiguration:
  tools:
    - toolSpec:
        name: "function_name"
        description: "What the function does"
        inputSchema:
          json:
            type: "object"
            properties:
              parameter_name:
                type: "string"
                description: "Parameter description"
                enum: ["value1", "value2"]  # Optional
            required: ["parameter_name"]
  
  toolChoice:
    auto: {}                    # Model decides when to use tools
    # any: {}                   # Model must use at least one tool
    # tool:                     # Force specific tool
    #   name: "function_name"
```

### Inference Configuration

```yaml
inferenceConfiguration:
  text:
    temperature: 0.1       # 0.0-1.0, controls randomness
    topP: 0.9             # 0.0-1.0, nucleus sampling
    topK: 50              # Top-K sampling
    maxTokens: 2048       # Maximum tokens to generate
    stopSequences:        # Sequences that stop generation
      - "Human:"
      - "Assistant:"
```

## Supported Foundation Models

| Model | Model ID |
|-------|----------|
| Claude 3 Sonnet | `anthropic.claude-3-sonnet-20240229-v1:0` |
| Claude 3 Haiku | `anthropic.claude-3-haiku-20240307-v1:0` |
| Claude 3 Opus | `anthropic.claude-3-opus-20240229-v1:0` |
| Claude 3.5 Sonnet | `anthropic.claude-3-5-sonnet-20240620-v1:0` |

## Auto-Generated IAM Permissions

Prompts automatically get IAM roles with these permissions:

### Bedrock Model Access
- `bedrock:InvokeModel` for the specified foundation model
- `bedrock:GetPrompt` and `bedrock:UpdatePrompt` for prompt management

### CloudWatch Logging
- `logs:CreateLogGroup`
- `logs:CreateLogStream`  
- `logs:PutLogEvents`

### KMS Access (if encryption key is specified)
- `kms:Decrypt` and `kms:GenerateDataKey` for the specified KMS key

## Agent Integration

### With Agent Resource

```yaml
kind: Agent
metadata:
  name: "customer-agent"
spec:
  foundationModel: "anthropic.claude-3-sonnet-20240229-v1:0"
  instruction: "You are a customer support agent"
  
  promptOverrides:
    - promptType: "ORCHESTRATION"
      prompt: "customer-support-prompt"
      variant: "production"
```

## Common Patterns

### Multi-Environment Prompts

```yaml
kind: Prompt
metadata:
  name: "environment-aware-prompt"
spec:
  defaultVariant: "production"
  
  inputVariables:
    - name: "query"
    - name: "environment"
  
  variants:
    - name: "production"
      templateType: "TEXT"
      templateConfiguration:
        text:
          text: |
            [PRODUCTION] Professional assistant
            Environment: {{environment}}
            Query: {{query}}
      
    - name: "development"
      templateType: "TEXT"
      templateConfiguration:
        text:
          text: |
            [DEBUG] Development mode
            Environment: {{environment}}
            Query: {{query}}
            Debug info will be included.
```

### Function Calling with Tools

```yaml
kind: Prompt
metadata:
  name: "function-calling-assistant"
spec:
  variants:
    - name: "production"
      templateType: "CHAT"
      templateConfiguration:
        chat:
          system:
            - text: "You are an assistant with access to company tools."
          
          toolConfiguration:
            tools:
              - toolSpec:
                  name: "search_orders"
                  description: "Search customer orders"
                  inputSchema:
                    json:
                      type: "object"
                      properties:
                        customer_id:
                          type: "string"
                        order_status:
                          type: "string"
                          enum: ["pending", "shipped", "delivered"]
                      required: ["customer_id"]
              
              - toolSpec:
                  name: "calculate_shipping"
                  description: "Calculate shipping cost"
                  inputSchema:
                    json:
                      type: "object"
                      properties:
                        weight:
                          type: "number"
                        destination:
                          type: "string"
                        shipping_method:
                          type: "string"
                          enum: ["standard", "express", "overnight"]
                      required: ["weight", "destination"]
            
            toolChoice:
              auto: {}
```

### Performance-Optimized Variants

```yaml
kind: Prompt
metadata:
  name: "performance-variants"
spec:
  defaultVariant: "balanced"
  
  variants:
    - name: "creative"
      modelId: "anthropic.claude-3-sonnet-20240229-v1:0"
      templateType: "TEXT"
      inferenceConfiguration:
        text:
          temperature: 0.8
          topP: 0.95
          maxTokens: 4096
    
    - name: "balanced"
      modelId: "anthropic.claude-3-sonnet-20240229-v1:0"
      templateType: "TEXT"
      inferenceConfiguration:
        text:
          temperature: 0.3
          topP: 0.9
          maxTokens: 2048
    
    - name: "fast"
      modelId: "anthropic.claude-3-haiku-20240307-v1:0"
      templateType: "TEXT"
      inferenceConfiguration:
        text:
          temperature: 0.1
          topP: 0.8
          maxTokens: 512
```

## Best Practices

### Template Design
1. **Define input variables** at both prompt and template levels
2. **Use clear system messages** for CHAT templates
3. **Include examples** in templates when tasks are complex
4. **Set appropriate constraints** for response format
5. **Test with different variants** for various use cases

### Variable Management
1. **Use descriptive variable names** (`customer_name` not `cn`)
2. **Define variables consistently** across templates
3. **Document variable purposes** in descriptions
4. **Validate variable usage** in templates
5. **Use template-level variables** for specific overrides

### Tool Configuration
1. **Provide clear tool descriptions** for better model understanding
2. **Use JSON Schema validation** for tool inputs
3. **Choose appropriate tool choice strategies** (auto, any, specific)
4. **Test tool calling behavior** thoroughly
5. **Handle tool errors gracefully** in your functions

### Parameter Tuning
1. **Lower temperature** (0.1-0.3) for consistent, factual responses
2. **Higher temperature** (0.7-0.9) for creative, varied responses  
3. **Adjust topP and topK** to control diversity
4. **Set maxTokens** based on expected response length
5. **Use stopSequences** to prevent unwanted continuation

### Security
1. **Use encryption keys** for sensitive prompts
2. **Validate input variables** before template rendering
3. **Implement proper access controls** for prompt variants
4. **Audit prompt changes** regularly
5. **Follow data classification** guidelines

## Variables and Templating

### Variable Syntax

Variables use double braces: `{{variable_name}}`

```yaml
text: |
  Hello {{customer_name}},
  Your order {{order_id}} status is: {{status}}
```

### Variable Scope

1. **Global variables** (prompt level): Available to all variants
2. **Template variables** (variant level): Specific to that template
3. **Template overrides**: Template-level variables override global ones

### Common Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `{{query}}` | User's question or request | "What is my order status?" |
| `{{context}}` | Relevant context information | Previous conversation, order details |
| `{{user_name}}` | User's name | "John Smith" |
| `{{user_id}}` | User identifier | "user-12345" |
| `{{session_id}}` | Session identifier | "session-67890" |

## Dependencies

- **Foundation Models**: Referenced models must be available in the region
- **KMS Keys**: If encryption is used, KMS key must exist and be accessible
- **Agent Resources**: If used with agents, agent must exist

## Generated Resources

- AWS Bedrock Prompt
- IAM Role (auto-generated)
- IAM Policy (auto-generated)
- Prompt variants
- KMS permissions (if encryption key is specified)

## Common Issues

### Template Variable Errors
```
Error: Undefined variable in template
```
**Solution**: Ensure all variables in templates are defined in inputVariables.

### Model Access Errors
```
Error: AccessDenied for foundation model
```
**Solution**: Verify the model is available in your region and account has access.

### Tool Schema Errors
```
Error: Invalid tool input schema
```
**Solution**: Ensure JSON schema is valid and follows OpenAPI specification.

### Encryption Key Errors
```
Error: AccessDenied for KMS key
```
**Solution**: Verify KMS key exists and IAM role has decrypt permissions.

## See Also

- [Agent Resource](agent.md)
- [IAM Management](../iam-management.md)
- [AWS Bedrock Prompts Documentation](https://docs.aws.amazon.com/bedrock/latest/userguide/prompt-management.html)