# Knowledge Base Resource

Vector knowledge bases with S3 data sources and chunking strategies for enhanced agent capabilities.

## Overview

Knowledge Base resources create AWS Bedrock knowledge bases that enable agents to access and retrieve information from external data sources. They support vector search capabilities with customizable chunking and embedding strategies.

## Basic Example

```yaml
kind: KnowledgeBase
metadata:
  name: "faq-kb"
  description: "Customer FAQ knowledge base"
spec:
  description: "Customer FAQ knowledge base for support agents"
  
  knowledgeBaseConfiguration:
    type: "VECTOR"
    vectorKnowledgeBaseConfiguration:
      embeddingModelArn: "arn:aws:bedrock:us-east-1::foundation-model/amazon.titan-embed-text-v1"
  
  storageConfiguration:
    type: "OPENSEARCH_SERVERLESS"
    opensearchServerlessConfiguration:
      collectionArn: "arn:aws:aoss:us-east-1:123456789012:collection/bedrock-kb"
      vectorIndexName: "bedrock-knowledge-base-index"
  
  dataSources:
    - name: "faq-documents"
      type: "S3"
      s3Configuration:
        bucketArn: "arn:aws:s3:::company-kb-documents"
        inclusionPrefixes: ["faq/"]
```

## Complete Example

```yaml
kind: KnowledgeBase
metadata:
  name: "comprehensive-kb"
  description: "Comprehensive knowledge base with multiple data sources"
spec:
  description: "Enterprise knowledge base with comprehensive document coverage"
  
  # Knowledge base configuration
  knowledgeBaseConfiguration:
    type: "VECTOR"
    vectorKnowledgeBaseConfiguration:
      embeddingModelArn: "arn:aws:bedrock:us-east-1::foundation-model/amazon.titan-embed-text-v1"
      # OR use Cohere embeddings
      # embeddingModelArn: "arn:aws:bedrock:us-east-1::foundation-model/cohere.embed-english-v3"
  
  # Storage configuration
  storageConfiguration:
    type: "OPENSEARCH_SERVERLESS"
    opensearchServerlessConfiguration:
      collectionArn: "arn:aws:aoss:us-east-1:123456789012:collection/enterprise-kb"
      vectorIndexName: "enterprise-knowledge-index"
      fieldMapping:
        vectorField: "bedrock-knowledge-base-default-vector"
        textField: "AMAZON_BEDROCK_TEXT_CHUNK"
        metadataField: "AMAZON_BEDROCK_METADATA"
  
  # Multiple data sources
  dataSources:
    # FAQ documents
    - name: "faq-documents"
      type: "S3"
      description: "Customer FAQ documents"
      s3Configuration:
        bucketArn: "arn:aws:s3:::company-knowledge-base"
        inclusionPrefixes: ["faq/", "help/"]
        exclusionPrefixes: ["drafts/", "archived/"]
      
      chunkingConfiguration:
        chunkingStrategy: "FIXED_SIZE"
        fixedSizeChunkingConfiguration:
          maxTokens: 512
          overlapPercentage: 20
    
    # Product documentation
    - name: "product-docs"
      type: "S3"
      description: "Product documentation and guides"
      s3Configuration:
        bucketArn: "arn:aws:s3:::company-knowledge-base"
        inclusionPrefixes: ["products/", "guides/"]
      
      chunkingConfiguration:
        chunkingStrategy: "HIERARCHICAL"
        hierarchicalChunkingConfiguration:
          levelConfigurations:
            - maxTokens: 1500
            - maxTokens: 300
          overlapTokens: 60
    
    # Policy documents
    - name: "policies"
      type: "S3"
      description: "Company policies and procedures"
      s3Configuration:
        bucketArn: "arn:aws:s3:::company-knowledge-base"
        inclusionPrefixes: ["policies/"]
      
      chunkingConfiguration:
        chunkingStrategy: "SEMANTIC"
        semanticChunkingConfiguration:
          maxTokens: 800
          bufferSize: 0
          breakpointPercentileThreshold: 95
  
  # Tags
  tags:
    Environment: "production"
    Team: "customer-support"
    DataClassification: "internal"
```

