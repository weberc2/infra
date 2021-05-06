resource "aws_cognito_user_pool" "self" {
  name              = var.name
  mfa_configuration = "OPTIONAL"

  account_recovery_setting {
    recovery_mechanism {
      name     = "verified_email"
      priority = 1
    }
  }

  admin_create_user_config {
    allow_admin_create_user_only = false
  }

  password_policy {
    minimum_length                   = 8
    require_lowercase                = true
    require_uppercase                = true
    require_numbers                  = true
    require_symbols                  = true
    temporary_password_validity_days = 2
  }

  software_token_mfa_configuration {
    enabled = true
  }

  schema {
    name                = "email"
    attribute_data_type = "String"
    mutable             = false
    required            = true

    string_attribute_constraints {
      max_length = "2048"
      min_length = "0"
    }
  }
  tags = var.tags
}

# https://docs.aws.amazon.com/cognito/latest/developerguide/cognito-user-pools-app-idp-settings.html
resource "aws_cognito_user_pool_client" "self" {
  name                                 = var.name
  user_pool_id                         = aws_cognito_user_pool.self.id
  generate_secret                      = false
  prevent_user_existence_errors        = "ENABLED"
  allowed_oauth_flows_user_pool_client = true
  allowed_oauth_flows                  = ["code", "implicit"]
  callback_urls                        = var.callback_urls
  supported_identity_providers         = ["COGNITO"]
  allowed_oauth_scopes                 = ["openid"]
  explicit_auth_flows = [
    "ALLOW_USER_PASSWORD_AUTH",
    "ALLOW_REFRESH_TOKEN_AUTH",
  ]
}

resource "aws_cognito_user_pool_domain" "self" {
  domain       = var.domain_name
  user_pool_id = aws_cognito_user_pool.self.id
}
