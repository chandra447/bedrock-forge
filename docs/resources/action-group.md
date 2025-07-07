# Action Group Resource

Action groups link Bedrock agents to Lambda functions, enabling agents to perform actions and access external systems.

## Overview

Action groups connect Bedrock agents to Lambda functions, allowing agents to execute functions based on user requests. They include function schemas that define available actions and their parameters.

## Basic Example

```yaml
kind: ActionGroup
metadata:
  name: "order-management"
  description: "Handle order lookups and updates"
spec:
  agentName: "customer-support"  # Reference to Agent resource
  description: "Provides order lookup and management capabilities"
  
  # Lambda executor
  actionGroupExecutor:
    lambda: "order-lookup-lambda"  # Reference to Lambda resource
  
  # Function schema
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

## Complete Example

```yaml
kind: ActionGroup
metadata:
  name: "customer-tools"
  description: "Comprehensive customer management tools"
spec:
  agentName: "customer-support"
  description: "Customer support tools for order management and account operations"
  
  # Can reference local Lambda or external ARN
  actionGroupExecutor:
    lambda: "customer-tools-lambda"
    # OR use external Lambda ARN
    # lambdaArn: "arn:aws:lambda:us-east-1:123456789012:function:external-function"
  
  # Define available functions
  functionSchema:
    functions:
      - name: "lookup_order"
        description: "Look up order details by order ID"
        parameters:
          order_id:
            type: "string"
            description: "The unique order identifier"
            required: true
      
      - name: "update_order_status"
        description: "Update the status of an order"
        parameters:
          order_id:
            type: "string"
            description: "The order ID to update"
            required: true
          status:
            type: "string"
            description: "New status for the order"
            enum: ["pending", "processing", "shipped", "delivered", "cancelled"]
            required: true
      
      - name: "get_customer_info"
        description: "Retrieve customer information"
        parameters:
          customer_id:
            type: "string"
            description: "Customer identifier"
            required: true
          include_orders:
            type: "boolean"
            description: "Whether to include order history"
            required: false
            default: false
      
      - name: "create_support_ticket"
        description: "Create a new support ticket"
        parameters:
          customer_id:
            type: "string"
            description: "Customer identifier"
            required: true
          issue_type:
            type: "string"
            description: "Type of issue"
            enum: ["billing", "technical", "general"]
            required: true
          description:
            type: "string"
            description: "Detailed description of the issue"
            required: true
          priority:
            type: "string"
            description: "Priority level"
            enum: ["low", "medium", "high", "urgent"]
            required: false
            default: "medium"
```

## Specification

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `agentName` | string | Name of the associated Agent resource |
| `actionGroupExecutor` | object | Lambda function configuration |

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `description` | string | Description of the action group |
| `functionSchema` | object | Function definitions and parameters |
| `apiSchema` | object | OpenAPI schema (alternative to functionSchema) |
| `tags` | object | Resource tags |

### Action Group Executor

```yaml
actionGroupExecutor:
  # Reference to local Lambda resource
  lambda: "lambda-name"
  
  # OR reference external Lambda ARN
  lambdaArn: "arn:aws:lambda:region:account:function:function-name"
```

### Function Schema

```yaml
functionSchema:
  functions:
    - name: "function_name"
      description: "Function description"
      parameters:
        param_name:
          type: "string"              # string, number, boolean, array, object
          description: "Parameter description"
          required: true              # true or false
          enum: ["value1", "value2"]  # Optional enum values
          default: "default_value"    # Optional default value
```

### Parameter Types

| Type | Description | Example |
|------|-------------|---------|
| `string` | Text value | `"hello world"` |
| `number` | Numeric value | `42` or `3.14` |
| `boolean` | True/false | `true` or `false` |
| `array` | List of values | `["item1", "item2"]` |
| `object` | Complex object | `{"key": "value"}` |

## Lambda Function Integration

### Local Lambda Reference

```yaml
# Lambda function definition
kind: Lambda
metadata:
  name: "order-tools"
spec:
  runtime: "python3.11"
  handler: "app.handler"
  description: "Order management functions"

---
# Action group referencing local Lambda
kind: ActionGroup
metadata:
  name: "order-actions"
spec:
  agentName: "customer-agent"
  actionGroupExecutor:
    lambda: "order-tools"  # References Lambda above
```

### External Lambda ARN

```yaml
kind: ActionGroup
metadata:
  name: "external-actions"
spec:
  agentName: "customer-agent"
  actionGroupExecutor:
    lambdaArn: "arn:aws:lambda:us-east-1:123456789012:function:existing-function"
```

## Function Schema Examples

### Simple Function

```yaml
functionSchema:
  functions:
    - name: "get_weather"
      description: "Get current weather for a location"
      parameters:
        location:
          type: "string"
          description: "City name or location"
          required: true
```

### Function with Multiple Parameters

```yaml
functionSchema:
  functions:
    - name: "book_appointment"
      description: "Book a customer appointment"
      parameters:
        customer_id:
          type: "string"
          description: "Customer identifier"
          required: true
        appointment_type:
          type: "string"
          description: "Type of appointment"
          enum: ["consultation", "service", "support"]
          required: true
        preferred_date:
          type: "string"
          description: "Preferred date (YYYY-MM-DD)"
          required: true
        preferred_time:
          type: "string"
          description: "Preferred time (HH:MM)"
          required: false
        notes:
          type: "string"
          description: "Additional notes"
          required: false
