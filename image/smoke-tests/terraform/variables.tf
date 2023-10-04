/*
 Copyright 2022 Google LLC

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

      https://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
 */

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

variable "client_image" {
  description = "Compute image for test NFS client"
  type        = string
  default     = "family/test-nfs-client"
}

variable "prefix" {
  description = "Prefix used for deployment. Used as the name or prefix for resources such as network, router, etc."
  type        = string
  default     = "smoke-tests"
}
