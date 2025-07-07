# Getting Started with Bedrock Forge

This guide walks you through creating your first AWS Bedrock agent using Bedrock Forge, from initial setup to deployment.

## Prerequisites

Before starting, ensure you have:

- **Go 1.21+** installed
- **AWS CLI** configured with appropriate credentials
- **Terraform 1.0+** installed
- **Git repository** for your project
- **AWS account** with Bedrock access enabled

## Step 1: Installation

### Clone and Build

```bash
# Clone the repository
git clone https://github.com/your-org/bedrock-forge
cd bedrock-forge

# Build the binary
go build -o bedrock-forge ./cmd/bedrock-forge

# Verify installation
./bedrock-forge version
```

### Add to PATH (Optional)

```bash
# Move binary to a directory in your PATH
sudo mv bedrock-forge /usr/local/bin/

# Now you can use it from anywhere
bedrock-forge version
```

## Step 2: Create Your First Agent

### Basic Agent

Create a simple agent that can answer questions:

```bash
# Create project directory
mkdir my-bedrock-project
cd my-bedrock-project

# Create agents directory
mkdir agents
```

Create `agents/my-first-agent.yml`:

```yaml
kind: Agent
metadata:
  name: "my-first-agent"
  description: "My first Bedrock agent"
spec:
  foundationModel: "anthropic.claude-3-sonnet-20240229-v1:0"
  instruction: |
    You are a helpful assistant. Provide clear, accurate, and friendly responses.
    Always be polite and professional in your interactions.
```

### Validate Your Configuration

```bash
# Validate the YAML syntax and dependencies
bedrock-forge validate .

# Scan for all resources
bedrock-forge scan .
```

Expected output:
```
✅ All YAML configurations are valid

📦 Resources Found:
├── Agent: my-first-agent
└── Total Resources Found: 1
```

### Generate Terraform

```bash
# Generate Terraform configuration
bedrock-forge generate . ./terraform

# Check generated files
ls -la terraform/
```

You should see:
- `main.tf` - Main Terraform configuration
- `variables.tf` - Variable definitions
- `outputs.tf` - Output values
- Auto-generated IAM roles and policies

### Deploy Your Agent

```bash
cd terraform

# Initialize Terraform
terraform init

# Review the deployment plan
terraform plan

# Deploy the agent
terraform apply
```

🎉 **Congratulations!** Your first Bedrock agent is now deployed and ready to use.

## Step 3: Add Lambda Functions

Let's enhance your agent with custom functions.

### Create Lambda Function

Create the directory structure:

```bash
mkdir -p lambdas/weather-function
```

Create `lambdas/weather-function/app.py`:

```python
import json
import random

def handler(event, context):
    """
    Simple weather function for demonstration
    """
    try:
        # Extract function details from Bedrock action group
        function_name = event.get('function', '')
        parameters = event.get('parameters', {})
        
        if function_name == 'get_weather':
            return get_weather(parameters)
        else:
            return {
                'statusCode': 400,
                'body': json.dumps({'error': 'Unknown function'})
            }
    
    except Exception as e:
        return {
            'statusCode': 500,
            'body': json.dumps({'error': str(e)})
        }

def get_weather(parameters):
    """Get weather for a location"""
    location = parameters.get('location', 'Unknown')
    
    # Simulate weather data (replace with real API call)
    weather_conditions = ['sunny', 'cloudy', 'rainy', 'snowy']
    temperature = random.randint(-10, 35)
    condition = random.choice(weather_conditions)
    
    weather_data = {
        'location': location,
        'temperature': f"{temperature}°C",
        'condition': condition,
        'humidity': f"{random.randint(30, 90)}%"
    }
    
    return {
        'statusCode': 200,
        'body': json.dumps(weather_data)
    }
```

Create `lambdas/weather-function/requirements.txt`:

```txt
# Add any Python dependencies here
boto3>=1.26.0
```

### Define Lambda Resource

Create `lambdas/weather-lambda.yml`:

```yaml
kind: Lambda
metadata:
  name: "weather-function"
  description: "Lambda function to get weather information"
spec:
  runtime: "python3.11"
  handler: "app.handler"
  description: "Weather information service"
  timeout: 30
  memorySize: 128
  
  environmentVariables:
    LOG_LEVEL: "INFO"
```

### Update Agent with Action Group

Update `agents/my-first-agent.yml`:

