# IAM Role Resource

Custom IAM roles for advanced scenarios requiring specific permissions.

## Overview

While Bedrock Forge automatically generates IAM roles for all resources, you can define custom IAM roles for enterprise scenarios requiring specific permissions, compliance requirements, or integration needs.

## When to Use Custom IAM Roles

- **Compliance Requirements**: Specific permission boundaries required by organization
- **Security Policies**: Least-privilege access patterns for sensitive environments
- **Integration Needs**: Access to specific AWS services beyond auto-generated permissions
- **Cross-Account Access**: Roles for multi-account setups
- **Custom Conditions**: Time-based, IP-based, or other conditional access

## Basic Example

```yaml
kind: IAMRole
metadata:
  name: "custom-agent-role"
  description: "Custom role for enterprise agent with specific permissions"
spec:
  assumeRolePolicy:
    version: "2012-10-17"
    statement:
      - effect: "Allow"
        principal:
          service: "bedrock.amazonaws.com"
        action: "sts:AssumeRole"
  
  policies:
    - policyArn: "arn:aws:iam::aws:policy/service-role/AmazonBedrockAgentResourcePolicy"
  
  inlinePolicies:
    - name: "CustomPermissions"
      policy:
        version: "2012-10-17"
        statement:
          - effect: "Allow"
            action: ["bedrock:InvokeModel"]
            resource: "arn:aws:bedrock:*::foundation-model/anthropic.claude-3-sonnet-20240229-v1:0"
```

## Complete Example

```yaml
kind: IAMRole
metadata:
  name: "enterprise-agent-role"
  description: "Enterprise agent role with restricted permissions"
spec:
  # Trust policy
  assumeRolePolicy:
    version: "2012-10-17"
    statement:
      - effect: "Allow"
        principal:
          service: "bedrock.amazonaws.com"
        action: "sts:AssumeRole"
        condition:
          StringEquals:
            "aws:RequestedRegion": "us-east-1"
  
  # AWS managed policies
  policies:
    - policyArn: "arn:aws:iam::aws:policy/service-role/AmazonBedrockAgentResourcePolicy"
  
  # Custom inline policies
  inlinePolicies:
    - name: "RestrictedBedrockAccess"
      policy:
        version: "2012-10-17"
        statement:
          # Only specific foundation models
          - effect: "Allow"
            action: 
              - "bedrock:InvokeModel"
              - "bedrock:InvokeModelWithResponseStream"
            resource: 
              - "arn:aws:bedrock:*::foundation-model/anthropic.claude-3-sonnet-20240229-v1:0"
              - "arn:aws:bedrock:*::foundation-model/anthropic.claude-3-haiku-20240307-v1:0"
          
          # Restricted Lambda access
          - effect: "Allow"
            action: ["lambda:InvokeFunction"]
            resource: "arn:aws:lambda:*:*:function:approved-*"
            condition:
              StringEquals:
                "lambda:FunctionTag/Environment": "production"
    
    - name: "CustomS3Access"
      policy:
        version: "2012-10-17"
        statement:
          - effect: "Allow"
            action: 
              - "s3:GetObject"
              - "s3:PutObject"
            resource: "arn:aws:s3:::company-data-bucket/*"
            condition:
              StringEquals:
                "s3:ExistingObjectTag/Department": "CustomerSupport"
    
    - name: "AuditLogging"
      policy:
        version: "2012-10-17"
        statement:
          - effect: "Allow"
            action: 
              - "logs:CreateLogGroup"
              - "logs:CreateLogStream"
              - "logs:PutLogEvents"
            resource: "arn:aws:logs:*:*:log-group:/aws/bedrock/agents/*"
  
  # Resource tags
  tags:
    Environment: "production"
    Team: "ai-platform"
    CostCenter: "engineering"
```

