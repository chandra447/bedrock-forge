# EventBridge rule for agent session events
resource "aws_cloudwatch_event_rule" "agent_session_events" {
  name        = "${var.agent_name_prefix}-session-events"
  description = "Capture agent session start/end events"

  event_pattern = jsonencode({
    source      = ["aws.bedrock-agent"]
    detail-type = ["Bedrock Agent Session State Change"]
    detail = {
      state = ["STARTED", "ENDED", "FAILED"]
    }
  })

  tags = {
    Environment = var.environment
    Purpose     = "agent-monitoring"
    ManagedBy   = "bedrock-forge"
    Agent       = var.agent_name_prefix
  }
}

# EventBridge target to send events to SNS
resource "aws_cloudwatch_event_target" "agent_session_sns" {
  rule      = aws_cloudwatch_event_rule.agent_session_events.name
  target_id = "AgentSessionSNSTarget"
  arn       = aws_sns_topic.agent_notifications.arn

  input_transformer {
    input_paths = {
      timestamp = "$.detail.timestamp"
      state     = "$.detail.state"
      agent     = "$.detail.agent-id"
      session   = "$.detail.session-id"
    }
    
    input_template = jsonencode({
      notification_type = "agent_session_event"
      timestamp        = "<timestamp>"
      agent_state      = "<state>"
      agent_id         = "<agent>"
      session_id       = "<session>"
      environment      = var.environment
      message          = "Agent session <state> for agent <agent> in session <session>"
    })
  }
}

# EventBridge rule for agent invocation events
resource "aws_cloudwatch_event_rule" "agent_invocation_events" {
  name        = "${var.agent_name_prefix}-invocation-events"
  description = "Capture agent function invocation events"

  event_pattern = jsonencode({
    source      = ["aws.bedrock-agent"]
    detail-type = ["Bedrock Agent Action Group Invocation"]
    detail = {
      status = ["SUCCESS", "FAILURE"]
    }
  })

  tags = {
    Environment = var.environment
    Purpose     = "agent-monitoring"
    ManagedBy   = "bedrock-forge"
    Agent       = var.agent_name_prefix
  }
}

# EventBridge target for invocation events
resource "aws_cloudwatch_event_target" "agent_invocation_sns" {
  rule      = aws_cloudwatch_event_rule.agent_invocation_events.name
  target_id = "AgentInvocationSNSTarget"
  arn       = aws_sns_topic.agent_notifications.arn

  input_transformer {
    input_paths = {
      timestamp    = "$.detail.timestamp"
      status       = "$.detail.status"
      action_group = "$.detail.action-group"
      function     = "$.detail.function"
    }
    
    input_template = jsonencode({
      notification_type = "agent_invocation_event"
      timestamp        = "<timestamp>"
      invocation_status = "<status>"
      action_group     = "<action_group>"
      function_name    = "<function>"
      environment      = var.environment
      message          = "Agent action group <action_group> function <function> completed with status <status>"
    })
  }
}

# Output EventBridge rule ARNs
output "agent_session_rule_arn" {
  description = "ARN of the agent session events rule"
  value       = aws_cloudwatch_event_rule.agent_session_events.arn
}

output "agent_invocation_rule_arn" {
  description = "ARN of the agent invocation events rule" 
  value       = aws_cloudwatch_event_rule.agent_invocation_events.arn
}