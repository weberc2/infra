resource "aws_s3_bucket_object" "data" {
  bucket  = "weberc2-${var.workload.environment}-contracts"
  key     = "${var.workload.environment}/${var.workload.system}"
  content = jsonencode(var.data)
}