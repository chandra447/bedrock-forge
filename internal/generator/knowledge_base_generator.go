package generator

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"

	"bedrock-forge/internal/models"
)

// generateKnowledgeBaseModule creates a module call for a KnowledgeBase resource
func (g *HCLGenerator) generateKnowledgeBaseModule(body *hclwrite.Body, resource models.BaseResource) error {
	knowledgeBase, ok := resource.Spec.(models.KnowledgeBaseSpec)
	if !ok {
		// Try to parse as map and convert to KnowledgeBaseSpec
		specMap, mapOk := resource.Spec.(map[string]interface{})
		if !mapOk {
			return fmt.Errorf("invalid knowledge base spec format")
		}

		// Convert map to KnowledgeBaseSpec
		specJSON, err := json.Marshal(specMap)
		if err != nil {
			return fmt.Errorf("failed to marshal knowledge base spec: %w", err)
		}

		if err := json.Unmarshal(specJSON, &knowledgeBase); err != nil {
			return fmt.Errorf("failed to unmarshal knowledge base spec: %w", err)
		}
	}

	resourceName := g.sanitizeResourceName(resource.Metadata.Name)

	// Create module block
	moduleBlock := body.AppendNewBlock("module", []string{resourceName})
	moduleBody := moduleBlock.Body()

	// Set module source
	moduleSource := fmt.Sprintf("%s//modules/bedrock-knowledge-base", g.config.ModuleRegistry)
	if g.config.ModuleVersion != "" {
		moduleSource += fmt.Sprintf("?ref=%s", g.config.ModuleVersion)
	}
	moduleBody.SetAttributeValue("source", cty.StringVal(moduleSource))

	// Set basic attributes
	moduleBody.SetAttributeValue("knowledge_base_name", cty.StringVal(resource.Metadata.Name))

	// Optional description
	if knowledgeBase.Description != "" {
		moduleBody.SetAttributeValue("description", cty.StringVal(knowledgeBase.Description))
	}

	// Knowledge base configuration
	if knowledgeBase.KnowledgeBaseConfiguration != nil {
		kbConfigValues := make(map[string]cty.Value)
		kbConfigValues["type"] = cty.StringVal(knowledgeBase.KnowledgeBaseConfiguration.Type)

		if knowledgeBase.KnowledgeBaseConfiguration.VectorKnowledgeBaseConfiguration != nil {
			vectorConfig := knowledgeBase.KnowledgeBaseConfiguration.VectorKnowledgeBaseConfiguration
			vectorValues := make(map[string]cty.Value)

			vectorValues["embedding_model_arn"] = cty.StringVal(vectorConfig.EmbeddingModelArn)

			if vectorConfig.EmbeddingModelConfiguration != nil {
				if vectorConfig.EmbeddingModelConfiguration.BedrockEmbeddingModelConfiguration != nil {
					bedrockConfig := vectorConfig.EmbeddingModelConfiguration.BedrockEmbeddingModelConfiguration
					if bedrockConfig.Dimensions > 0 {
						vectorValues["embedding_model_configuration"] = cty.ObjectVal(map[string]cty.Value{
							"bedrock_embedding_model_configuration": cty.ObjectVal(map[string]cty.Value{
								"dimensions": cty.NumberIntVal(int64(bedrockConfig.Dimensions)),
							}),
						})
					}
				}
			}

			kbConfigValues["vector_knowledge_base_configuration"] = cty.ObjectVal(vectorValues)
		}

		moduleBody.SetAttributeValue("knowledge_base_configuration", cty.ObjectVal(kbConfigValues))
	}

	// Storage configuration
	if knowledgeBase.StorageConfiguration != nil {
		storageValues := make(map[string]cty.Value)
		storageValues["type"] = cty.StringVal(knowledgeBase.StorageConfiguration.Type)

		// Enhanced OpenSearch Serverless configuration (new approach)
		if knowledgeBase.StorageConfiguration.OpenSearchServerless != nil {
			osConfig := knowledgeBase.StorageConfiguration.OpenSearchServerless
			osValues := make(map[string]cty.Value)

			// Determine collection ARN based on configuration
			if osConfig.CollectionArn != nil {
				// Use existing collection ARN
				osValues["collection_arn"] = cty.StringVal(*osConfig.CollectionArn)
			} else if osConfig.CollectionName != nil && !osConfig.CollectionName.IsEmpty() {
				// Reference auto-created collection by name
				if collectionArn, err := g.resolveReferenceToOutput(*osConfig.CollectionName, models.OpenSearchServerlessKind, "collection_arn"); err == nil {
					osValues["collection_arn"] = cty.StringVal(collectionArn)
				} else {
					// Fallback to direct reference for backward compatibility
					collectionResourceName := g.sanitizeResourceName(osConfig.CollectionName.String())
					osValues["collection_arn"] = cty.StringVal(fmt.Sprintf("${aws_opensearchserverless_collection.%s.arn}", collectionResourceName))
					g.logger.WithError(err).WithField("collection", osConfig.CollectionName.String()).Warn("Failed to resolve OpenSearch Serverless collection reference")
				}
			}

			osValues["vector_index_name"] = cty.StringVal(osConfig.VectorIndexName)

			// Field mapping
			fieldMappingValues := make(map[string]cty.Value)
			fieldMappingValues["vector_field"] = cty.StringVal(osConfig.FieldMapping.VectorField)
			fieldMappingValues["text_field"] = cty.StringVal(osConfig.FieldMapping.TextField)
			fieldMappingValues["metadata_field"] = cty.StringVal(osConfig.FieldMapping.MetadataField)

			osValues["field_mapping"] = cty.ObjectVal(fieldMappingValues)

			storageValues["opensearch_serverless_configuration"] = cty.ObjectVal(osValues)
		} else if knowledgeBase.StorageConfiguration.OpensearchServerlessConfiguration != nil {
			// Legacy OpenSearch Serverless configuration (backward compatibility)
			osConfig := knowledgeBase.StorageConfiguration.OpensearchServerlessConfiguration
			osValues := make(map[string]cty.Value)

			osValues["collection_arn"] = cty.StringVal(osConfig.CollectionArn)
			osValues["vector_index_name"] = cty.StringVal(osConfig.VectorIndexName)

			// Field mapping
			fieldMappingValues := make(map[string]cty.Value)
			fieldMappingValues["vector_field"] = cty.StringVal(osConfig.FieldMapping.VectorField)
			fieldMappingValues["text_field"] = cty.StringVal(osConfig.FieldMapping.TextField)
			fieldMappingValues["metadata_field"] = cty.StringVal(osConfig.FieldMapping.MetadataField)

			osValues["field_mapping"] = cty.ObjectVal(fieldMappingValues)

			storageValues["opensearch_serverless_configuration"] = cty.ObjectVal(osValues)
		}

		moduleBody.SetAttributeValue("storage_configuration", cty.ObjectVal(storageValues))
	}

	// Data sources configuration
	if len(knowledgeBase.DataSources) > 0 {
		dataSourceList := make([]cty.Value, 0, len(knowledgeBase.DataSources))

		for _, dataSource := range knowledgeBase.DataSources {
			dsValues := make(map[string]cty.Value)
			dsValues["name"] = cty.StringVal(dataSource.Name)
			dsValues["type"] = cty.StringVal(dataSource.Type)

			// S3 configuration
			if dataSource.S3Configuration != nil {
				s3Values := make(map[string]cty.Value)
				s3Values["bucket_arn"] = cty.StringVal(dataSource.S3Configuration.BucketArn)

				// Always include both prefix types for consistency
				if len(dataSource.S3Configuration.InclusionPrefixes) > 0 {
					prefixes := make([]cty.Value, 0, len(dataSource.S3Configuration.InclusionPrefixes))
					for _, prefix := range dataSource.S3Configuration.InclusionPrefixes {
						prefixes = append(prefixes, cty.StringVal(prefix))
					}
					s3Values["inclusion_prefixes"] = cty.ListVal(prefixes)
				} else {
					s3Values["inclusion_prefixes"] = cty.NullVal(cty.List(cty.String))
				}

				if len(dataSource.S3Configuration.ExclusionPrefixes) > 0 {
					prefixes := make([]cty.Value, 0, len(dataSource.S3Configuration.ExclusionPrefixes))
					for _, prefix := range dataSource.S3Configuration.ExclusionPrefixes {
						prefixes = append(prefixes, cty.StringVal(prefix))
					}
					s3Values["exclusion_prefixes"] = cty.ListVal(prefixes)
				} else {
					s3Values["exclusion_prefixes"] = cty.NullVal(cty.List(cty.String))
				}

				dsValues["s3_configuration"] = cty.ObjectVal(s3Values)
			} else {
				// Ensure s3_configuration is always present for consistency
				dsValues["s3_configuration"] = cty.NullVal(cty.Object(map[string]cty.Type{
					"bucket_arn":         cty.String,
					"inclusion_prefixes": cty.List(cty.String),
					"exclusion_prefixes": cty.List(cty.String),
				}))
			}

			// Chunking configuration
			if dataSource.ChunkingConfiguration != nil {
				chunkingValues := make(map[string]cty.Value)
				chunkingValues["chunking_strategy"] = cty.StringVal(dataSource.ChunkingConfiguration.ChunkingStrategy)

				// Always include both chunking configuration types for consistency
				if dataSource.ChunkingConfiguration.FixedSizeChunkingConfiguration != nil {
					fixedSizeConfig := dataSource.ChunkingConfiguration.FixedSizeChunkingConfiguration
					chunkingValues["fixed_size_chunking_configuration"] = cty.ObjectVal(map[string]cty.Value{
						"max_tokens":         cty.NumberIntVal(int64(fixedSizeConfig.MaxTokens)),
						"overlap_percentage": cty.NumberIntVal(int64(fixedSizeConfig.OverlapPercentage)),
					})
				} else {
					chunkingValues["fixed_size_chunking_configuration"] = cty.NullVal(cty.Object(map[string]cty.Type{
						"max_tokens":         cty.Number,
						"overlap_percentage": cty.Number,
					}))
				}

				if dataSource.ChunkingConfiguration.SemanticChunkingConfiguration != nil {
					semanticConfig := dataSource.ChunkingConfiguration.SemanticChunkingConfiguration
					chunkingValues["semantic_chunking_configuration"] = cty.ObjectVal(map[string]cty.Value{
						"max_tokens":                      cty.NumberIntVal(int64(semanticConfig.MaxTokens)),
						"buffer_size":                     cty.NumberIntVal(int64(semanticConfig.BufferSize)),
						"breakpoint_percentile_threshold": cty.NumberIntVal(int64(semanticConfig.BreakpointPercentileThreshold)),
					})
				} else {
					chunkingValues["semantic_chunking_configuration"] = cty.NullVal(cty.Object(map[string]cty.Type{
						"max_tokens":                      cty.Number,
						"buffer_size":                     cty.Number,
						"breakpoint_percentile_threshold": cty.Number,
					}))
				}

				dsValues["chunking_configuration"] = cty.ObjectVal(chunkingValues)
			} else {
				// Ensure chunking_configuration is always present for consistency
				dsValues["chunking_configuration"] = cty.NullVal(cty.Object(map[string]cty.Type{
					"chunking_strategy": cty.String,
					"fixed_size_chunking_configuration": cty.Object(map[string]cty.Type{
						"max_tokens":         cty.Number,
						"overlap_percentage": cty.Number,
					}),
					"semantic_chunking_configuration": cty.Object(map[string]cty.Type{
						"max_tokens":                      cty.Number,
						"buffer_size":                     cty.Number,
						"breakpoint_percentile_threshold": cty.Number,
					}),
				}))
			}

			// Vector ingestion configuration
			if dataSource.VectorIngestionConfiguration != nil && dataSource.VectorIngestionConfiguration.ChunkingConfiguration != nil {
				vectorIngestionValues := make(map[string]cty.Value)
				chunkingConfig := dataSource.VectorIngestionConfiguration.ChunkingConfiguration

				chunkingValues := make(map[string]cty.Value)
				chunkingValues["chunking_strategy"] = cty.StringVal(chunkingConfig.ChunkingStrategy)

				if chunkingConfig.SemanticChunkingConfiguration != nil {
					semanticConfig := chunkingConfig.SemanticChunkingConfiguration
					chunkingValues["semantic_chunking_configuration"] = cty.ObjectVal(map[string]cty.Value{
						"max_tokens":                      cty.NumberIntVal(int64(semanticConfig.MaxTokens)),
						"buffer_size":                     cty.NumberIntVal(int64(semanticConfig.BufferSize)),
						"breakpoint_percentile_threshold": cty.NumberIntVal(int64(semanticConfig.BreakpointPercentileThreshold)),
					})
				} else {
					chunkingValues["semantic_chunking_configuration"] = cty.NullVal(cty.Object(map[string]cty.Type{
						"max_tokens":                      cty.Number,
						"buffer_size":                     cty.Number,
						"breakpoint_percentile_threshold": cty.Number,
					}))
				}

				vectorIngestionValues["chunking_configuration"] = cty.ObjectVal(chunkingValues)
				dsValues["vector_ingestion_configuration"] = cty.ObjectVal(vectorIngestionValues)
			} else {
				// Ensure vector_ingestion_configuration is always present for consistency
				dsValues["vector_ingestion_configuration"] = cty.NullVal(cty.Object(map[string]cty.Type{
					"chunking_configuration": cty.Object(map[string]cty.Type{
						"chunking_strategy": cty.String,
						"semantic_chunking_configuration": cty.Object(map[string]cty.Type{
							"max_tokens":                      cty.Number,
							"buffer_size":                     cty.Number,
							"breakpoint_percentile_threshold": cty.Number,
						}),
					}),
				}))
			}

			// Custom transformation
			if dataSource.CustomTransformation != nil {
				customTransValues := make(map[string]cty.Value)

				if dataSource.CustomTransformation.TransformationLambda != nil {
					lambdaValues := make(map[string]cty.Value)
					lambdaValues["lambda_arn"] = cty.StringVal(dataSource.CustomTransformation.TransformationLambda.LambdaArn)
					customTransValues["transformation_lambda"] = cty.ObjectVal(lambdaValues)
				} else {
					customTransValues["transformation_lambda"] = cty.NullVal(cty.Object(map[string]cty.Type{
						"lambda_arn": cty.String,
					}))
				}

				if dataSource.CustomTransformation.IntermediateStorage != nil {
					storageValues := make(map[string]cty.Value)
					if dataSource.CustomTransformation.IntermediateStorage.S3Location != nil {
						s3Values := make(map[string]cty.Value)
						s3Values["uri"] = cty.StringVal(dataSource.CustomTransformation.IntermediateStorage.S3Location.URI)
						storageValues["s3_location"] = cty.ObjectVal(s3Values)
					} else {
						storageValues["s3_location"] = cty.NullVal(cty.Object(map[string]cty.Type{
							"uri": cty.String,
						}))
					}
					customTransValues["intermediate_storage"] = cty.ObjectVal(storageValues)
				} else {
					customTransValues["intermediate_storage"] = cty.NullVal(cty.Object(map[string]cty.Type{
						"s3_location": cty.Object(map[string]cty.Type{
							"uri": cty.String,
						}),
					}))
				}

				dsValues["custom_transformation"] = cty.ObjectVal(customTransValues)
			} else {
				// Ensure custom_transformation is always present for consistency
				dsValues["custom_transformation"] = cty.NullVal(cty.Object(map[string]cty.Type{
					"transformation_lambda": cty.Object(map[string]cty.Type{
						"lambda_arn": cty.String,
					}),
					"intermediate_storage": cty.Object(map[string]cty.Type{
						"s3_location": cty.Object(map[string]cty.Type{
							"uri": cty.String,
						}),
					}),
				}))
			}

			dataSourceList = append(dataSourceList, cty.ObjectVal(dsValues))
		}

		moduleBody.SetAttributeValue("data_sources", cty.ListVal(dataSourceList))
	}

	// Tags
	if len(knowledgeBase.Tags) > 0 {
		tagValues := make(map[string]cty.Value)
		for key, value := range knowledgeBase.Tags {
			tagValues[key] = cty.StringVal(value)
		}
		moduleBody.SetAttributeValue("tags", cty.ObjectVal(tagValues))
	}

	body.AppendNewline()

	g.logger.WithField("knowledge_base", resource.Metadata.Name).Info("Generated knowledge base module")
	return nil
}
