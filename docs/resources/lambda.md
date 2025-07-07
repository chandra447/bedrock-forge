# Lambda Resource

AWS Lambda functions with automatic code packaging and IAM role generation.

## Overview

Lambda resources define AWS Lambda functions with automatic code packaging, dependency management, and IAM role generation. They are commonly used as executors for Bedrock action groups.

## Basic Example

```yaml
kind: Lambda
metadata:
  name: "order-lookup"
  description: "Lambda function to look up customer orders"
spec:
  runtime: "python3.11"
  handler: "app.handler"
  description: "Lambda function to look up customer orders"
  # IAM role is automatically generated!
```

## Complete Example

```yaml
kind: Lambda
metadata:
  name: "customer-tools"
  description: "Customer management tools Lambda function"
spec:
  runtime: "python3.11"
  handler: "app.handler"
  description: "Comprehensive customer management tools"
  timeout: 60
  memorySize: 512
  
  # Environment variables
  environmentVariables:
    LOG_LEVEL: "INFO"
    API_URL: "https://api.company.com"
    DATABASE_URL: "postgresql://user:pass@host:5432/db"
    REGION: "us-east-1"
  
  # VPC configuration
  vpcConfig:
    securityGroupIds: ["sg-12345678", "sg-87654321"]
    subnetIds: ["subnet-12345678", "subnet-87654321"]
  
  # Dead letter queue
  deadLetterConfig:
    targetArn: "arn:aws:sqs:us-east-1:123456789012:dlq-queue"
  
  # Reserved concurrency
  reservedConcurrencyLimit: 10
  
  # Layers
  layers:
    - "arn:aws:lambda:us-east-1:123456789012:layer:shared-utils:1"
    - "arn:aws:lambda:us-east-1:123456789012:layer:database-drivers:2"
  
  # File system configuration
  fileSystemConfig:
    arn: "arn:aws:elasticfilesystem:us-east-1:123456789012:file-system/fs-12345678"
    localMountPath: "/mnt/efs"
  
  # Tracing
  tracingConfig:
    mode: "Active"
  
  # Tags
  tags:
    Environment: "production"
    Team: "customer-support"
    Application: "bedrock-agents"
```

## Specification

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `runtime` | string | Lambda runtime environment |
| `handler` | string | Function handler (e.g., "app.handler") |

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `description` | string | Function description |
| `timeout` | number | Function timeout in seconds (1-900) |
| `memorySize` | number | Memory allocation in MB (128-10240) |
| `environmentVariables` | object | Environment variables |
| `vpcConfig` | object | VPC configuration |
| `deadLetterConfig` | object | Dead letter queue configuration |
| `reservedConcurrencyLimit` | number | Reserved concurrency limit |
| `layers` | array | Lambda layer ARNs |
| `fileSystemConfig` | object | EFS file system configuration |
| `tracingConfig` | object | X-Ray tracing configuration |
| `tags` | object | Resource tags |

### Supported Runtimes

| Runtime | Description |
|---------|-------------|
| `python3.9` | Python 3.9 |
| `python3.10` | Python 3.10 |
| `python3.11` | Python 3.11 |
| `python3.12` | Python 3.12 |
| `nodejs18.x` | Node.js 18.x |
| `nodejs20.x` | Node.js 20.x |
| `java11` | Java 11 |
| `java17` | Java 17 |
| `java21` | Java 21 |
| `dotnet6` | .NET 6 |
| `dotnet8` | .NET 8 |
| `go1.x` | Go 1.x |
| `ruby3.2` | Ruby 3.2 |
| `ruby3.3` | Ruby 3.3 |

### Environment Variables

```yaml
environmentVariables:
  LOG_LEVEL: "INFO"
  API_URL: "https://api.example.com"
  DATABASE_HOST: "db.example.com"
  TIMEOUT: "30"
  ENABLE_CACHE: "true"
```

### VPC Configuration

