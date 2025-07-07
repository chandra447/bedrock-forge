# IAM Management

Bedrock Forge automatically generates IAM roles and policies for all Bedrock resources, with options for custom configurations in enterprise scenarios.

## Overview

**🎉 IAM roles are automatically generated for all Bedrock resources!** No manual configuration required.

Every Bedrock resource gets appropriate IAM permissions automatically:
- **Agents**: Foundation model access, Lambda invocation, knowledge base access, CloudWatch logging
- **Lambda Functions**: Execution roles with VPC access and CloudWatch logging
- **Action Groups**: Inherit permissions from associated agent roles
- **Knowledge Bases**: S3 access and OpenSearch operations

## Auto-Generated Permissions

### Agent IAM Roles

When you create an Agent, Bedrock Forge automatically generates:

1. **IAM Role**: `bedrock-agent-{agent-name}-role`
2. **IAM Policy**: `bedrock-agent-{agent-name}-policy`

#### Foundation Model Access
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "bedrock:InvokeModel",
        "bedrock:InvokeModelWithResponseStream"
      ],
      "Resource": "*"
    }
  ]
}
```

#### Lambda Invocation (for Action Groups)
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "lambda:InvokeFunction"
      ],
      "Resource": "arn:aws:lambda:*:*:function:*"
    }
  ]
}
```

#### Knowledge Base Access
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "bedrock:Retrieve",
        "bedrock:RetrieveAndGenerate"
      ],
      "Resource": "*"
    }
  ]
}
```

#### CloudWatch Logging
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "arn:aws:logs:*:*:*"
    }
  ]
}
```

### Lambda IAM Roles

Lambda functions get execution roles with:

#### Basic Execution
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "arn:aws:logs:*:*:*"
    }
  ]
}
```

#### VPC Access (if configured)
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ec2:CreateNetworkInterface",
        "ec2:DescribeNetworkInterfaces",
        "ec2:DeleteNetworkInterface"
      ],
      "Resource": "*"
    }
  ]
}
```

### Knowledge Base IAM Roles

Knowledge bases get roles with:

#### S3 Access
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetObject",
        "s3:ListBucket"
      ],
      "Resource": [
        "arn:aws:s3:::knowledge-base-bucket",
        "arn:aws:s3:::knowledge-base-bucket/*"
      ]
    }
  ]
}
```

#### OpenSearch Access
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "aoss:APIAccessAll"
      ],
      "Resource": "arn:aws:aoss:*:*:collection/*"
    }
  ]
}
```

## Custom IAM Roles

For enterprise scenarios requiring specific permissions, you can define custom IAM roles.

### When to Use Custom Roles

- **Compliance Requirements**: Specific permission boundaries
- **Security Policies**: Least-privilege access patterns
- **Integration Needs**: Access to specific AWS services
- **Cross-Account Access**: Roles for multi-account setups

### Custom Role Definition

```yaml
kind: IAMRole
metadata:
  name: "custom-agent-role"
  description: "Custom role for enterprise agent"
spec:
  assumeRolePolicy:
    version: "2012-10-17"
    statement:
      - effect: "Allow"
        principal:
          service: "bedrock.amazonaws.com"
        action: "sts:AssumeRole"
  
  # AWS managed policies
  policies:
    - policyArn: "arn:aws:iam::aws:policy/service-role/AmazonBedrockAgentResourcePolicy"
  
  # Custom inline policies
  inlinePolicies:
    - name: "CustomBedrockPermissions"
      policy:
        version: "2012-10-17"
        statement:
          - effect: "Allow"
            action: 
              - "bedrock:InvokeModel"
              - "bedrock:InvokeModelWithResponseStream"
            resource: 
              - "arn:aws:bedrock:*::foundation-model/anthropic.claude-3-sonnet-20240229-v1:0"
              - "arn:aws:bedrock:*::foundation-model/anthropic.claude-3-haiku-20240307-v1:0"
          
          - effect: "Allow"
            action: ["lambda:InvokeFunction"]
            resource: "arn:aws:lambda:*:*:function:customer-support-*"
    
    - name: "CustomS3Access"
      policy:
        version: "2012-10-17"
        statement:
          - effect: "Allow"
            action: 
              - "s3:GetObject"
              - "s3:PutObject"
            resource: "arn:aws:s3:::company-data-bucket/*"
```

