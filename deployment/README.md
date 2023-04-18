# Knfsd Deployment

This directory contains a [Terraform Module](https://www.terraform.io/docs/modules/index.html) for deploying Knfsd on Google Cloud.

The `main` branch may be updated at any time with the latest changes which could be breaking. You should always configure your module to use a release. This can be configured in the modules Terraform Configuration block.

```
source = "github.com/GoogleCloudPlatform/knfsd-cache-utils//deployment/terraform-module-knfsd?ref=v1.0.0-beta4"
```

## Prerequisites

Before continuing with the deployment and configuration of Knfsd you should review the deployment [prerequisites](prerequisites.md).

## Features / Special Configurations

There are a number of optional features and special configurations that can be enabled and configured for Knfsd. If you are planning on using any of these features/configurations then please review the appropriate documentation section.

* [Metrics](metrics.md) - System and proxy metrics for monitoring and observing Knfsd
* [Autoscaling](autoscaling.md) - Automatic scale up of Knfsd in response to the number of connected clients
* [Agent](agent.md) - A lightweight API that provides information on Knfsd nodes
* [Fanout Architecture](fanout.md) - Documentation on how to deploy Knfsd in the fanout approach

## Usage

Basic usage of this module is as follows:

```terraform
module "nfs_proxy" {

    source = "github.com/GoogleCloudPlatform/knfsd-cache-utils//deployment/terraform-module-knfsd?ref=v1.0.0-beta4"

    # Google Cloud Project Configuration
    PROJECT                        = "my-gcp-project"
    REGION                         = "us-west1"
    ZONE                           = "us-west1-a"

    # Network Configuration
    NETWORK                        = "my-vpc"
    SUBNETWORK                     = "my-subnet"
    AUTO_CREATE_FIREWALL_RULES     = false
    TRAFFIC_DISTRIBUTION_MODE      = "dns_round_robin"
    DNS_NAME                       = "rendercluster1.gcp.example.com."
    ASSIGN_STATIC_IPS              = true

    # Knfsd Proxy Configuration
    PROXY_IMAGENAME                = "knfsd-base-image"
    EXPORT_MAP                     = "10.0.5.5;/remoteexport;/remoteexport"
    PROXY_BASENAME                 = "rendercluster1"
    KNFSD_NODES                    = 3

}

# Prints the DNS name of KNFSD proxy
output "dns_name" {
    value = module.nfs_proxy.dns_name
}
```

Edit the above [configuration variables](#Configuration-Variables) to match your desired configuration.

### Deploy Knfsd

Once you have created your `deploy.tf` and created you can deploy Knfsd with:

```sh
terraform init
terraform apply
```

## Configuration Variables

### Google Cloud Project Configuration

| Variable | Description                                                                                                                                                                     | Required | Default |
| -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- | ------- |
| PROJECT  | The Google Cloud Project that the Knfsd Cluster is being deployed to. If it is not provided, the provider project is used.                                                      | False    | N/A     |
| REGION   | The [Google Cloud Region](https://cloud.google.com/compute/docs/regions-zones) to use for deployment of regional resources. If it is not provided, the provider region is used. | False    | N/A     |
| ZONE     | The [Google Cloud Zone](https://cloud.google.com/compute/docs/regions-zones) to use for deployment of zonal resources. If it is not provided, the provider zone is used.        | False    | N/A     |

### Network Configuration

| Variable                            | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        | Required | Default                              |
| ----------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | -------- | ------------------------------------ |
| NETWORK                             | The network name (VPC) to use for the deployment of the Knfsd Compute Engine Instances.                                                                                                                                                                                                                                                                                                                                                                                                                            | False    | `default`                            |
| SUBNETWORK                          | The subnetwork name (subnet) to use for the deployment of the Knfsd Compute Engine Instances.                                                                                                                                                                                                                                                                                                                                                                                                                      | False    | `default`                            |
| SUBNETWORK_PROJECT                  | The project that the subnetwork exists in. This only needs to be set if using a Shared VPC, where the subnetwork exists in a different project. Otherwise it defaults to the provider project.                                                                                                                                                                                                                                                                                                                     | False    | null                                 |
| AUTO_CREATE_FIREWALL_RULES          | Should firewall rules automatically be created to allow [healthcheck connectivity](https://cloud.google.com/load-balancing/docs/health-check-concepts#ip-ranges)?                                                                                                                                                                                                                                                                                                                                                  | False    | `true`                               |
| TRAFFIC_DISTRIBUTION_MODE           | The [client traffic distribution mode](./traffic-distribution.md) used to distribute traffic between proxy instances in the KNFSD proxy cluster. Can be either `dns_round_robin`, `loadbalancer`, or `none`. The recommended option is `dns_round_robin`. If using `none` you will need to provide your own solution to handle traffic distribution.                                                                                                                                                               | True     |                                      |
| LOADBALANCER_IP                     | The IP address to use for the Internal Load Balancer when `TRAFFIC_DISTRIBUTION_MODE = "loadbalancer"`. If not specified, a random IP address will be assigned within the subnet.                                                                                                                                                                                                                                                                                                                                  | False    | null                                 |
| ENABLE_UDP                          | Create a load balancer to support UDP traffic to the NFS proxy instances (when `TRAFFIC_DISTRIBUTION_MODE = "loadbalancer"`). UDP is not recommended for the main NFS traffic as it can cause data corruption. However, this maybe useful for older clients that default to using UDP for the mount protocol.                                                                                                                                                                                                      | False    | `false`                              |
| DNS_NAME                            | The fully qualified DNS name (FQDN) to use for the KNFSD proxy cluster when `TRAFFIC_DISTRIBUTION_MODE = "round_robin_dns"`. This must end with a period, such as `rendercluster1.gcp.example.com.` or `rendercluster1.knfsd.internal.`.                                                                                                                                                                                                                                                                           | False    | `"{PROXY_BASENAME}.knfsd.internal."` |
| ASSIGN_STATIC_IPS                   | If set to `true`, configures the MIG to use [stateful IP addresses](https://cloud.google.com/compute/docs/instance-groups/configuring-stateful-ip-addresses-in-migs). If an instance is replaced due to an update or failing health check the new instance will keep the same IP address as the original instance.                                                                                                                                                                                                 | False    | `false`                              |
| ENABLE_HIGH_BANDWIDTH_CONFIGURATION | If set to `true` enables [gVNIC](https://cloud.google.com/compute/docs/networking/using-gvnic) and [Tier 1 Bandwidth](https://cloud.google.com/compute/docs/networking/configure-vm-with-high-bandwidth-configuration) for higher egress. When enabled, only N2, N2D, C2 or C2D VM's are supported. You should also make sure you [assign enough vCPU's](https://cloud.google.com/compute/docs/networking/configure-vm-with-high-bandwidth-configuration#bandwidth-tiers) to take advantage of this configuration. | False    | null                                 |

### Health Check Configuration

| Variable                          | Description                                                                                                                                                                                                                             | Required | Default |
| --------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- | ------- |
| ENABLE_AUTOHEALING_HEALTHCHECKS   | When `true`, instances that fail their health check will be be replaced by the Managed Instance Group. Note, even when this is `false` the health checks still used by the load balancer to direct NFS traffic to healthy instances.    | False    | `true`  |
| HEALTHCHECK_INITIAL_DELAY_SECONDS | Initial delay before a failing health check will replace an proxy instance. This allows the proxy time to start up. Note, this only applies to the initial boot. If you reboot a proxy instance this initial interval *does not* apply. | False    | `600`   |
| HEALTHCHECK_INTERVAL_SECONDS      | How frequently (in seconds) to probe if a proxy instance is healthy. This is measured from the start of one probe, to the start of the next probe.                                                                                      | False    | `60`    |
| HEALTHCHECK_TIMEOUT_SECONDS       | How long (in seconds) to wait for a response from a probe. Must be less than or equal to `HEALTHCHECK_INTERVAL_SECONDS`.                                                                                                                | False    | `2`     |
| HEALTHCHECK_HEALTHY_THRESHOLD     | Number of sequential successful probe results for a proxy instance to be considered healthy.                                                                                                                                            | False    | `3`     |
| HEALTHCHECK_UNHEALTHY_THRESHOLD   | Number of sequential failed probe results for a proxy instance to be considered unhealthy.                                                                                                                                              | False    | `3`     |

**NOTE:** `HEALTHCHECK_INITIAL_DELAY_SECONDS` only applies to the first time the proxy starts up. If you reboot the proxy the standard health checks intervals will apply. The time allowed for a reboot is `HEALTHCHECK_INTERVAL_SECONDS * (HEALTHCHECK_UNHEALTHY_THRESHOLD - 1) + HEALTHCHECK_TIMEOUT_SECONDS`, with the default values this is `60 seconds * (3 probes - 1) + 2 seconds = 122 seconds` (effectively 2 minutes).

Increasing `HEALTHCHECK_INTERVAL_SECONDS` and/or `HEALTHCHECK_UNHEALTHY_THRESHOLD` will allow more time to reboot a proxy instance. However, it will also delay the system from detecting unhealthy instances.

### Export Configuration

| Variable                | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                     | Required                                                                             | Default |
| ----------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------ | ------- |
| EXPORT_MAP              | A list of NFS Exports to mount from on-premise and re-export in the format `<SOURCE_IP>;<SOURCE_EXPORT>;<TARGET_EXPORT>`.<br><br> For example to mount `10.0.0.1/export` from on-premise and re-export as `10.100.100.1/reexport` you would set the `EXPORT_MAP` variable to `10.100.100.1;/export;/reexport`.<br><br>You can specify multiple re-exports using a comma, for example `10.100.100.1;/assets;/assetscache,10.100.100.1;/textures;/texturescache`. | `EXPORT_MAP`, `EXPORT_HOST_AUTO_DETECT` or NetApp Auto-Discovery must be configured. | N/A     |
| EXPORT_HOST_AUTO_DETECT | A list of IP addresses or hostnames of NFS Filers that respond to the `showmount` command. Knfsd will automatically detect and re-export mounts from this filer. Exports paths on the cache will match the export path on the source filer.<br><br> You can specify multiple filers using a comma, for example `10.100.100.1,10.100.200.1` however you must ensure that these hosts are not exporting the same exports.                                         | `EXPORT_MAP`, `EXPORT_HOST_AUTO_DETECT` or NetApp Auto-Discovery must be configured. | N/A     |
| EXCLUDED_EXPORTS        | A list of filter patterns to be excluded from auto-discovery (see [Filter Patterns](filter-patterns.md)). Auto-discovery will ignore any exports that match any of the exclude patterns. Does not apply to mounts specified in the `EXPORT_MAP`. Paths filtered from auto-discovery can be explicitly exported using `EXPORT_MAP`, this can be used to change the export path.                                                                                  | False                                                                                | `[]`    |
| INCLUDED_EXPORTS        | If set, auto-discovery will only include paths matching a filter pattern from the include list (see [Filter Patterns](filter-patterns.md)). Does not apply to mounts specified in the `EXPORT_MAP`. Paths filtered from auto-discovery can be explicitly exported using `EXPORT_MAP`, this can be used to change the export path.                                                                                                                               | False                                                                                | `[]`    |

### NetApp Exports Auto-Discovery Configuration

If using the NetApp Exports Auto-Discovery feature, please also read the [NetApp specific docs](netapp-docs.md).

| Variable                  | Description                                                                                                                                                                                                                                                                           | Required                                 | Default                               |
| ------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------- | ------------------------------------- |
| ENABLE_NETAPP_AUTO_DETECT | Enables automatic discovery of exports using the NetApp REST API.                                                                                                                                                                                                                     | False                                    | `false`                               |
| NETAPP_HOST               | DNS or IP of the NetApp server. This is the DNS or IP name clients use when mounting the NFS shares.                                                                                                                                                                                  | If `ENABLE_NETAPP_AUTO_DETECT` is `true` |                                       |
| NETAPP_URL                | URL of the NetApp REST API. This *must* include the API version and end with a slash, for example `https://netapp.example/api/v1/`.                                                                                                                                                   | If `ENABLE_NETAPP_AUTO_DETECT` is `true` |                                       |
| NETAPP_USER               | The username used to authenticate with the NetApp REST API.                                                                                                                                                                                                                           | If `ENABLE_NETAPP_AUTO_DETECT` is `true` |                                       |
| NETAPP_SECRET             | The name of a GCP Secret containing the NetApp REST API password.                                                                                                                                                                                                                     | If `ENABLE_NETAPP_AUTO_DETECT` is `true` |                                       |
| NETAPP_SECRET_PROJECT     | The GCP project containing the secret.                                                                                                                                                                                                                                                | False                                    | The project the cluster is running in |
| NETAPP_SECRET_VERSION     | The version of the secret.                                                                                                                                                                                                                                                            | False                                    | `latest`                              |
| NETAPP_CA                 | PEM encoded certificate containing the root certificate for the NetApp REST API. This can also include intermediate certificates to provide the full certificate chain. To read this from a file use the [Terraform file function](https://www.terraform.io/language/functions/file). | If `ENABLE_NETAPP_AUTO_DETECT` is `true` |                                       |
| NETAPP_ALLOW_COMMON_NAME  | Allows using the Common Name (CN) field of the certificate as a DNS name when the certificate does not include a Subject Alternate Name (SAN) field.                                                                                                                                  | False                                    | `false`                               |

### Knfsd Proxy Configuration

| Variable                         | Description                                                                                                                                                                                                                                                                                                                                                                                                                             | Required | Default                         |
| -------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- | ------------------------------- |
| PROXY_BASENAME                   | A nickname to use for this Knfsd deployment (used to ensure uniquely named resources for multiple deployments).                                                                                                                                                                                                                                                                                                                         | False    | `nfsproxy`                      |
| EXPORT_CIDR                      | The CIDR to use in `/etc/exports` of the Knfsd Node for filesystem re-export.                                                                                                                                                                                                                                                                                                                                                           | False    | `10.0.0.0/8`                    |
| PROXY_IMAGENAME                  | The name of the Knfsd base [image](https://cloud.google.com/compute/docs/images).                                                                                                                                                                                                                                                                                                                                                       | True     | N/A                             |
| KNFSD_NODES                      | The number of Knfsd nodes to deploy as part of the cluster.                                                                                                                                                                                                                                                                                                                                                                             | False    | 3                               |
| PROXY_LABELS                     | GCP Labels to apply to the proxy VM instances                                                                                                                                                                                                                                                                                                                                                                                           | False    | `{ vm-type = "nfs-proxy" }`     |
| SERVICE_LABEL                    | The Service Label to use for the Forwarding Rule.                                                                                                                                                                                                                                                                                                                                                                                       | False    | `dns`                           |
| VFS_CACHE_PRESSURE               | The value to set for `vfs_cache_pressure` Rule.                                                                                                                                                                                                                                                                                                                                                                                         | False    | `100`                           |
| READ_AHEAD                       | The number of bytes to read ahead. Must be a multiple of the kernel page size (8 KiB for 5.11). The kernel will round this down to the nearest page.                                                                                                                                                                                                                                                                                    | False    | `8388608`                       |
| ENABLE_AUTOHEALING_HEALTHCHECKS  | Should failed healthchecks lead to instance replacement?                                                                                                                                                                                                                                                                                                                                                                                | False    | `true`                          |
| ENABLE_STACKDRIVER_METRICS       | Should Knfsd metrics be exported into Stackdriver?                                                                                                                                                                                                                                                                                                                                                                                      | False    | `true`                          |
| METRICS_AGENT_CONFIG             | Custom YAML configuration for the metrics agent. The configuration *is not* validated by Terraform, when using a custom config check the proxy startup log. See the custom configuration section in the [metrics documentation](metrics.md) for more details.                                                                                                                                                                           | False    | `""`                            |
| ROUTE_METRICS_PRIVATE_GOOGLEAPIS | Override the IP address used for `monitoring.googleapis.com` and `logging.googleapis.com` to an IP range in the `private.googleapis.com` range. See [here](https://cloud.google.com/vpc/docs/configure-private-google-access-hybrid#config-choose-domain) for more details.                                                                                                                                                             | False    |
| CUSTOM_PRE_STARTUP_SCRIPT        | Optional bash script to run before the [proxy-startup.sh](proxy-startup.sh) script. For example `file("/home/ben/myscript.sh")`.                                                                                                                                                                                                                                                                                                        | False    | empty script                    |
| CUSTOM_POST_STARTUP_SCRIPT       | The path to a bash script to run after the [proxy-startup.sh](proxy-startup.sh) script. For example `file("/home/ben/myscript.sh")`.                                                                                                                                                                                                                                                                                                    | False    | empty script                    |
| MACHINE_TYPE                     | The GCP Machine type to use for the Knfsd cache. Currently only N1 instances can be used.                                                                                                                                                                                                                                                                                                                                               | False    | `n1-highmem-16`                 |
| MIG_MAX_UNAVAILABLE_PERCENT      | The maximum number of instances that can be unavailable during automated MIG updates ([see docs](https://cloud.google.com/compute/docs/instance-groups/rolling-out-updates-to-managed-instance-groups#max_unavailable)). Defaults to 100% to ensure consistent cache instances within the MIG.                                                                                                                                          | False    | `100`                           |
| MIG_REPLACEMENT_METHOD           | The instance replacement method for managed instance groups. Valid values are: `RECREATE`, `SUBSTITUTE`.<br><br>If `SUBSTITUTE` (default), the group replaces VM instances with new instances that have randomly generated names. If `RECREATE`, instance names are preserved. You must also set `MIG_MAX_UNAVAILABLE_PERCENT` to be greater than 0 (default is already `100` so this only applies if you have modified this variable). | False    | `SUBSTITUTE` or `RECREATE`      |
| MIG_MINIMAL_ACTION               | Minimal action to be taken on an instance. You can specify either RESTART to restart existing instances or REPLACE to delete and create new instances from the target template. If you specify a RESTART, the Updater will attempt to perform that action only. However, if the Updater determines that the minimal action you specify is not enough to perform the update, it might perform a more disruptive action.                  | False    | `RESTART`                       |
| ENABLE_KNFSD_AGENT               | Should the [Knfsd Agent](../../image/knfsd-agent/README.md) be started at Proxy Startup?                                                                                                                                                                                                                                                                                                                                                | False    | `true`                          |
| SERVICE_ACCOUNT                  | Service account the NFS proxy compute instances will run with.                                                                                                                                                                                                                                                                                                                                                                          | False    | See service account notes below |

The default `MIG_REPLACEMENT_METHOD` depends on `ASSIGN_STATIC_IPS`:

* When `ASSIGN_STATIC_IPS = false` then the default `MIG_REPLACEMENT_METHOD` is `SUBSTITUTE`.
* When `ASSIGN_STATIC_IPS = true` then the default `MIG_REPLACEMENT_METHOD` is `RECREATE`.

#### Service Account Notes

The default `SERVICE_ACCOUNT` depends on `ENABLE_STACKDRIVER_METRICS`.

* When `false` the NFS proxy instances will not have a service account
* When `true` the NFS proxy instances will use the default compute service account with the following scopes:
  * `https://www.googleapis.com/auth/logging.write`
  * `https://www.googleapis.com/auth/monitoring.write`

If a custom service account is assigned the compute instances will use the `https://www.googleapis.com/auth/cloud-platform` scope. This allows the proxy instances access to any GCP API permitted by IAM.

The service account will need the following project level IAM permissions:

* Logs Writer (`roles/logging.logWriter`)
* Monitoring Metric Writer (`roles/monitoring.metricWriter`)

### Cachefilesd Configuration

| Variable                            | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                 | Required | Default     |
| ----------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- | ----------- |
| CACHEFILESD_DISK_TYPE               | The disk type to use for the Cachefiles directory. Can be either `local-ssd`, `pd-standard`, `pd-balanced` or `pd-ssd`. Local SSDs provide the highest performance, but persistent disk allows for larger volumes.                                                                                                                                                                                                                                          | False    | `local-ssd` |
| LOCAL_SSDS                          | (Only used if `CACHEFILESD_DISK_TYPE` = `local-ssd`) The number of Local SSDs to assign to each cache instance. This can be either 0 to 8, 16, or 24 local SSDs for up to 9TB of capacity ([see here](https://cloud.google.com/compute/docs/disks/local-ssd#choosing_a_valid_number_of_local_ssds)). If you are setting this to 24 Local SSDs you should also change the `MACHINE_TYPE` variable to an instance with 32 CPU's, for example `n1-highmem-32`. | False    | `4`         |
| CACHEFILESD_PERSISTENT_DISK_SIZE_GB | (Only used if `CACHEFILESD_DISK_TYPE` = `pd-standard`, `pd-balanced` or `pd-ssd`), what size should the persistent disk be in GB? Can be set between `100` and `64000`. For large volumes, consider larger instance types (see [here](https://cloud.google.com/compute/docs/disks/performance)).                                                                                                                                                            | False    | `1500`      |

### Mount Options

These mount options are for the proxy to the source server.

| Variable          | Description                                                                                                                                                                                                                            | Required | Default   |
| ----------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- | --------- |
| NCONNECT_VALUE    | The number of TCP connections to use when connecting to the source.                                                                                                                                                                    | False    | `16`      |
| ACREGMIN          | The minimum time (in seconds) that the NFS client caches attributes of a regular file.                                                                                                                                                 | False    | `600`     |
| ACREGMAX          | The maximum time (in seconds) that the NFS client caches attributes of a regular file.                                                                                                                                                 | False    | `600`     |
| ACDIRMIN          | The minimum time (in seconds) that the NFS client caches attributes of a directory.                                                                                                                                                    | False    | `600`     |
| ACDIRMAX          | The maximum time (in seconds) that the NFS client caches attributes of a directory. This can be reduced to improve the cache coherency for `readdir` operationns (e.g `ls`) at the cost of increasing metadata requests to the source. | False    | `600`     |
| RSIZE             | The maximum number of bytes the proxy will read from the source in a single request. The actual value will be negotiated with the source server to determine the maximum value support by both machines.                               | False    | `1048576` |
| WSIZE             | The maximum number of bytes the proxy will write to the source in a single request. The actual value will be negotiated with the source server to determine the maximum value support by both machines.                                | False    | `1048576` |
| MOUNT_OPTIONS     | Any additional NFS mount options not covered by existing variables. These options will be applied to all NFS mounts.                                                                                                                   | False    | `""`      |
| NFS_MOUNT_VERSION | The mount version to use for NFS client mounts (`vers` option). Acceptable values are `3`, `4`, `4.0`, `4.1`, `4.2`.                                                                                                                   | False    | `3`       |

### NFS Kernel Server Options

| Variable              | Description                                                                                                                                                                                                                                                                                                                                          | Required | Default       |
| --------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- | ------------- |
| DISABLED_NFS_VERSIONS | The versions of NFS that should be disabled in `nfs-kernel-server`. Explicitly disabling unwanted NFS versions prevents clients from accidentally auto-negotiating an undesired NFS version. Specify multiple versions to disable with a comma separated list. Acceptable values are `3`, `4`, `4.0`, `4.1`, `4.2`. NFS Version 2 is always diabled. | False    | `4.0,4.1,4.2` |
| NUM_NFS_THREADS       | The number of NFS Threads to use for KNFSD.                                                                                                                                                                                                                                                                                                          | False    | `512`         |

### Export Options

| Variable       | Description                                                                       | Required | Default |
| -------------- | --------------------------------------------------------------------------------- | -------- | ------- |
| NOHIDE         | When `true`, adds the `nohide` option to all the exports.                         | False    | `true`  |
| EXPORT_OPTIONS | Any custom NFS exports options. These options will be applied to all NFS exports. | False    | `""`    |


### Autoscaling Configuration

| Variable                                    | Description                                                                                                                            | Required | Default |
| ------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------- | -------- | ------- |
| ENABLE_KNFSD_AUTOSCALING                    | Should autoscaling be enabled for Knfsd? **You MUST set the `ENABLE_STACKDRIVER_METRICS` variable to `true` if enabling autoscaling**. | False    | False   |
| KNFSD_AUTOSCALING_NFS_CONNECTIONS_THRESHOLD | The number of Client Connections to Knfsd that should be targeted for each instance (exceeding will trigger a scale-up).               | False    | `250`   |
| KNFSD_AUTOSCALING_MIN_INSTANCES             | The minimum number of Knfsd instances to set regardless of the traffic volumes.                                                        | False    | `1`     |
| KNFSD_AUTOSCALING_MAX_INSTANCES             | The maximum number of Knfsd instances to set regardless of the traffic volumes.                                                        | False    | `10`    |

## Outputs

| Output                             | Description                                                                                                                        |
| ---------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------- |
| `nfsproxy_loadbalancer_ipaddress`  | The internal IP address of the load balancer (when `TRAFFIC_DISTRIBUTION_MODE = "loadbalancer"`).                                  |
| `nfsproxy_loadbalancer_dnsaddress` | The internal DNS name of the load balancer (when `TRAFFIC_DISTRIBUTION_MODE = "loadbalancer"`).                                    |
| `dns_name`                         | The internal DNS name of the KNFSD proxy instance group (when `TRAFFIC_DISTRIBUTION_MODE` is `dns_round_robin` or `loadbalancer`). |
| `instance_group`                   | Full URL (self link) of the KNFSD proxy instance group.                                                                            |
| `instance_group_manager`           | Full URL (self link) of the KNFSD proxy instance group manager.                                                                    |
| `instance_group_name`              | Name of the KNFSD proxy instance group.                                                                                            |

## Caveats

### Excluding nested mounts

If you exclude a nested mount but still export the parent mount you may get I/O errors when accessing the nested mount.

The exact behaviour will depend on how the source server has exported the nested mount.

If the source server exports the mount with the `crossmnt`, or `nohide` options then trying to access the nested mount, or list the directory containing the nested mount will result in I/O errors.

If the source server exports the mount without `crossmnt`, or `hide` options then the directory for the nested mount will be visible, but empty.

It is advised that if you exclude a nested mount, you also exclude the parent mount. You may however exclude a parent mount but include a nested mount.

For example, if you have the following mounts:

```text
/assets
/assets/common
/assets/common/textures
```

You could exclude `/assets`, but still export `/assets/common` and `/assets/common/textures`. You could also export only `/assets/common/textures`.

However, exporting `/assets` but excluding `/assets/common` could cause errors.

Exporting `/assets` and `/assets/common/textures`, but excluding `/assets/common` will likely fail, and can have unintended side-effects as the proxy will try to create the directory `/assets/common`.

### Combining auto-discovery and explicit mounts

While auto-discovery and explicit mounts can be combined the system does not have any special handling for duplicate paths.

As such it is not recommended to combine multiple auto-discovery methods, or explicit (`EXPORT_MAP`).

The behaviour of duplicates is undefined. The system might overwrite one mount with another, or it may error.

### Limitations on export names

The proxy cannot re-export any path that matches a symlink on the local server.

The most likely symlinks that will cause conflicts are:

* `/bin`
* `/lib`
* `/lib32`
* `/lib64`
* `/libx32`
* `/sbin`

The proxy will fail to start if it attempts to export a path that matches a symlink. Check the logs for errors such as:

```text
ERROR: Cannot mount 10.0.0.2:/bin because /bin matches a symlink
```

If you are providing a manual export list, specify a different path for the export, such as `10.0.0.2;/bin;/binaries`.

If you're using auto-discovery add the path to the list of excluded exports, for example `EXCLUDED_EXPORTS = ["/bin"]`

For a full list of symlinks, start a compute instance using the proxy image (without the standard startup script) and run the command:

```bash
find / -type l
```

Most of the symlinks listed are unlikely to cause issues, such as `/usr/lib/x86_64-linux-gnu/libc.so.6`.
