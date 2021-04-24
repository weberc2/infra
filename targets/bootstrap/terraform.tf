terraform {
  required_providers {
    aws = {
      # The "hashicorp" namespace is the new home for the HashiCorp-maintained
      # provider plugins.
      #
      # source is not required for the hashicorp/* namespace as a measure of
      # backward compatibility for commonly-used providers, but recommended for
      # explicitness.
      source  = "hashicorp/aws"
      version = "~> 3.37.0"
    }
  }
}

provider "aws" {
  region = "us-east-2"
}
