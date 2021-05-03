terraform {
  backend "s3" {
    bucket = "weberc2-terraform-state"
    key    = "targets/prd/lambda-support"
    region = "us-east-2"
  }
}

provider "aws" {
  region = "us-east-2"
}
