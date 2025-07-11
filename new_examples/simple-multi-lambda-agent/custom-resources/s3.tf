# S3 bucket for document storage
resource "aws_s3_bucket" "document_storage_bucket" {
  bucket = "document-storage-bucket-${random_id.bucket_suffix.hex}"
  
  tags = merge({
    Environment = var.environment
    Project     = var.project_name
    ManagedBy   = "bedrock-forge"
  }, var.additional_tags)
}

# Random ID for unique bucket naming
resource "random_id" "bucket_suffix" {
  byte_length = 4
}

# S3 bucket versioning
resource "aws_s3_bucket_versioning" "document_storage_versioning" {
  bucket = aws_s3_bucket.document_storage_bucket.id
  versioning_configuration {
    status = "Enabled"
  }
}

# S3 bucket public access block
resource "aws_s3_bucket_public_access_block" "document_storage_public_access_block" {
  bucket = aws_s3_bucket.document_storage_bucket.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# S3 bucket server-side encryption
resource "aws_s3_bucket_server_side_encryption_configuration" "document_storage_encryption" {
  bucket = aws_s3_bucket.document_storage_bucket.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

# Output S3 bucket information
output "bucket_name" {
  value       = aws_s3_bucket.document_storage_bucket.bucket
  description = "Name of the S3 bucket for document storage"
}

output "bucket_arn" {
  value       = aws_s3_bucket.document_storage_bucket.arn
  description = "ARN of the S3 bucket for document storage"
}

output "bucket_id" {
  value       = aws_s3_bucket.document_storage_bucket.id
  description = "ID of the S3 bucket for document storage"
}

output "bucket_domain_name" {
  value       = aws_s3_bucket.document_storage_bucket.bucket_domain_name
  description = "Domain name of the S3 bucket"
}