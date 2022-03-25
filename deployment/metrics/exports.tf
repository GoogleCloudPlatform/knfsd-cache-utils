resource "google_monitoring_metric_descriptor" "exports_total_read_bytes" {
  description  = "Total bytes read by the NFS clients"
  display_name = "NFS Export Total Read Bytes"
  type         = "custom.googleapis.com/knfsd/exports/total_read_bytes"
  metric_kind  = "CUMULATIVE"
  value_type   = "INT64"
  unit         = "By"
}

resource "google_monitoring_metric_descriptor" "exports_total_write_bytes" {
  description  = "Total bytes wrote by the NFS clients"
  display_name = "NFS Export Total Write Bytes"
  type         = "custom.googleapis.com/knfsd/exports/total_write_bytes"
  metric_kind  = "CUMULATIVE"
  value_type   = "INT64"
  unit         = "By"
}
