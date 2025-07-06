package generator

import "bedrock-forge/internal/packager"

// GenerationContext holds shared data for the generation process
type GenerationContext struct {
	LambdaPackages map[string]*packager.LambdaPackage
	SchemaPackages map[string]*packager.SchemaPackage
}

// NewGenerationContext creates a new generation context
func NewGenerationContext() *GenerationContext {
	return &GenerationContext{
		LambdaPackages: make(map[string]*packager.LambdaPackage),
		SchemaPackages: make(map[string]*packager.SchemaPackage),
	}
}

// GetLambdaS3URI returns the S3 URI for a Lambda package
func (ctx *GenerationContext) GetLambdaS3URI(lambdaName string) string {
	if pkg, exists := ctx.LambdaPackages[lambdaName]; exists {
		return pkg.S3URI
	}
	return ""
}

// GetSchemaS3URI returns the S3 URI for a schema package
func (ctx *GenerationContext) GetSchemaS3URI(actionGroupName string) string {
	if pkg, exists := ctx.SchemaPackages[actionGroupName]; exists {
		return pkg.S3URI
	}
	return ""
}

// GetSchemaS3Location returns the S3 bucket and key for a schema
func (ctx *GenerationContext) GetSchemaS3Location(actionGroupName string) (bucket, key string) {
	if pkg, exists := ctx.SchemaPackages[actionGroupName]; exists {
		return pkg.S3Bucket, pkg.S3Key
	}
	return "", ""
}