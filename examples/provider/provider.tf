variable "ovh_application_key" {
  type        = string
  description = "OVH Application Key"
  sensitive   = true
}

variable "ovh_application_secret" {
  type        = string
  description = "OVH Application Secret"
  sensitive   = true
}

variable "ovh_consumer_key" {
  type        = string
  description = "OVH Consumer Key"
  sensitive   = true
}

variable "ovh_endpoint" {
  type        = string
  description = "OVH Endpoint"
  default = "ovh-eu"
}

terraform {
  required_providers {
    cdcovhns = {
      source = "cdc/cdcovhns"
    }
  }
}

provider "cdcovhns" {
  endpoint           = var.ovh_endpoint
  application_key    = var.ovh_application_key
  application_secret = var.ovh_application_secret
  consumer_key       = var.ovh_consumer_key
}
