output "name" {
  value = "${var.environment}-${var.system}"
}

output "tags" {
  value = merge({
    environment = var.environment
    management  = "terraform"
    system      = var.system
  }, var.additional_tags)
}
