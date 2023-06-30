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

variable "EXPORT_MAP" {
  type     = string
  nullable = false
  default  = ""
}

variable "EXPORT_HOST_AUTO_DETECT" {
  type     = string
  nullable = false
  default  = ""
}

variable "EXCLUDED_EXPORTS" {
  type     = list(string)
  nullable = false
  default  = []
}

variable "INCLUDED_EXPORTS" {
  type     = list(string)
  nullable = false
  default  = []
}

variable "EXPORT_CIDR" {
  type     = string
  nullable = false
  default  = "10.0.0.0/8"
}

variable "PROJECT" {
  type     = string
  nullable = false
  default  = ""
}

variable "SUBNETWORK_PROJECT" {
  type    = string
  default = ""
}

variable "REGION" {
  type     = string
  nullable = false
  default  = ""
}

variable "ZONE" {
  type     = string
  nullable = false
  default  = ""
}

variable "NETWORK" {
  type    = string
  default = "default"
}

variable "SUBNETWORK" {
  type    = string
  default = "default"
}

variable "ASSIGN_STATIC_IPS" {
  type     = bool
  nullable = false
  default  = false
}

variable "PROXY_BASENAME" {
  type     = string
  nullable = false
  default  = "nfsproxy"
  validation {
    condition     = var.PROXY_BASENAME != ""
    error_message = "PROXY_BASENAME is required."
  }
}

variable "PROXY_LABELS" {
  type = map(string)
  default = {
    vm-type = "nfs-proxy",
  }
}

variable "PROXY_IMAGENAME" {
  type     = string
  nullable = false
  validation {
    condition     = var.PROXY_IMAGENAME != ""
    error_message = "PROXY_IMAGENAME is required."
  }
}

variable "KNFSD_NODES" {
  type     = number
  nullable = false
  default  = 3
}

variable "AUTO_CREATE_FIREWALL_RULES" {
  type     = bool
  nullable = false
  default  = true
}

variable "TRAFFIC_DISTRIBUTION_MODE" {
  type     = string
  nullable = false
  validation {
    condition     = contains(["dns_round_robin", "loadbalancer", "none"], var.TRAFFIC_DISTRIBUTION_MODE)
    error_message = "Valid values for TRAFFIC_DISTRIBUTION_MODE are 'dns_round_robin', 'loadbalancer', and 'none'."
  }
}

variable "LOADBALANCER_IP" {
  type     = string
  nullable = true
  default  = null
}

variable "DNS_NAME" {
  type     = string
  nullable = false
  default  = ""

  validation {
    condition     = var.DNS_NAME == "" || endswith(var.DNS_NAME, ".")
    error_message = "DNS_NAME must end with tailing dot, for example \"knfsd.example.\" (note the tailing dot)."
  }
}

variable "SERVICE_LABEL" {
  type    = string
  default = "dns"
}

variable "SERVICE_ACCOUNT" {
  type     = string
  nullable = false
  default  = ""
}

variable "NCONNECT_VALUE" {
  type     = string
  nullable = false
  default  = "16"
}

variable "ACDIRMIN" {
  type     = number
  nullable = false
  default  = 600
}

variable "ACDIRMAX" {
  type     = number
  nullable = false
  default  = 600
}

variable "ACREGMIN" {
  type     = number
  nullable = false
  default  = 600
}

variable "ACREGMAX" {
  type     = number
  nullable = false
  default  = 600
}

variable "RSIZE" {
  type     = number
  nullable = false
  default  = 1048576
}

variable "WSIZE" {
  type     = number
  nullable = false
  default  = 1048576
}

variable "NOHIDE" {
  type     = bool
  nullable = false
  default  = true
}

variable "MOUNT_OPTIONS" {
  type     = string
  nullable = false
  default  = ""
}

variable "EXPORT_OPTIONS" {
  type     = string
  nullable = false
  default  = ""
}

variable "AUTO_REEXPORT" {
  type     = bool
  nullable = false
  default  = false
}

variable "FSID_MODE" {
  type     = string
  nullable = false
  default  = "static"
  validation {
    condition     = contains(["static", "local", "external"], var.FSID_MODE)
    error_message = "Valid values for FSID_MODE are 'static', 'local', or 'external'."
  }
}

variable "FSID_DATABASE_DEPLOY" {
  type     = bool
  nullable = false
  default  = true
}

variable "FSID_DATABASE_PRIVATE_IP" {
  type     = bool
  nullable = true
  default  = null
}

variable "FSID_DATABASE_CONFIG" {
  type     = string
  nullable = false
  default  = ""
}

variable "VFS_CACHE_PRESSURE" {
  type     = string
  nullable = false
  default  = "100"
}

variable "READ_AHEAD" {
  type     = number
  nullable = false
  default  = 8388608
}

variable "ENABLE_UDP" {
  type     = bool
  nullable = false
  default  = false
}

variable "ENABLE_AUTOHEALING_HEALTHCHECKS" {
  type     = bool
  nullable = false
  default  = true
}

variable "HEALTHCHECK_INITIAL_DELAY_SECONDS" {
  type     = number
  nullable = false
  default  = 600
}

variable "HEALTHCHECK_INTERVAL_SECONDS" {
  type     = number
  nullable = false
  default  = 60
}

