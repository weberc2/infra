terraform {
  backend "s3" {
    bucket = "weberc2-terraform-state"
    key    = "targets/remote-state-test"
    region = "us-east-2"
  }
}

provider "aws" {
  region = "us-east-2"
}
