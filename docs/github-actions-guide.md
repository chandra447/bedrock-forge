# GitHub Actions Integration Guide

This guide provides comprehensive information on setting up and using Bedrock Forge with GitHub Actions for automated AWS Bedrock agent deployments.

## Table of Contents

- [Overview](#overview)
- [Setup Process](#setup-process)
- [Security Configuration](#security-configuration)
- [Workflow Configuration](#workflow-configuration)
- [Environment Management](#environment-management)
- [Advanced Usage](#advanced-usage)
- [Monitoring and Troubleshooting](#monitoring-and-troubleshooting)

## Overview

The Bedrock Forge GitHub Actions workflow provides:
- **Automated Deployment**: Deploy Bedrock agents on code changes
- **Multi-Environment Support**: Separate deployments for dev, staging, and production
- **Security Best Practices**: AWS OIDC authentication without long-lived credentials
- **Terraform State Management**: S3 backend with DynamoDB locking
- **Deployment Approvals**: Environment protection with approval workflows

## Setup Process

### 1. Prerequisites

Before setting up the GitHub Actions workflow, ensure you have:

- AWS account with appropriate permissions
- GitHub repository with admin access
- Terraform state S3 bucket
- DynamoDB table for state locking (optional but recommended)

### 2. Copy Workflow Files

Copy the Bedrock Forge workflow to your repository:

```bash
# Create workflows directory
mkdir -p .github/workflows

# Copy the main workflow
cp /path/to/bedrock-forge/.github/workflows/bedrock-forge-deploy.yml .github/workflows/

# Copy documentation and setup scripts
cp -r /path/to/bedrock-forge/.github/workflows/README.md .github/workflows/
cp /path/to/bedrock-forge/.github/workflows/setup-aws-oidc.sh .github/workflows/
```

### 3. AWS OIDC Configuration

#### Automated Setup

Use the provided setup script for automated OIDC configuration:

```bash
# Make the script executable
chmod +x .github/workflows/setup-aws-oidc.sh

# Set environment variables
export AWS_ACCOUNT_ID="123456789012"
export GITHUB_ORG="your-org"
export GITHUB_REPO="your-repo"
export ROLE_NAME="BedrockForgeDeploymentRole"

# Run the setup script
./.github/workflows/setup-aws-oidc.sh
```

#### Manual Setup

If you prefer manual setup:

1. **Create OIDC Provider**:
```bash
aws iam create-open-id-connect-provider \
  --url https://token.actions.githubusercontent.com \
  --client-id-list sts.amazonaws.com \
  --thumbprint-list 6938fd4d98bab03faadb97b34396831e3780aea1
```

2. **Create Trust Policy**:
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

3. **Create IAM Role**:
```bash
aws iam create-role \
  --role-name BedrockForgeDeploymentRole \
  --assume-role-policy-document file://trust-policy.json
```

4. **Attach Policies**:
```bash
# Attach the deployment policy (created by setup script)
aws iam attach-role-policy \
  --role-name BedrockForgeDeploymentRole \
  --policy-arn arn:aws:iam::123456789012:policy/BedrockForgeDeploymentPolicy
```

### 4. GitHub Repository Configuration

#### Repository Variables

Configure the following variables in your GitHub repository settings (Settings → Secrets and variables → Actions → Variables):

| Variable | Description | Example |
|----------|-------------|---------|
| `AWS_DEPLOYMENT_ROLE` | IAM role ARN for deployments | `arn:aws:iam::123456789012:role/BedrockForgeDeploymentRole` |
| `AWS_REGION` | AWS region for deployments | `us-east-1` |
| `TF_STATE_BUCKET` | S3 bucket for Terraform state | `your-terraform-state-bucket` |
| `TF_STATE_KEY_PREFIX` | Prefix for state file keys | `bedrock-forge` |
| `TF_STATE_LOCK_TABLE` | DynamoDB table for state locking | `terraform-state-lock` |

#### Environment Secrets (Optional)

For environment-specific configurations, you can set secrets at the environment level:

- `AWS_DEPLOYMENT_ROLE_DEV`
- `AWS_DEPLOYMENT_ROLE_STAGING`
- `AWS_DEPLOYMENT_ROLE_PROD`

### 5. Terraform Backend Setup

#### S3 Bucket Creation

```bash
# Create S3 bucket for Terraform state
aws s3 mb s3://your-terraform-state-bucket

# Enable versioning
aws s3api put-bucket-versioning \
  --bucket your-terraform-state-bucket \
  --versioning-configuration Status=Enabled

# Enable encryption
aws s3api put-bucket-encryption \
  --bucket your-terraform-state-bucket \
  --server-side-encryption-configuration '{
    "Rules": [
      {
        "ApplyServerSideEncryptionByDefault": {
          "SSEAlgorithm": "AES256"
        }
      }
    ]
  }'
```

#### DynamoDB Table for State Locking

```bash
# Create DynamoDB table for state locking
aws dynamodb create-table \
  --table-name terraform-state-lock \
  --attribute-definitions AttributeName=LockID,AttributeType=S \
  --key-schema AttributeName=LockID,KeyType=HASH \
  --billing-mode PAY_PER_REQUEST \
  --tags Key=Purpose,Value=TerraformStateLocking Key=Project,Value=BedrockForge
```

## Security Configuration

### AWS IAM Permissions

The deployment role requires comprehensive permissions for Bedrock services:

#### Bedrock Permissions
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "BedrockAgentManagement",
      "Effect": "Allow",
      "Action": [
        "bedrock:CreateAgent",
        "bedrock:UpdateAgent",
        "bedrock:DeleteAgent",
        "bedrock:GetAgent",
        "bedrock:ListAgents",
        "bedrock:CreateAgentActionGroup",
        "bedrock:UpdateAgentActionGroup",
        "bedrock:DeleteAgentActionGroup",
        "bedrock:CreateAgentAlias",
        "bedrock:UpdateAgentAlias",
        "bedrock:DeleteAgentAlias"
      ],
      "Resource": "*"
    }
  ]
}
```

#### Lambda Permissions
```json
{
  "Sid": "LambdaManagement",
  "Effect": "Allow",
  "Action": [
    "lambda:CreateFunction",
    "lambda:UpdateFunctionCode",
    "lambda:UpdateFunctionConfiguration",
    "lambda:DeleteFunction",
    "lambda:GetFunction",
    "lambda:AddPermission",
    "lambda:RemovePermission"
  ],
  "Resource": "*"
}
```

#### IAM Permissions
```json
{
  "Sid": "IAMManagement",
  "Effect": "Allow",
  "Action": [
    "iam:CreateRole",
    "iam:UpdateRole",
    "iam:DeleteRole",
    "iam:GetRole",
    "iam:AttachRolePolicy",
    "iam:DetachRolePolicy",
    "iam:PutRolePolicy",
    "iam:DeleteRolePolicy",
    "iam:PassRole"
  ],
  "Resource": "*"
}
```

### Environment Protection

#### Branch Protection Rules

Configure branch protection for your main branch:

```yaml
# .github/branch-protection.yml
protection_rules:
  main:
    required_reviews: 2
    dismiss_stale_reviews: true
    require_code_owner_reviews: true
    required_status_checks:
      - "Validate Configuration"
      - "Security Scan"
```

#### Environment Approval Workflows

Configure environment protection rules in GitHub:

1. Go to Settings → Environments
2. Create environments: `dev`, `staging`, `prod`
3. Configure protection rules:
   - **Development**: No restrictions
   - **Staging**: Require 1 reviewer
   - **Production**: Require 2 reviewers + deployment branch restriction

## Workflow Configuration

### Basic Workflow

```yaml
# .github/workflows/deploy.yml
name: Deploy Bedrock Agents

on:
  push:
    branches: [ main ]
    paths:
      - 'agents/**'
      - 'lambdas/**'
      - 'action-groups/**'
      - 'knowledge-bases/**'
      - 'iam-roles/**'
  
  pull_request:
    branches: [ main ]
    paths:
      - 'agents/**'
      - 'lambdas/**'
      - 'action-groups/**'
      - 'knowledge-bases/**'
      - 'iam-roles/**'
  
  workflow_dispatch:
    inputs:
      environment:
        description: 'Environment to deploy'
        required: true
        default: 'dev'
        type: choice
        options:
          - dev
          - staging
          - prod
      
      force_deploy:
        description: 'Force deployment (skip validation)'
        required: false
        default: false
        type: boolean

jobs:
  deploy:
    uses: ./.github/workflows/bedrock-forge-deploy.yml
    with:
      environment: ${{ github.event.inputs.environment || 'dev' }}
      force_deploy: ${{ github.event.inputs.force_deploy || false }}
      working_directory: '.'
    secrets: inherit
```

### Advanced Workflow with Matrix Strategy

```yaml
name: Multi-Region Deployment

on:
  workflow_dispatch:
    inputs:
      environment:
        description: 'Environment'
        required: true
        type: choice
        options: [staging, prod]

jobs:
  deploy:
    strategy:
      matrix:
        region: [us-east-1, us-west-2, eu-west-1]
        include:
          - region: us-east-1
            primary: true
          - region: us-west-2
            primary: false
          - region: eu-west-1
            primary: false
    
    uses: ./.github/workflows/bedrock-forge-deploy.yml
    with:
      environment: ${{ github.event.inputs.environment }}
      aws_region: ${{ matrix.region }}
      is_primary_region: ${{ matrix.primary }}
    secrets: inherit
```

### Custom Validation Workflow

```yaml
name: Custom Validation

on:
  pull_request:
    branches: [ main ]

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Bedrock Forge
        run: |
          curl -L https://github.com/your-org/bedrock-forge/releases/latest/download/bedrock-forge-linux-amd64 -o bedrock-forge
          chmod +x bedrock-forge
      
      - name: Validate Configuration
        run: |
          ./bedrock-forge validate .
      
      - name: Check Security Policies
        run: |
          # Custom security validation
          python scripts/validate-iam-policies.py
      
      - name: Cost Estimation
        run: |
          # Generate cost estimates
          ./bedrock-forge generate . ./terraform
          terraform plan -out=plan.tfplan
          terraform show -json plan.tfplan | python scripts/cost-estimate.py
```

## Environment Management

### Environment-Specific Configurations

#### Development Environment
```yaml
# environments/dev.yml
metadata:
  environment: "dev"

terraform:
  backend:
    key: "dev/terraform.tfstate"

variables:
  log_level: "DEBUG"
  retain_logs: false
  instance_size: "small"

tags:
  Environment: "development"
  CostCenter: "engineering"
  AutoShutdown: "true"
```

#### Production Environment
```yaml
# environments/prod.yml
metadata:
  environment: "prod"

terraform:
  backend:
    key: "prod/terraform.tfstate"

variables:
  log_level: "WARN"
  retain_logs: true
  instance_size: "large"
  enable_monitoring: true
  backup_enabled: true

tags:
  Environment: "production"
  CostCenter: "operations"
  AutoShutdown: "false"
  Compliance: "required"
```

### Deployment Strategies

#### Blue-Green Deployment
```yaml
name: Blue-Green Deployment

jobs:
  deploy-blue:
    if: github.event.inputs.deployment_type == 'blue-green'
    uses: ./.github/workflows/bedrock-forge-deploy.yml
    with:
      environment: ${{ github.event.inputs.environment }}-blue
      deployment_slot: "blue"
    secrets: inherit
  
  validate-blue:
    needs: deploy-blue
    runs-on: ubuntu-latest
    steps:
      - name: Run Integration Tests
        run: |
          python tests/integration_tests.py \
            --endpoint ${{ needs.deploy-blue.outputs.agent_endpoint }} \
            --environment blue
  
  switch-traffic:
    needs: validate-blue
    runs-on: ubuntu-latest
    environment: ${{ github.event.inputs.environment }}
    steps:
      - name: Switch Traffic to Blue
        run: |
          # Update load balancer or API Gateway routing
          aws apigateway update-stage \
            --rest-api-id ${{ vars.API_GATEWAY_ID }} \
            --stage-name prod \
            --patch-ops op=replace,path=/variables/deployment_slot,value=blue
```

#### Canary Deployment
```yaml
name: Canary Deployment

jobs:
  deploy-canary:
    uses: ./.github/workflows/bedrock-forge-deploy.yml
    with:
      environment: ${{ github.event.inputs.environment }}-canary
      traffic_percentage: 10
    secrets: inherit
  
  monitor-canary:
    needs: deploy-canary
    runs-on: ubuntu-latest
    steps:
      - name: Monitor Metrics
        run: |
          python scripts/monitor-canary.py \
            --duration 30m \
            --error-threshold 1% \
            --latency-threshold 500ms
  
  promote-canary:
    needs: monitor-canary
    if: success()
    uses: ./.github/workflows/bedrock-forge-deploy.yml
    with:
      environment: ${{ github.event.inputs.environment }}
      traffic_percentage: 100
    secrets: inherit
```

## Advanced Usage

### Custom Pre/Post Deployment Steps

```yaml
name: Custom Deployment Pipeline

jobs:
  pre-deployment:
    runs-on: ubuntu-latest
    steps:
      - name: Backup Current State
        run: |
          # Create backup of current deployment
          aws bedrock describe-agent --agent-id ${{ vars.AGENT_ID }} > backup/agent-state.json
      
      - name: Notify Team
        run: |
          # Send Slack notification
          curl -X POST -H 'Content-type: application/json' \
            --data '{"text":"Starting deployment to ${{ github.event.inputs.environment }}"}' \
            ${{ secrets.SLACK_WEBHOOK_URL }}
  
  deploy:
    needs: pre-deployment
    uses: ./.github/workflows/bedrock-forge-deploy.yml
    with:
      environment: ${{ github.event.inputs.environment }}
    secrets: inherit
  
  post-deployment:
    needs: deploy
    runs-on: ubuntu-latest
    if: always()
    steps:
      - name: Run Smoke Tests
        run: |
          python tests/smoke_tests.py \
            --agent-id ${{ needs.deploy.outputs.agent_id }} \
            --environment ${{ github.event.inputs.environment }}
      
      - name: Update Documentation
        run: |
          # Update deployment documentation
          python scripts/update-docs.py \
            --deployment-id ${{ github.run_id }} \
            --environment ${{ github.event.inputs.environment }}
      
      - name: Notify Team
        if: always()
        run: |
          STATUS="${{ job.status }}"
          MESSAGE="Deployment to ${{ github.event.inputs.environment }} completed with status: $STATUS"
          curl -X POST -H 'Content-type: application/json' \
            --data "{\"text\":\"$MESSAGE\"}" \
            ${{ secrets.SLACK_WEBHOOK_URL }}
```

### Integration with External Systems

#### ServiceNow Integration
```yaml
name: ServiceNow Integration

jobs:
  create-change-request:
    if: github.event.inputs.environment == 'prod'
    runs-on: ubuntu-latest
    outputs:
      change_request_id: ${{ steps.create-cr.outputs.change_request_id }}
    steps:
      - name: Create Change Request
        id: create-cr
        run: |
          # Create ServiceNow change request
          RESPONSE=$(curl -X POST \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer ${{ secrets.SERVICENOW_TOKEN }}" \
            -d '{
              "short_description": "Bedrock Agent Deployment to Production",
              "description": "Automated deployment via GitHub Actions",
              "category": "Software",
              "priority": "3"
            }' \
            "${{ vars.SERVICENOW_URL }}/api/now/table/change_request")
          
          CHANGE_ID=$(echo $RESPONSE | jq -r '.result.number')
          echo "change_request_id=$CHANGE_ID" >> $GITHUB_OUTPUT
  
  deploy:
    needs: create-change-request
    if: always() && (needs.create-change-request.result == 'success' || github.event.inputs.environment != 'prod')
    uses: ./.github/workflows/bedrock-forge-deploy.yml
    with:
      environment: ${{ github.event.inputs.environment }}
      change_request_id: ${{ needs.create-change-request.outputs.change_request_id }}
    secrets: inherit
  
  update-change-request:
    needs: [create-change-request, deploy]
    if: always() && needs.create-change-request.outputs.change_request_id
    runs-on: ubuntu-latest
    steps:
      - name: Update Change Request
        run: |
          STATUS="${{ needs.deploy.result }}"
          curl -X PATCH \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer ${{ secrets.SERVICENOW_TOKEN }}" \
            -d "{\"state\": \"$([ \"$STATUS\" == \"success\" ] && echo \"3\" || echo \"4\")\", \"work_notes\": \"Deployment $STATUS\"}" \
            "${{ vars.SERVICENOW_URL }}/api/now/table/change_request/${{ needs.create-change-request.outputs.change_request_id }}"
```

## Monitoring and Troubleshooting

### Workflow Monitoring

#### CloudWatch Integration
```yaml
name: Enhanced Monitoring

jobs:
  deploy:
    uses: ./.github/workflows/bedrock-forge-deploy.yml
    with:
      environment: ${{ github.event.inputs.environment }}
    secrets: inherit
  
  setup-monitoring:
    needs: deploy
    runs-on: ubuntu-latest
    steps:
      - name: Create CloudWatch Dashboard
        run: |
          aws cloudwatch put-dashboard \
            --dashboard-name "BedrockAgent-${{ github.event.inputs.environment }}" \
            --dashboard-body file://monitoring/dashboard.json
      
      - name: Set Up Alarms
        run: |
          # Create CloudWatch alarms for agent metrics
          aws cloudwatch put-metric-alarm \
            --alarm-name "BedrockAgent-ErrorRate-${{ github.event.inputs.environment }}" \
            --alarm-description "High error rate for Bedrock agent" \
            --metric-name "ErrorRate" \
            --namespace "AWS/Bedrock" \
            --statistic "Average" \
            --period 300 \
            --threshold 5.0 \
            --comparison-operator "GreaterThanThreshold" \
            --evaluation-periods 2
```

#### Prometheus/Grafana Integration
```yaml
name: Metrics Collection

jobs:
  deploy:
    uses: ./.github/workflows/bedrock-forge-deploy.yml
    with:
      environment: ${{ github.event.inputs.environment }}
    secrets: inherit
  
  configure-metrics:
    needs: deploy
    runs-on: ubuntu-latest
    steps:
      - name: Update Prometheus Config
        run: |
          # Add new targets to Prometheus configuration
          python scripts/update-prometheus-config.py \
            --agent-id ${{ needs.deploy.outputs.agent_id }} \
            --environment ${{ github.event.inputs.environment }}
      
      - name: Import Grafana Dashboard
        run: |
          # Import pre-built Grafana dashboard
          curl -X POST \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer ${{ secrets.GRAFANA_API_KEY }}" \
            -d @monitoring/grafana-dashboard.json \
            "${{ vars.GRAFANA_URL }}/api/dashboards/db"
```

### Troubleshooting Common Issues

#### 1. Authentication Failures
```yaml
name: Debug Authentication

jobs:
  debug-auth:
    runs-on: ubuntu-latest
    steps:
      - name: Test AWS Authentication
        run: |
          echo "Testing AWS authentication..."
          aws sts get-caller-identity
          echo "Current AWS region: $(aws configure get region)"
          echo "Available regions: $(aws ec2 describe-regions --query 'Regions[].RegionName' --output text)"
      
      - name: Test OIDC Token
        run: |
          echo "OIDC Token (first 50 chars): ${ACTIONS_ID_TOKEN_REQUEST_TOKEN:0:50}..."
          echo "OIDC URL: $ACTIONS_ID_TOKEN_REQUEST_URL"
```

#### 2. Terraform State Issues
```yaml
name: Debug Terraform State

jobs:
  debug-state:
    runs-on: ubuntu-latest
    steps:
      - name: Check S3 Backend
        run: |
          echo "Checking S3 bucket access..."
          aws s3 ls s3://${{ vars.TF_STATE_BUCKET }}/ || echo "Cannot access bucket"
          
          echo "Checking state file..."
          aws s3 ls s3://${{ vars.TF_STATE_BUCKET }}/${{ vars.TF_STATE_KEY_PREFIX }}/ || echo "No state files found"
      
      - name: Check DynamoDB Lock Table
        run: |
          echo "Checking DynamoDB lock table..."
          aws dynamodb describe-table --table-name ${{ vars.TF_STATE_LOCK_TABLE }} || echo "Lock table not accessible"
```

#### 3. Resource Validation
```yaml
name: Debug Resource Configuration

jobs:
  debug-resources:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Validate YAML Syntax
        run: |
          find . -name "*.yml" -o -name "*.yaml" | xargs -I {} sh -c 'echo "Validating {}" && python -c "import yaml; yaml.safe_load(open(\"{}\"))"'
      
      - name: Check Resource Dependencies
        run: |
          # Custom dependency validation
          python scripts/validate-dependencies.py --verbose
      
      - name: Generate Terraform (Dry Run)
        run: |
          ./bedrock-forge generate . ./terraform-debug --dry-run
```

### Performance Optimization

#### Parallel Deployments
```yaml
name: Optimized Deployment

jobs:
  validate:
    runs-on: ubuntu-latest
    outputs:
      agents: ${{ steps.scan.outputs.agents }}
      lambdas: ${{ steps.scan.outputs.lambdas }}
    steps:
      - uses: actions/checkout@v4
      - name: Scan Resources
        id: scan
        run: |
          # Scan and categorize resources for parallel processing
          AGENTS=$(./bedrock-forge scan . --type agent --json)
          LAMBDAS=$(./bedrock-forge scan . --type lambda --json)
          echo "agents=$AGENTS" >> $GITHUB_OUTPUT
          echo "lambdas=$LAMBDAS" >> $GITHUB_OUTPUT
  
  deploy-lambdas:
    needs: validate
    if: fromJSON(needs.validate.outputs.lambdas).count > 0
    strategy:
      matrix:
        lambda: ${{ fromJSON(needs.validate.outputs.lambdas).items }}
    runs-on: ubuntu-latest
    steps:
      - name: Deploy Lambda
        run: |
          ./bedrock-forge generate ./lambdas/${{ matrix.lambda }} ./terraform-lambda-${{ matrix.lambda }}
          cd terraform-lambda-${{ matrix.lambda }}
          terraform init && terraform apply -auto-approve
  
  deploy-agents:
    needs: [validate, deploy-lambdas]
    strategy:
      matrix:
        agent: ${{ fromJSON(needs.validate.outputs.agents).items }}
    runs-on: ubuntu-latest
    steps:
      - name: Deploy Agent
        run: |
          ./bedrock-forge generate ./agents/${{ matrix.agent }} ./terraform-agent-${{ matrix.agent }}
          cd terraform-agent-${{ matrix.agent }}
          terraform init && terraform apply -auto-approve
```

## Best Practices

### 1. Security
- Use OIDC authentication instead of long-lived credentials
- Implement least-privilege IAM policies
- Enable branch protection and required reviews
- Use environment-specific secrets and variables
- Regularly rotate access tokens and keys

### 2. Deployment Strategy
- Use pull requests for code review
- Implement automated testing in CI pipeline
- Use environment promotion (dev → staging → prod)
- Implement rollback procedures
- Monitor deployments with health checks

### 3. Resource Management
- Use consistent naming conventions
- Implement proper tagging strategy
- Monitor costs and set budget alerts
- Clean up unused resources regularly
- Document resource dependencies

### 4. Monitoring and Alerting
- Set up CloudWatch dashboards
- Configure alerts for critical metrics
- Implement log aggregation
- Use distributed tracing for complex workflows
- Regular health checks and synthetic monitoring

### 5. Documentation
- Keep deployment documentation up to date
- Document environment-specific configurations
- Maintain runbooks for common issues
- Version control all configuration changes
- Regular team training on deployment procedures

---

This comprehensive guide should help teams successfully implement and operate Bedrock Forge with GitHub Actions for enterprise-grade AWS Bedrock agent deployments.