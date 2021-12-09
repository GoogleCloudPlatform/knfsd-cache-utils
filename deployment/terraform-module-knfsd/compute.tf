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

# Instance Template for the KNFSD nodes
resource "google_compute_instance_template" "nfsproxy-template" {
  project          = var.PROJECT
  region           = var.REGION
  name_prefix      = var.PROXY_BASENAME
  machine_type     = var.MACHINE_TYPE
  min_cpu_platform = "Intel Skylake"
  can_ip_forward   = false
  tags             = ["knfsd-cache-server"]

  lifecycle {
    create_before_destroy = true
  }

  disk {
    source_image = var.PROXY_IMAGENAME
    auto_delete  = true
    boot         = true
    disk_size_gb = "100"
  }

  dynamic "disk" {
    for_each = range(1, var.LOCAL_SSDS + 1)
    content {
      interface    = "NVME"
      disk_type    = "local-ssd"
      type         = "SCRATCH"
      mode         = "READ_WRITE"
      device_name  = "local-ssd-${disk.value}"
      disk_size_gb = 375
    }
  }


  network_interface {
    network            = var.NETWORK
    subnetwork         = var.SUBNETWORK
    subnetwork_project = var.PROJECT
  }

  metadata = {
    # mounts
    EXPORT_MAP              = var.EXPORT_MAP
    EXPORT_HOST_AUTO_DETECT = var.EXPORT_HOST_AUTO_DETECT
    DISCO_MOUNT_EXPORT_MAP  = var.DISCO_MOUNT_EXPORT_MAP
    EXCLUDED_EXPORTS        = var.EXCLUDED_EXPORTS
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
    NCONNECT       = var.NCONNECT_VALUE
    ACDIRMIN       = var.ACDIRMIN
    ACDIRMAX       = var.ACDIRMAX
    ACREGMIN       = var.ACREGMIN
    ACREGMAX       = var.ACREGMAX
    RSIZE          = var.RSIZE
    WSIZE          = var.WSIZE
    MOUNT_OPTIONS  = var.MOUNT_OPTIONS
    EXPORT_OPTIONS = var.EXPORT_OPTIONS

    # system
    NFS_KERNEL_SERVER_CONF = file("${path.module}/resources/nfs-kernel-server-conf")
    NUM_NFS_THREADS        = var.NUM_NFS_THREADS
    VFS_CACHE_PRESSURE     = var.VFS_CACHE_PRESSURE
    DISABLED_NFS_VERSIONS  = var.DISABLED_NFS_VERSIONS
    READ_AHEAD_KB          = floor(var.READ_AHEAD / 1024)
    LOADBALANCER_IP        = google_compute_address.nfsproxy_static.address
    serial-port-enable     = "TRUE"

    # metrics
    ENABLE_STACKDRIVER_METRICS  = var.ENABLE_STACKDRIVER_METRICS
    COLLECTD_METRICS_CONFIG     = file("${path.module}/resources/monitoring/knfsd.conf")
    COLLECTD_METRICS_SCRIPT     = file("${path.module}/resources/monitoring/knfsd.sh")
    COLLECTD_ROOT_EXPORT_SCRIPT = file("${path.module}/resources/monitoring/export-root.sh")

    # scripts / software
    startup-script             = file("${path.module}/resources/proxy-startup.sh")
    CUSTOM_PRE_STARTUP_SCRIPT  = var.CUSTOM_PRE_STARTUP_SCRIPT
    CUSTOM_POST_STARTUP_SCRIPT = var.CUSTOM_POST_STARTUP_SCRIPT
    ENABLE_KNFSD_AGENT         = var.ENABLE_KNFSD_AGENT
  }

  labels = {
    vm-type = "nfs-proxy"
  }

  scheduling {
    automatic_restart   = true
    on_host_maintenance = "MIGRATE"
    preemptible         = false
  }

  # We use a dnaymic block for service_account here as we only want to assign an SA if we have metrics enabled.
  # If we do not have metrics enabled there is no need for an SA
  dynamic "service_account" {
    for_each = local.enable_service_account ? [1] : []
    content {
      email  = var.SERVICE_ACCOUNT
      scopes = local.scopes
    }
  }
}

