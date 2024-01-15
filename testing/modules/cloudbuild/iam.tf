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

data "google_project" "this" {
  project_id = var.project
}

locals {
  build_service_account = "${data.google_project.this.number}@cloudbuild.gserviceaccount.com"
}

resource "google_project_iam_member" "build" {
  for_each = toset([
    # Build image (and test terraform)
    "roles/compute.admin",

    # Deploy and test terraform
    "roles/dns.admin",
    "roles/file.editor",
    "roles/cloudsql.admin",
  ])

  project = var.project
  role    = each.key
  member  = "serviceAccount:${local.build_service_account}"

  depends_on = [google_project_service.services]
}

resource "google_project_iam_member" "build_assign_permissions" {
  project = var.project
  role    = "roles/iam.securityAdmin"
  member  = "serviceAccount:${local.build_service_account}"

  condition {
    title       = "KNFSD roles only"
    description = "Allow granting roles required to deploy KNFSD"
    expression  = <<-EOT
      api.getAttribute('iam.googleapis.com/modifiedGrantsByRole', []).hasOnly([
        'roles/logging.logWriter',
        'roles/monitoring.metricWriter',
        'roles/cloudsql.client',
        'roles/cloudsql.instanceUser',
      ])
    EOT
  }

  depends_on = [google_project_service.services]
}

resource "google_service_account_iam_member" "build_use_proxy" {
  service_account_id = google_service_account.proxy.id
  role               = "roles/iam.serviceAccountUser"
  member             = "serviceAccount:${local.build_service_account}"
}

resource "google_service_account" "proxy" {
  project     = var.project
  account_id  = "knfsd-proxy"
  description = "KNFSD proxy cluster service account"
}

resource "google_project_iam_member" "proxy" {
  for_each = toset([
    "roles/logging.logWriter",
    "roles/monitoring.metricWriter",
  ])

  project = var.project
  role    = each.key
  member  = "serviceAccount:${google_service_account.proxy.email}"
}
