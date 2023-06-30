/*
 * Copyright 2020 Google Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

variable "project" {
  type     = string
  nullable = false
  default  = ""
}

variable "networks" {
  type     = set(string)
  nullable = true
  default  = null
}

variable "instance_group" {
  type     = string
  nullable = false
  validation {
    condition     = var.instance_group != ""
    error_message = "instance_group is required."
  }
}

variable "proxy_basename" {
  type     = string
  nullable = false
  validation {
    condition     = var.proxy_basename != ""
    error_message = "proxy_basename is required."
  }
}

variable "dns_name" {
  type     = string
  nullable = false
  default  = ""
  validation {
    condition     = var.dns_name == "" || endswith(var.dns_name, ".")
    error_message = "dns_name must end with tailing dot, for example \"knfsd.example.\" (note the tailing dot)."
  }
}

variable "knfsd_nodes" {
  type     = number
  nullable = false
}
