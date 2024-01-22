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

variable "source_image" {
  description = "Compute image for the source NFS server"
  type        = string
  default     = "family/nfs"
}

variable "name" {
  description = "Name of deployment. Used as the name or prefix for resources such as network, router, etc."
  type        = string
}

variable "network" {
  description = "network to use for source instance"
  type        = string
}

variable "subnetwork" {
  description = "subnet to use for source instance"
  type        = string
}
