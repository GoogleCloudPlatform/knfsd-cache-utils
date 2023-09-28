// Copyright 2023 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

variable "build_instance_name" {
  type    = string
  default = "packer-nfs-proxy-{{uuid}}"
}

variable "network_project" {
  type    = string
  default = ""
}

variable "subnetwork" {
  type    = string
  default = "default"
}

variable "network_tags" {
  type    = set(string)
  default = ["packer"]
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
