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

  // Adding custom labels to identify the proxy group can help with filtering
  // metrics in Cloud Monitoring, and logs.
  // Including the default vm-type = "nfs-proxy" label for compatibility with
  // the example dashboard.
  PROXY_LABELS = {
    vm-type    = "nfs-proxy"
    deployment = var.name
    component  = "proxy"
  }

  NETWORK    = var.network
  SUBNETWORK = var.subnetwork

  KNFSD_NODES = 1
  LOCAL_SSDS  = 4

  PROXY_BASENAME  = "${var.name}-proxy"
  PROXY_IMAGENAME = var.proxy_image

  // Normally for production you should create the firewall rules yourself.
  // For the basic test, create the firewall rules automatically.
  // You cannot deploy more than one proxy cluster with
  // AUTO_CREATE_FIREWALL_RULES set to true in the same GCP project.
  AUTO_CREATE_FIREWALL_RULES = true

  TRAFFIC_DISTRIBUTION_MODE = "dns_round_robin"
  ASSIGN_STATIC_IPS         = true

  // For simplicity of the basic test using FSID_MODE static.
  // This should not be used in production or with more than a single node.
  FSID_MODE = "static"

  EXPORT_MAP = var.export_map
}
