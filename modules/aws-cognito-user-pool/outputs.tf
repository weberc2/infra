output "user_pool" {
  value = aws_cognito_user_pool.self
}

# Can't output the whole client because its `client_secret` attribute is a
# secret.
output "user_pool_client_id" {
  value = aws_cognito_user_pool_client.self.id
}

output "aws_cognito_user_pool_domain" {
  value = aws_cognito_user_pool_domain.self
}
