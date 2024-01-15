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

locals {
  // format the service ranges in standard CIDR (IP/prefix) format.
  service_ranges = {
    for k, sa in google_compute_global_address.service_addresses :
    k => "${sa.address}/${sa.prefix_length}"
  }
}

resource "google_project_service" "servicenetworking" {
  project            = var.project
  service            = "servicenetworking.googleapis.com"
  disable_on_destroy = false
}

resource "google_compute_network" "build" {
  project                 = var.project
  name                    = var.network
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "build" {
  project = google_compute_network.build.project
  network = google_compute_network.build.id
  region  = var.region
  name    = var.network

  ip_cidr_range            = "10.0.0.0/20"
  private_ip_google_access = true
}

resource "google_compute_global_address" "service_addresses" {
  for_each = toset(["worker-pool", "database"])
  project  = var.project
  name     = "${var.network}-${each.key}"

  purpose       = "VPC_PEERING"
  address_type  = "INTERNAL"
  prefix_length = 20
  network       = google_compute_network.build.name
}

resource "google_service_networking_connection" "private_vpc_connection" {
  network = google_compute_network.build.id
  service = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [
    for _, a in google_compute_global_address.service_addresses : a.name
  ]
  depends_on = [
    google_project_service.servicenetworking,
  ]
}

resource "google_compute_firewall" "allow-ssh" {
  project       = var.project
  name          = "${var.network}-cloudbuild-ssh"
  network       = google_compute_network.build.id
  source_ranges = [local.service_ranges["worker-pool"]]
  allow {
    protocol = "tcp"
    ports    = ["22"]
  }
}