## Specification

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `knowledgeBaseConfiguration` | object | Knowledge base type and embedding configuration |
| `storageConfiguration` | object | Vector storage configuration |

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `description` | string | Knowledge base description |
| `dataSources` | array | Data source configurations |
| `tags` | object | Resource tags |

### Knowledge Base Configuration

```yaml
knowledgeBaseConfiguration:
  type: "VECTOR"  # Currently only VECTOR is supported
  vectorKnowledgeBaseConfiguration:
    embeddingModelArn: "arn:aws:bedrock:region::foundation-model/model-id"
```

#### Supported Embedding Models

| Model | ARN |
|-------|-----|
| Amazon Titan Embed Text v1 | `arn:aws:bedrock:us-east-1::foundation-model/amazon.titan-embed-text-v1` |
| Amazon Titan Embed Text v2 | `arn:aws:bedrock:us-east-1::foundation-model/amazon.titan-embed-text-v2:0` |
| Cohere Embed English v3 | `arn:aws:bedrock:us-east-1::foundation-model/cohere.embed-english-v3` |
| Cohere Embed Multilingual v3 | `arn:aws:bedrock:us-east-1::foundation-model/cohere.embed-multilingual-v3` |

### Storage Configuration

#### OpenSearch Serverless

```yaml
storageConfiguration:
  type: "OPENSEARCH_SERVERLESS"
  opensearchServerlessConfiguration:
    collectionArn: "arn:aws:aoss:region:account:collection/collection-name"
    vectorIndexName: "vector-index-name"
    fieldMapping:  # Optional
      vectorField: "bedrock-knowledge-base-default-vector"
      textField: "AMAZON_BEDROCK_TEXT_CHUNK"
      metadataField: "AMAZON_BEDROCK_METADATA"
```

#### Pinecone (if supported)

```yaml
storageConfiguration:
  type: "PINECONE"
  pineconeConfiguration:
    connectionString: "https://your-index.pinecone.io"
    credentialsSecretArn: "arn:aws:secretsmanager:region:account:secret:pinecone-creds"
    namespace: "bedrock-namespace"
    fieldMapping:
      textField: "text"
      metadataField: "metadata"
```

### Data Sources

#### S3 Data Source

```yaml
dataSources:
  - name: "source-name"
    type: "S3"
    description: "Data source description"
    s3Configuration:
      bucketArn: "arn:aws:s3:::bucket-name"
      inclusionPrefixes: ["folder1/", "folder2/"]  # Optional
      exclusionPrefixes: ["drafts/", "temp/"]      # Optional
    
    chunkingConfiguration:  # Optional
      chunkingStrategy: "FIXED_SIZE"  # "FIXED_SIZE", "HIERARCHICAL", "SEMANTIC", "NONE"
      # Configuration based on strategy (see below)
```

### Chunking Strategies

#### Fixed Size Chunking

```yaml
chunkingConfiguration:
  chunkingStrategy: "FIXED_SIZE"
  fixedSizeChunkingConfiguration:
    maxTokens: 512          # 20-8192
    overlapPercentage: 20   # 1-99
```

#### Hierarchical Chunking

```yaml
chunkingConfiguration:
  chunkingStrategy: "HIERARCHICAL"
  hierarchicalChunkingConfiguration:
    levelConfigurations:
      - maxTokens: 1500     # Parent chunks
      - maxTokens: 300      # Child chunks
    overlapTokens: 60       # Overlap between chunks
```

#### Semantic Chunking

```yaml
chunkingConfiguration:
  chunkingStrategy: "SEMANTIC"
  semanticChunkingConfiguration:
    maxTokens: 800                    # Maximum tokens per chunk
    bufferSize: 0                     # Buffer size for semantic boundaries
    breakpointPercentileThreshold: 95 # Percentile threshold for breakpoints
```

#### No Chunking

```yaml
chunkingConfiguration:
  chunkingStrategy: "NONE"
  # Documents are ingested as-is without chunking
```

## Supported File Types

Knowledge bases support various document formats:

