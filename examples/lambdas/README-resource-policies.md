# Lambda Resource-Based Policies for Bedrock Agents

This document explains how bedrock-forge automatically manages Lambda resource-based policies to allow Bedrock agents to invoke Lambda functions in action groups.

## Zero-Configuration Automatic Permission Management

When you create a Lambda function in bedrock-forge, the system **automatically** handles all permissions:

1. **Creates default Bedrock permissions** on the Lambda function allowing `bedrock.amazonaws.com` to invoke it
2. **Generates agent-specific permissions** when agents reference the Lambda function in action groups
3. **Sets proper source ARN restrictions** to limit access to specific agents

## Intelligent Permission Detection

The Lambda generator **automatically detects** which agents reference each Lambda function and creates appropriate permissions:

### When Agents Reference the Lambda:
If any agent references your Lambda function, bedrock-forge creates **agent-specific** permissions like this:

```json
{
  "Sid": "AllowBedrockAgent_customer_support",
  "Effect": "Allow",
  "Principal": {
    "Service": "bedrock.amazonaws.com"
  },
  "Action": "lambda:InvokeFunction",
  "Condition": {
    "StringEquals": {
      "aws:SourceArn": "${module.customer_support.agent_arn}"
    }
  }
}
```

### When No Agents Reference the Lambda:
If no agents reference your Lambda function yet, bedrock-forge creates a **general** permission:

```json
{
  "Sid": "AllowBedrockAgentInvoke",
  "Effect": "Allow",
  "Principal": {
    "Service": "bedrock.amazonaws.com"
  },
  "Action": "lambda:InvokeFunction"
}
```

This ensures the Lambda is ready for future agent integration while maintaining security.

## Simple Lambda Configuration

**Most users don't need to configure anything!** Just create your Lambda function:

```yaml
kind: Lambda
metadata:
  name: "order-lookup"
  description: "Lambda function to look up customer orders"
spec:
  runtime: "python3.9"
  handler: "app.handler"
  code:
    source: "directory"
  environment:
    ORDER_API_URL: "https://api.company.com/orders"
  timeout: 30
  memorySize: 256
# No resource policy needed - bedrock-forge handles everything automatically!
```

## Advanced: Custom Resource Policies (Optional)

For advanced scenarios, you can still define custom resource-based policies:

```yaml
kind: Lambda
metadata:
  name: "my-function"
spec:
  runtime: "python3.9"
  handler: "app.handler"
  # ... other configuration
  
  # Optional: Advanced resource policy configuration
  resourcePolicy:
    # Control default behavior (default: true)
    allowBedrockAgents: true
    
    # Additional custom statements for enterprise scenarios
    statements:
      - sid: "AllowSpecificAccount"
        effect: "Allow"
        principal:
          Service: "bedrock.amazonaws.com"
        action: "lambda:InvokeFunction"
        condition:
          StringEquals:
            "aws:SourceAccount": "123456789012"
```

## Configuration Options

### allowBedrockAgents
- **Type**: boolean
- **Default**: true
- **Description**: Whether to include the default Bedrock agent permission

### statements
- **Type**: array of policy statements
- **Description**: Additional custom policy statements to include

### Policy Statement Structure
- **sid**: Statement identifier
- **effect**: "Allow" or "Deny"
- **principal**: Principal (Service, AWS, Federated, etc.)
- **action**: Action or list of actions
- **condition**: Optional condition block

## Security Best Practices

1. **Use agent-specific permissions**: The automatically generated agent-specific permissions are more secure than broad Bedrock service permissions

2. **Add account restrictions**: Use conditions to restrict access to specific AWS accounts:
   ```yaml
   condition:
     StringEquals:
       "aws:SourceAccount": "123456789012"
   ```

3. **Use source ARN patterns**: Restrict access to specific agent ARN patterns:
   ```yaml
   condition:
     StringLike:
       "aws:SourceArn": "arn:aws:bedrock:us-east-1:123456789012:agent/AGENT123*"
   ```

4. **Monitor invocations**: Enable CloudTrail and CloudWatch to monitor Lambda invocations from Bedrock agents

## Examples

- `lambda-with-resource-policy.yml`: Shows custom resource policy configuration
- `order-lookup/lambda.yml`: Basic Lambda without custom policies (uses defaults)
- `product-search-api/lambda.yml`: Another basic example

## How It Works (Behind the Scenes)

1. **Dependency Detection**: The Lambda generator scans all agents to see which ones reference this Lambda function
2. **Intelligent Policy Creation**: 
   - If agents reference the Lambda → creates agent-specific permissions with source ARN restrictions
   - If no agents reference it → creates general Bedrock permission for future use
3. **Automatic Security**: Agent-specific permissions use `aws:SourceArn` conditions to restrict access to only the specific agents
4. **Zero Configuration**: Users just define basic Lambda properties - no policy configuration needed

## Benefits

✅ **Zero Configuration**: No manual policy setup required  
✅ **Maximum Security**: Agent-specific permissions with source ARN restrictions  
✅ **Automatic Detection**: Intelligent analysis of agent-Lambda relationships  
✅ **Future-Proof**: General permissions for Lambda functions not yet used by agents  
✅ **Enterprise Ready**: Optional custom policies for advanced scenarios  

**Bottom Line**: Just create your Lambda function and reference it in your agent's action groups. bedrock-forge handles all the security automatically!