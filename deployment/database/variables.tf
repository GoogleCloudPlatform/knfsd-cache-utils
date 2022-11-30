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
  type    = string
  default = ""
}

variable "region" {
  type    = string
  default = ""
}

variable "zone" {
  type    = string
  default = ""
}

variable "availability_type" {
  type    = string
  default = "ZONAL"
  validation {
    condition     = contains(["REGIONAL", "ZONAL"], var.availability_type)
    error_message = "Valid values for availability_type are REGIONAL or ZONAL."
  }
}

variable "name_prefix" {
  type    = string
  default = "fsids"
}

variable "name" {
  type    = string
  default = ""
}

variable "tier" {
  type    = string
  default = "db-custom-1-3840"
}

variable "deletion_protection" {
  type    = bool
  default = true
}

variable "proxy_service_account" {
  type = string
}
