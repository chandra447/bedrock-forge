# SNS Topic for agent notifications
resource "aws_sns_topic" "agent_notifications" {
  name = var.sns_topic_name
  
  tags = {
    Environment = var.environment
    Purpose     = "agent-notifications"
    ManagedBy   = "bedrock-forge"
    Agent       = var.agent_name_prefix
  }
}

# SNS Topic Policy to allow EventBridge to publish
resource "aws_sns_topic_policy" "agent_notifications_policy" {
  arn = aws_sns_topic.agent_notifications.arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "AllowEventBridgePublish"
        Effect = "Allow"
        Principal = {
          Service = "events.amazonaws.com"
        }
        Action   = "sns:Publish"
        Resource = aws_sns_topic.agent_notifications.arn
      }
    ]
  })
}

# Output the topic ARN for use in other resources
output "sns_topic_arn" {
  description = "ARN of the agent notifications SNS topic"
  value       = aws_sns_topic.agent_notifications.arn
}

output "sns_topic_name" {
  description = "Name of the agent notifications SNS topic"
  value       = aws_sns_topic.agent_notifications.name
}