# Healthcheck on port 2049, used for monitoring the NFS Health Status
resource "google_compute_health_check" "autohealing" {
  project             = var.PROJECT
  name                = "${var.PROXY_BASENAME}-autohealing-health-check"
  check_interval_sec  = 5
  timeout_sec         = 2
  healthy_threshold   = 3
  unhealthy_threshold = 3

  tcp_health_check {
    port = "2049"
  }

}

# Instance Group Manager for the Knfsd Nodes
resource "google_compute_instance_group_manager" "proxy-group" {
  project            = var.PROJECT
  name               = "${var.PROXY_BASENAME}-group"
  base_instance_name = var.PROXY_BASENAME
  zone               = var.ZONE
  // Set the Target Size to null if autoscaling is enabled
  target_size = (var.ENABLE_KNFSD_AUTOSCALING == true ? null : var.KNFSD_NODES)

  update_policy {
    type                    = "PROACTIVE"
    minimal_action          = var.MIG_MINIMAL_ACTION
    max_unavailable_percent = var.MIG_MAX_UNAVAILABLE_PERCENT
    replacement_method      = var.MIG_REPLACEMENT_METHOD
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
      initial_delay_sec = 600
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

# Load Balancer backend service for the Knfsd Cluster
resource "google_compute_region_backend_service" "nfsproxy" {
  project               = var.PROJECT
  region                = var.REGION
  name                  = "${var.PROXY_BASENAME}-backend-service"
  health_checks         = [google_compute_health_check.autohealing.self_link]
  load_balancing_scheme = "INTERNAL"
  session_affinity      = "CLIENT_IP"
  protocol              = "TCP"
  timeout_sec           = 10
  backend {
    description = "Load Balancer backend for nfsProxy managed instance group."
    group       = google_compute_instance_group_manager.proxy-group.instance_group
  }
}

# Load Balancer backend service for the Knfsd Cluster
resource "google_compute_region_backend_service" "nfsproxy_udp" {
  count                 = var.ENABLE_UDP ? 1 : 0
  project               = var.PROJECT
  region                = var.REGION
  name                  = "${var.PROXY_BASENAME}-backend-service-udp"
  health_checks         = [google_compute_health_check.autohealing.self_link]
  load_balancing_scheme = "INTERNAL"
  session_affinity      = "CLIENT_IP"
  protocol              = "UDP"
  timeout_sec           = 10
  backend {
    description = "Load Balancer backend for NFS proxy managed instance group."
    group       = google_compute_instance_group_manager.proxy-group.instance_group
  }
}

# Load Balancer forwarding rule service for the Knfsd Cluster
resource "google_compute_forwarding_rule" "default" {
  project               = var.PROJECT
  region                = var.REGION
  name                  = var.PROXY_BASENAME
  load_balancing_scheme = "INTERNAL"
  backend_service       = google_compute_region_backend_service.nfsproxy.self_link
  ip_protocol           = "TCP"
  ip_address            = google_compute_address.nfsproxy_static.address
  all_ports             = true
  network               = var.NETWORK
  subnetwork            = var.SUBNETWORK
  service_label         = var.SERVICE_LABEL
}

resource "google_compute_forwarding_rule" "udp" {
  count                 = var.ENABLE_UDP ? 1 : 0
  project               = var.PROJECT
  region                = var.REGION
  name                  = "${var.PROXY_BASENAME}-udp"
  load_balancing_scheme = "INTERNAL"
  backend_service       = google_compute_region_backend_service.nfsproxy_udp[0].self_link
  ip_protocol           = "UDP"
  ip_address            = google_compute_address.nfsproxy_static.address
  all_ports             = true
  network               = var.NETWORK
  subnetwork            = var.SUBNETWORK
  service_label         = var.SERVICE_LABEL
}
