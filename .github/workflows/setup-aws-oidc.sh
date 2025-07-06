#!/bin/bash

# Setup AWS OIDC for GitHub Actions
# This script helps configure AWS IAM for secure GitHub Actions authentication

set -e

# Configuration
AWS_ACCOUNT_ID=${AWS_ACCOUNT_ID:-$(aws sts get-caller-identity --query Account --output text)}
GITHUB_ORG=${GITHUB_ORG:-"your-org"}
GITHUB_REPO=${GITHUB_REPO:-"your-repo"}
ROLE_NAME=${ROLE_NAME:-"BedrockForgeDeploymentRole"}
OIDC_PROVIDER_URL="https://token.actions.githubusercontent.com"
OIDC_AUDIENCE="sts.amazonaws.com"

echo "🔧 Setting up AWS OIDC for GitHub Actions"
echo "Account ID: $AWS_ACCOUNT_ID"
echo "GitHub Repo: $GITHUB_ORG/$GITHUB_REPO"
echo "Role Name: $ROLE_NAME"
echo

# Check if OIDC provider exists
echo "📋 Checking OIDC provider..."
if aws iam get-open-id-connect-provider --open-id-connect-provider-arn "arn:aws:iam::$AWS_ACCOUNT_ID:oidc-provider/token.actions.githubusercontent.com" >/dev/null 2>&1; then
    echo "✅ OIDC provider already exists"
else
    echo "📝 Creating OIDC provider..."
    aws iam create-open-id-connect-provider \
        --url "$OIDC_PROVIDER_URL" \
        --client-id-list "$OIDC_AUDIENCE" \
        --thumbprint-list "6938fd4d98bab03faadb97b34396831e3780aea1" \
        --tags Key=Purpose,Value=GitHubActions Key=Project,Value=BedrockForge
    echo "✅ OIDC provider created"
fi

# Create trust policy
echo "📝 Creating trust policy..."
cat > trust-policy.json << EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::$AWS_ACCOUNT_ID:oidc-provider/token.actions.githubusercontent.com"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "token.actions.githubusercontent.com:aud": "sts.amazonaws.com"
        },
        "StringLike": {
          "token.actions.githubusercontent.com:sub": "repo:$GITHUB_ORG/$GITHUB_REPO:*"
        }
      }
    }
  ]
}
EOF

