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

// The IP Address of the Internal Load Balancer
output "nfsproxy_loadbalancer_ipaddress" {
  description = "The internal ip address for the nfsProxy load balancer:"
  value       = one(google_compute_address.nfsproxy_static.*.address)
}

# The Internal DNS name of the Internal Load Balancer
output "nfsproxy_loadbalancer_dnsaddress" {
  description = "The internal dns entry address for the nfsProxy load balancer:"
  value       = one(module.loadbalancer.*.dns_name)
}

output "instance_group" {
  description = "Full URL of the KNFSD proxy instance group."
  value       = google_compute_instance_group_manager.proxy-group.instance_group
}

output "instance_group_manager" {
  description = "Full URL of the KNFSD proxy instance group manager."
  value       = google_compute_instance_group_manager.proxy-group.self_link
}

output "instance_group_name" {
  description = "Name of the KNFSD proxy instance group."
  value       = google_compute_instance_group_manager.proxy-group.name
}
