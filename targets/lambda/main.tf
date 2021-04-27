module "workload" {
  source      = "../../modules/workload"
  environment = "inf"
  system      = "lambda"
}

module "bucket" {
  source = "../../modules/aws-s3-bucket"
  name   = "${module.workload.name}-code-artifacts"
  tags   = module.workload.tags
}