# Create deployment policy
echo "📝 Creating deployment policy..."
cat > deployment-policy.json << EOF
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
        "bedrock:GetAgentActionGroup",
        "bedrock:ListAgentActionGroups",
        "bedrock:CreateAgentAlias",
        "bedrock:UpdateAgentAlias",
        "bedrock:DeleteAgentAlias",
        "bedrock:GetAgentAlias",
        "bedrock:ListAgentAliases",
        "bedrock:AssociateAgentKnowledgeBase",
        "bedrock:DisassociateAgentKnowledgeBase",
        "bedrock:GetAgentKnowledgeBase",
        "bedrock:ListAgentKnowledgeBases"
      ],
      "Resource": "*"
    },
    {
      "Sid": "BedrockKnowledgeBaseManagement",
      "Effect": "Allow",
      "Action": [
        "bedrock:CreateKnowledgeBase",
        "bedrock:UpdateKnowledgeBase",
        "bedrock:DeleteKnowledgeBase",
        "bedrock:GetKnowledgeBase",
        "bedrock:ListKnowledgeBases",
        "bedrock:CreateDataSource",
        "bedrock:UpdateDataSource",
        "bedrock:DeleteDataSource",
        "bedrock:GetDataSource",
        "bedrock:ListDataSources",
        "bedrock:StartIngestionJob",
        "bedrock:StopIngestionJob",
        "bedrock:GetIngestionJob",
        "bedrock:ListIngestionJobs"
      ],
      "Resource": "*"
    },
    {
      "Sid": "BedrockGuardrailManagement",
      "Effect": "Allow",
      "Action": [
        "bedrock:CreateGuardrail",
        "bedrock:UpdateGuardrail",
        "bedrock:DeleteGuardrail",
        "bedrock:GetGuardrail",
        "bedrock:ListGuardrails",
        "bedrock:CreateGuardrailVersion",
        "bedrock:DeleteGuardrailVersion",
        "bedrock:GetGuardrailVersion",
        "bedrock:ListGuardrailVersions"
      ],
      "Resource": "*"
    },
    {
      "Sid": "BedrockPromptManagement",
      "Effect": "Allow",
      "Action": [
        "bedrock:CreatePrompt",
        "bedrock:UpdatePrompt",
        "bedrock:DeletePrompt",
        "bedrock:GetPrompt",
        "bedrock:ListPrompts",
        "bedrock:CreatePromptVersion",
        "bedrock:DeletePromptVersion",
        "bedrock:GetPromptVersion",
        "bedrock:ListPromptVersions"
      ],
      "Resource": "*"
    },
    {
      "Sid": "LambdaManagement",
      "Effect": "Allow",
      "Action": [
        "lambda:CreateFunction",
        "lambda:UpdateFunctionCode",
        "lambda:UpdateFunctionConfiguration",
        "lambda:DeleteFunction",
        "lambda:GetFunction",
        "lambda:ListFunctions",
        "lambda:AddPermission",
        "lambda:RemovePermission",
        "lambda:GetPolicy",
        "lambda:CreateAlias",
        "lambda:UpdateAlias",
        "lambda:DeleteAlias",
        "lambda:GetAlias",
        "lambda:ListAliases",
        "lambda:TagResource",
        "lambda:UntagResource",
        "lambda:ListTags"
      ],
      "Resource": "*"
    },
    {
      "Sid": "IAMManagement",
      "Effect": "Allow",
      "Action": [
        "iam:CreateRole",
        "iam:UpdateRole",
        "iam:DeleteRole",
        "iam:GetRole",
        "iam:ListRoles",
        "iam:AttachRolePolicy",
        "iam:DetachRolePolicy",
        "iam:PutRolePolicy",
        "iam:DeleteRolePolicy",
        "iam:GetRolePolicy",
        "iam:ListRolePolicies",
        "iam:ListAttachedRolePolicies",
        "iam:PassRole",
        "iam:TagRole",
        "iam:UntagRole",
        "iam:ListRoleTags"
      ],
      "Resource": "*"
    },
    {
      "Sid": "S3Management",
      "Effect": "Allow",
      "Action": [
        "s3:CreateBucket",
        "s3:DeleteBucket",
        "s3:GetBucketLocation",
        "s3:GetBucketPolicy",
        "s3:PutBucketPolicy",
        "s3:DeleteBucketPolicy",
        "s3:GetBucketVersioning",
        "s3:PutBucketVersioning",
        "s3:GetBucketEncryption",
        "s3:PutBucketEncryption",
        "s3:GetObject",
        "s3:PutObject",
        "s3:DeleteObject",
        "s3:ListBucket",
        "s3:GetBucketTagging",
        "s3:PutBucketTagging"
      ],
      "Resource": "*"
    },
    {
      "Sid": "OpenSearchServerless",
      "Effect": "Allow",
      "Action": [
        "aoss:CreateCollection",
        "aoss:UpdateCollection",
        "aoss:DeleteCollection",
        "aoss:GetCollection",
        "aoss:ListCollections",
        "aoss:CreateSecurityConfig",
        "aoss:UpdateSecurityConfig",
        "aoss:DeleteSecurityConfig",
        "aoss:GetSecurityConfig",
        "aoss:ListSecurityConfigs",
        "aoss:CreateAccessPolicy",
        "aoss:UpdateAccessPolicy",
        "aoss:DeleteAccessPolicy",
        "aoss:GetAccessPolicy",
        "aoss:ListAccessPolicies",
        "aoss:CreateSecurityPolicy",
        "aoss:UpdateSecurityPolicy",
        "aoss:DeleteSecurityPolicy",
        "aoss:GetSecurityPolicy",
        "aoss:ListSecurityPolicies"
      ],
      "Resource": "*"
    },
    {
      "Sid": "CloudWatchLogs",
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents",
        "logs:DescribeLogGroups",
        "logs:DescribeLogStreams",
        "logs:PutRetentionPolicy",
        "logs:TagLogGroup",
        "logs:UntagLogGroup"
      ],
      "Resource": "*"
    },
    {
      "Sid": "DynamoDBStateLocking",
      "Effect": "Allow",
      "Action": [
        "dynamodb:GetItem",
        "dynamodb:PutItem",
        "dynamodb:DeleteItem",
        "dynamodb:CreateTable",
        "dynamodb:DescribeTable",
        "dynamodb:TagResource",
        "dynamodb:UntagResource"
      ],
      "Resource": "*"
    }
  ]
}
EOF

