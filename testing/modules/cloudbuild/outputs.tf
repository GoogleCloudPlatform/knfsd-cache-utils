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

output "network" {
  value = google_compute_network.build.id
}

output "subnetwork" {
  value = google_compute_subnetwork.build.id
}

output "worker_pool" {
  value = google_cloudbuild_worker_pool.pool.id
}

output "docker_repository_url" {
  value = "${google_artifact_registry_repository.docker_repository.location}-docker.pkg.dev/${google_artifact_registry_repository.docker_repository.project}/${google_artifact_registry_repository.docker_repository.repository_id}"
}

output "proxy_service_account" {
  value = google_service_account.proxy.email
}