## Specification

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `assumeRolePolicy` | object | Trust policy defining who can assume the role |

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `policies` | array | List of AWS managed policy ARNs |
| `inlinePolicies` | array | Custom inline policies |
| `tags` | object | Resource tags |
| `maxSessionDuration` | number | Maximum session duration in seconds (3600-43200) |

### Assume Role Policy

```yaml
assumeRolePolicy:
  version: "2012-10-17"
  statement:
    - effect: "Allow"
      principal:
        service: "bedrock.amazonaws.com"  # AWS service
        # OR
        aws: "arn:aws:iam::ACCOUNT:user/USER"  # AWS account/user
      action: "sts:AssumeRole"
      condition:  # Optional conditions
        StringEquals:
          "aws:RequestedRegion": "us-east-1"
```

### Managed Policies

```yaml
policies:
  - policyArn: "arn:aws:iam::aws:policy/service-role/AmazonBedrockAgentResourcePolicy"
  - policyArn: "arn:aws:iam::aws:policy/CloudWatchLogsFullAccess"
```

### Inline Policies

```yaml
inlinePolicies:
  - name: "PolicyName"
    policy:
      version: "2012-10-17"
      statement:
        - effect: "Allow"
          action: ["service:Action"]
          resource: "arn:aws:service:*:*:resource/*"
          condition:
            StringEquals:
              "service:Tag/Key": "Value"
```

## Common Patterns

### Least Privilege Agent Role

```yaml
kind: IAMRole
metadata:
  name: "least-privilege-agent-role"
spec:
  assumeRolePolicy:
    version: "2012-10-17"
    statement:
      - effect: "Allow"
        principal:
          service: "bedrock.amazonaws.com"
        action: "sts:AssumeRole"
  
  inlinePolicies:
    - name: "MinimalPermissions"
      policy:
        version: "2012-10-17"
        statement:
          # Only specific foundation model
          - effect: "Allow"
            action: ["bedrock:InvokeModel"]
            resource: "arn:aws:bedrock:us-east-1::foundation-model/anthropic.claude-3-sonnet-20240229-v1:0"
          
          # Only specific Lambda function
          - effect: "Allow"
            action: ["lambda:InvokeFunction"]
            resource: "arn:aws:lambda:us-east-1:123456789012:function:my-specific-function"
          
          # Minimal logging
          - effect: "Allow"
            action: ["logs:CreateLogStream", "logs:PutLogEvents"]
            resource: "arn:aws:logs:us-east-1:123456789012:log-group:/aws/bedrock/agents/my-agent:*"
```

### Cross-Account Access Role

```yaml
kind: IAMRole
metadata:
  name: "cross-account-agent-role"
spec:
  assumeRolePolicy:
    version: "2012-10-17"
    statement:
      - effect: "Allow"
        principal:
          service: "bedrock.amazonaws.com"
        action: "sts:AssumeRole"
      
      # Allow access from trusted account
      - effect: "Allow"
        principal:
          aws: "arn:aws:iam::TRUSTED-ACCOUNT-ID:root"
        action: "sts:AssumeRole"
        condition:
          StringEquals:
            "sts:ExternalId": "unique-external-id"
  
  inlinePolicies:
    - name: "CrossAccountPermissions"
      policy:
        version: "2012-10-17"
        statement:
          - effect: "Allow"
            action: ["bedrock:InvokeModel"]
            resource: "*"
          
          # Access resources in trusted account
          - effect: "Allow"
            action: ["s3:GetObject"]
            resource: "arn:aws:s3:::trusted-account-bucket/*"
```

### Time-Restricted Role

