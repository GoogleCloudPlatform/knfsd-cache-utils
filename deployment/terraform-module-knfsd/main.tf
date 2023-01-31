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

terraform {
  required_version = ">=1.2.0"
}

locals {
  enable_service_account = var.SERVICE_ACCOUNT != "" || var.ENABLE_STACKDRIVER_METRICS
  scopes = (
    var.SERVICE_ACCOUNT != "" ? ["cloud-platform"] :
    var.ENABLE_STACKDRIVER_METRICS ? ["logging-write", "monitoring-write"] :
    []
  )
  CULLING_LAST_ACCESS_DEFAULT = var.CACHEFILESD_DISK_TYPE == "local-ssd" ? "${var.LOCAL_SSDS}h" : "6h"
  MIG_REPLACEMENT_METHOD_DEFAULT = var.ASSIGN_STATIC_IPS ? "RECREATE" : "SUBSTITUTE"
}
