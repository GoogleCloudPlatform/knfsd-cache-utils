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

# Create firewall rules instead of using AUTO_CREATE_FIREWALL_RULES to avoid
# conflicts with other cache deployments to the same project.

resource "google_compute_firewall" "allow-iap-tunnel" {
  name     = "${var.prefix}-allow-iap-tunnel"
  network  = google_compute_network.this.id
  priority = 1000

  allow {
    protocol = "TCP"
    ports    = ["22"]
  }

  source_ranges = ["35.235.240.0/20"]
}

resource "google_compute_firewall" "allow-internal" {
  name     = "${var.prefix}-allow-internal"
  network  = google_compute_network.this.id
  priority = 1000

  allow {
    # TODO: make this more specific once mountd is pinned to a specific port
    protocol = "all"
  }

  source_ranges = ["10.0.0.0/8"]
}

# Firewall rule to allow healthchecks from the GCP Healthcheck ranges
resource "google_compute_firewall" "allow-tcp-healthcheck" {
  name     = "${var.prefix}-allow-nfs-tcp-healthcheck"
  network  = google_compute_network.this.id
  priority = 1000

  allow {
    protocol = "tcp"
    ports    = ["2049"]
  }

  source_ranges = ["130.211.0.0/22", "35.191.0.0/16", "209.85.152.0/22", "209.85.204.0/22"]
  target_tags   = ["knfsd-cache-server"]
}
