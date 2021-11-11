variable "project" {
  description = "GCP project ID"
  type        = string
}

variable "region" {
  description = "GCP region to run benchmarks"
  type        = string
}

variable "zone" {
  description = "GCP zone to run benchmarks"
  type        = string
}

variable "proxy_image" {
  description = "Compute image for the NFS proxy"
  type        = string
}

variable "prefix" {
  description = "Prefix used for deployment. Used as the name or prefix for resources such as network, router, etc."
  type        = string
  default     = "smoke-tests"
}
