variable "environment" {
  type        = string
  description = "The environment with which the tagged resources are associated."
}

variable "system" {
  type        = string
  description = "The system name."
}

variable "additional_tags" {
  type        = map(string)
  description = "(Optional) Additional tags. Defaults to `{}`"
  default     = {}
}
