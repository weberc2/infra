variable "name" {
  type        = string
  description = "(Required) The name for the API gateway and IAM role resources. This should be unique for the AWS account."
}

variable "lambda" {
  description = "(Required) The arguments to the lambda function. See ../aws-lambda-function for more information."
}

variable "tags" {
  type        = map(string)
  description = "(Optional) The tags to apply to the resources. Defaults to `{}`."
  default     = {}
}
