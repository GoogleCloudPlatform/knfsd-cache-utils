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
  type     = string
  nullable = false
}

variable "region" {
  type     = string
  nullable = false
}

variable "zone" {
  type     = string
  nullable = false
}

variable "name" {
  type     = string
  nullable = false
}

variable "network" {
  type     = string
  nullable = false
  default  = "default"
}

variable "subnetwork" {
  type     = string
  nullable = false
  default  = "default"
}

variable "proxy_image" {
  type     = string
  nullable = false
}

variable "export_map" {
  type     = string
  nullable = false
}
