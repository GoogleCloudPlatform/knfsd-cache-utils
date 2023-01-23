variable "project" {
  type    = string
  default = ""
}

variable "zone" {
  type    = string
  default = ""
}

variable "machine_type" {
  type    = string
  default = "c2-standard-16"
}

variable "network_project" {
  type    = string
  default = ""
}

variable "subnetwork" {
  type    = string
  default = "default"
}

variable "omit_external_ip" {
  type    = bool
  default = true
}

variable "image_name" {
  type    = string
  default = ""
}

variable "image_family" {
  type    = string
  default = "nfs-proxy"
}

variable "image_storage_location" {
  type    = string
  default = ""
}

variable "use_iap" {
  type    = bool
  default = true
}

variable "use_internal_ip" {
  type    = bool
  default = true
}

variable "skip_create_image" {
  type        = bool
  default     = false
  description = "Skip creating the image. Useful when testing the changes to the build scripts."
}
