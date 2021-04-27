resource "aws_s3_bucket" "self" {
  bucket = "weberc2-${var.name}"

  server_side_encryption_configuration {
    rule {
      apply_server_side_encryption_by_default {
        sse_algorithm = "aws:kms"
      }
    }
  }
}
