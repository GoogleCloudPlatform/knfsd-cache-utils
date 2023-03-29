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

# Static IP used for the Load Balancer. This can be manually set via the
# LOADBALANCER_IP variable, otherwise it defaults to null which allocates a
# random IP address.
# This cannot be moved to the loadbalancer module as it creates a cycle between
# the managed instance group and the loadbalancer module. The module needs the
# MIG's self_link, but the MIG needs the load balancer's IP address.
resource "google_compute_address" "nfsproxy_static" {
  # This module will be made optional later, set the count to 1 so that the
  # correct refactoring can be created and tested.
  count = 1

  project      = var.PROJECT
  region       = var.REGION
  name         = "${var.PROXY_BASENAME}-static-ip"
  address_type = "INTERNAL"
  subnetwork   = var.SUBNETWORK
  address      = var.LOADBALANCER_IP
  purpose      = "SHARED_LOADBALANCER_VIP"

  lifecycle {
    # Cannot change the purpose of an address that is in use.
    # Originally this resource was deployed with a purpose of "GCE_ENDPOINT".
    # To support both TCP and UDP the purpose needs to be changed to "SHARED_LOADBALANCER_VIP".
    ignore_changes = [purpose]
  }
}


module "loadbalancer" {
  # This module will be made optional later, set the count to 1 so that the
  # correct refactoring can be created and tested.
  count  = 1
  source = "./modules/loadbalancer"

  PROJECT        = var.PROJECT
  REGION         = var.REGION
  PROXY_BASENAME = var.PROXY_BASENAME
  NETWORK        = var.NETWORK
  SUBNETWORK     = var.SUBNETWORK
  SERVICE_LABEL  = var.SERVICE_LABEL
  IP_ADDRESS     = google_compute_address.nfsproxy_static[0].address
  ENABLE_UDP     = var.ENABLE_UDP
  HEALTH_CHECK   = google_compute_health_check.autohealing.self_link
  INSTANCE_GROUP = google_compute_instance_group_manager.proxy-group.instance_group
}

moved {
  from = google_compute_address.nfsproxy_static
  to   = google_compute_address.nfsproxy_static[0]
}

moved {
  from = google_compute_region_backend_service.nfsproxy
  to   = module.loadbalancer[0].google_compute_region_backend_service.nfsproxy
}

moved {
  from = google_compute_region_backend_service.nfsproxy_udp
  to   = module.loadbalancer[0].google_compute_region_backend_service.nfsproxy_udp
}

moved {
  from = google_compute_forwarding_rule.default
  to   = module.loadbalancer[0].google_compute_forwarding_rule.default
}

moved {
  from = google_compute_forwarding_rule.udp
  to   = module.loadbalancer[0].google_compute_forwarding_rule.udp
}
