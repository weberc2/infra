variable "name" {
  type        = string
  description = "(Required) The name to apply to the resources in the module. This should be unique for the AWS account in which the modules are being created."
}

variable "tags" {
  type        = map(string)
  description = "(Optional) The tags to apply to the function and its execution role. Defaults to `{}`."
  default     = {}
}

variable "code" {
  type = object({
    filename         = string
    source_code_hash = string
    runtime          = string
    handler          = string
  })
  description = "(Required) An object containing the path of the zip file containing the lambda source code, the base64-encoded sha256 of the file, the lambda runtime, and the name of the handler within the file (this is specific to the runtime). This is intended to be created with archive_file. See https://registry.terraform.io/providers/hashicorp/archive/latest/docs/data-sources/archive_file."
}

variable "environment" {
  type        = map(string)
  description = "(Optional) The lambda function's execution environment. Defaults to `{}`."
  default     = {}
}

variable "memory_size" {
  type        = number
  description = "(Optional) The amount of memory available to the lambda function. Defaults to 128. See [Limits](https://docs.aws.amazon.com/lambda/latest/dg/gettingstarted-limits.html.)"
  default     = null
}
