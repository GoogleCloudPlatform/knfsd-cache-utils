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

variable "PROJECT" {
  type    = string
  default = ""
}

variable "REGION" {
  type    = string
  default = ""
}

variable "PROXY_BASENAME" {
  type = string
}

variable "NETWORK" {
  type = string
}

variable "SUBNETWORK" {
  type = string
}

variable "SERVICE_LABEL" {
  type    = string
  default = "dns"
}

variable "IP_ADDRESS" {
  type = string
}

variable "ENABLE_UDP" {
  type    = bool
  default = false
}

variable "HEALTH_CHECK" {
  type = string
}

variable "INSTANCE_GROUP" {
  type = string
}
