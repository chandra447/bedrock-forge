# EventBridge rule for capturing Bedrock agent events
resource "aws_cloudwatch_event_rule" "agent_events" {
  name        = var.eventbridge_rule_name
  description = "Capture Bedrock agent execution events"

  event_pattern = jsonencode({
    source      = ["aws.bedrock-agent"]
    detail-type = ["Bedrock Agent Execution"]
    detail = {
      state = ["COMPLETED", "FAILED"]
    }
  })

  tags = {
    Environment = var.environment
    Project     = var.project_name
    Purpose     = "BedrockAgentEvents"
    ManagedBy   = "bedrock-forge"
  }
}

# EventBridge target to send events to SNS
resource "aws_cloudwatch_event_target" "sns_target" {
  rule      = aws_cloudwatch_event_rule.agent_events.name
  target_id = "SendToSNS"
  arn       = aws_sns_topic.agent_notifications.arn

  input_transformer {
    input_paths = {
      agent_id   = "$.detail.agentId"
      state      = "$.detail.state"
      timestamp  = "$.detail.timestamp"
    }
    input_template = <<EOF
{
  "message": "Bedrock Agent <agent_id> execution <state> at <timestamp>",
  "agentId": "<agent_id>",
  "state": "<state>",
  "timestamp": "<timestamp>",
  "environment": "${var.environment}"
}
EOF
  }
}

# Output EventBridge rule ARN
output "eventbridge_rule_arn" {
  description = "ARN of the agent events EventBridge rule"
  value       = aws_cloudwatch_event_rule.agent_events.arn
}

output "eventbridge_rule_name" {
  description = "Name of the agent events EventBridge rule"
  value       = aws_cloudwatch_event_rule.agent_events.name
}