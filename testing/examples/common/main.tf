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

locals {
  source = module.source.network_ip
}

module "source" {
  source  = "../../modules/source"
  name    = "${var.name}-source"
  project = var.project
  zone    = var.zone
  network = var.network
  subnet  = var.subnetwork

  labels = {
    deployment = var.name
    component  = "source"
  }

  image       = var.source_image
  nfs_image   = "source-files"
  capacity_gb = 0
}
