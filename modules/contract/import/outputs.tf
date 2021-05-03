output "data" {
    value = jsondecode(data.aws_s3_bucket_object.data.body)
}