terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
  required_version = ">= 1.0"
}

provider "aws" {
  default_tags {
    tags = {
      Environment = "dev"
      ManagedBy   = "bedrock-forge"
      Project     = "bedrock-project"
    }
  }
}

variable "project_name" {
  description = "Name of the project"
  type        = string
  default     = "bedrock-project"
}
variable "environment" {
  description = "Environment name"
  type        = string
  default     = "dev"
}

module "content_safety_guardrail" {
  source         = "git::https://github.com/company/bedrock-terraform-modules//modules/bedrock-guardrail?ref=v1.0.0"
  guardrail_name = "content-safety-guardrail"
  description    = "Enterprise content safety guardrail"
  content_policy_config = {
    filters_config = [{
      input_strength  = "HIGH"
      output_strength = "HIGH"
      type            = "SEXUAL"
      }, {
      input_strength  = "MEDIUM"
      output_strength = "HIGH"
      type            = "VIOLENCE"
      }, {
      input_strength  = "HIGH"
      output_strength = "HIGH"
      type            = "HATE"
      }, {
      input_strength  = "MEDIUM"
      output_strength = "HIGH"
      type            = "INSULTS"
      }, {
      input_strength  = "MEDIUM"
      output_strength = "MEDIUM"
      type            = "MISCONDUCT"
      }, {
      input_strength  = "HIGH"
      output_strength = "NONE"
      type            = "PROMPT_ATTACK"
    }]
  }
  sensitive_information_policy_config = {
    pii_entities_config = [{
      action = "BLOCK"
      type   = "EMAIL"
      }, {
      action = "ANONYMIZE"
      type   = "PHONE"
      }, {
      action = "BLOCK"
      type   = "SSN"
    }]
  }
  contextual_grounding_policy_config = {
    filters_config = [{
      threshold = 0.8
      type      = "GROUNDING"
      }, {
      threshold = 0.7
      type      = "RELEVANCE"
    }]
  }
  topic_policy_config = {
    topics_config = [{
      definition = "Discussions about financial investments or trading advice"
      examples   = ["Should I buy this stock?", "What's the best crypto to invest in?"]
      name       = "Investment Advice"
      type       = "DENY"
    }]
  }
  word_policy_config = {
    managed_word_lists_config = [{
      type = "PROFANITY"
    }]
    words_config = [{
      text = "competitor_name"
    }]
  }
}

module "custom_orchestration_prompt" {
  source          = "git::https://github.com/company/bedrock-terraform-modules//modules/bedrock-prompt?ref=v1.0.0"
  prompt_name     = "custom-orchestration-prompt"
  description     = "Custom orchestration prompt for customer support"
  default_variant = "production"
  variants = [{
    inference_configuration = {
      text = {
        max_tokens     = 2048
        stop_sequences = ["Human:", "Assistant:"]
        temperature    = 0.1
        top_p          = 0.9
      }
    }
    model_id = "anthropic.claude-3-sonnet-20240229-v1:0"
    name     = "production"
    template_configuration = {
      text = "You are a customer support agent. Always be helpful and professional.\n\nInstructions:\n1. Greet the customer warmly\n2. Listen to their concern\n3. Provide accurate information\n4. Offer additional assistance\n\nCustomer Query: {{query}}\n\nContext: {{context}}\n"
    }
    template_type = "TEXT"
    }, {
    inference_configuration = {
      text = {
        max_tokens     = 1024
        stop_sequences = null
        temperature    = 0.2
        top_p          = 0.95
      }
    }
    model_id = "anthropic.claude-3-sonnet-20240229-v1:0"
    name     = "development"
    template_configuration = {
      text = "[DEBUG MODE] Customer Support Agent\n\nQuery: {{query}}\nContext: {{context}}\n\nDebug info will be included in responses.\n"
    }
    template_type = "TEXT"
  }]
}

module "order_lookup" {
  source        = "git::https://github.com/company/bedrock-terraform-modules//modules/lambda-function?ref=v1.0.0"
  function_name = "order-lookup"
  runtime       = "python3.9"
  handler       = "app.handler"
  description   = "Lambda function to look up customer orders"
  code = {
    source = "directory"
  }
  environment_variables = {
    LOG_LEVEL     = "INFO"
    ORDER_API_URL = "https://api.company.com/orders"
  }
  timeout     = 30
  memory_size = 256
  tags = {
    Function = "OrderLookup"
    Team     = "CustomerSupport"
  }
  create_role = true
  lambda_resource_policy_statements = [{
    actions = ["lambda:InvokeFunction"]
    effect  = "Allow"
    principals = [{
      identifiers = ["bedrock.amazonaws.com"]
      type        = "Service"
    }]
    sid = "AllowBedrockAgentInvoke"
  }]
}

module "product_search_api" {
  source        = "git::https://github.com/company/bedrock-terraform-modules//modules/lambda-function?ref=v1.0.0"
  function_name = "product-search-api"
  runtime       = "python3.9"
  handler       = "app.handler"
  description   = "FastAPI-based product search service"
  code = {
    source = "directory"
  }
  environment_variables = {
    LOG_LEVEL      = "INFO"
    PRODUCT_DB_URL = "https://api.company.com/products"
  }
  timeout     = 30
  memory_size = 512
  tags = {
    Framework = "FastAPI"
    Function  = "ProductSearch"
  }
  create_role = true
  lambda_resource_policy_statements = [{
    actions = ["lambda:InvokeFunction"]
    effect  = "Allow"
    principals = [{
      identifiers = ["bedrock.amazonaws.com"]
      type        = "Service"
    }]
    sid = "AllowBedrockAgentInvoke"
  }]
}