```

### Function with Complex Parameters

```yaml
functionSchema:
  functions:
    - name: "process_order"
      description: "Process a customer order"
      parameters:
        order_data:
          type: "object"
          description: "Order information"
          required: true
          properties:
            customer_id:
              type: "string"
              required: true
            items:
              type: "array"
              description: "List of items"
              required: true
            shipping_address:
              type: "object"
              description: "Shipping address"
              required: true
            payment_method:
              type: "string"
              enum: ["credit_card", "debit_card", "paypal"]
              required: true
```

## OpenAPI Schema Alternative

Instead of `functionSchema`, you can use OpenAPI specifications:

```yaml
kind: ActionGroup
metadata:
  name: "api-actions"
spec:
  agentName: "customer-agent"
  actionGroupExecutor:
    lambda: "api-lambda"
  
  # OpenAPI schema instead of functionSchema
  apiSchema:
    openapi: "3.0.0"
    info:
      title: "Customer API"
      version: "1.0.0"
    paths:
      /orders/{orderId}:
        get:
          summary: "Get order details"
          parameters:
            - name: "orderId"
              in: "path"
              required: true
              schema:
                type: "string"
          responses:
            200:
              description: "Order details"
```

## Auto-Generated IAM Permissions

Action groups inherit IAM permissions from their associated agent roles. The agent's automatically generated role includes:

- `lambda:InvokeFunction` for the referenced Lambda function
- All standard agent permissions (foundation models, logging, etc.)

## Agent Integration

### Inline Action Groups (Recommended)

```yaml
kind: Agent
metadata:
  name: "customer-agent"
spec:
  foundationModel: "anthropic.claude-3-sonnet-20240229-v1:0"
  instruction: "You are a customer support agent"
  
  # Inline action groups
  actionGroups:
    - name: "order-tools"
      description: "Order management tools"
      actionGroupExecutor:
        lambda: "order-lambda"
      functionSchema:
        functions:
          - name: "lookup_order"
            description: "Look up order by ID"
            parameters:
              order_id:
                type: "string"
                required: true
```

### Separate Action Group Resources

```yaml
kind: Agent
metadata:
  name: "customer-agent"
spec:
  foundationModel: "anthropic.claude-3-sonnet-20240229-v1:0"
  instruction: "You are a customer support agent"

---
kind: ActionGroup
metadata:
  name: "order-tools"
spec:
  agentName: "customer-agent"
  actionGroupExecutor:
    lambda: "order-lambda"
  functionSchema:
    functions:
      - name: "lookup_order"
        description: "Look up order by ID"
        parameters:
          order_id:
            type: "string"
            required: true
```

## Best Practices

### Function Design
1. **Use descriptive names** for functions and parameters
2. **Include detailed descriptions** for better agent understanding
3. **Use appropriate parameter types** for validation
4. **Define enum values** for constrained parameters
5. **Set sensible defaults** for optional parameters

### Schema Management
1. **Keep schemas focused** - one action group per domain
2. **Use consistent naming** conventions across functions
3. **Validate schemas** before deployment
4. **Version control** schema changes
5. **Test with different parameter combinations**

### Performance
1. **Optimize Lambda functions** for quick responses
2. **Use appropriate timeout values** based on function complexity
3. **Implement proper error handling** in Lambda functions
4. **Monitor function execution** times and costs
5. **Use Lambda layers** for common dependencies

### Security
1. **Validate all inputs** in Lambda functions
2. **Use least privilege** IAM roles
3. **Implement proper logging** for auditing
4. **Sanitize sensitive data** in function responses
5. **Use encryption** for data at rest and in transit

## Common Patterns

### CRUD Operations

```yaml
functionSchema:
  functions:
    - name: "create_record"
      description: "Create a new record"
      parameters:
        data:
          type: "object"
          required: true
    
    - name: "read_record"
      description: "Read a record by ID"
      parameters:
        id:
          type: "string"
          required: true
    
    - name: "update_record"
      description: "Update an existing record"
      parameters:
        id:
          type: "string"
          required: true
        data:
          type: "object"
          required: true
    
    - name: "delete_record"
      description: "Delete a record by ID"
      parameters:
        id:
          type: "string"
          required: true
```

### Search and Filter

```yaml
functionSchema:
  functions:
    - name: "search_products"
      description: "Search for products"
      parameters:
        query:
          type: "string"
          description: "Search query"
          required: true
        category:
          type: "string"
          description: "Product category"
          enum: ["electronics", "clothing", "books", "home"]
          required: false
        price_range:
          type: "object"
          description: "Price range filter"
          required: false
          properties:
            min:
              type: "number"
            max:
              type: "number"
        sort_by:
          type: "string"
          description: "Sort criteria"
          enum: ["price_low", "price_high", "rating", "newest"]
          default: "rating"
          required: false
```

## Dependencies

- **Agent**: Must reference an existing Agent resource
- **Lambda**: Referenced Lambda function must exist (for local references)
- **IAM**: Agent's IAM role must have permissions to invoke the Lambda function

## Generated Resources

- AWS Bedrock Action Group
- Lambda function permissions (automatically granted to agent role)

## See Also

- [Agent Resource](agent.md)
- [Lambda Resource](lambda.md)
- [IAM Management](../iam-management.md)
- [AWS Bedrock Action Groups Documentation](https://docs.aws.amazon.com/bedrock/latest/userguide/agents-action-groups.html)