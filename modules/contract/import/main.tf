module "target_workload" {
  source      = "../../workload"
  environment = var.environment
  system      = var.target_system
}

data "aws_s3_bucket_object" "data" {
  bucket = module.target_workload.contract_bucket
  key    = module.target_workload.contract_key
}
