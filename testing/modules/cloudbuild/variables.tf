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
  type        = string
  description = "GCP Project ID to configure for building images"
}

variable "region" {
  type        = string
  description = "GCP Region that will be used to build images."
}

variable "network" {
  type        = string
  default     = "knfsd-build"
  description = "Name of private VPC network to create for use by Cloud Build."
}

variable "worker_pool" {
  type        = string
  default     = "knfsd-build"
  description = "Name of Cloud Build private worker pool to create."
}

variable "docker_repository" {
  type        = string
  default     = "knfsd-docker"
  description = "Name of the docker repository to create. This is used to store docker images used by Cloud Build."
}
