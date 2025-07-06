# Shared Lambda Action Group Example

This example demonstrates how to create an Action Group that uses an existing Lambda function shared by another team or service.

## Overview

Instead of defining a new Lambda function in the same project, this Action Group references an existing Lambda function using its ARN. This is useful when:

- Another team maintains a shared service Lambda function
- You want to reuse existing Lambda functions across multiple Bedrock agents
- The Lambda function is managed in a different AWS account or deployment pipeline

## Configuration

The key difference is using `lambdaArn` instead of `lambda` in the `actionGroupExecutor`:

```yaml
actionGroupExecutor:
  # Use lambdaArn for existing Lambda functions
  lambdaArn: "arn:aws:lambda:us-east-1:123456789012:function:inventory-team-prod-inventory-service"
```

## Important Security Considerations

When using shared Lambda functions, ensure that:

1. **IAM Permissions**: The Bedrock agent's execution role has permission to invoke the shared Lambda function
2. **Resource Policy**: The Lambda function's resource policy allows the Bedrock agent to invoke it
3. **Cross-Account Access**: If the Lambda is in a different AWS account, proper cross-account permissions are configured

### Required IAM Permissions

Add this policy to your Bedrock agent's execution role:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "lambda:InvokeFunction"
      ],
      "Resource": [
        "arn:aws:lambda:us-east-1:123456789012:function:inventory-team-prod-inventory-service"
      ]
    }
  ]
}
```

### Lambda Resource Policy

The shared Lambda function should have a resource policy allowing the agent role:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:iam::123456789012:role/bedrock-agent-execution-role"
      },
      "Action": "lambda:InvokeFunction",
      "Resource": "arn:aws:lambda:us-east-1:123456789012:function:inventory-team-prod-inventory-service"
    }
  ]
}
```

## Files

- `action-group.yml`: Action Group configuration using shared Lambda ARN
- `openapi.json`: OpenAPI schema defining the Lambda function's API
- `README.md`: This documentation file

## Usage

1. Replace the example ARN with your actual shared Lambda function ARN
2. Update the OpenAPI schema to match your Lambda function's interface
3. Ensure proper IAM permissions are configured
4. Deploy using `bedrock-forge generate`

## Validation

This Action Group will pass validation because:
- No dependency validation is performed for `lambdaArn` (external resource)
- The OpenAPI schema is provided for API definition
- All required fields are properly configured

The system automatically detects when `lambdaArn` is used and skips dependency validation for the Lambda function.