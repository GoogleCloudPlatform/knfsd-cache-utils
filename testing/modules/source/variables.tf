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
  type = string
}

variable "zone" {
  type = string
}

variable "network" {
  type = string
}

variable "subnet" {
  type = string
}

variable "name" {
  description = "Instance name of the source server"
  type        = string
}

variable "labels" {
  type    = map(string)
  default = {}
}

variable "image" {
  # Avoid using proxy image for the source server to avoid possible issues
  # with the latest NFS versions. This way the source server uses older
  # stable versions and the only component being tested with newer versions
  # is the proxy.
  description = "Image for NFS server (boot disk)"
  type        = string
}

variable "nfs_image" {
  description = "Disk image for NFS share"
  type        = string
  default     = ""
}

variable "capacity_gb" {
  description = "Size of the source NFS share in GB"
  type        = number
  default     = 1024
}

variable "latency_ms" {
  description = "Latency (delay) to apply in milliseconds"
  type        = number
  default     = 0
}

variable "rate_limit_mbit" {
  description = "Rate limit to apply in megabits per second, 0 means unlimited"
  type        = number
  default     = 0
}