| File Type | Extensions | Notes |
|-----------|------------|-------|
| Text | `.txt` | Plain text documents |
| PDF | `.pdf` | Portable Document Format |
| Word | `.doc`, `.docx` | Microsoft Word documents |
| Markdown | `.md`, `.markdown` | Markdown formatted text |
| HTML | `.html`, `.htm` | Web documents |
| CSV | `.csv` | Comma-separated values |
| Excel | `.xls`, `.xlsx` | Microsoft Excel spreadsheets |

## Auto-Generated IAM Permissions

Knowledge bases automatically get IAM roles with these permissions:

### S3 Data Source Access
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

### OpenSearch Serverless Access
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

### Bedrock Model Access
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "bedrock:InvokeModel"
      ],
      "Resource": "arn:aws:bedrock:*::foundation-model/*"
    }
  ]
}
```

## Agent Integration

### Knowledge Base Association

```yaml
# Knowledge base
kind: KnowledgeBase
metadata:
  name: "product-kb"
spec:
  description: "Product knowledge base"
  knowledgeBaseConfiguration:
    type: "VECTOR"
    vectorKnowledgeBaseConfiguration:
      embeddingModelArn: "arn:aws:bedrock:us-east-1::foundation-model/amazon.titan-embed-text-v1"
  storageConfiguration:
    type: "OPENSEARCH_SERVERLESS"
    opensearchServerlessConfiguration:
      collectionArn: "arn:aws:aoss:us-east-1:123456789012:collection/product-kb"
      vectorIndexName: "product-index"

---
# Agent knowledge base association
kind: AgentKnowledgeBaseAssociation
metadata:
  name: "agent-kb-association"
spec:
  agentName: "customer-agent"
  knowledgeBaseName: "product-kb"
  description: "Associate product knowledge base with customer agent"
  knowledgeBaseState: "ENABLED"
```

### Using with Existing Collections

```yaml
kind: KnowledgeBase
metadata:
  name: "existing-collection-kb"
spec:
  description: "Knowledge base using existing OpenSearch collection"
  
  knowledgeBaseConfiguration:
    type: "VECTOR"
    vectorKnowledgeBaseConfiguration:
      embeddingModelArn: "arn:aws:bedrock:us-east-1::foundation-model/amazon.titan-embed-text-v1"
  
  # Reference existing OpenSearch Serverless collection
  storageConfiguration:
    type: "OPENSEARCH_SERVERLESS"
    opensearchServerlessConfiguration:
      collectionArn: "arn:aws:aoss:us-east-1:123456789012:collection/existing-collection"
      vectorIndexName: "existing-index"
      fieldMapping:
        vectorField: "custom-vector-field"
        textField: "custom-text-field"
        metadataField: "custom-metadata-field"
```

## Common Patterns

### Multi-Source Knowledge Base

```yaml
kind: KnowledgeBase
metadata:
  name: "multi-source-kb"
spec:
  description: "Knowledge base with multiple document types"
  
  knowledgeBaseConfiguration:
    type: "VECTOR"
    vectorKnowledgeBaseConfiguration:
      embeddingModelArn: "arn:aws:bedrock:us-east-1::foundation-model/amazon.titan-embed-text-v1"
  
  storageConfiguration:
    type: "OPENSEARCH_SERVERLESS"
    opensearchServerlessConfiguration:
      collectionArn: "arn:aws:aoss:us-east-1:123456789012:collection/multi-source"
      vectorIndexName: "multi-source-index"
  
  dataSources:
    # Technical documentation
    - name: "tech-docs"
      type: "S3"
      s3Configuration:
        bucketArn: "arn:aws:s3:::company-docs"
        inclusionPrefixes: ["technical/"]
      chunkingConfiguration:
        chunkingStrategy: "HIERARCHICAL"
        hierarchicalChunkingConfiguration:
          levelConfigurations:
            - maxTokens: 1500
            - maxTokens: 300
          overlapTokens: 60
    
    # FAQ documents
    - name: "faq"
      type: "S3"
      s3Configuration:
        bucketArn: "arn:aws:s3:::company-docs"
        inclusionPrefixes: ["faq/"]
      chunkingConfiguration:
        chunkingStrategy: "FIXED_SIZE"
        fixedSizeChunkingConfiguration:
          maxTokens: 512
          overlapPercentage: 20
    
    # Policies (no chunking for legal accuracy)
    - name: "policies"
      type: "S3"
      s3Configuration:
        bucketArn: "arn:aws:s3:::company-docs"
        inclusionPrefixes: ["policies/"]
      chunkingConfiguration:
        chunkingStrategy: "NONE"
