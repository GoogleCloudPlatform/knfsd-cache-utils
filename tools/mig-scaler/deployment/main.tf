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
  service_account = google_service_account.mig_scaler.email
}

resource "google_service_account" "mig_scaler" {
  project      = var.project
  account_id   = "mig-scaler"
  display_name = "Scales up client MIGs"
}

resource "google_project_service" "workflows" {
  project            = var.project
  service            = "workflows.googleapis.com"
  disable_on_destroy = false
}

module "iam" {
  source          = "./modules/iam"
  project         = var.project
  service_account = local.service_account
}

module "workflow" {
  source          = "./modules/workflow"
  project         = var.project
  region          = var.region
  name            = "mig-scaler"
  service_account = local.service_account
  depends_on = [
    google_project_service.workflows
  ]
}