variable "HEALTHCHECK_TIMEOUT_SECONDS" {
  type     = number
  nullable = false
  default  = 2
}

variable "HEALTHCHECK_HEALTHY_THRESHOLD" {
  type     = number
  nullable = false
  default  = 3
}

variable "HEALTHCHECK_UNHEALTHY_THRESHOLD" {
  type     = number
  nullable = false
  default  = 3
}

variable "NUM_NFS_THREADS" {
  type     = number
  nullable = false
  default  = 512
}

variable "ENABLE_STACKDRIVER_METRICS" {
  type     = bool
  nullable = false
  default  = true
}

variable "METRICS_AGENT_CONFIG" {
  type     = string
  nullable = false
  default  = ""
}

variable "ROUTE_METRICS_PRIVATE_GOOGLEAPIS" {
  type     = bool
  nullable = false
  default  = false
}

variable "ENABLE_KNFSD_AUTOSCALING" {
  type     = bool
  nullable = false
  default  = false
}

variable "KNFSD_AUTOSCALING_MIN_INSTANCES" {
  type     = number
  nullable = false
  default  = 1
}

variable "KNFSD_AUTOSCALING_MAX_INSTANCES" {
  type     = number
  nullable = false
  default  = 10
}

variable "KNFSD_AUTOSCALING_NFS_CONNECTIONS_THRESHOLD" {
  type     = number
  nullable = false
  default  = 250
}

variable "CUSTOM_PRE_STARTUP_SCRIPT" {
  type     = string
  nullable = false
  default  = "echo 'Running default pre startup script. No action taken.'"
}

variable "CUSTOM_POST_STARTUP_SCRIPT" {
  type     = string
  nullable = false
  default  = "echo 'Running default post startup script. No action taken.'"
}

variable "MACHINE_TYPE" {
  type    = string
  default = "n1-highmem-16"
}

variable "MIG_MINIMAL_ACTION" {
  type    = string
  default = "RESTART"
}

variable "MIG_MAX_UNAVAILABLE_PERCENT" {
  type    = number
  default = "100"
}

variable "MIG_REPLACEMENT_METHOD" {
  type    = string
  default = ""
}

variable "ENABLE_KNFSD_AGENT" {
  type     = bool
  nullable = false
  default  = true
}

variable "DISABLED_NFS_VERSIONS" {
  type     = string
  nullable = false
  default  = "4.0,4.1,4.2"
}

variable "ENABLE_NETAPP_AUTO_DETECT" {
  type     = bool
  nullable = false
  default  = false
}

variable "NETAPP_HOST" {
  type     = string
  nullable = false
  default  = ""
}

variable "NETAPP_URL" {
  type     = string
  nullable = false
  default  = ""
}

variable "NETAPP_USER" {
  type     = string
  nullable = false
  default  = ""
}

variable "NETAPP_SECRET" {
  type     = string
  nullable = false
  default  = ""
}

variable "NETAPP_SECRET_PROJECT" {
  type     = string
  nullable = false
  default  = ""
}

variable "NETAPP_SECRET_VERSION" {
  type     = string
  nullable = false
  default  = ""
}

variable "NETAPP_CA" {
  type     = string
  nullable = false
  default  = ""
}

variable "NETAPP_ALLOW_COMMON_NAME" {
  type     = bool
  nullable = false
  default  = false
}

variable "CACHEFILESD_DISK_TYPE" {
  type     = string
  nullable = false
  default  = "local-ssd"

  validation {
    condition     = contains(["local-ssd", "pd-ssd", "pd-balanced", "pd-standard"], var.CACHEFILESD_DISK_TYPE)
    error_message = "Valid values for CACHEFILESD_DISK_TYPE are 'local-ssd', 'pd-ssd', 'pd-balanced', or 'pd-standard'."
  }
}

variable "LOCAL_SSDS" {
  type     = number
  nullable = false
  default  = 4

  validation {
    condition     = contains([0, 1, 2, 3, 4, 5, 6, 7, 8, 16, 24], var.LOCAL_SSDS)
    error_message = "Valid values for LOCAL_SSDS are 0-8, 16 or 24."
  }
}

variable "CACHEFILESD_PERSISTENT_DISK_SIZE_GB" {
  type     = number
  nullable = false
  default  = 1500

  validation {
    condition     = var.CACHEFILESD_PERSISTENT_DISK_SIZE_GB >= 10 && var.CACHEFILESD_PERSISTENT_DISK_SIZE_GB <= 64000
    error_message = "CACHEFILESD_PERSISTENT_DISK_SIZE_GB must be between 10 and 6400."
  }
}

variable "NFS_MOUNT_VERSION" {
  type     = string
  nullable = false
  default  = "3"

  validation {
    condition     = contains(["3", "4", "4.0", "4.1", "4.2"], var.NFS_MOUNT_VERSION)
    error_message = "Valid values for NFS_MOUNT_VERSION are '3', '4', '4.0', '4.1', '4.2'."
  }
}

variable "ENABLE_HIGH_BANDWIDTH_CONFIGURATION" {
  type     = bool
  nullable = false
  default  = false
}

variable "ENABLE_GVNIC" {
  type     = bool
  nullable = false
  default  = false
}
