module "workload" {
  source      = "../../modules/workload"
  environment = "prd"
  system      = "lambda-support"
}

module "bucket" {
  source = "../../modules/aws-s3-bucket"
  name   = "${module.workload.name}-code-artifacts"
  tags   = module.workload.tags
}

module "contract_export" {
  source   = "../../modules/contract/export"
  workload = module.workload
  data = {
    bucket_name = module.bucket.bucket.id
    bucket_arn  = module.bucket.bucket.arn
  }
}
