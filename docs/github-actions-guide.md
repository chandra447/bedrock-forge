# GitHub Actions Integration Guide

This guide shows how to use the Bedrock Forge reusable GitHub Actions workflow to deploy your AWS Bedrock agents and resources.

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Configuration Reference](#configuration-reference)
- [Authentication Setup](#authentication-setup)
- [Advanced Usage](#advanced-usage)
- [Troubleshooting](#troubleshooting)

## Overview

The Bedrock Forge reusable GitHub Actions workflow lets you deploy AWS Bedrock agents directly from your repository, just like using any other GitHub Action. Simply reference the workflow and provide your AWS role - no complex setup required.

Key features:
- **Simple Integration**: Use like `actions/checkout` - just specify the action and required parameters
- **Automated Deployment**: Deploy Bedrock agents on code changes
- **Multi-Environment Support**: Separate deployments for dev, staging, and production
- **Secure by Default**: Uses AWS IAM roles for authentication
- **Terraform State Management**: S3 backend with DynamoDB locking
- **OpenSearch Serverless**: Automated creation and configuration
- **Flexible Configuration**: Customizable versions, regions, and deployment options

## Quick Start

### 1. Prerequisites

Before using the Bedrock Forge workflow, ensure you have:

- AWS account with appropriate permissions
- GitHub repository with your Bedrock YAML configurations
- AWS IAM role configured for GitHub Actions
- Terraform state S3 bucket (optional but recommended)
- DynamoDB table for state locking (optional but recommended)

### 2. Create Workflow File

Create `.github/workflows/deploy-bedrock-agents.yml` in your repository. Just like using `actions/checkout`, you specify the action and its required parameters:

```yaml
name: Deploy Bedrock Agents

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]
  workflow_dispatch:
    inputs:
      environment:
        description: 'Environment to deploy'
        required: true
        default: 'dev'
        type: choice
        options: [dev, staging, prod]
      dry_run:
        description: 'Run in dry-run mode (plan only)'
        required: false
        default: false
        type: boolean

jobs:
  deploy:
    uses: your-org/bedrock-forge/.github/workflows/bedrock-forge-deploy.yml@main
    with:
      aws_role: 'arn:aws:iam::123456789012:role/BedrockForgeDeploymentRole'
      aws_region: 'us-east-1'
      environment: ${{ inputs.environment || 'dev' }}
      dry_run: ${{ inputs.dry_run || false }}
      tf_state_bucket: 'my-terraform-state-bucket'
```

**Note:** Replace `your-org/bedrock-forge` with the actual repository where you publish this action.

## Configuration Reference

### Required Inputs

| Input | Description | Example |
|-------|-------------|---------|
| `aws_role` | AWS IAM Role ARN to assume | `arn:aws:iam::123456789012:role/BedrockForgeRole` |

### Optional Inputs

| Input | Description | Default | Example |
|-------|-------------|---------|---------|
| `environment` | Deployment environment | `dev` | `prod` |
| `aws_region` | AWS region | `us-east-1` | `us-west-2` |
| `aws_session_name` | AWS session name | `bedrock-forge-deploy` | `my-deploy-session` |
| `terraform_version` | Terraform version | `1.5.0` | `1.6.0` |
| `go_version` | Go version | `1.21` | `1.22` |
| `bedrock_forge_version` | Bedrock Forge version/ref | `main` | `v1.0.0` |
| `source_path` | Path to YAML configs | `.` | `./configs` |
| `tf_state_bucket` | S3 bucket for state | None | `my-tf-state` |
| `tf_state_key_prefix` | State key prefix | `bedrock-forge` | `my-project` |
| `tf_state_lock_table` | DynamoDB lock table | None | `terraform-locks` |
| `dry_run` | Plan only mode | `false` | `true` |

### Secrets (Alternative Authentication)

| Secret | Description | When Required |
|--------|-------------|---------------|
| `AWS_ACCESS_KEY_ID` | AWS Access Key ID | When not using IAM roles |
| `AWS_SECRET_ACCESS_KEY` | AWS Secret Access Key | When not using IAM roles |

## Authentication Setup

### AWS IAM Role Setup

Configure an AWS IAM role with the required permissions and trust policy for GitHub Actions:

**Required IAM Permissions**:
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "bedrock:*",
        "lambda:*",
        "aoss:*",
        "iam:CreateRole",
        "iam:DeleteRole",
        "iam:AttachRolePolicy",
        "iam:DetachRolePolicy",
        "iam:PutRolePolicy",
        "iam:DeleteRolePolicy",
        "iam:GetRole",
        "iam:PassRole",
        "s3:GetObject",
        "s3:PutObject",
        "s3:DeleteObject",
        "s3:ListBucket",
        "dynamodb:GetItem",
        "dynamodb:PutItem",
        "dynamodb:DeleteItem",
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "*"
    }
  ]
}
```

**Note**: For specific instructions on setting up GitHub OIDC with AWS, refer to the [official AWS documentation](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_providers_create_oidc_github.html).

### Terraform Backend Setup (Optional)

#### S3 Bucket for State
```bash
# Create S3 bucket
aws s3 mb s3://my-terraform-state-bucket

# Enable versioning
aws s3api put-bucket-versioning \
  --bucket my-terraform-state-bucket \
  --versioning-configuration Status=Enabled
```

#### DynamoDB for State Locking
```bash
aws dynamodb create-table \
  --table-name terraform-locks \
  --attribute-definitions AttributeName=LockID,AttributeType=S \
  --key-schema AttributeName=LockID,KeyType=HASH \
  --billing-mode PAY_PER_REQUEST
```

## Advanced Usage

### Multi-Environment Deployment

```yaml
name: Multi-Environment Deploy

on:
  push:
    branches: [main]

jobs:
  deploy-dev:
    uses: your-org/bedrock-forge/.github/workflows/bedrock-forge-deploy.yml@main
    with:
      aws_role: 'arn:aws:iam::123456789012:role/BedrockForgeDevRole'
      environment: 'dev'
      tf_state_bucket: 'dev-terraform-state'

  deploy-staging:
    needs: deploy-dev
    if: github.ref == 'refs/heads/main'
    uses: your-org/bedrock-forge/.github/workflows/bedrock-forge-deploy.yml@main
    with:
      aws_role: 'arn:aws:iam::123456789012:role/BedrockForgeStagingRole'
      environment: 'staging'
      tf_state_bucket: 'staging-terraform-state'

  deploy-prod:
    needs: deploy-staging
    if: github.ref == 'refs/heads/main'
    environment: production  # Requires manual approval
    uses: your-org/bedrock-forge/.github/workflows/bedrock-forge-deploy.yml@main
    with:
      aws_role: 'arn:aws:iam::123456789012:role/BedrockForgeProdRole'
      environment: 'prod'
      tf_state_bucket: 'prod-terraform-state'
```

### Cross-Region Deployment

```yaml
name: Cross-Region Deploy

on:
  workflow_dispatch:
    inputs:
      regions:
        description: 'Regions to deploy (comma-separated)'
        required: true
        default: 'us-east-1,us-west-2'

jobs:
  matrix-deploy:
    strategy:
      matrix:
        region: ${{ fromJson(format('["{0}"]', join(split(inputs.regions, ','), '","'))) }}
    
    uses: your-org/bedrock-forge/.github/workflows/bedrock-forge-deploy.yml@main
    with:
      aws_role: 'arn:aws:iam::123456789012:role/BedrockForgeRole'
      aws_region: ${{ matrix.region }}
      environment: 'prod'
      tf_state_bucket: 'terraform-state-${{ matrix.region }}'
```

### Dry Run for Pull Requests

```yaml
name: PR Validation

on:
  pull_request:
    branches: [main]

jobs:
  validate:
    uses: your-org/bedrock-forge/.github/workflows/bedrock-forge-deploy.yml@main
    with:
      aws_role: 'arn:aws:iam::123456789012:role/BedrockForgeReadOnlyRole'
      environment: 'dev'
      dry_run: true  # Only plan, don't apply
      tf_state_bucket: 'dev-terraform-state'
```

### Example Repository Structure

```
my-bedrock-project/
├── .github/
│   └── workflows/
│       └── deploy-bedrock-agents.yml
├── agents/
│   ├── customer-support.yml
│   └── sales-assistant.yml
├── lambdas/
│   ├── order-lookup/
│   │   ├── lambda.yml
│   │   ├── app.py
│   │   └── requirements.txt
│   └── product-search/
│       ├── lambda.yml
│       └── index.js
├── knowledge-bases/
│   └── company-docs.yml
├── opensearch/
│   └── vector-store.yml
└── README.md
```

## Troubleshooting

### Common Issues

1. **Permission Denied**
   - Ensure your AWS role has all required permissions
   - Check if the role trust policy allows your GitHub repository
   - Verify the role ARN is correct

2. **Terraform State Lock**
   - Check if DynamoDB table exists and is accessible
   - Ensure the table name matches your configuration
   - Look for stuck locks in DynamoDB console

3. **Build Failures**
   - Verify Go version compatibility with your Lambda functions
   - Check if your YAML files are valid
   - Ensure all required dependencies are included

4. **Resource Conflicts**
   - Use unique resource names across environments
   - Check for naming conflicts in AWS console
   - Verify region-specific resource availability

### Debugging Tips

1. **Enable Debug Logs**: Set repository variables:
   - `ACTIONS_STEP_DEBUG: true`
   - `ACTIONS_RUNNER_DEBUG: true`

2. **Check Workflow Outputs**: Review the job summaries for detailed information

3. **Validate Locally**: Test your configurations before committing:
   ```bash
   # Build Bedrock Forge locally
   go build -o bedrock-forge ./cmd/bedrock-forge
   
   # Validate your configurations
   ./bedrock-forge validate ./your-configs
   
   # Test generation
   ./bedrock-forge generate ./your-configs ./output
   ```

4. **Test AWS Authentication**:
   ```bash
   aws sts get-caller-identity
   aws bedrock list-agents
   ```

### Getting Help

- Check the [Bedrock Forge documentation](../README.md)
- Review example configurations in the `examples/` directory
- File issues on the [GitHub repository](https://github.com/your-org/bedrock-forge/issues)

## Security Best Practices

1. **Use IAM roles instead of access keys** when possible
2. **Limit role permissions** to only what's needed
3. **Use separate roles per environment**
4. **Enable CloudTrail** for audit logging
5. **Store sensitive data in GitHub secrets**, not in YAML files
6. **Use environment protection rules** for production deployments
7. **Regularly review and rotate credentials**
8. **Monitor deployment activities** and set up alerts

## Best Practices

### Repository Organization
- Keep configurations in organized directories (`agents/`, `lambdas/`, etc.)
- Use consistent naming conventions
- Version control all configuration changes
- Document environment-specific settings

### Deployment Strategy
- Use pull requests for code review
- Implement automated testing in CI pipeline
- Use environment promotion (dev → staging → prod)
- Enable manual approval for production deployments
- Monitor deployments with health checks

### Cost Management
- Use appropriate instance sizes per environment
- Implement resource tagging for cost tracking
- Set up budget alerts
- Clean up unused resources regularly
- Monitor usage patterns and optimize accordingly

---

This guide provides everything you need to use the Bedrock Forge reusable GitHub Actions workflow to deploy your AWS Bedrock agents and resources efficiently and securely.