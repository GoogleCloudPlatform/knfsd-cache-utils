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

output "project" {
  description = "project where GCP resources are created"
  value       = var.project
}

output "region" {
  description = "GCP region where resources were created"
  value       = var.region
}

output "zone" {
  description = "GCP zone where resources were created"
  value       = var.zone
}

output "proxy_instance_group" {
  description = "Name of the KNFSD proxy instance group."
  value       = module.proxy.instance_group_name
}

output "proxy_host" {
  description = "Proxy host DNS name"
  value       = module.proxy.dns_name
}
