/*
 * Copyright 2022 Google Inc.
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

resource "random_id" "name" {
  prefix      = "${var.name_prefix}-"
  byte_length = 4
  keepers = {
    tier   = var.tier,
    region = var.region,
    zone   = var.zone
  }
}

locals {
  name = coalesce(var.name, random_id.name.hex)
}

resource "google_project_service" "sqladmin" {
  project            = var.project
  service            = "sqladmin.googleapis.com"
  disable_on_destroy = false
}

resource "google_sql_database_instance" "fsids" {
  project = var.project
  region  = var.region
  name    = local.name

  database_version = "POSTGRES_14"

  deletion_protection = var.deletion_protection

  settings {
    tier      = var.tier
    disk_size = "10"
    disk_type = "PD_SSD"

    database_flags {
      name  = "cloudsql.iam_authentication"
      value = "on"
    }

    backup_configuration {
      enabled                        = true
      point_in_time_recovery_enabled = true
      transaction_log_retention_days = 1
      backup_retention_settings {
        retained_backups = 2
      }
    }

    ip_configuration {
      ipv4_enabled = true
    }

    location_preference {
      zone = var.zone
    }
  }
}

resource "google_sql_database" "fsids" {
  project  = var.project
  name     = "fsids"
  instance = google_sql_database_instance.fsids.name
}

resource "google_sql_user" "knfsd" {
  # Due to length limit on Postgresql user names the .gserviceaccount.com suffix is omitted
  # https://cloud.google.com/sql/docs/postgres/add-manage-iam-users#creating-a-database-user
  project  = var.project
  name     = trimsuffix(var.proxy_service_account, ".gserviceaccount.com")
  instance = google_sql_database_instance.fsids.name
  type     = "CLOUD_IAM_SERVICE_ACCOUNT"
}

resource "google_project_iam_member" "client" {
  project = var.project
  role    = "roles/cloudsql.client"
  member  = "serviceAccount:${var.proxy_service_account}"
}

resource "google_project_iam_member" "user" {
  project = var.project
  role    = "roles/cloudsql.instanceUser"
  member  = "serviceAccount:${var.proxy_service_account}"
}