```

### Multilingual Knowledge Base

```yaml
kind: KnowledgeBase
metadata:
  name: "multilingual-kb"
spec:
  description: "Multilingual knowledge base"
  
  knowledgeBaseConfiguration:
    type: "VECTOR"
    vectorKnowledgeBaseConfiguration:
      # Use multilingual embedding model
      embeddingModelArn: "arn:aws:bedrock:us-east-1::foundation-model/cohere.embed-multilingual-v3"
  
  storageConfiguration:
    type: "OPENSEARCH_SERVERLESS"
    opensearchServerlessConfiguration:
      collectionArn: "arn:aws:aoss:us-east-1:123456789012:collection/multilingual"
      vectorIndexName: "multilingual-index"
  
  dataSources:
    - name: "english-docs"
      type: "S3"
      s3Configuration:
        bucketArn: "arn:aws:s3:::company-docs"
        inclusionPrefixes: ["en/"]
    
    - name: "spanish-docs"
      type: "S3"
      s3Configuration:
        bucketArn: "arn:aws:s3:::company-docs"
        inclusionPrefixes: ["es/"]
    
    - name: "french-docs"
      type: "S3"
      s3Configuration:
        bucketArn: "arn:aws:s3:::company-docs"
        inclusionPrefixes: ["fr/"]
```

## Best Practices

### Data Organization
1. **Structure your S3 buckets** with clear prefixes for different document types
2. **Use consistent naming conventions** for files and folders
3. **Separate documents by language** if using multilingual models
4. **Keep documents up to date** and remove outdated content
5. **Use meaningful file names** that describe content

### Chunking Strategy
1. **Choose appropriate chunk sizes** based on document types
2. **Use hierarchical chunking** for structured documents
3. **Use semantic chunking** for natural language content
4. **Consider overlap** to maintain context across chunks
5. **Test different strategies** to find optimal performance

### Performance Optimization
1. **Choose the right embedding model** for your content type
2. **Optimize chunk sizes** for your specific use case
3. **Monitor query performance** and adjust as needed
4. **Use appropriate field mappings** for OpenSearch
5. **Consider data source organization** for efficient retrieval

### Security
1. **Use appropriate S3 bucket policies** to control access
2. **Encrypt data at rest** in S3 and OpenSearch
3. **Use IAM roles** instead of access keys
4. **Audit knowledge base access** regularly
5. **Follow data classification** guidelines for sensitive content

## Dependencies

- **S3 Buckets**: Referenced S3 buckets and objects must exist
- **OpenSearch Collection**: OpenSearch Serverless collection must exist
- **Embedding Models**: Referenced embedding models must be available in the region
- **IAM Permissions**: Appropriate permissions for S3 and OpenSearch access

## Generated Resources

- AWS Bedrock Knowledge Base
- IAM Role (service role)
- IAM Policy (service policy)
- Knowledge Base Data Sources
- Vector index in OpenSearch (if auto-created)

## Common Issues

### Ingestion Failures
```
Error: Failed to ingest documents from S3
```
**Solution**: Check S3 permissions and verify document formats are supported.

### OpenSearch Connection Issues
```
Error: Cannot connect to OpenSearch collection
```
**Solution**: Verify collection ARN and ensure proper network access policies.

### Embedding Model Access
```
Error: AccessDenied for embedding model
```
**Solution**: Ensure the knowledge base role has permissions to invoke the embedding model.

### Chunking Problems
```
Error: Document too large for chunking
```
**Solution**: Adjust chunk size parameters or consider splitting large documents.

## See Also

- [Agent Resource](agent.md)
- [OpenSearch Serverless Resource](opensearch-serverless.md)
- [IAM Management](../iam-management.md)
- [AWS Bedrock Knowledge Bases Documentation](https://docs.aws.amazon.com/bedrock/latest/userguide/knowledge-base.html)