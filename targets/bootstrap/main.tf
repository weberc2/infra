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
  name     = "TerraformStateLock"
  hash_key = "LockID"

  # I only get 25 free capacity units each month for both reading and writing,
  # so I don't want to use them all here.
  read_capacity  = 4
  write_capacity = 4


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
