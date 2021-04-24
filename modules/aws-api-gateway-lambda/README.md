# README

Creates a lambda fronted by an API gateway. This module manages creating and
wiring together the lambda, the API gateway, and IAM roles for each.

## Example


```hcl
data "archive_file" "code" {
  type                    = "zip"
  source_content_filename = "lambda.py"
  source_content          = <<EOF
import json
import logging

# Setup logging that works for both AWS lambda and local execution
logging.basicConfig(level = logging.INFO)
logger = logging.getLogger()

def handler(event, context):
    logger.info("Receiving event:", json.dumps(event))
    return event
EOF
  output_path             = "test.zip"
}

module "lambda" {
  source = "../../modules/aws-api-gateway-lambda"
  name   = "test-lambda"
  lambda = {
    code = {
      filename         = data.archive_file.code.output_path
      source_code_hash = data.archive_file.code.output_base64sha256
      runtime          = "python3.8"
      handler          = "lambda.handler"
    }
    environment = {}
    memory_size = 128
  }
}

```
