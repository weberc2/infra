variable "name" {
  type        = string
  description = "(Required) The name of the S3 bucket."
}

variable "tags" {
  type        = map(string)
  description = "(Required) The tags to associate with the S3 bucket."
}
