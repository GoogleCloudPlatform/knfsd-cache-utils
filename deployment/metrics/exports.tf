/*
 Copyright 2022 Google LLC

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

      https://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
 */

resource "google_monitoring_metric_descriptor" "exports_total_operations" {
  project      = var.project
  description  = "Total NFS operations received from NFS clients"
  display_name = "NFS Export Total Operations"
  type         = "custom.googleapis.com/knfsd/exports/total_operations"
  metric_kind  = "CUMULATIVE"
  value_type   = "INT64"
  unit         = "{operations}"
}

resource "google_monitoring_metric_descriptor" "exports_total_read_bytes" {
  project      = var.project
  description  = "Total bytes read by the NFS clients"
  display_name = "NFS Export Total Read Bytes"
  type         = "custom.googleapis.com/knfsd/exports/total_read_bytes"
  metric_kind  = "CUMULATIVE"
  value_type   = "INT64"
  unit         = "By"
}

resource "google_monitoring_metric_descriptor" "exports_total_write_bytes" {
  project      = var.project
  description  = "Total bytes wrote by the NFS clients"
  display_name = "NFS Export Total Write Bytes"
  type         = "custom.googleapis.com/knfsd/exports/total_write_bytes"
  metric_kind  = "CUMULATIVE"
  value_type   = "INT64"
  unit         = "By"
}
