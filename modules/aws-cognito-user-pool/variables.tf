variable "name" {
  type        = string
  description = "(Required) The name of the AWS Cognito User Pool resource. This should be unique for the AWS account."
}

variable "domain_name" {
  type        = string
  description = "(Required) The domain name to use for the signin and signup pages. Example: my-domain.example.com. To use the default Cognito domain, omit the `example.com` portion. If a custom domain name is provided, a corresponding key must be provided via `certificate_arn`."
}

variable "certificate_arn" {
  type        = string
  description = "(Optional) The certificate corresponding to custom domain names. Only required if a custom domain name is provided via `domain_name`. See [here](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/cognito_user_pool_domain#certificate_arn) for additional details."
  default     = null
}

variable "callback_urls" {
  type        = list(string)
  description = "(Optional) A list of URLs to which the user will be redirected after successfully authenticating"
  default     = []
}

variable "tags" {
  type        = map(string)
  description = "(Optional) The tags to associate with the AWS Cognito User Pool resource."
  default     = {}
}
