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

variable "EXPORT_MAP" {
  type = ""
}

variable "DISCO_MOUNT_EXPORT_MAP" {
  type    = string
  default = ""
}

variable "EXPORT_CIDR" {
  type    = string
  default = "10.0.0.0/8"
}

variable "PROJECT" {
  type = string
}

variable "REGION" {
  default = "us-west1"
  type    = string
}

variable "ZONE" {
  default = "us-west1-a"
  type    = string
}

variable "NETWORK" {
  default = "default"
  type    = string
}

variable "SUBNETWORK" {
  default = "default"
  type    = string
}

variable "PROXY_BASENAME" {
  default = "nfsproxy"
  type    = string
}

variable "PROXY_IMAGENAME" {
  type = string
}

variable "KNFSD_NODES" {
  default = 3
  type    = number
}

variable "AUTO_CREATE_FIREWALL_RULES" {
  default = true
  type    = bool
}

variable "LOADBALANCER_IP" {
  default = null
  type    = string
}

variable "SERVICE_LABEL" {
  default = "dns"
  type    = string
}

variable "NCONNECT_VALUE" {
  default = "16"
  type    = string
}

variable "VFS_CACHE_PRESSURE" {
  default = "100"
  type    = string
}

variable "ENABLE_AUTOHEALING_HEALTHCHECKS" {
  default = true
  type    = bool
}

variable "NUM_NFS_THREADS" {
  default = 512
  type    = number
}

variable "ENABLE_STACKDRIVER_METRICS" {
  default = true
  type    = bool
}

variable "ENABLE_KNFSD_AUTOSCALING" {
  default = false
  type    = bool
}

variable "KNFSD_AUTOSCALING_MIN_INSTANCES" {
  default = 1
  type    = number
}

variable "KNFSD_AUTOSCALING_MAX_INSTANCES" {
  default = 10
  type    = number
}

variable "KNFSD_AUTOSCALING_NFS_CONNECTIONS_THRESHOLD" {
  default = 250
  type    = number
}
