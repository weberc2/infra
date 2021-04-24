output "api_gateway_role" {
  value = aws_iam_role.self
}

output "api_gateway" {
  value = aws_apigatewayv2_api.self
}

output "function" {
  value = module.lambda.function
}

output "function_role" {
  value = module.lambda.role
}
