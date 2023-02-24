resource "google_monitoring_metric_descriptor" "fsid_request_count" {
  project      = var.project
  description  = "Number of requests received by the KNFSD FSID daemon."
  display_name = "knfsd-fsidd request count"
  type         = "custom.googleapis.com/knfsd/fsid/request/count"
  metric_kind  = "CUMULATIVE"
  value_type   = "INT64"
  unit         = "1"

  labels {
    key         = "command"
    description = "The command that was requested, such as \"get_fsid\"."
  }

  labels {
    key         = "result"
    description = "The result of the request, such as \"ok\"."
  }
}

resource "google_monitoring_metric_descriptor" "fsid_request_duration" {
  project      = var.project
  description  = "Total duration of requests received (including retries) by the KNFSD FSID daemon."
  display_name = "knfsd-fsidd request duration"
  type         = "custom.googleapis.com/knfsd/fsid/request/duration"
  metric_kind  = "CUMULATIVE"
  value_type   = "DISTRIBUTION"
  unit         = "ms"

  labels {
    key         = "command"
    description = "The command that was requested, such as \"get_fsid\"."
  }

  labels {
    key         = "result"
    description = "The result of the request, such as \"ok\"."
  }
}

resource "google_monitoring_metric_descriptor" "fsid_request_retries" {
  project      = var.project
  description  = "Number of times each request was retried."
  display_name = "knfsd fsidd request retries"
  type         = "custom.googleapis.com/knfsd/fsid/request/retries"
  metric_kind  = "CUMULATIVE"
  value_type   = "DISTRIBUTION"
  unit         = "{retries}"

  labels {
    key         = "command"
    description = "The command that was requested, such as \"get_fsid\"."
  }

  labels {
    key         = "result"
    description = "The result of the request, such as \"ok\"."
  }
}

resource "google_monitoring_metric_descriptor" "fsid_operation_count" {
  project      = var.project
  description  = "Number of operations performed by the KNFSD FSID daemon. Each attempt to handle a request is one operation."
  display_name = "knfsd-fsidd operation count"
  type         = "custom.googleapis.com/knfsd/fsid/operation/count"
  metric_kind  = "CUMULATIVE"
  value_type   = "INT64"
  unit         = "1"

  labels {
    key         = "command"
    description = "The command that was requested, such as \"get_fsid\"."
  }

  labels {
    key         = "result"
    description = "The result of the request, such as \"ok\"."
  }

  labels {
    key         = "retry"
    description = "The retry count for this operation."
  }
}

resource "google_monitoring_metric_descriptor" "fsid_operation_duration" {
  project      = var.project
  description  = "Duration of each operation performed by the KNFSD FSID daemon. Each attempt to handle a request is one operation."
  display_name = "knfsd-fsidd operation duration"
  type         = "custom.googleapis.com/knfsd/fsid/operation/duration"
  metric_kind  = "CUMULATIVE"
  value_type   = "DISTRIBUTION"
  unit         = "ms"

  labels {
    key         = "command"
    description = "The command that was requested, such as \"get_fsid\"."
  }

  labels {
    key         = "result"
    description = "The result of the request, such as \"ok\"."
  }

  labels {
    key         = "retry"
    description = "The retry count for this operation."
  }
}


resource "google_monitoring_metric_descriptor" "fsid_sql_query_count" {
  project      = var.project
  description  = "Number of SQL queries executed by the KNFSD FSID daemon."
  display_name = "knfsd-fsidd SQL query count"
  type         = "custom.googleapis.com/knfsd/fsid/sql/query/count"
  metric_kind  = "CUMULATIVE"
  value_type   = "INT64"
  unit         = "1"

  labels {
    key         = "query"
    description = "The query that was executed, such as \"get_fsid\"."
  }

  labels {
    key         = "result"
    description = "The result of the query, such as \"ok\"."
  }
}

resource "google_monitoring_metric_descriptor" "fsid_sql_query_duration" {
  project      = var.project
  description  = "Duration of SQL queries executed by the KNFSD FSID daemon."
  display_name = "knfsd-fsidd SQL query duration"
  type         = "custom.googleapis.com/knfsd/fsid/sql/query/duration"
  metric_kind  = "CUMULATIVE"
  value_type   = "DISTRIBUTION"
  unit         = "ms"

  labels {
    key         = "query"
    description = "The query that was executed, such as \"get_fsid\"."
  }

  labels {
    key         = "result"
    description = "The result of the query, such as \"ok\"."
  }
}
