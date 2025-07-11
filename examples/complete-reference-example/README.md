# Complete Reference Example

This example demonstrates the **NEW reference syntax** in bedrock-forge, showcasing both object and string syntax patterns for cross-resource references.

## 🎯 New Reference Syntax

Bedrock-forge now supports two reference formats:

### Object Syntax (Recommended)
```yaml
lambda: {ref: lambda-name}
guardrail:
  name: {ref: guardrail-name}
```

### String Syntax (Legacy)
```yaml
lambda: "lambda-name"
guardrail:
  name: "guardrail-name"
```

## Architecture

This example creates a complete customer support system with:

1. ~~**Custom Infrastructure** (CustomResources) - SNS topic and EventBridge rules~~ (Temporarily disabled)
2. ~~**OpenSearch Serverless** collection for vector storage~~ (Temporarily disabled) 
3. **Lambda Functions** for order lookup
4. ~~**Knowledge Base** with company documentation~~ (Temporarily disabled)
5. **Guardrail** for content safety ✅
6. **Prompt Templates** for orchestration ✅
7. **Agent** with action groups ✅
8. ~~**Agent-Knowledge Base Association** for RAG capabilities~~ (Temporarily disabled)

## Reference Syntax Patterns

### Object Syntax (Recommended)
```yaml
lambda: {ref: order-lookup-function}
guardrail:
  name: {ref: content-safety-guardrail}
```

### String Syntax (Legacy)
```yaml
lambda: "order-lookup-function"
guardrail:
  name: "content-safety-guardrail"
```

## Deployment Order

The resources will be deployed in the following order based on dependencies:

1. CustomResources (infrastructure)
2. OpenSearchServerless
3. Guardrail
4. Prompt
5. Lambda functions
6. KnowledgeBase
7. Agent
8. AgentKnowledgeBaseAssociation

## Working Files

- `simple-agent.yml` - **WORKING EXAMPLE** showcasing all reference patterns ✅
- `03-guardrail.yml` - Content safety policies
- `04-prompt.yml` - Custom prompt templates  
- `05-lambda.yml` - Function definitions
- `07-agent.yml` - Main conversational agent (complex version)
- `08-association.yml` - Agent-KB linking

## Disabled Files (Type Issues)

- `01-infrastructure.yml.disabled` - Custom SNS and EventBridge resources
- `02-opensearch.yml` - Vector storage collection
- `06-knowledge-base.yml` - Vector knowledge base

## Quick Start

To see the reference syntax in action:

```bash
# Test the working example
bedrock-forge generate examples/complete-reference-example/simple-agent.yml ./output

# Check the generated Terraform
cat output/main.tf | grep "module\."
```

You'll see references like:
- `{ref: customer-support-safety}` → `${module.customer_support_safety.guardrail_id}`
- `{ref: order-lookup-function}` → `${module.order_lookup_function.lambda_function_arn}`
- `{ref: customer-support-orchestration}` → `${module.customer_support_orchestration.prompt_arn}`

## Usage

```bash
# Validate configuration
bedrock-forge validate .

# Generate Terraform
bedrock-forge generate . ./output

# Deploy (requires AWS credentials)
cd output && terraform init && terraform apply
```

This example demonstrates all reference patterns and dependency management capabilities of bedrock-forge.