```yaml
vpcConfig:
  securityGroupIds: 
    - "sg-12345678"
    - "sg-87654321"
  subnetIds:
    - "subnet-12345678"
    - "subnet-87654321"
    - "subnet-13579024"
```

### Dead Letter Configuration

```yaml
deadLetterConfig:
  targetArn: "arn:aws:sqs:us-east-1:123456789012:my-dlq"
  # OR
  targetArn: "arn:aws:sns:us-east-1:123456789012:my-topic"
```

### File System Configuration

```yaml
fileSystemConfig:
  arn: "arn:aws:elasticfilesystem:us-east-1:123456789012:file-system/fs-12345678"
  localMountPath: "/mnt/efs"
```

### Tracing Configuration

```yaml
tracingConfig:
  mode: "Active"    # "Active" or "PassThrough"
```

## Code Packaging

Bedrock Forge automatically packages Lambda function code based on runtime:

### Python Functions

**Directory Structure:**
```
lambdas/
├── my-function/
│   ├── app.py              # Main handler file
│   ├── requirements.txt    # Python dependencies
│   ├── utils.py           # Additional modules
│   └── config/
│       └── settings.py
```

**app.py Example:**
```python
import json
import boto3
from utils import helper_function

def handler(event, context):
    """
    Lambda function handler for Bedrock action groups
    """
    try:
        # Extract action group information
        action_group = event.get('actionGroup', '')
        function_name = event.get('function', '')
        parameters = event.get('parameters', {})
        
        # Route to appropriate function
        if function_name == 'lookup_order':
            return lookup_order(parameters)
        elif function_name == 'update_order':
            return update_order(parameters)
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

def lookup_order(parameters):
    """Look up order by ID"""
    order_id = parameters.get('order_id')
    
    # Your order lookup logic here
    order = {
        'id': order_id,
        'status': 'shipped',
        'items': ['item1', 'item2']
    }
    
    return {
        'statusCode': 200,
        'body': json.dumps(order)
    }
```

**requirements.txt:**
```txt
boto3>=1.26.0
requests>=2.28.0
psycopg2-binary>=2.9.0
```

### Node.js Functions

**Directory Structure:**
```
lambdas/
├── my-function/
│   ├── index.js           # Main handler file
│   ├── package.json       # Node.js dependencies
│   ├── utils.js          # Additional modules
│   └── config/
│       └── settings.js
```

**index.js Example:**
```javascript
const AWS = require('aws-sdk');

exports.handler = async (event, context) => {
    try {
        const actionGroup = event.actionGroup || '';
        const functionName = event.function || '';
        const parameters = event.parameters || {};
        
        // Route to appropriate function
        switch (functionName) {
            case 'lookup_order':
                return await lookupOrder(parameters);
            case 'update_order':
                return await updateOrder(parameters);
            default:
                return {
                    statusCode: 400,
                    body: JSON.stringify({ error: 'Unknown function' })
                };
        }
    } catch (error) {
        return {
            statusCode: 500,
            body: JSON.stringify({ error: error.message })
        };
    }
};

async function lookupOrder(parameters) {
    const orderId = parameters.order_id;
    
    // Your order lookup logic here
    const order = {
        id: orderId,
        status: 'shipped',
        items: ['item1', 'item2']
    };
    
    return {
        statusCode: 200,
        body: JSON.stringify(order)
    };
}
```

**package.json:**
```json
{
  "name": "my-function",
  "version": "1.0.0",
  "description": "Lambda function for Bedrock action groups",
  "main": "index.js",
  "dependencies": {
    "aws-sdk": "^2.1400.0",
    "axios": "^1.4.0"
  }
}
```

### Java Functions

**Directory Structure:**
```
lambdas/
├── my-function/
│   ├── pom.xml
│   └── src/
│       └── main/
│           └── java/
│               └── com/
│                   └── example/
│                       └── Handler.java
```

