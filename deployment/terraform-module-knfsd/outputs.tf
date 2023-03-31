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

output "nfsproxy_loadbalancer_ipaddress" {
  description = "The internal IP address of the load balancer."
  value       = one(google_compute_address.nfsproxy_static.*.address)
}

output "nfsproxy_loadbalancer_dnsaddress" {
  description = "The internal DNS name of the load balancer."
  value       = one(module.loadbalancer.*.dns_name)
}

output "dns_name" {
  description = "The internal DNS name of the KNFSD proxy instance group."
  value = (
    var.TRAFFIC_DISTRIBUTION_MODE == "loadbalancer" ? one(module.loadbalancer.*.dns_name) :
    var.TRAFFIC_DISTRIBUTION_MODE == "dns_round_robin" ? one(module.dns_round_robin.*.dns_name) :
    null
  )
}

output "instance_group" {
  description = "Full URL (self link) of the KNFSD proxy instance group."
  value       = google_compute_instance_group_manager.proxy-group.instance_group
}

output "instance_group_manager" {
  description = "Full URL (self link) of the KNFSD proxy instance group manager."
  value       = google_compute_instance_group_manager.proxy-group.self_link
}

output "instance_group_name" {
  description = "Name of the KNFSD proxy instance group."
  value       = google_compute_instance_group_manager.proxy-group.name
}
