locals {
  account_alias = "weberc2"
}

resource "aws_s3_bucket" "state" {
  bucket = "weberc2-terraform-state"

  server_side_encryption_configuration {
    rule {
      apply_server_side_encryption_by_default {
        sse_algorithm = "aws:kms"
      }
    }
  }
}

resource "aws_dynamodb_table" "lock" {
  name           = "TerraformStateLock"
  hash_key       = "LockID"
  read_capacity  = 25
  write_capacity = 25


  attribute {
    name = "LockID"
    type = "S"
  }
}

resource "aws_iam_user" "terraform" {
  name = "terraform"
  path = "/system/"
}

resource "aws_iam_user_policy" "manage_state" {
  name = "manage-terraform-state"
  user = aws_iam_user.terraform.name
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "s3:ListBucket"
        ],
        Effect   = "Allow",
        Resource = "arn:aws:s3:::${aws_s3_bucket.state.bucket}"
      },
      {
        Action = [
          "s3:GetObject",
          "s3:PutObject"
        ],
        Effect   = "Allow",
        Resource = "arn:aws:s3:::${aws_s3_bucket.state.bucket}/*"
      },
      {
        Effect = "Allow",
        Action = [
          "dynamodb:GetItem",
          "dynamodb:PutItem",
          "dynamodb:DeleteItem"
        ],
        Resource = aws_dynamodb_table.lock.arn
      }
    ]
  })
}
