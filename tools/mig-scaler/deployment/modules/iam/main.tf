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

resource "google_project_iam_custom_role" "mig_scaler" {
  # The only default role with enough permissions to run the workflow is
  # roles/compute.admin. Create a custom role with less privileges.
  project     = var.project
  stage       = "BETA"
  role_id     = "migscaler"
  title       = "MIG Scaler"
  description = "Used by the mig-scaler workflow to scale up MIGs"
  permissions = [
    "compute.instanceGroupManagers.get",
    "compute.instanceGroupManagers.update",
  ]
}

resource "google_project_iam_member" "single_project" {
  project = var.project
  role    = google_project_iam_custom_role.mig_scaler.id
  member  = "serviceAccount:${var.service_account}"
}
