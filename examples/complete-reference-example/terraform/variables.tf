# Variables for the custom infrastructure
variable "sns_topic_name" {
  description = "Name of the SNS topic for agent notifications"
  type        = string
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
}

variable "agent_name_prefix" {
  description = "Prefix for agent-related resource names"
  type        = string
}