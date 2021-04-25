module "tags" {
  source      = "../../modules/default-tags"
  environment = "prd"
}

module "user_pool" {
  source = "../../modules/aws-cognito-user-pool"

  name          = "auth-service"
  domain_name   = "weberc2"
  callback_urls = ["https://weberc2.github.io"]
  tags          = module.tags.tags
}
