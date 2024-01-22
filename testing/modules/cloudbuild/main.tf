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

terraform {
  required_version = ">= 1.6"

  required_providers {
    google = {
      version = ">= 4.84.0"
    }
  }
}

resource "google_project_service" "services" {
  for_each = toset([
    "cloudbuild.googleapis.com",
    "servicenetworking.googleapis.com",
    "artifactregistry.googleapis.com",
  ])

  project = var.project
  service = each.key
}

resource "google_cloudbuild_worker_pool" "pool" {
  project  = var.project
  name     = var.worker_pool
  location = var.region
  worker_config {
    disk_size_gb   = 100
    machine_type   = "e2-standard-2"
    no_external_ip = false
  }
  network_config {
    peered_network          = google_compute_network.build.id
    peered_network_ip_range = local.service_ranges["worker-pool"]
  }
  depends_on = [google_service_networking_connection.private_vpc_connection]
}

resource "google_artifact_registry_repository" "docker_repository" {
  project       = var.project
  location      = var.region
  format        = "DOCKER"
  repository_id = "knfsd-docker"
  description   = "Docker repository for knfsd images"
  depends_on    = [google_project_service.services]
}
