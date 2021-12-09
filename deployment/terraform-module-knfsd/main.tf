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

locals {
  enable_service_account = var.SERVICE_ACCOUNT != "" || var.ENABLE_STACKDRIVER_METRICS
  scopes = (
    var.SERVICE_ACCOUNT != "" ? ["cloud-platform"] :
    var.ENABLE_STACKDRIVER_METRICS ? ["logging-write", "monitoring-write"] :
    []
  )
}

# Static IP used for the Load Balancer. This can be manually set via the LOADBALANCER_IP variable, otherwise it defaults to null which allocates a random IP address
resource "google_compute_address" "nfsproxy_static" {
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
