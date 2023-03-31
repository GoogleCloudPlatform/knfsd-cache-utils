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
  type    = string
  default = ""
}

variable "EXPORT_HOST_AUTO_DETECT" {
  type    = string
  default = ""
}

variable "EXCLUDED_EXPORTS" {
  type    = list(string)
  default = []
}

variable "INCLUDED_EXPORTS" {
  type    = list(string)
  default = []
}

variable "EXPORT_CIDR" {
  type    = string
  default = "10.0.0.0/8"
}

variable "PROJECT" {
  type    = string
  default = ""
}

variable "SUBNETWORK_PROJECT" {
  type    = string
  default = ""
}

variable "REGION" {
  type    = string
  default = ""
}

variable "ZONE" {
  type    = string
  default = ""
}

variable "NETWORK" {
  default = "default"
  type    = string
}

variable "SUBNETWORK" {
  default = "default"
  type    = string
}

variable "ASSIGN_STATIC_IPS" {
  default = false
  type    = bool
}

variable "PROXY_BASENAME" {
  default = "nfsproxy"
  type    = string
}

variable "PROXY_LABELS" {
  default = {
    vm-type = "nfs-proxy",
  }
  type = map(string)
}

variable "PROXY_IMAGENAME" {
  type = string
}

variable "KNFSD_NODES" {
  default = 3
  type    = number
}

variable "AUTO_CREATE_FIREWALL_RULES" {
  default = true
  type    = bool
}

variable "TRAFFIC_DISTRIBUTION_MODE" {
  type = string
  validation {
    condition     = contains(["dns_round_robin", "loadbalancer", "none"], var.TRAFFIC_DISTRIBUTION_MODE)
    error_message = "Valid values for TRAFFIC_DISTRIBUTION_MODE are 'dns_round_robin', 'loadbalancer', and 'none'."
  }
}

variable "LOADBALANCER_IP" {
  default = null
  type    = string
}

variable "DNS_NAME" {
  default = ""
  type    = string

  validation {
    condition     = var.DNS_NAME == "" || endswith(var.DNS_NAME, ".")
    error_message = "DNS_NAME must end with tailing dot, for example \"knfsd.example.\" (note the tailing dot)."
  }
}

variable "SERVICE_LABEL" {
  default = "dns"
  type    = string
}

variable "SERVICE_ACCOUNT" {
  default = ""
  type    = string
}

variable "NCONNECT_VALUE" {
  default = "16"
  type    = string
}

variable "ACDIRMIN" {
  default = 600
  type    = number
}

variable "ACDIRMAX" {
  default = 600
  type    = number
}

variable "ACREGMIN" {
  default = 600
  type    = number
}

variable "ACREGMAX" {
  default = 600
  type    = number
}

variable "RSIZE" {
  default = 1048576
  type    = number
}

variable "WSIZE" {
  default = 1048576
  type    = number
}

variable "NOHIDE" {
  default = true
  type    = bool
}

variable "MOUNT_OPTIONS" {
  default = ""
  type    = string
}

variable "EXPORT_OPTIONS" {
  default = ""
  type    = string
}


variable "VFS_CACHE_PRESSURE" {
  default = "100"
  type    = string
}

variable "READ_AHEAD" {
  default = 8388608
  type    = number
}

variable "ENABLE_UDP" {
  default = false
  type    = bool
}

variable "ENABLE_AUTOHEALING_HEALTHCHECKS" {
  default = true
  type    = bool
}

variable "HEALTHCHECK_INITIAL_DELAY_SECONDS" {
  type    = number
  default = 600
}

variable "HEALTHCHECK_INTERVAL_SECONDS" {
  type    = number
  default = 60
}

variable "HEALTHCHECK_TIMEOUT_SECONDS" {
  type    = number
  default = 2
}

variable "HEALTHCHECK_HEALTHY_THRESHOLD" {
  type    = number
  default = 3
}

variable "HEALTHCHECK_UNHEALTHY_THRESHOLD" {
  type    = number
  default = 3
}

variable "NUM_NFS_THREADS" {
  default = 512
  type    = number
}