### Using Custom Roles with Agents

```yaml
kind: Agent
metadata:
  name: "enterprise-agent"
spec:
  foundationModel: "anthropic.claude-3-sonnet-20240229-v1:0"
  instruction: "You are an enterprise assistant"
  # Custom IAM role reference
  iamRole: "custom-agent-role"
```

## Enterprise Patterns

### Least Privilege Access

```yaml
kind: IAMRole
metadata:
  name: "restricted-agent-role"
spec:
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
  
  inlinePolicies:
    - name: "RestrictedPermissions"
      policy:
        version: "2012-10-17"
        statement:
          # Only specific foundation models
          - effect: "Allow"
            action: ["bedrock:InvokeModel"]
            resource: "arn:aws:bedrock:us-east-1::foundation-model/anthropic.claude-3-sonnet-20240229-v1:0"
          
          # Only specific Lambda functions
          - effect: "Allow"
            action: ["lambda:InvokeFunction"]
            resource: "arn:aws:lambda:us-east-1:123456789012:function:approved-*"
          
          # Time-based access
          - effect: "Allow"
            action: ["logs:*"]
            resource: "*"
            condition:
              DateGreaterThan:
                "aws:CurrentTime": "2024-01-01T00:00:00Z"
              DateLessThan:
                "aws:CurrentTime": "2024-12-31T23:59:59Z"
```

### Cross-Account Access

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
      
      # Allow cross-account access
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

### Multi-Environment Roles

```yaml
kind: IAMRole
metadata:
  name: "multi-env-agent-role"
spec:
  assumeRolePolicy:
    version: "2012-10-17"
    statement:
      - effect: "Allow"
        principal:
          service: "bedrock.amazonaws.com"
        action: "sts:AssumeRole"
  
  inlinePolicies:
    - name: "EnvironmentSpecificPermissions"
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

## Best Practices

### Security
1. **Use auto-generated roles** for most scenarios
2. **Implement least privilege** for custom roles
3. **Use resource-specific ARNs** instead of wildcards
4. **Add condition blocks** for additional security
5. **Regular role auditing** for compliance

### Management
1. **Document custom roles** with clear descriptions
2. **Version control** role definitions
3. **Test permissions** in development first
4. **Monitor role usage** with CloudTrail
5. **Rotate credentials** regularly

### Enterprise Compliance
1. **Follow organizational policies** for role naming
2. **Include required tags** on all resources
3. **Implement approval workflows** for custom roles
4. **Use SCPs** for additional boundaries
5. **Regular compliance audits** of generated roles

## Troubleshooting

### Common Issues

#### Permission Denied Errors
```
Error: AccessDenied: User is not authorized to perform: bedrock:InvokeModel
```
**Solution**: Ensure the agent's IAM role has foundation model permissions.

#### Lambda Invocation Failures
```
Error: AccessDenied: is not authorized to perform: lambda:InvokeFunction
```
**Solution**: Verify Lambda function ARN matches the role's permissions.

#### Knowledge Base Access Issues
```
Error: AccessDenied: is not authorized to perform: bedrock:Retrieve
```
**Solution**: Check knowledge base permissions in the agent's role.

### Debugging IAM Issues

1. **Check CloudTrail logs** for detailed error messages
2. **Use IAM policy simulator** to test permissions
3. **Verify resource ARNs** in policies
4. **Check condition blocks** for time/region restrictions
5. **Review trust relationships** for role assumptions

## See Also

- [Agent Resource](resources/agent.md)
- [Lambda Resource](resources/lambda.md)
- [Knowledge Base Resource](resources/knowledge-base.md)
- [IAM Role Resource](resources/iam-role.md)