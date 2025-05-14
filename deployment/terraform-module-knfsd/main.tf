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
  required_version = ">=1.5.0"
}

locals {
  enable_service_account = var.SERVICE_ACCOUNT != "" || var.ENABLE_STACKDRIVER_METRICS
  scopes = (
    var.SERVICE_ACCOUNT != "" ? ["cloud-platform"] :
    var.ENABLE_STACKDRIVER_METRICS ? ["logging-write", "monitoring-write"] :
    []
  )
  MIG_REPLACEMENT_METHOD_DEFAULT = var.ASSIGN_STATIC_IPS ? "RECREATE" : "SUBSTITUTE"
  deploy_fsid_database           = var.FSID_MODE == "external" && var.FSID_DATABASE_DEPLOY

  # Check if network/subnetwork are using simple names. If so, convert them to
  # IDs so that they can be used with resources such as Cloud SQL.
  network    = strcontains(var.NETWORK, "/") ? var.NETWORK : "projects/${var.PROJECT}/global/networks/${var.NETWORK}"
  subnetwork = strcontains(var.SUBNETWORK, "/") ? var.SUBNETWORK : "projects/${var.PROJECT}/regions/${var.REGION}/subnetworks/${var.SUBNETWORK}"
}

# Validate that SERVICE_ACCOUNT is set when deploying an external database.
# This provides a better error message with more context than the default
# error message.
resource "null_resource" "validate_fsid_database" {
  count = local.deploy_fsid_database ? 1 : 0
  lifecycle {
    precondition {
      condition     = var.SERVICE_ACCOUNT != ""
      error_message = "SERVICE_ACCOUNT is required when deploying an external fsid database. See FSID_MODE and FSID_DATABASE_DEPLOY."
    }

    precondition {
      condition     = var.FSID_DATABASE_PRIVATE_IP != null
      error_message = "FSID_DATABASE_PRIVATE_IP is required when deploying an external fsid database. See FSID_MODE and FSID_DATABASE_PRIVATE_IP."
    }
  }
}

module "fsid_database" {
  source                = "../database"
  count                 = local.deploy_fsid_database ? 1 : 0
  project               = var.PROJECT
  region                = var.REGION
  zone                  = var.ZONE
  name_prefix           = "${var.PROXY_BASENAME}-fsids"
  proxy_service_account = var.SERVICE_ACCOUNT

  # For simplicity, deploy with either a public IP or private IP but not both.
  enable_public_ip = !var.FSID_DATABASE_PRIVATE_IP
  private_network  = var.FSID_DATABASE_PRIVATE_IP ? var.NETWORK : ""

  # Simplify creating and destroying proxy cluster instances.
  deletion_protection = false

  # Modules do not support lifecycle pre/post conditions. Simulate this by
  # making the module depend on a null_resource and place the precondition on
  # the null resource.
  depends_on = [
    null_resource.validate_fsid_database,
  ]
}
