# SNS Topic for agent notifications
resource "aws_sns_topic" "agent_notifications" {
  name = var.sns_topic_name

  tags = {
    Environment = var.environment
    Project     = var.project_name
    Purpose     = "BedrockAgentNotifications"
    ManagedBy   = "bedrock-forge"
  }
}

# SNS Topic Policy
resource "aws_sns_topic_policy" "agent_notifications_policy" {
  arn = aws_sns_topic.agent_notifications.arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "events.amazonaws.com"
        }
        Action   = "sns:Publish"
        Resource = aws_sns_topic.agent_notifications.arn
        Condition = {
          StringEquals = {
            "aws:SourceAccount" = data.aws_caller_identity.current.account_id
          }
        }
      }
    ]
  })
}

# Data source for current AWS account
data "aws_caller_identity" "current" {}

# Output SNS topic ARN for use in other resources
output "sns_topic_arn" {
  description = "ARN of the agent notifications SNS topic"
  value       = aws_sns_topic.agent_notifications.arn
}

output "sns_topic_name" {
  description = "Name of the agent notifications SNS topic"
  value       = aws_sns_topic.agent_notifications.name
}