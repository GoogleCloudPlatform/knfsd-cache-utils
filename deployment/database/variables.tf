/*
 * Copyright 2022 Google Inc.
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
  validation {
    condition     = var.project != ""
    error_message = "project is required."
  }
}

variable "region" {
  type     = string
  nullable = false
  validation {
    condition     = var.region != ""
    error_message = "region is required."
  }
}

variable "zone" {
  type     = string
  nullable = false
  validation {
    condition     = var.zone != ""
    error_message = "zone is required."
  }
}

variable "availability_type" {
  type     = string
  nullable = false
  default  = "ZONAL"
  validation {
    condition     = contains(["REGIONAL", "ZONAL"], var.availability_type)
    error_message = "Valid values for availability_type are REGIONAL or ZONAL."
  }
}

variable "name_prefix" {
  type     = string
  nullable = false
  default  = "fsids"
}

variable "name" {
  type    = string
  default = ""
}

variable "tier" {
  type     = string
  nullable = false
  default  = "db-custom-1-3840"
}

variable "deletion_protection" {
  type     = bool
  nullable = false
  default  = true
}

variable "proxy_service_account" {
  type     = string
  nullable = false
  validation {
    condition     = var.proxy_service_account != ""
    error_message = "proxy_service_account is required."
  }
}

variable "enable_public_ip" {
  type     = bool
  nullable = false
  default  = false
}

variable "private_network" {
  type     = string
  nullable = false
  default  = ""
}

variable "allocated_ip_range" {
  type     = string
  nullable = false
  default  = ""
}
