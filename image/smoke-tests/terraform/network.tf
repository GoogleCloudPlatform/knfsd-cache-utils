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

resource "google_compute_network" "this" {
  project                 = var.project
  name                    = var.prefix
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "this" {
  project = google_compute_network.this.project
  network = google_compute_network.this.id
  region  = var.region
  name    = var.prefix

  ip_cidr_range = "10.0.0.0/20"

  private_ip_google_access = true
}

resource "google_compute_router" "this" {
  name    = var.prefix
  network = google_compute_network.this.id
  region  = var.region
}

resource "google_compute_router_nat" "this" {
  name   = var.prefix
  router = google_compute_router.this.name
  region = google_compute_router.this.region

  nat_ip_allocate_option             = "AUTO_ONLY"
  source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"
}
