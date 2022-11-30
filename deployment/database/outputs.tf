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

output "name" {
  value = google_sql_database_instance.fsids.name
}

output "sql_user" {
  value = google_sql_user.knfsd.name
}

output "database_name" {
  value = google_sql_database.fsids.name
}

output "connection_name" {
  value = google_sql_database_instance.fsids.connection_name
}

output "public_ip_address" {
  value = google_sql_database_instance.fsids.public_ip_address
}

output "private_ip_address" {
  value = google_sql_database_instance.fsids.private_ip_address
}

output "ip_address" {
  value = google_sql_database_instance.fsids.ip_address
}

output "server_ca_cert" {
  value = google_sql_database_instance.fsids.server_ca_cert
}