module "faq_kb" {
  source              = "git::https://github.com/company/bedrock-terraform-modules//modules/bedrock-knowledge-base?ref=v1.0.0"
  knowledge_base_name = "faq-kb"
  description         = "Customer FAQ knowledge base"
  knowledge_base_configuration = {
    type = "VECTOR"
    vector_knowledge_base_configuration = {
      embedding_model_arn = "arn:aws:bedrock:us-east-1::foundation-model/amazon.titan-embed-text-v1"
      embedding_model_configuration = {
        bedrock_embedding_model_configuration = {
          dimensions = 1536
        }
      }
    }
  }
  storage_configuration = {
    opensearch_serverless_configuration = {
      collection_arn = "arn:aws:aoss:us-east-1:123456789012:collection/bedrock-kb"
      field_mapping = {
        metadata_field = "metadata"
        text_field     = "text"
        vector_field   = "vector"
      }
      vector_index_name = "bedrock-knowledge-base-index"
    }
    type = "OPENSEARCH_SERVERLESS"
  }
  data_sources = [{
    chunking_configuration = {
      chunking_strategy = "FIXED_SIZE"
      fixed_size_chunking_configuration = {
        max_tokens         = 512
        overlap_percentage = 20
      }
    }
    custom_transformation = {
      intermediate_storage = {
        s3_location = {
          uri = "s3://company-kb-temp/transformations/"
        }
      }
      transformation_lambda = {
        lambda_arn = "arn:aws:lambda:us-east-1:123456789012:function:kb-preprocessor"
      }
    }
    name = "faq-documents"
    s3_configuration = {
      bucket_arn         = "arn:aws:s3:::company-kb-documents"
      inclusion_prefixes = ["faq/"]
    }
    type = "S3"
    vector_ingestion_configuration = {
      chunking_configuration = {
        chunking_strategy = "SEMANTIC"
        semantic_chunking_configuration = {
          breakpoint_percentile_threshold = 95
          buffer_size                     = 1
          max_tokens                      = 300
        }
      }
    }
  }]
}

module "order_management" {
  source            = "git::https://github.com/company/bedrock-terraform-modules//modules/bedrock-action-group?ref=v1.0.0"
  action_group_name = "order-management"
  description       = "Provides order lookup and management capabilities"
  action_group_executor = {
    lambda = "$${module.order_lookup.lambda_function_arn}"
  }
  api_schema = {
    s3 = {
      s3_bucket_name = "bedrock-artifacts"
      s3_object_key  = "bedrock-forge/schemas/order-management/openapi.json"
    }
  }
  function_schema = {
    functions = [{
      description = "Look up order details by order ID"
      name        = "lookup_order"
      parameters = {
        order_id = {
          description = "The unique order identifier"
          required    = true
          type        = "string"
        }
      }
    }]
  }
}

module "product_search" {
  source            = "git::https://github.com/company/bedrock-terraform-modules//modules/bedrock-action-group?ref=v1.0.0"
  action_group_name = "product-search"
  description       = "Provides product search and filtering capabilities"
  action_group_executor = {
    lambda = "$${module.product_search_api.lambda_function_arn}"
  }
  api_schema = {
    s3 = {
      s3_bucket_name = "bedrock-schemas"
      s3_object_key  = "action-groups/product-search/openapi.json"
    }
  }
}

module "customer_support" {
  source                      = "git::https://github.com/company/bedrock-terraform-modules//modules/bedrock-agent?ref=v1.0.0"
  name                        = "customer-support"
  foundation_model            = "anthropic.claude-3-sonnet-20240229-v1:0"
  instruction                 = "You are a helpful customer support agent..."
  description                 = "Customer support agent for order management"
  idle_session_ttl_in_seconds = 3600
  guardrail = {
    guardrail_id      = "$${module.content_safety_guardrail.guardrail_id}"
    guardrail_version = "$${module.content_safety_guardrail.guardrail_version}"
    mode              = "pre"
    name              = "content-safety-guardrail"
    version           = "1"
  }
  knowledge_bases = [{
    description       = "Customer FAQ knowledge base"
    knowledge_base_id = "$${module.faq_kb.knowledge_base_id}"
    name              = "faq-kb"
  }]
  prompt_overrides = [{
    prompt_arn  = "arn:aws:bedrock:us-east-1:123456789012:prompt/custom-preprocessing"
    prompt_type = "PRE_PROCESSING"
    variant     = "v1"
    }, {
    prompt_arn  = "$${module.custom_orchestration_prompt.prompt_arn}"
    prompt_type = "ORCHESTRATION"
    variant     = "production"
  }]
  memory_configuration = {
    enabled_memory_types = ["SESSION_SUMMARY"]
    storage_days         = 30
  }
}

output "customer_support_agent_id" {
  description = "ID of the customer-support agent"
  value       = module.customer_support.agent_id
}
output "customer_support_agent_arn" {
  description = "ARN of the customer-support agent"
  value       = module.customer_support.agent_arn
}

