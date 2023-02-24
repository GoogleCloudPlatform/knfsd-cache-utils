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

locals {
  FSID_DATABASE_CONFIG = (
    # this module deployed an external fsid database, so generate our own config
    local.deploy_fsid_database ? templatefile("${path.module}/resources/knfsd-fsidd.conf.tftpl", {
      connection_name    = module.fsid_database.0.connection_name,
      sql_user           = module.fsid_database.0.sql_user,
      database_name      = module.fsid_database.0.database_name,
      private_ip_address = module.fsid_database.0.private_ip_address,
      enable_metrics     = var.ENABLE_STACKDRIVER_METRICS,
    }) :

    # otherwise use an external config
    var.FSID_DATABASE_CONFIG
  )
}

# Instance Template for the KNFSD nodes
resource "google_compute_instance_template" "nfsproxy-template" {

  provider = google-beta # Required due to network_performance_config being in beta provider only

  project          = var.PROJECT
  region           = var.REGION
  name_prefix      = var.PROXY_BASENAME
  machine_type     = var.MACHINE_TYPE
  min_cpu_platform = lower(split(var.MACHINE_TYPE, "-")[0]) == "n1" ? "Intel Skylake" : null // Set Skylake only if N1
  can_ip_forward   = false
  tags             = ["knfsd-cache-server"]
  labels           = var.PROXY_LABELS

  disk {
    source_image = var.PROXY_IMAGENAME
    auto_delete  = true
    boot         = true
    disk_size_gb = 100
  }

  # Configuration for Persistent Disk for FS-Cache directory
  dynamic "disk" {

    for_each = var.CACHEFILESD_DISK_TYPE != "local-ssd" ? [1] : []
    content {
      disk_type    = var.CACHEFILESD_DISK_TYPE
      type         = "PERSISTENT"
      mode         = "READ_WRITE"
      device_name  = "pd-fscache"
      disk_size_gb = var.CACHEFILESD_PERSISTENT_DISK_SIZE_GB
    }
  }

  # Configuration for Local SSDs for FS-Cache directory
  dynamic "disk" {
    for_each = var.CACHEFILESD_DISK_TYPE == "local-ssd" ? range(1, var.LOCAL_SSDS + 1) : []
    content {
      interface    = "NVME"
      disk_type    = "local-ssd"
      type         = "SCRATCH"
      mode         = "READ_WRITE"
      device_name  = "local-ssd-${disk.value}"
      disk_size_gb = 375
    }
  }

  network_performance_config {
    total_egress_bandwidth_tier = var.ENABLE_HIGH_BANDWIDTH_CONFIGURATION ? "TIER_1" : "DEFAULT"
  }

  network_interface {
    network            = var.NETWORK
    subnetwork         = var.SUBNETWORK
    subnetwork_project = var.SUBNETWORK_PROJECT != "" ? var.SUBNETWORK_PROJECT : null
    nic_type           = var.ENABLE_HIGH_BANDWIDTH_CONFIGURATION ? "GVNIC" : "VIRTIO_NET"
  }

  metadata = {
    # mounts
    EXPORT_MAP              = var.EXPORT_MAP
    EXPORT_HOST_AUTO_DETECT = var.EXPORT_HOST_AUTO_DETECT
    EXCLUDED_EXPORTS        = join("\n", var.EXCLUDED_EXPORTS)
    INCLUDED_EXPORTS        = join("\n", var.INCLUDED_EXPORTS)
    EXPORT_CIDR             = var.EXPORT_CIDR

    # NetApp auto-discovery
    ENABLE_NETAPP_AUTO_DETECT = var.ENABLE_NETAPP_AUTO_DETECT
    NETAPP_HOST               = var.NETAPP_HOST
    NETAPP_URL                = var.NETAPP_URL
    NETAPP_USER               = var.NETAPP_USER
    NETAPP_SECRET             = var.NETAPP_SECRET
    NETAPP_SECRET_PROJECT     = var.NETAPP_SECRET_PROJECT
    NETAPP_SECRET_VERSION     = var.NETAPP_SECRET_VERSION
    NETAPP_CA                 = var.NETAPP_CA
    NETAPP_ALLOW_COMMON_NAME  = var.NETAPP_ALLOW_COMMON_NAME

    # mount options
    NCONNECT          = var.NCONNECT_VALUE
    ACDIRMIN          = var.ACDIRMIN
    ACDIRMAX          = var.ACDIRMAX
    ACREGMIN          = var.ACREGMIN
    ACREGMAX          = var.ACREGMAX
    RSIZE             = var.RSIZE
    WSIZE             = var.WSIZE
    NOHIDE            = var.NOHIDE
    MOUNT_OPTIONS     = var.MOUNT_OPTIONS
    EXPORT_OPTIONS    = var.EXPORT_OPTIONS
    NFS_MOUNT_VERSION = var.NFS_MOUNT_VERSION

    # auto-reexport nested mounts
    AUTO_REEXPORT        = var.AUTO_REEXPORT
    FSID_MODE            = var.FSID_MODE
    FSID_DATABASE_CONFIG = local.FSID_DATABASE_CONFIG

    # system
    NFS_KERNEL_SERVER_CONF = file("${path.module}/resources/nfs-kernel-server.conf")
    NUM_NFS_THREADS        = var.NUM_NFS_THREADS
    VFS_CACHE_PRESSURE     = var.VFS_CACHE_PRESSURE
    DISABLED_NFS_VERSIONS  = var.DISABLED_NFS_VERSIONS
    READ_AHEAD_KB          = floor(var.READ_AHEAD / 1024)
    LOADBALANCER_IP        = one(google_compute_address.nfsproxy_static.*.address)
    serial-port-enable     = "TRUE"

    # metrics
    ENABLE_STACKDRIVER_METRICS       = var.ENABLE_STACKDRIVER_METRICS
    METRICS_AGENT_CONFIG             = var.METRICS_AGENT_CONFIG
    ROUTE_METRICS_PRIVATE_GOOGLEAPIS = var.ROUTE_METRICS_PRIVATE_GOOGLEAPIS

    # scripts / software
    startup-script             = file("${path.module}/resources/proxy-startup.sh")
    CUSTOM_PRE_STARTUP_SCRIPT  = var.CUSTOM_PRE_STARTUP_SCRIPT
    CUSTOM_POST_STARTUP_SCRIPT = var.CUSTOM_POST_STARTUP_SCRIPT
    ENABLE_KNFSD_AGENT         = var.ENABLE_KNFSD_AGENT
  }

  scheduling {
    automatic_restart   = true
    on_host_maintenance = "MIGRATE"
    preemptible         = false
  }

  # We use a dynamic block for service_account here as we only want to assign an SA if we have metrics enabled.
  # If we do not have metrics enabled there is no need for an SA
  dynamic "service_account" {
    for_each = local.enable_service_account ? [1] : []
    content {
      email  = var.SERVICE_ACCOUNT
      scopes = local.scopes
    }
  }

  lifecycle {
    create_before_destroy = true

    # Most of these preconditions are conditional. Terraform 1.20 does not have
    # support for some kind of "enabled" or "when" condition, so write the
    # conditions in the form:
    #   condition = (when ? check : true)
    #
    # This roughly translates to
    #   if when then
    #     check precondition
    #   else
    #     skip precondition

    precondition {
      # AUTO_REEXPORT requires an fsid service to be enabled
      condition = (
        var.AUTO_REEXPORT
        ? contains(["local", "external"], var.FSID_MODE)
        : true
      )
      error_message = "FSID_MODE must be either \"local\" or \"external\" when AUTO_REEXPORT is enabled."
    }

    # If this module has created the database ensure that the user did not try
    # and provide their own FSID_DATABASE_CONFIG, as that configuration will be
    # ignored.
    # This just avoids two possible errors:
    #   * This module ignores the custom configuration leading to confusion.
    #   * This module uses the custom configuration leading to an unused Cloud
    #     SQL being deployed.
    precondition {
      condition = (
        var.FSID_MODE == "external" && var.FSID_DATABASE_DEPLOY
        ? var.FSID_DATABASE_CONFIG == ""
        : true
      )
      error_message = "Can only provide a custom FSID_DATABASE_CONFIG when using a custom external fsid database (FSID_MODE = \"external\" and FSID_DATABASE_DEPLOY = false)."
    }

    # Again to avoid confusion, if not using an external database do not allow
    # the user to provide their own FSID_DATABASE_CONFIG.
    precondition {
      condition = (
        var.FSID_MODE != "external"
        ? var.FSID_DATABASE_CONFIG == ""
        : true
      )
      error_message = "Can only provide a custom FSID_DATABASE_CONFIG when using a custom external fsid database (FSID_MODE = \"external\" and FSID_DATABASE_DEPLOY = false)."
    }

    # When using an custom external database then FSID_DATABASE_DEPLOY must be provided
    precondition {
      condition = (
        var.FSID_MODE == "external" && !var.FSID_DATABASE_DEPLOY
        ? var.FSID_DATABASE_CONFIG != ""
        : true
      )
      error_message = "Must specify a database configuration (FSID_DATABASE_CONFIG) when using a custom external database (FSID_MODE = \"external\" and FISD_DATABASE_DEPLOY = false)."
    }

    # Bug check: This should not occur and indicates a bug in the Terraform script.
    # Fail early during terraform plan, otherwise the proxy will deploy and enter
    # a reboot loop.
    precondition {
      condition = (
        local.deploy_fsid_database
        ? local.FSID_DATABASE_CONFIG != ""
        : true
      )
      error_message = "BUG: database configuration not set for external fsid database."
    }

    # Bug check: This should not occur and indicates a bug in the Terraform script.
    # Check that if the script deployed a Cloud SQL database, that database will
    # be used by the proxy otherwise its just wasting money.
    # Including this here as this is the likely place people will update when
    # changing how FSID_MODE is handled as this is the main proxy validation.
    # This acts as a cross check for the local.deploy_fsid_database logic in
    # case only one is updated.
    precondition {
      condition = (
        local.deploy_fsid_database
        ? var.FSID_MODE == "external"
        : true
      )
      error_message = "BUG: deployed Cloud SQL database, but that database is not in use."
    }
  }
}

