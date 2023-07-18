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

data "google_compute_instance_group" "proxy" {
  project   = var.project
  self_link = var.instance_group

  lifecycle {
    postcondition {
      condition     = self.instances != null
      error_message = "No compute instances found for instance group \"${var.instance_group}\"."
    }

    postcondition {
      condition     = length(self.instances) == var.knfsd_nodes
      error_message = "Incorrect number of compute instances found, expected ${var.knfsd_nodes} but was ${length(self.instances)}."
    }
  }
}

data "google_compute_instance" "proxy" {
  count     = var.knfsd_nodes
  self_link = local.instances[count.index]
}

locals {
  instances = tolist(data.google_compute_instance_group.proxy.instances)

  ip_addresses = toset([
    for vm in data.google_compute_instance.proxy :
    vm.network_interface[0].network_ip
  ])
}

resource "google_dns_managed_zone" "proxy" {
  project     = var.project
  name        = var.proxy_basename
  dns_name    = coalesce(var.dns_name, "${var.proxy_basename}.knfsd.internal.")
  description = "Internal DNS for KNFSD proxies"
  visibility  = "private"

  private_visibility_config {
    dynamic "networks" {
      for_each = coalesce(var.networks, [data.google_compute_instance_group.proxy.network])
      content {
        network_url = networks.value
      }
    }
  }

  lifecycle {
    precondition {
      condition     = length(local.ip_addresses) == var.knfsd_nodes
      error_message = "Incorrect number of IP addresses found, expected ${var.knfsd_nodes} but was ${length(local.ip_addresses)}."
    }
  }
}

resource "google_dns_record_set" "proxy" {
  count = var.knfsd_nodes > 0 ? 1 : 0

  project      = var.project
  managed_zone = google_dns_managed_zone.proxy.name

  name = google_dns_managed_zone.proxy.dns_name
  type = "A"
  ttl  = 300

  routing_policy {
    dynamic "wrr" {
      for_each = local.ip_addresses
      iterator = ip

      content {
        weight  = 1
        rrdatas = [ip.value]
      }
    }
  }
}

moved {
  from = google_dns_record_set.proxy
  to   = google_dns_record_set.proxy[0]
}
