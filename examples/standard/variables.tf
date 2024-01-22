/*
 * Copyright 2024 Google Inc.
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
  description = "name of the project where knfsd proxy resources gets created"
  type        = string
  nullable    = false
}

variable "region" {
  description = "REGION where to host the knfsd proxy components"
  type        = string
  nullable    = false
}

variable "zone" {
  description = "ZONE, where to host the knfsd proxy components"
  type        = string
  nullable    = false
}

variable "name" {
  description = "NAME prefix for the components"
  type        = string
  nullable    = false
}

variable "network" {
  description = "Fully qualified NETWORK path"
  default     = "default"
  type        = string
}

variable "subnetwork" {
  description = "Fully qualified SUBNET path"
  default     = "default"
  type        = string
}

variable "proxy_image" {
  description = "Compute image for the NFS proxy"
  type        = string
  nullable    = false
}

variable "proxy_service_account" {
  type     = string
  nullable = false
}

variable "export_map" {
  description = "SOURCE export map"
  type        = string
  nullable    = false
}
