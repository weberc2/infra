module "workload" {
  source = "../../modules/workload"

  environment = var.name
  system      = "environment"
}

module "contracts" {
  source = "../../modules/aws-s3-bucket"

  name = "${var.name}-contracts"
  tags = module.workload.tags
}