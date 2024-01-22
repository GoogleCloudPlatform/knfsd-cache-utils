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

module "proxy" {
  source = "../../deployment/terraform-module-knfsd"

  PROJECT = var.project
  REGION  = var.region
  ZONE    = var.zone

  PROXY_LABELS = {
    vm-type    = "nfs-proxy"
    deployment = var.name
    component  = "proxy"
  }

  NETWORK    = var.network
  SUBNETWORK = var.subnetwork

  KNFSD_NODES = 3
  LOCAL_SSDS  = 4

  PROXY_BASENAME  = "${var.name}-proxy"
  PROXY_IMAGENAME = var.proxy_image
  SERVICE_ACCOUNT = var.proxy_service_account

  AUTO_CREATE_FIREWALL_RULES      = false
  TRAFFIC_DISTRIBUTION_MODE       = "dns_round_robin"
  ASSIGN_STATIC_IPS               = true
  ENABLE_AUTOHEALING_HEALTHCHECKS = true

  FSID_MODE                = "external"
  FSID_DATABASE_PRIVATE_IP = true

  EXPORT_MAP = var.export_map
}
