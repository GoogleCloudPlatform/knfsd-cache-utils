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

resource "google_compute_disk" "source" {
  project = var.project
  zone    = var.zone
  name    = "${var.name}-nfs"
  image   = var.nfs_image
  size    = var.capacity_gb == 0 ? null : var.capacity_gb
  type    = "pd-ssd"
}

resource "google_compute_instance" "source" {
  project = var.project
  zone    = var.zone
  name    = var.name
  labels  = var.labels
  tags    = ["nfs-server"]

  machine_type     = "n1-standard-16"
  min_cpu_platform = "Intel Skylake"

  boot_disk {
    auto_delete = true
    initialize_params {
      image = var.image
      size  = 200
    }
  }

  attached_disk {
    source      = google_compute_disk.source.self_link
    device_name = "nfs"
  }

  network_interface {
    network    = var.network
    subnetwork = var.subnet
  }

  metadata_startup_script = file("${path.module}/scripts/startup")
  metadata = {
    "delay" = var.latency_ms == 0 ? "" : "${var.latency_ms}ms"
    "rate"  = var.rate_limit_mbit == 0 ? "" : "${var.rate_limit_mbit}MBit"
  }
}