```yaml
kind: Agent
metadata:
  name: "my-first-agent"
  description: "My first Bedrock agent with weather capabilities"
spec:
  foundationModel: "anthropic.claude-3-sonnet-20240229-v1:0"
  instruction: |
    You are a helpful assistant with access to weather information.
    Use the get_weather function to provide current weather data for any location.
    Always be polite and provide clear, useful responses.
  
  # Add inline action group
  actionGroups:
    - name: "weather-tools"
      description: "Weather information tools"
      actionGroupExecutor:
        lambda: "weather-function"
      functionSchema:
        functions:
          - name: "get_weather"
            description: "Get current weather information for a location"
            parameters:
              location:
                type: "string"
                description: "City name or location to get weather for"
                required: true
```

### Deploy Updated Agent

```bash
# Validate the updated configuration
bedrock-forge validate .

# Generate updated Terraform
bedrock-forge generate . ./terraform

# Deploy changes
cd terraform
terraform plan
terraform apply
```

Now your agent can answer weather questions like:
- "What's the weather in New York?"
- "Can you check the weather in London?"

## Step 4: Add a Knowledge Base

Let's add a knowledge base so your agent can access company information.

### Prepare Knowledge Base Data

Create some sample documents:

```bash
mkdir -p knowledge-base-data/faq
```

Create `knowledge-base-data/faq/company-info.txt`:

```txt
Company Information

Our company was founded in 2020 and specializes in AI-powered customer support solutions.

Business Hours:
- Monday to Friday: 9 AM to 6 PM EST
- Saturday: 10 AM to 4 PM EST
- Sunday: Closed

Contact Information:
- Email: support@company.com
- Phone: 1-800-555-0123

Services:
- AI Agent Development
- Customer Support Automation
- Technical Consulting
```

Create `knowledge-base-data/faq/products.txt`:

```txt
Product Information

AI Assistant Pro:
- Advanced conversational AI
- Multi-language support
- Custom integrations
- Starting at $99/month

Business Intelligence Suite:
- Real-time analytics
- Custom dashboards
- API access
- Starting at $199/month

Enterprise Platform:
- Full white-label solution
- Dedicated support
- Custom development
- Contact for pricing
```

### Upload to S3

```bash
# Create S3 bucket for knowledge base
aws s3 mb s3://your-company-knowledge-base

# Upload documents
aws s3 sync knowledge-base-data/ s3://your-company-knowledge-base/
```

### Create OpenSearch Collection

Create `infrastructure/opensearch-collection.yml`:

```yaml
kind: CustomModule
metadata:
  name: "opensearch-kb"
  description: "OpenSearch Serverless collection for knowledge base"
spec:
  source: "./modules/opensearch-serverless"
  
  variables:
    collection_name: "bedrock-knowledge-base"
    
    # Access policy for Bedrock
    access_policies:
      - name: "bedrock-access"
        type: "data"
        description: "Access policy for Bedrock knowledge base"
        policy:
          Rules:
            - ResourceType: "collection"
              Resource: ["collection/bedrock-knowledge-base"]
              Permission: ["aoss:*"]
            - ResourceType: "index"
              Resource: ["index/bedrock-knowledge-base/*"]
              Permission: ["aoss:*"]
```

### Define Knowledge Base

Create `knowledge-bases/company-kb.yml`:

```yaml
kind: KnowledgeBase
metadata:
  name: "company-knowledge-base"
  description: "Company information and FAQ knowledge base"
spec:
  description: "Comprehensive company information for customer support"
  
  knowledgeBaseConfiguration:
    type: "VECTOR"
    vectorKnowledgeBaseConfiguration:
      embeddingModelArn: "arn:aws:bedrock:us-east-1::foundation-model/amazon.titan-embed-text-v1"
  
  storageConfiguration:
    type: "OPENSEARCH_SERVERLESS"
    opensearchServerlessConfiguration:
      collectionArn: "arn:aws:aoss:us-east-1:123456789012:collection/bedrock-knowledge-base"
      vectorIndexName: "company-kb-index"
  
  dataSources:
    - name: "company-documents"
      type: "S3"
      description: "Company FAQ and information documents"
      s3Configuration:
        bucketArn: "arn:aws:s3:::your-company-knowledge-base"
        inclusionPrefixes: ["faq/"]
      
      chunkingConfiguration:
        chunkingStrategy: "FIXED_SIZE"
        fixedSizeChunkingConfiguration:
          maxTokens: 512
          overlapPercentage: 20
```

### Associate Knowledge Base with Agent

Create `associations/agent-kb-association.yml`:

```yaml
kind: AgentKnowledgeBaseAssociation
metadata:
  name: "agent-company-kb"
  description: "Associate company knowledge base with agent"
spec:
  agentName: "my-first-agent"
  knowledgeBaseName: "company-knowledge-base"
  description: "Company information access for customer support"
  knowledgeBaseState: "ENABLED"
```

### Deploy Complete Setup

```bash
# Validate all configurations
bedrock-forge validate .

# Generate Terraform
bedrock-forge generate . ./terraform

# Deploy infrastructure and knowledge base
cd terraform
terraform plan
terraform apply
```

Now your agent can answer questions about your company using the knowledge base!

## Step 5: Test Your Agent

### Using AWS Console

1. Go to the AWS Bedrock console
2. Navigate to "Agents"
3. Find your agent "my-first-agent"
4. Click "Test" to open the test interface
5. Try these questions:
   - "What are your business hours?"
   - "What's the weather in Seattle?"
   - "Tell me about your products"

### Using AWS CLI

```bash
# Get agent details
aws bedrock-agent get-agent --agent-id <your-agent-id>

# Test the agent (requires agent alias)
aws bedrock-agent-runtime invoke-agent \
  --agent-id <your-agent-id> \
  --agent-alias-id <alias-id> \
  --session-id "test-session-1" \
  --input-text "What are your business hours?"
```

## Step 6: Set Up CI/CD (Optional)

For production deployments, set up GitHub Actions:

### Copy Workflow

```bash
mkdir -p .github/workflows
cp .github/workflows/bedrock-forge-deploy.yml .github/workflows/
```

### Configure Repository Secrets

In your GitHub repository settings, add:

- `AWS_ROLE_ARN`: Your AWS deployment role ARN
- `TF_STATE_BUCKET`: S3 bucket for Terraform state
- `TF_STATE_LOCK_TABLE`: DynamoDB table for state locking

### Create Workflow

Create `.github/workflows/deploy.yml`:

```yaml
name: Deploy Bedrock Agents

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  deploy:
    uses: your-org/bedrock-forge/.github/workflows/bedrock-forge-deploy.yml@main
    with:
      environment: ${{ github.ref == 'refs/heads/main' && 'production' || 'development' }}
      aws_region: us-east-1
      aws_role: ${{ vars.AWS_DEPLOYMENT_ROLE }}
      tf_state_bucket: ${{ vars.TF_STATE_BUCKET }}
      validation_only: ${{ github.event_name == 'pull_request' }}
    secrets: inherit
```

## Next Steps

### Advanced Features

1. **Add Guardrails**: Implement content safety and compliance
2. **Multiple Environments**: Set up dev/staging/prod deployments
3. **Monitoring**: Add CloudWatch dashboards and alerts
4. **Custom Modules**: Integrate with existing infrastructure

### Best Practices

1. **Version Control**: Keep all configurations in Git
2. **Environment Variables**: Use Terraform variables for environment-specific values
3. **Testing**: Test agents thoroughly before production deployment
4. **Documentation**: Document your agent's capabilities and limitations
5. **Monitoring**: Set up alerts for agent performance and errors

### Resources

- [Resource Reference](resources/) - Detailed documentation for each resource type
- [IAM Management](iam-management.md) - Understanding auto-generated permissions
- [GitHub Actions Guide](github-actions-guide.md) - Setting up CI/CD
- [Enterprise Setup](enterprise-setup.md) - Multi-environment deployments

## Common Issues

### Agent Not Responding
**Symptoms**: Agent doesn't respond or gives generic answers
**Solutions**:
- Check agent instruction clarity
- Verify foundation model permissions
- Review CloudWatch logs for errors

### Lambda Function Errors
**Symptoms**: Action group functions fail
**Solutions**:
- Check Lambda function logs
- Verify function handler path
- Ensure proper IAM permissions

### Knowledge Base Not Working
**Symptoms**: Agent can't access knowledge base information
**Solutions**:
- Verify S3 data source permissions
- Check OpenSearch collection configuration
- Ensure knowledge base association is enabled

### Deployment Failures
**Symptoms**: Terraform apply fails
**Solutions**:
- Check AWS credentials and permissions
- Verify resource naming conflicts
- Review Terraform plan output

## Getting Help

- **Documentation**: [docs/](../docs/)
- **Examples**: [examples/](../examples/)
- **Issues**: [GitHub Issues](https://github.com/your-org/bedrock-forge/issues)
- **Discussions**: [GitHub Discussions](https://github.com/your-org/bedrock-forge/discussions)

---

**Congratulations!** You've successfully created and deployed your first Bedrock agent with Bedrock Forge. You now have a foundation for building more sophisticated AI agents with custom capabilities.