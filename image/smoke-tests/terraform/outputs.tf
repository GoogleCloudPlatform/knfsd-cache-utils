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

output "project" {
  value = var.project
}

output "zone" {
  value = var.zone
}

output "source_host" {
  value = local.source_host
}

output "proxy_host" {
  value = local.proxy_host
}

output "proxy_mig" {
  value = module.proxy.instance_group_name
}

output "client_instance" {
  value = google_compute_instance.client.name
}