# Create or update the role
echo "📝 Creating deployment role..."
if aws iam get-role --role-name "$ROLE_NAME" >/dev/null 2>&1; then
    echo "⚠️  Role $ROLE_NAME already exists, updating trust policy..."
    aws iam update-assume-role-policy \
        --role-name "$ROLE_NAME" \
        --policy-document file://trust-policy.json
else
    echo "📝 Creating new role..."
    aws iam create-role \
        --role-name "$ROLE_NAME" \
        --assume-role-policy-document file://trust-policy.json \
        --tags Key=Purpose,Value=GitHubActions Key=Project,Value=BedrockForge
fi

# Create and attach the deployment policy
POLICY_NAME="BedrockForgeDeploymentPolicy"
echo "📝 Creating deployment policy..."

if aws iam get-policy --policy-arn "arn:aws:iam::$AWS_ACCOUNT_ID:policy/$POLICY_NAME" >/dev/null 2>&1; then
    echo "⚠️  Policy $POLICY_NAME already exists, creating new version..."
    aws iam create-policy-version \
        --policy-arn "arn:aws:iam::$AWS_ACCOUNT_ID:policy/$POLICY_NAME" \
        --policy-document file://deployment-policy.json \
        --set-as-default
else
    echo "📝 Creating new policy..."
    aws iam create-policy \
        --policy-name "$POLICY_NAME" \
        --policy-document file://deployment-policy.json \
        --tags Key=Purpose,Value=GitHubActions Key=Project,Value=BedrockForge
fi

# Attach policy to role
echo "📝 Attaching policy to role..."
aws iam attach-role-policy \
    --role-name "$ROLE_NAME" \
    --policy-arn "arn:aws:iam::$AWS_ACCOUNT_ID:policy/$POLICY_NAME"

# Clean up temporary files
rm -f trust-policy.json deployment-policy.json

echo
echo "✅ AWS OIDC setup completed successfully!"
echo
echo "📋 Next steps:"
echo "1. Add these repository variables to your GitHub repository:"
echo "   - AWS_DEPLOYMENT_ROLE: arn:aws:iam::$AWS_ACCOUNT_ID:role/$ROLE_NAME"
echo "   - AWS_REGION: $(aws configure get region || echo "us-east-1")"
echo "   - TF_STATE_BUCKET: your-terraform-state-bucket"
echo "   - TF_STATE_KEY_PREFIX: bedrock-forge"
echo "   - TF_STATE_LOCK_TABLE: terraform-state-lock"
echo
echo "2. Create S3 bucket for Terraform state:"
echo "   aws s3 mb s3://your-terraform-state-bucket"
echo
echo "3. Create DynamoDB table for state locking:"
echo "   aws dynamodb create-table \\"
echo "     --table-name terraform-state-lock \\"
echo "     --attribute-definitions AttributeName=LockID,AttributeType=S \\"
echo "     --key-schema AttributeName=LockID,KeyType=HASH \\"
echo "     --billing-mode PAY_PER_REQUEST"
echo
echo "4. Test the workflow by pushing to your repository"
echo