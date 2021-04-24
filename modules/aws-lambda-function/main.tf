resource "aws_iam_role" "self" {
  name        = var.name
  tags        = var.tags
  description = "Execution role for the ${var.name} lambda function"

  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Action = "sts:AssumeRole"
        Principal = {
          Service = "lambda.amazonaws.com"
        },
        Effect = "Allow"
      }
    ]
  })
}

resource "aws_iam_role_policy" "self" {
  name = "lambda-execution"
  role = aws_iam_role.self.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents",
        ]
        Effect   = "Allow"
        Resource = "*"
      },
    ]
  })
}

resource "aws_lambda_function" "self" {
  function_name    = var.name
  role             = aws_iam_role.self.arn
  filename         = var.code.filename
  source_code_hash = var.code.source_code_hash
  runtime          = var.code.runtime
  handler          = var.code.handler
  tags             = var.tags
  memory_size      = var.memory_size

  dynamic "environment" {
    for_each = var.environment
    content {
      variables = environment.value
    }
  }
}
