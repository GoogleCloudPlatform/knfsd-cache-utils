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

provider "google" {
  project = var.project
  region  = var.region
  zone    = var.zone
}

locals {
  source_host = google_filestore_instance.source.networks[0].ip_addresses[0]
  proxy_host  = module.proxy.dns_name
}

resource "google_filestore_instance" "source" {
  project  = var.project
  name     = "${var.prefix}-source"
  tier     = "BASIC_HDD"
  location = var.zone

  networks {
    network = var.network
    modes   = ["MODE_IPV4"]
  }

  file_shares {
    name        = "files"
    capacity_gb = 1024
  }
}

module "proxy" {
  source = "../../../deployment/terraform-module-knfsd"

  PROJECT = var.project
  REGION  = var.region
  ZONE    = var.zone

  NETWORK            = var.network
  SUBNETWORK         = var.subnetwork

  # SUBNETWORK_PROJECT = google_compute_subnetwork.this.project

  AUTO_CREATE_FIREWALL_RULES = false
  TRAFFIC_DISTRIBUTION_MODE  = "dns_round_robin"
  ASSIGN_STATIC_IPS          = true

  PROXY_BASENAME  = "${var.prefix}-proxy"
  PROXY_IMAGENAME = var.proxy_image

  # The smoke tests rely on using a single node so that the test client reliably
  # connects to a specific instance. Also, the smoke tests only create a single
  # client so they'd only ever connect to one instance.
  KNFSD_NODES = 1

  # Smoke tests only need a single SSD as we're not reading that much data.
  LOCAL_SSDS = 1

  EXPORT_MAP = "${local.source_host};/files;/files"
}

resource "google_compute_instance" "client" {
  project = var.project
  zone    = var.zone

  name         = "${var.prefix}-client"
  machine_type = "n1-standard-1"
  tags         = ["nfs-client"]

  boot_disk {
    initialize_params {
      image = var.client_image
    }
  }

  network_interface {
    network    = var.network
    subnetwork = var.subnetwork
  }

  metadata = {
    "source_host" = local.source_host,
    "proxy_host" = local.proxy_host,
  }
}
