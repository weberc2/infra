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

resource "aws_s3_bucket" "bad" {}

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

resource "aws_iam_user_policy_attachment" "admin" {
  user       = aws_iam_user.terraform.name
  policy_arn = "arn:aws:iam::aws:policy/AdministratorAccess"
}
