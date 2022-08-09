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
  source_ip = google_filestore_instance.source.networks[0].ip_addresses[0]
  proxy_ip  = module.proxy.nfsproxy_loadbalancer_ipaddress
}

resource "google_filestore_instance" "source" {
  project = var.project
  name    = "${var.prefix}-source"
  tier    = "BASIC_HDD"
  zone    = var.zone

  networks {
    network = google_compute_network.this.name
    modes   = ["MODE_IPV4"]
  }

  file_shares {
    name        = "files"
    capacity_gb = 1024
  }
}

module "proxy" {
  source = "github.com/GoogleCloudPlatform/knfsd-cache-utils//deployment/terraform-module-knfsd?ref=v0.8.0"

  PROJECT = var.project
  REGION  = var.region
  ZONE    = var.zone

  NETWORK    = google_compute_network.this.name
  SUBNETWORK = google_compute_subnetwork.this.name

  AUTO_CREATE_FIREWALL_RULES = false

  PROXY_BASENAME  = "${var.prefix}-proxy"
  PROXY_IMAGENAME = var.proxy_image

  # The smoke tests rely on using a single node so that
  KNFSD_NODES = 1

  EXPORT_MAP = "${local.source_ip};/files;/files"
}

resource "google_compute_instance" "client" {
  project = var.project
  zone    = var.zone

  name         = "${var.prefix}-client"
  machine_type = "n1-standard-1"

  boot_disk {
    initialize_params {
      image = "family/ubuntu-2004-lts"
    }
  }

  network_interface {
    network    = google_compute_network.this.id
    subnetwork = google_compute_subnetwork.this.id
  }
}
