output "name" {
  value = "${var.environment}-${var.system}"
}

output "environment" {
  value = var.environment
}

output "system" {
  value = var.system
}

output "contract_bucket" {
  value = "weberc2-${var.environment}-contracts"
}

output "contract_key" {
  value = var.system
}

output "tags" {
  value = merge({
    environment = var.environment
    management  = "terraform"
    system      = var.system
  }, var.additional_tags)
}
