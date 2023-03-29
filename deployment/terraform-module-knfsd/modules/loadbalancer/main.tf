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

terraform {
  required_version = ">=1.2.0"
}

# Load Balancer backend service for the Knfsd Cluster
resource "google_compute_region_backend_service" "nfsproxy" {
  project               = var.PROJECT
  region                = var.REGION
  name                  = "${var.PROXY_BASENAME}-backend-service"
  health_checks         = [var.HEALTH_CHECK]
  load_balancing_scheme = "INTERNAL"
  session_affinity      = "CLIENT_IP"
  protocol              = "TCP"
  timeout_sec           = 10
  backend {
    description = "Load Balancer backend for nfsProxy managed instance group."
    group       = var.INSTANCE_GROUP
  }
}

# Load Balancer backend service for the Knfsd Cluster
resource "google_compute_region_backend_service" "nfsproxy_udp" {
  count                 = var.ENABLE_UDP ? 1 : 0
  project               = var.PROJECT
  region                = var.REGION
  name                  = "${var.PROXY_BASENAME}-backend-service-udp"
  health_checks         = [var.HEALTH_CHECK]
  load_balancing_scheme = "INTERNAL"
  session_affinity      = "CLIENT_IP"
  protocol              = "UDP"
  timeout_sec           = 10
  backend {
    description = "Load Balancer backend for NFS proxy managed instance group."
    group       = var.INSTANCE_GROUP
  }
}

# Load Balancer forwarding rule service for the Knfsd Cluster
resource "google_compute_forwarding_rule" "default" {
  project               = var.PROJECT
  region                = var.REGION
  name                  = var.PROXY_BASENAME
  load_balancing_scheme = "INTERNAL"
  backend_service       = google_compute_region_backend_service.nfsproxy.self_link
  ip_protocol           = "TCP"
  ip_address            = var.IP_ADDRESS
  all_ports             = true
  network               = var.NETWORK
  subnetwork            = var.SUBNETWORK
  service_label         = var.SERVICE_LABEL
}

resource "google_compute_forwarding_rule" "udp" {
  count                 = var.ENABLE_UDP ? 1 : 0
  project               = var.PROJECT
  region                = var.REGION
  name                  = "${var.PROXY_BASENAME}-udp"
  load_balancing_scheme = "INTERNAL"
  backend_service       = google_compute_region_backend_service.nfsproxy_udp[0].self_link
  ip_protocol           = "UDP"
  ip_address            = var.IP_ADDRESS
  all_ports             = true
  network               = var.NETWORK
  subnetwork            = var.SUBNETWORK
  service_label         = var.SERVICE_LABEL
}