variable "ENABLE_STACKDRIVER_METRICS" {
  default = true
  type    = bool
}

variable "METRICS_AGENT_CONFIG" {
  default = ""
  type    = string
}

variable "ROUTE_METRICS_PRIVATE_GOOGLEAPIS" {
  default = false
  type    = bool
}

variable "ENABLE_KNFSD_AUTOSCALING" {
  default = false
  type    = bool
}

variable "KNFSD_AUTOSCALING_MIN_INSTANCES" {
  default = 1
  type    = number
}

variable "KNFSD_AUTOSCALING_MAX_INSTANCES" {
  default = 10
  type    = number
}

variable "KNFSD_AUTOSCALING_NFS_CONNECTIONS_THRESHOLD" {
  default = 250
  type    = number
}

variable "CUSTOM_PRE_STARTUP_SCRIPT" {
  default = "echo 'Running default pre startup script. No action taken.'"
  type    = string
}

variable "CUSTOM_POST_STARTUP_SCRIPT" {
  default = "echo 'Running default post startup script. No action taken.'"
  type    = string
}

variable "MACHINE_TYPE" {
  default = "n1-highmem-16"
  type    = string
}

variable "MIG_MINIMAL_ACTION" {
  default = "RESTART"
  type    = string
}

variable "MIG_MAX_UNAVAILABLE_PERCENT" {
  default = "100"
  type    = number
}

variable "MIG_REPLACEMENT_METHOD" {
  default = ""
  type    = string
}

variable "ENABLE_KNFSD_AGENT" {
  default = true
  type    = bool
}

variable "DISABLED_NFS_VERSIONS" {
  default = "4.0,4.1,4.2"
  type    = string
}

variable "ENABLE_NETAPP_AUTO_DETECT" {
  default = false
  type    = bool
}

variable "NETAPP_HOST" {
  type    = string
  default = ""
}

variable "NETAPP_URL" {
  type    = string
  default = ""
}

variable "NETAPP_USER" {
  type    = string
  default = ""
}

variable "NETAPP_SECRET" {
  type    = string
  default = ""
}

variable "NETAPP_SECRET_PROJECT" {
  type    = string
  default = ""
}

variable "NETAPP_SECRET_VERSION" {
  type    = string
  default = ""
}

variable "NETAPP_CA" {
  type    = string
  default = ""
}

variable "NETAPP_ALLOW_COMMON_NAME" {
  type    = bool
  default = false
}

variable "CACHEFILESD_DISK_TYPE" {
  type    = string
  default = "local-ssd"

  validation {
    condition     = contains(["local-ssd", "pd-ssd", "pd-balanced", "pd-standard"], var.CACHEFILESD_DISK_TYPE)
    error_message = "Valid values for CACHEFILESD_DISK_TYPE are 'local-ssd', 'pd-ssd', 'pd-balanced, 'pd-standard'."
  }
}

variable "LOCAL_SSDS" {
  default = 4
  type    = number

  validation {
    condition     = contains([0, 1, 2, 3, 4, 5, 6, 7, 8, 16, 24], var.LOCAL_SSDS)
    error_message = "Valid values for LOCAL_SSDS are 0-8, 16 or 24."
  }
}

variable "CACHEFILESD_PERSISTENT_DISK_SIZE_GB" {
  type    = number
  default = 1500

  validation {
    condition     = var.CACHEFILESD_PERSISTENT_DISK_SIZE_GB >= 10 && var.CACHEFILESD_PERSISTENT_DISK_SIZE_GB <= 64000
    error_message = "CACHEFILESD_PERSISTENT_DISK_SIZE_GB must be between 10 and 6400."
  }
}

variable "NFS_MOUNT_VERSION" {
  type    = string
  default = "3"

  validation {
    condition     = contains(["3", "4", "4.0", "4.1", "4.2"], var.NFS_MOUNT_VERSION)
    error_message = "Valid values for NFS_MOUNT_VERSION are '3', '4', '4.0', '4.1', '4.2'."
  }


}

variable "ENABLE_HIGH_BANDWIDTH_CONFIGURATION" {
  type    = bool
  default = false
}
