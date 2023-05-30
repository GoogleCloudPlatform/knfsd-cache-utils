# Metrics

These deployment scripts can optionally configure the exporting a range of metrics from each Knfsd node into [Google Cloud Operations](https://cloud.google.com/products/operations) (formerly Stackdriver). These are exported via a combination of the [Google Cloud Ops Agent](https://cloud.google.com/monitoring/agent/ops-agent) and the [knfsd-metrics-agent](../image/resources/knfsd-metrics-agent/README.md) which are both installed as part of the [build scripts](/image).

These metrics can be enabled via the `ENABLE_STACKDRIVER_METRICS` variable. **If you wish to use auto-scaling then metrics must be enabled**.

## Metrics Prerequisites

The following additional prerequisites must be met if you wish to enable metrics:

| Prerequisite                                                                                                             | Details                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| ------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| [Metric Descriptors and Dashboard Import](metrics)                                                                       | If this is the first time you are deploying Knfsd in a Google Cloud Project you need to setup the Metric Descriptors and import the Knfsd Monitoring Dashboard. This is achieved via a standalone Terraform configuration and the process is described in the [metrics](metrics) directory.                                                                                                                                                                                                                                                                                                                                          |
| [Private Google Access](https://cloud.google.com/vpc/docs/configure-private-google-access)                               | You must have [Private Google Access](https://cloud.google.com/vpc/docs/configure-private-google-access) enabled on the subnet that you will be using for the Knfsd nodes. This is required to allow connectivity to the Monitoring API for VM's without a Public IP. You should also ensure you have the default `0.0.0.0/0` route configured and pointing to the default internet gateway with appropriate firewall rules to allow outbound connectivity to the Google Cloud API's. You can optionally force routing over the `private.googleapis.com` range by setting the `ROUTE_METRICS_PRIVATE_GOOGLEAPIS` variable to `true`. |
| [Service Account Permissions](https://cloud.google.com/compute/docs/access/service-accounts#service_account_permissions) | A Service Account needs to be configured for the Knfsd Nodes with the `logging-write` and `monitoring-write` scopes. This is performed automatically by the Terraform Module when you have metrics enabled. By default, the [Compute Engine Default Service Account](https://cloud.google.com/compute/docs/access/service-accounts#default_service_account) will be used.                                                                                                                                                                                                                                                            |

## Exported Metrics

The following custom metrics are exported currently:

| Metric Name                                                    | Description                                                                                                     |
| -------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------- |
| **custom.googleapis.com/knfsd/nfs_connections**                | The number of NFS Clients connected to the Knfsd filer (used for autoscaling).                                  |
| **custom.googleapis.com/knfsd/nfs_inode_cache_active_objects** | The number of active objects in the Linux NFS inode Cache.                                                      |
| **custom.googleapis.com/knfsd/dentry_cache_active_objects**    | The number of active objects in the Linux Dentry Cache.                                                         |
| **custom.googleapis.com/knfsd/nfs_inode_cache_objsize**        | The total size of the objects in the Linux NFS inode Cache in bytes.                                            |
| **custom.googleapis.com/knfsd/dentry_cache_objsize**           | The total size of the objects in the Linux Dentry Cache in bytes.                                               |
| **custom.googleapis.com/knfsd/nfsiostat_mount_read_exe**       | The average read operation EXE per NFS client mount over the past 60 seconds (Knfsd --> Source Filer).          |
| **custom.googleapis.com/knfsd/nfsiostat_mount_read_rtt**       | The average read operation RTT per NFS client mount over the past 60 seconds (Knfsd --> Source Filer).          |
| **custom.googleapis.com/knfsd/nfsiostat_mount_write_exe**      | The average write operation EXE per NFS client mount over the past 60 seconds (Knfsd --> Source Filer).         |
| **custom.googleapis.com/knfsd/nfsiostat_mount_write_rtt**      | The average write operation RTT per NFS client mount over the past 60 seconds (Knfsd --> Source Filer)..        |
| **custom.googleapis.com/knfsd/nfsiostat_ops_per_second**       | The number of NFS operations per second per NFS client mount over the past 60 seconds (Knfsd --> Source Filer). |
| **custom.googleapis.com/knfsd/nfsiostat_rpc_backlog**          | The RPC Backlog per NFS client mount over the past 60 seconds (Knfsd --> Source Filer).                         |
| **custom.googleapis.com/knfsd/mount/read_bytes**               | The total number of bytes read from the source NFS server.                                                      |
| **custom.googleapis.com/knfsd/mount/write_bytes**              | The total number of bytes wrote to the source NFS server.                                                       |
| **custom.googleapis.com/knfsd/mount/operation/requests**       | The total number of NFS requests sent to the source NFS server.                                                 |
| **custom.googleapis.com/knfsd/mount/operation/sent_bytes**     | The total number of bytes sent to the source NFS server. This includes the RPC protocol headers.                |
| **custom.googleapis.com/knfsd/mount/operation/received_bytes** | The total number of bytes received from the source NFS server. This includes the RPC protocol headers.          |
| **custom.googleapis.com/knfsd/mount/operation/major_timeouts** | The total number of RPC major timeouts (`timeo`, default 60 seconds) between the proxy and source NFS servers.  |
| **custom.googleapis.com/knfsd/mount/operation/errors**         | The total number of RPC errors between the proxy and the source NFS servers.                                    |
| **custom.googleapis.com/knfsd/exports/total_operations**       | The total number of NFS operations received from NFS clients.                                                   |
| **custom.googleapis.com/knfsd/exports/total_read_bytes**       | The total number of bytes read by NFS clients.                                                                  |
| **custom.googleapis.com/knfsd/exports/total_write_bytes**      | The total number of bytes wrote by NFS clients.                                                                 |
| **custom.googleapis.com/knfsd/fscache_oldest_file**            | The age of the oldest file in FS-Cache. This metric is not enabled by default.                                  |

## Dashboards

The Knfsd Monitoring Dashboard is created automatically by the metrics initialisation Terraform that is detailed in the [Metrics Prerequisites](#metricsprerequisites).

Once ran, you can then access the dashboard from [https://console.cloud.google.com/monitoring/dashboards/](https://console.cloud.google.com/monitoring/dashboards/)

## Custom Configuration

The metrics can be configured using the `METRICS_AGENT_CONFIG` variable in the Terraform module, or by customizing the metrics config when building the image.

Configuring the metrics using Terraform is the simplest option. You can provide the metrics configuration using a file or directly inline using heredoc.

Providing the metrics config from a file:

```terraform
module "nfs_proxy" {
  source = "github.com/GoogleCloudPlatform/knfsd-cache-utils//deployment/terraform-module-knfsd?ref=v0.11.0"

  METRICS_AGENT_CONFIG = file("metrics-config.yaml")
}
```

Providing the metrics config inline using heredoc syntax:

```terraform
module "nfs_proxy" {
  source = "github.com/GoogleCloudPlatform/knfsd-cache-utils//deployment/terraform-module-knfsd?ref=v0.11.0"

  METRICS_AGENT_CONFIG = <<-EOT
    receivers:
      mounts:
        collection_interval: 5m
  EOT
}
```

See the [knfsd-metrics-agent README](../image/resources/knfsd-metrics-agent/README.md) for details how to configure the metrics agent.
