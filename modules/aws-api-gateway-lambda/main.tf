module "lambda" {
  source      = "../aws-lambda-function"
  name        = "${var.name}-function"
  tags        = var.tags
  code        = var.lambda.code
  environment = var.lambda.environment
  memory_size = var.lambda.memory_size
}

resource "aws_iam_role" "self" {
  name        = "${var.name}-api-gateway"
  description = "API Gateway role to invoke lambda function"
  tags        = var.tags

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect    = "Allow"
        Principal = { Service = ["apigateway.amazonaws.com"] }
        Action    = ["sts:AssumeRole"]
      }
    ]
  })
}

resource "aws_iam_role_policy" "self" {
  name = "invoke-lambda"
  role = aws_iam_role.self.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action   = "lambda:InvokeFunction"
        Effect   = "Allow"
        Resource = module.lambda.function.arn
      }
    ]
  })
}

resource "aws_apigatewayv2_api" "self" {
  name            = var.name
  protocol_type   = "HTTP"
  target          = module.lambda.function.arn
  credentials_arn = aws_iam_role.self.arn
  tags            = var.tags
}
