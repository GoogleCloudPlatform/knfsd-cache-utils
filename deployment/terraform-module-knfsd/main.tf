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


# Static IP used for the Load Balancer. This can be manually set via the LOADBALANCER_IP variable, otherwise it defaults to null which allocates a random IP address
resource "google_compute_address" "nfsproxy_static" {
  name         = "${var.PROXY_BASENAME}-static-ip"
  address_type = "INTERNAL"
  subnetwork   = var.SUBNETWORK
  address      = var.LOADBALANCER_IP
}