# Healthcheck on port 2049, used for monitoring the NFS Health Status
resource "google_compute_health_check" "autohealing" {
  project             = var.PROJECT
  name                = "${var.PROXY_BASENAME}-autohealing-health-check"
  check_interval_sec  = var.HEALTHCHECK_INTERVAL_SECONDS
  timeout_sec         = var.HEALTHCHECK_TIMEOUT_SECONDS
  healthy_threshold   = var.HEALTHCHECK_HEALTHY_THRESHOLD
  unhealthy_threshold = var.HEALTHCHECK_UNHEALTHY_THRESHOLD

  tcp_health_check {
    port = "2049"
  }

  depends_on = [
    # Ensure that the firewall rules are not deleted while the health check
    # still exists. Otherwise when removing clusters, Terraform may delete the
    # firewall rule causing the proxy group to start replacing instances.
    # Terraform will then get stuck waiting for the instance group to complete
    # the changes before removing the instance group.
    google_compute_firewall.allow-tcp-healthcheck
  ]
}

# Instance Group Manager for the Knfsd Nodes
resource "google_compute_instance_group_manager" "proxy-group" {
  provider = google-beta # required to support stateful_internal_ip

  project            = var.PROJECT
  name               = "${var.PROXY_BASENAME}-group"
  base_instance_name = var.PROXY_BASENAME
  zone               = var.ZONE
  // Set the Target Size to null if autoscaling is enabled
  target_size = (var.ENABLE_KNFSD_AUTOSCALING == true ? null : var.KNFSD_NODES)

  # when using static IPs, wait for all the instances to be updated so that the
  # IPs of the Compute Instances can be fetched using the instance_ips module.
  wait_for_instances        = var.ASSIGN_STATIC_IPS
  wait_for_instances_status = "UPDATED"

  update_policy {
    type                    = "PROACTIVE"
    minimal_action          = var.MIG_MINIMAL_ACTION
    max_unavailable_percent = var.MIG_MAX_UNAVAILABLE_PERCENT
    replacement_method      = coalesce(var.MIG_REPLACEMENT_METHOD, local.MIG_REPLACEMENT_METHOD_DEFAULT)
  }

  version {
    name              = "v1"
    instance_template = google_compute_instance_template.nfsproxy-template.self_link
  }

  # We use a dynamic block for auto_healing_policies here as we only want to assign a healthcheck if the ENABLE_AUTOHEALING_HEALTHCHECKS is set
  dynamic "auto_healing_policies" {
    for_each = var.ENABLE_AUTOHEALING_HEALTHCHECKS ? [1] : []
    content {
      health_check      = google_compute_health_check.autohealing.self_link
      initial_delay_sec = var.HEALTHCHECK_INITIAL_DELAY_SECONDS
    }
  }

  dynamic "stateful_internal_ip" {
    for_each = toset(var.ASSIGN_STATIC_IPS ? ["nic0"] : [])
    content {
      interface_name = stateful_internal_ip.value
      delete_rule    = "ON_PERMANENT_INSTANCE_DELETION"
    }
  }
}

# Firewall rule to allow healthchecks from the GCP Healthcheck ranges
resource "google_compute_firewall" "allow-tcp-healthcheck" {

  // Count is used here to determine if the firewall rules should automatically be created.
  // If var.AUTO_CREATE_FIREWALL_RULES is true then we want 1 firewall rule, else 0
  count = var.AUTO_CREATE_FIREWALL_RULES ? 1 : 0

  project  = var.PROJECT
  name     = "allow-nfs-tcp-healthcheck"
  network  = var.NETWORK
  priority = 1000

  allow {
    protocol = "tcp"
    ports    = ["2049"]
  }
  source_ranges = ["130.211.0.0/22", "35.191.0.0/16", "209.85.152.0/22", "209.85.204.0/22"]
  target_tags   = ["knfsd-cache-server"]

}
