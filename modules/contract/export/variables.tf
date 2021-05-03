variable "workload" {
  type = object({
    environment = string
    system      = string
  })
  description = "(Required) The workload module for the caller's system."
}

variable "data" {
  type        = map(any)
  description = "(Required) The data to store in the contract instance"
}