```yaml
kind: IAMRole
metadata:
  name: "time-restricted-agent-role"
spec:
  assumeRolePolicy:
    version: "2012-10-17"
    statement:
      - effect: "Allow"
        principal:
          service: "bedrock.amazonaws.com"
        action: "sts:AssumeRole"
  
  inlinePolicies:
    - name: "TimeRestrictedPermissions"
      policy:
        version: "2012-10-17"
        statement:
          - effect: "Allow"
            action: ["bedrock:InvokeModel"]
            resource: "*"
            condition:
              # Only during business hours (UTC)
              DateGreaterThan:
                "aws:CurrentTime": "2024-01-01T09:00:00Z"
              DateLessThan:
                "aws:CurrentTime": "2024-12-31T17:00:00Z"
              # Only on weekdays
              ForAnyValue:StringEquals:
                "aws:RequestedRegion": ["us-east-1", "us-west-2"]
```

### Environment-Specific Role

```yaml
kind: IAMRole
metadata:
  name: "env-specific-agent-role"
spec:
  assumeRolePolicy:
    version: "2012-10-17"
    statement:
      - effect: "Allow"
        principal:
          service: "bedrock.amazonaws.com"
        action: "sts:AssumeRole"
  
  inlinePolicies:
    - name: "EnvironmentPermissions"
      policy:
        version: "2012-10-17"
        statement:
          - effect: "Allow"
            action: ["bedrock:InvokeModel"]
            resource: "*"
          
          # Development environment
          - effect: "Allow"
            action: ["lambda:InvokeFunction"]
            resource: "arn:aws:lambda:*:*:function:dev-*"
            condition:
              StringEquals:
                "aws:RequestedRegion": "us-west-2"
          
          # Production environment
          - effect: "Allow"
            action: ["lambda:InvokeFunction"]
            resource: "arn:aws:lambda:*:*:function:prod-*"
            condition:
              StringEquals:
                "aws:RequestedRegion": "us-east-1"
```

## Using Custom Roles

### With Agents

```yaml
kind: Agent
metadata:
  name: "custom-agent"
spec:
  foundationModel: "anthropic.claude-3-sonnet-20240229-v1:0"
  instruction: "You are a custom agent"
  iamRole: "custom-agent-role"  # Reference to custom role
```

### With Lambda Functions

```yaml
kind: Lambda
metadata:
  name: "custom-lambda"
spec:
  runtime: "python3.11"
  handler: "app.handler"
  iamRole: "custom-lambda-role"  # Reference to custom role
```

## Best Practices

### Security
1. **Follow least privilege principle** - only grant necessary permissions
2. **Use resource-specific ARNs** instead of wildcards when possible
3. **Add condition blocks** for additional security constraints
4. **Regular audit** of custom roles and their usage
5. **Use AWS managed policies** when they meet your requirements

### Compliance
1. **Document role purpose** in metadata description
2. **Include required tags** per organizational policy
3. **Version control** role definitions
4. **Implement approval workflows** for role changes
5. **Regular compliance reviews** of permissions

### Management
1. **Use descriptive names** that indicate role purpose
2. **Group related permissions** in separate inline policies
3. **Avoid overly broad permissions** even if convenient
4. **Test roles** in development environments first
5. **Monitor role usage** with CloudTrail

## Generated Resources

- AWS IAM Role
- AWS IAM Policies (inline)
- Policy attachments (for managed policies)

## Common Issues

### Trust Policy Errors
```
Error: Cannot assume role - invalid trust policy
```
**Solution**: Ensure the trust policy allows the correct AWS service to assume the role.

### Permission Boundaries
```
Error: AccessDenied due to permission boundary
```
**Solution**: Check if organizational permission boundaries are restricting access.

### Resource ARN Mismatches
```
Error: AccessDenied - resource ARN doesn't match policy
```
**Solution**: Verify resource ARNs in policies match actual resource names.

### Condition Block Issues
```
Error: AccessDenied - condition not met
```
**Solution**: Check condition blocks for time, region, or tag restrictions.

## See Also

- [Agent Resource](agent.md)
- [Lambda Resource](lambda.md)
- [IAM Management](../iam-management.md)
- [AWS IAM Documentation](https://docs.aws.amazon.com/iam/)