**Handler.java Example:**
```java
package com.example;

import com.amazonaws.services.lambda.runtime.Context;
import com.amazonaws.services.lambda.runtime.RequestHandler;
import com.fasterxml.jackson.databind.ObjectMapper;
import java.util.Map;

public class Handler implements RequestHandler<Map<String, Object>, Map<String, Object>> {
    
    private final ObjectMapper objectMapper = new ObjectMapper();
    
    @Override
    public Map<String, Object> handleRequest(Map<String, Object> event, Context context) {
        try {
            String actionGroup = (String) event.get("actionGroup");
            String functionName = (String) event.get("function");
            Map<String, Object> parameters = (Map<String, Object>) event.get("parameters");
            
            switch (functionName) {
                case "lookup_order":
                    return lookupOrder(parameters);
                case "update_order":
                    return updateOrder(parameters);
                default:
                    return Map.of(
                        "statusCode", 400,
                        "body", "{\"error\":\"Unknown function\"}"
                    );
            }
        } catch (Exception e) {
            return Map.of(
                "statusCode", 500,
                "body", "{\"error\":\"" + e.getMessage() + "\"}"
            );
        }
    }
    
    private Map<String, Object> lookupOrder(Map<String, Object> parameters) {
        String orderId = (String) parameters.get("order_id");
        
        // Your order lookup logic here
        Map<String, Object> order = Map.of(
            "id", orderId,
            "status", "shipped",
            "items", List.of("item1", "item2")
        );
        
        return Map.of(
            "statusCode", 200,
            "body", objectMapper.writeValueAsString(order)
        );
    }
}
```

## Auto-Generated IAM Permissions

Lambda functions automatically get IAM roles with these permissions:

### Basic Execution Role
- `logs:CreateLogGroup`
- `logs:CreateLogStream`
- `logs:PutLogEvents`

### VPC Access (if VPC config is specified)
- `ec2:CreateNetworkInterface`
- `ec2:DescribeNetworkInterfaces`
- `ec2:DeleteNetworkInterface`

### Additional Permissions (based on configuration)
- X-Ray tracing permissions (if tracing is enabled)
- EFS access permissions (if file system is configured)
- SQS/SNS permissions (if dead letter queue is configured)

## Action Group Integration

### Direct Integration

```yaml
# Lambda function
kind: Lambda
metadata:
  name: "order-tools"
spec:
  runtime: "python3.11"
  handler: "app.handler"

---
# Action group using the Lambda
kind: ActionGroup
metadata:
  name: "order-actions"
spec:
  agentName: "customer-agent"
  actionGroupExecutor:
    lambda: "order-tools"  # References Lambda above
```

### Inline with Agent

```yaml
# Lambda function
kind: Lambda
metadata:
  name: "customer-tools"
spec:
  runtime: "python3.11"
  handler: "app.handler"

---
# Agent with inline action group
kind: Agent
metadata:
  name: "customer-agent"
spec:
  foundationModel: "anthropic.claude-3-sonnet-20240229-v1:0"
  instruction: "You are a customer support agent"
  actionGroups:
    - name: "tools"
      description: "Customer tools"
      actionGroupExecutor:
        lambda: "customer-tools"
      functionSchema:
        functions:
          - name: "lookup_order"
            description: "Look up order"
            parameters:
              order_id:
                type: "string"
                required: true
```

## Common Patterns

### Simple API Client

```python
import json
import requests
import os

def handler(event, context):
    """Simple API client Lambda"""
    api_url = os.environ.get('API_URL')
    api_key = os.environ.get('API_KEY')
    
    function_name = event.get('function', '')
    parameters = event.get('parameters', {})
    
    if function_name == 'get_user_info':
        user_id = parameters.get('user_id')
        
        response = requests.get(
            f"{api_url}/users/{user_id}",
            headers={'Authorization': f'Bearer {api_key}'}
        )
        
        return {
            'statusCode': 200,
            'body': json.dumps(response.json())
        }
    
    return {
        'statusCode': 400,
        'body': json.dumps({'error': 'Unknown function'})
    }
```

### Database Integration

