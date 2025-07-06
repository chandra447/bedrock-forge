# Bedrock Forge GitHub Actions Workflow

This directory contains the GitHub Actions workflow for automatically deploying AWS Bedrock agents using Bedrock Forge.

## Quick Start

1. **Copy the workflow file** to your repository:
   ```bash
   mkdir -p .github/workflows
   cp bedrock-forge-deploy.yml .github/workflows/
   ```

2. **Configure repository variables** in your GitHub repository settings:
   - `AWS_DEPLOYMENT_ROLE`: ARN of the IAM role for deployments
   - `AWS_REGION`: AWS region (default: us-east-1)
   - `TF_STATE_BUCKET`: S3 bucket for Terraform state
   - `TF_STATE_KEY_PREFIX`: S3 key prefix for state files
   - `TF_STATE_LOCK_TABLE`: DynamoDB table for state locking

3. **Set up AWS OIDC** for secure authentication (recommended):
   ```bash
   # Create OIDC provider and deployment role
   aws iam create-open-id-connect-provider \
     --url https://token.actions.githubusercontent.com \
     --client-id-list sts.amazonaws.com
   ```

## Workflow Features

### 🔍 Validation Phase
- Validates all YAML configurations
- Scans for resources and optimizes pipeline
- Fails fast on configuration errors

### 📦 Packaging Phase
- Automatically packages Lambda functions
- Discovers and uploads OpenAPI schemas
- Generates unique S3 keys with versioning

### 🚀 Deployment Phase
- Generates Terraform configuration
- Configures S3 backend with state locking
- Plans and applies infrastructure changes
- Supports multiple environments

### 🧹 Cleanup Phase
- Provides deployment summary
- Reports success/failure status
- Uploads deployment artifacts

## Environment Support

The workflow supports multiple environments through:

- **Manual dispatch**: Select environment from dropdown
- **Environment-specific variables**: Different configs per environment
- **State isolation**: Separate Terraform state per environment
- **Approval workflows**: Configure deployment approvals

## Repository Variables

### Required Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `AWS_DEPLOYMENT_ROLE` | IAM role ARN for deployments | `arn:aws:iam::123456789012:role/BedrockForgeDeploymentRole` |
| `TF_STATE_BUCKET` | S3 bucket for Terraform state | `company-terraform-state` |
| `TF_STATE_KEY_PREFIX` | S3 key prefix for state files | `bedrock-forge` |
| `TF_STATE_LOCK_TABLE` | DynamoDB table for state locking | `terraform-state-lock` |

### Optional Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `AWS_REGION` | AWS region | `us-east-1` |
| `BEDROCK_FORGE_VERSION` | Bedrock Forge version | `latest` |
| `TF_VERSION` | Terraform version | `1.5.0` |

## IAM Role Setup

Create a deployment role with the following trust policy:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::123456789012:oidc-provider/token.actions.githubusercontent.com"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "token.actions.githubusercontent.com:aud": "sts.amazonaws.com"
        },
        "StringLike": {
          "token.actions.githubusercontent.com:sub": "repo:your-org/your-repo:*"
        }
      }
    }
  ]
}
```

### Required Permissions

The deployment role needs permissions for:

- **Bedrock**: Agent, knowledge base, and guardrail management
- **Lambda**: Function deployment and configuration
- **IAM**: Role and policy management
- **S3**: Bucket access for artifacts and state
- **DynamoDB**: State locking (if using)
- **OpenSearch**: Serverless collection management
- **CloudWatch**: Logging and monitoring

## Terraform State Management

### S3 Backend Configuration

The workflow automatically configures Terraform to use S3 backend:

```hcl
terraform {
  backend "s3" {
    bucket         = "company-terraform-state"
    key            = "bedrock-forge/dev/terraform.tfstate"
    region         = "us-east-1"
    dynamodb_table = "terraform-state-lock"
    encrypt        = true
  }
}
```

### State File Organization

State files are organized by environment:
- `bedrock-forge/dev/terraform.tfstate`
- `bedrock-forge/staging/terraform.tfstate`
- `bedrock-forge/prod/terraform.tfstate`

## Trigger Conditions

The workflow runs on:

1. **Push to main/master**: Automatic deployment to dev
2. **Pull requests**: Validation and planning only
3. **Manual dispatch**: Deploy to any environment

## Deployment Outputs

After successful deployment, the workflow provides:

- **Resource ARNs**: Agent, Lambda, and knowledge base ARNs
- **Deployment summary**: Environment, region, and state location
- **Terraform outputs**: All module outputs in JSON format

## Example Repository Structure

```
your-repo/
├── .github/workflows/
│   └── bedrock-forge-deploy.yml
├── agents/
│   └── customer-support.yml
├── lambdas/
│   └── order-lookup/
│       ├── app.py
│       ├── lambda.yml
│       └── requirements.txt
├── action-groups/
│   └── order-management/
│       ├── action-group.yml
│       └── openapi.json
├── knowledge-bases/
│   └── faq-kb.yml
└── forge.yml
```

## Troubleshooting

### Common Issues

1. **AWS credentials**: Ensure OIDC provider and role are correctly configured
2. **Terraform state**: Check S3 bucket permissions and DynamoDB table
3. **Resource validation**: Run `bedrock-forge validate` locally first
4. **Lambda packaging**: Ensure requirements.txt is present for Python functions

### Debug Mode

Enable debug logging by setting repository variable:
```
ACTIONS_STEP_DEBUG=true
```

## Advanced Configuration

### Custom Terraform Modules

Override default module sources in `forge.yml`:

```yaml
modules:
  registry: "git::https://github.com/your-org/bedrock-terraform-modules"
  version: "v2.0.0"
```

### Pre-deployment Hooks

Add custom steps before deployment:

```yaml
- name: Custom validation
  run: |
    # Your custom validation logic
    ./scripts/validate-business-rules.sh
```

### Multi-region Deployment

Deploy to multiple regions by matrix strategy:

```yaml
strategy:
  matrix:
    region: [us-east-1, us-west-2, eu-west-1]
```

## Security Best Practices

1. **Use OIDC**: Avoid long-lived AWS credentials
2. **Least privilege**: Grant minimal required permissions
3. **Environment protection**: Enable approval workflows for production
4. **State encryption**: Always encrypt Terraform state
5. **Secret management**: Use GitHub secrets for sensitive values

## Support

For issues and questions:
- Check the [troubleshooting guide](../docs/troubleshooting.md)
- Review workflow logs for detailed error messages
- Validate configurations locally before pushing