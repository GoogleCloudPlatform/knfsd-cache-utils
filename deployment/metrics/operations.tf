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

resource "google_monitoring_metric_descriptor" "operation_requests" {
  project      = var.project
  description  = "Number of requests"
  display_name = "NFS Number of requests"
  type         = "custom.googleapis.com/knfsd/mount/operation/requests"
  metric_kind  = "CUMULATIVE"
  value_type   = "INT64"
  unit         = "{requests}"

  dynamic "labels" {
    for_each = local.mount_operation_labels
    content {
      key         = labels.key
      value_type  = "STRING"
      description = labels.value
    }
  }
}

resource "google_monitoring_metric_descriptor" "operation_sent_bytes" {
  project      = var.project
  description  = "Total bytes sent for these operations, including RPC headers and payload"
  display_name = "NFS Total bytes sent"
  type         = "custom.googleapis.com/knfsd/mount/operation/sent_bytes"
  metric_kind  = "CUMULATIVE"
  value_type   = "INT64"
  unit         = "By"

  dynamic "labels" {
    for_each = local.mount_operation_labels
    content {
      key         = labels.key
      value_type  = "STRING"
      description = labels.value
    }
  }
}

resource "google_monitoring_metric_descriptor" "operation_received_bytes" {
  project      = var.project
  description  = "Total bytes received for these operations, including RPC headers and payload"
  display_name = "NFS Total bytes received"
  type         = "custom.googleapis.com/knfsd/mount/operation/received_bytes"
  metric_kind  = "CUMULATIVE"
  value_type   = "INT64"
  unit         = "By"

  dynamic "labels" {
    for_each = local.mount_operation_labels
    content {
      key         = labels.key
      value_type  = "STRING"
      description = labels.value
    }
  }
}

resource "google_monitoring_metric_descriptor" "operation_major_timeouts" {
  project      = var.project
  description  = "Number of times a request has had a major timeout"
  display_name = "NFS Major Timeouts"
  type         = "custom.googleapis.com/knfsd/mount/operation/major_timeouts"
  metric_kind  = "CUMULATIVE"
  value_type   = "INT64"
  unit         = "{timeouts}"

  dynamic "labels" {
    for_each = local.mount_operation_labels
    content {
      key         = labels.key
      value_type  = "STRING"
      description = labels.value
    }
  }
}

resource "google_monitoring_metric_descriptor" "operation_errors" {
  project      = var.project
  description  = "Number of requests that complete with tk_status < 0"
  display_name = "NFS Errors"
  type         = "custom.googleapis.com/knfsd/mount/operation/errors"
  metric_kind  = "CUMULATIVE"
  value_type   = "INT64"
  unit         = "{errors}"

  dynamic "labels" {
    for_each = local.mount_operation_labels
    content {
      key         = labels.key
      value_type  = "STRING"
      description = labels.value
    }
  }
}