```python
import json
import psycopg2
import os

def handler(event, context):
    """Database integration Lambda"""
    db_url = os.environ.get('DATABASE_URL')
    
    function_name = event.get('function', '')
    parameters = event.get('parameters', {})
    
    conn = psycopg2.connect(db_url)
    cursor = conn.cursor()
    
    try:
        if function_name == 'get_orders':
            customer_id = parameters.get('customer_id')
            
            cursor.execute(
                "SELECT * FROM orders WHERE customer_id = %s",
                (customer_id,)
            )
            orders = cursor.fetchall()
            
            return {
                'statusCode': 200,
                'body': json.dumps({
                    'orders': [dict(zip([col[0] for col in cursor.description], row)) 
                              for row in orders]
                })
            }
    
    finally:
        cursor.close()
        conn.close()
    
    return {
        'statusCode': 400,
        'body': json.dumps({'error': 'Unknown function'})
    }
```

### Multi-Service Integration

```python
import json
import boto3
import requests

def handler(event, context):
    """Multi-service integration Lambda"""
    s3 = boto3.client('s3')
    dynamodb = boto3.resource('dynamodb')
    
    function_name = event.get('function', '')
    parameters = event.get('parameters', {})
    
    if function_name == 'process_document':
        bucket = parameters.get('bucket')
        key = parameters.get('key')
        
        # Download from S3
        response = s3.get_object(Bucket=bucket, Key=key)
        content = response['Body'].read()
        
        # Process document (example: extract text)
        processed_data = process_document_content(content)
        
        # Store in DynamoDB
        table = dynamodb.Table('processed_documents')
        table.put_item(Item={
            'document_id': key,
            'processed_data': processed_data,
            'timestamp': datetime.utcnow().isoformat()
        })
        
        return {
            'statusCode': 200,
            'body': json.dumps({
                'document_id': key,
                'status': 'processed'
            })
        }
```

## Best Practices

### Performance
1. **Optimize cold starts** by minimizing imports and initialization
2. **Use appropriate memory settings** based on workload
3. **Implement connection pooling** for database connections
4. **Use Lambda layers** for common dependencies
5. **Monitor and tune timeout values**

### Security
1. **Use IAM roles** instead of hard-coded credentials
2. **Encrypt sensitive environment variables**
3. **Validate all inputs** from action groups
4. **Use VPC** for accessing private resources
5. **Implement proper error handling**

### Reliability
1. **Configure dead letter queues** for failed invocations
2. **Implement idempotency** for critical operations
3. **Use appropriate retry logic**
4. **Monitor function metrics** and set up alerts
5. **Test with various input scenarios**

### Cost Optimization
1. **Right-size memory allocation**
2. **Use reserved concurrency** to control costs
3. **Implement efficient code paths**
4. **Clean up temporary resources**
5. **Monitor costs** and optimize regularly

## Dependencies

- **Code Directory**: Function code must exist in the specified directory
- **Dependencies**: Runtime-specific dependency files must be present
- **VPC Resources**: VPC, subnets, and security groups must exist (if VPC config is used)
- **Layers**: Referenced Lambda layers must exist
- **Dead Letter Queue**: Target SQS queue or SNS topic must exist (if configured)

## Generated Resources

- AWS Lambda Function
- IAM Role (execution role)
- IAM Policy (execution policy)
- Lambda function code package (ZIP file)
- S3 upload (for function code)

## Common Issues

### Package Size Limits
```
Error: Package size too large
```
**Solution**: Use Lambda layers for large dependencies or optimize package size.

### VPC Configuration Issues
```
Error: Function cannot connect to VPC
```
**Solution**: Ensure subnets have routes to internet and NAT gateway for external access.

### Permission Errors
```
Error: AccessDenied
```
**Solution**: Check IAM role permissions and trust relationships.

### Cold Start Performance
```
Error: Function timeout
```
**Solution**: Optimize initialization code and increase timeout if necessary.

## See Also

- [Action Group Resource](action-group.md)
- [Agent Resource](agent.md)
- [IAM Management](../iam-management.md)
- [AWS Lambda Documentation](https://docs.aws.amazon.com/lambda/)