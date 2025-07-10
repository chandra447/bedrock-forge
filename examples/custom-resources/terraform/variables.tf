# Variables for custom infrastructure resources

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
  default     = "dev"
}

variable "project_name" {
  description = "Project name for tagging"
  type        = string
  default     = "bedrock-agent"
}

variable "sns_topic_name" {
  description = "Name of the SNS topic for agent notifications"
  type        = string
  default     = "agent-notifications"
}

variable "eventbridge_rule_name" {
  description = "Name of the EventBridge rule for agent events"
  type        = string
  default     = "agent-events"
}