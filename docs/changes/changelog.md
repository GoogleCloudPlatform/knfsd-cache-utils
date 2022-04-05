# Next

* Pin to last GCP image that includes 5.13 kernel
* Custom KNFSD culling agent

## Pin to last GCP image that includes 5.13 kernel

The 5.13 HWE kernel packages have been removed from the Ubuntu 20.04
sources list.

Use a GCP image that has the 5.13 kernel pre-installed until the 5.15
kernel can be tested.

## Custom KNFSD culling agent

cachefilesd will periodically stop culling files when the cache is full.

This is because FS-Cache indicates that the file is still in use. It appears that NFS is not releasing the file while NFS still has the file in the slab cache. Dropping the slab cache forces NFS to release the files allowing cachefilesd to resume culling.

When custom culling is enabled culling is disabled in cachefilesd and the agent takes over deleting files when the cache is over a certain threshold.

While this is undefined behaviour in FS-Cache, in testing FS-Cache would drop its state and resume caching of files after dropping the dentries and inode cache (triggered writing 2 to /proc/sys/vm/drop_caches).

The culling threshold *MUST* be greater than `bstop` and `fstop` in `/etc/cachefilesd.conf`, otherwise the cache can fill before the culling threshold is reached.

# v0.7.1

* Add new `SUBNETWORK_PROJECT` configuration option for explicitly defining which project the Knfsd subnet belongs to

## Add new `SUBNETWORK_PROJECT` configuration option for explicitly defining which project the Knfsd subnet belongs to

When defined, the new `SUBNETWORK_PROJECT` environment variable explicitly sets the `subnetwork_project` variable in the Terraform `compute_instance_template` resource.

This only needs to be set if using a Shared VPC, where the subnetwork exists in a different project. Otherwise it defaults to the provider project.

# v0.7.0

* Add new `ROUTE_METRICS_PRIVATE_GOOGLEAPIS` configuration option for sending metrics and logs over `private.googleapis.com` IP addresses

## Add new `ROUTE_METRICS_PRIVATE_GOOGLEAPIS` configuration option for sending metrics over `private.googleapis.com` IP addresses

When set to `true`, the new `ROUTE_METRICS_PRIVATE_GOOGLEAPIS` variable will trigger an addition to the `/etc/hosts` file of each Knfsd Node for the following FQDN's:

* monitoring.googleapis.com
* logging.googleapis.com
* cloudtrace.googleapis.com

The IP used (`199.36.153.11`) is from the range defined in the [Private Google Access docs](https://cloud.google.com/vpc/docs/configure-private-google-access-hybrid#config-choose-domain). This ensures that metrics and logs are shipped over a predictable IP address range that is only routable from within Google Cloud.

For most use-cases this will not be required, however this is beneficial when the default internet (`0.0.0.0/0`) route has been removed from the VPC and a specific, predictable CIDR range is required for shipping logs and metrics to Google Cloud Operations.

# v0.6.4

* Fix missing proxy to source metrics

## Fix missing proxy to source metrics

The `mount` metrics were missing (or reported as zero) from the `knfsd-metrics-agent`.

# v0.6.3

This release only updates documentation, there are no changes to the main knfsd
proxy image or deployment scripts.

* Update tutorial scripts to work with the current proxy image

## Update tutorial scripts to work with the current proxy image

The tutorial scripts were out of date and no longer worked with the proxy image. Update the scripts in preparation for a new tutorial.

# v0.6.2

* Increase health check interval to 60 seconds
* Add parameters to configure health checks
* Support deploying metrics as a Terraform module
* Remove per-mount stats (aggregate by source server)
* Exclude support when resolving knfsd proxy instance name

## Increase health check interval to 60 seconds

This allows 2 minutes (with the default health check values) to reboot a knfsd proxy instance without the managed instance group replacing the instance.

## Add parameters to configure health checks

This allows overriding various parameters used by the health checks. For example, if you do not encounter the culling issue you might want to reduce the `HEALTHCHECK_INTERVAL_SECONDS` so that failed instances are detected more quickly.

If you have a lot of volumes, or high latency between the source and the proxy causing a startup time slower than 10 minutes (600 seconds), you might want to increase the `HEALTHCHECK_INITIAL_DELAY_SECONDS`. Conversely, if you know your proxy starts up in less than 5 minutes, you can reduce the initial delay so that instances that fail to start up correctly are detected and replaced more quickly.

## Support deploying metrics as a Terraform module

Support deploying the metrics as a Terraform module so that the metrics can be deployed without needing to clone the Terraform configuration from git.

## Remove per-mount stats (aggregate by source server)

Reporting on the stats per-mount generates a lot of logging data when the source has 50 or more volumes.

Secondly, the stats cannot be reliably aggregated later because multiple mounts can share the same NFS client. All mounts sharing the same NFS clients will have identical stats that are an aggregate of all the mounts sharing the NFS client. If these per-mount stats are then summed together on a dashboard it leads to the stats being multiplied by the number of mounts that share the same NFS client.

However, because some mounts might have a separate NFS client, and thus separate stats, it becomes impossible to view an accurate total on a dashboard when the stats are reported per-mount.

## Exclude support when resolving knfsd proxy instance name

Support a list of excluded servers and/or local paths when `query_proxy_instances` is enabled.

When collecting metrics from clients with `query_proxy_instances` enabled by default the collector will probe every NFS server that is mounted by the client. This can cause issues if the client is mounting a mixture of knfsd proxies and other NFS servers.

The metrics collector now has an `exclude` configuration section for `query_proxy_instances`.

# v0.6.1

* Use latest 5.13 HWE kernel
* Use metric labels for mount stats
* Update dashboard to use new metrics

## Use latest 5.13 HWE kernel

The image build script started failing with the error:

```text
E: Version '5.13.0.39.44~20.04.24' for 'linux-generic-hwe-20.04' was not found
E: Version '5.13.0.39.44~20.04.24' for 'linux-image-generic-hwe-20.04' was not found
E: Version '5.13.0.39.44~20.04.24' for 'linux-headers-generic-hwe-20.04' was not found
```

The `linux-image-hwe-20.04` package only keeps the binaries for the latest HWE kernel. As such the HWE kernels cannot be pinned to a specific kernel version.

## Use metric labels for mount stats

The mount stats were previously using resource labels for `server`, `path` and `instance`. This was intended to reduce the volume of data being logged by reducing repeated values.

However, GCP Cloud Monitoring does not support custom resource labels. This is likely to be a common issue with other reporting systems either handling resource labels differently, or ignoring them completely.

To avoid issues the labels for `server`, `path` and `instance` are now reported as metric level labels.

## Update dashboard to use new metrics

This adds new graphs using the new metrics to show the total read/write throughput between:

* KNFSD Proxy and Source.
* Clients and KNFSD Proxy.

The new dashboard also corrects an issue where the total number of operations from the KNFSD Proxy to the Source were being under reported. This is because the metric agent only parses a single `xprt` (transport) entry.

# v0.6.0

* Revert to Ubuntu 20.04 with kernel 5.13
* Increase how much space is culled by cachefilesd
* Abort mounting export after 3 attempts
* Custom GCP labels for proxy VM instances

## Revert to Ubuntu 20.04 with kernel 5.13

5.17 is currently has too high a performance degradation in the new FS-Cache implementation. Currently observing a maximum of 40 MB/s per thread.

Though the total throughput can still reach the maximum network speed (e.g. 1 GB/s) in aggregate the performance hit to individual clients shows a significant performance drop in workloads such as rendering.

## Increase how much space is culled by cachefilesd

Increase the `frun` and `brun` limits from 10% to 20%. This causes cachefilesd to reclaim more space once culling begins. The goal is to reduce how often cachefilesd needs to cull space when reading uncached data.

## Abort mounting export after 3 attempts

Only try to mount the same export a maximum of 3 times (with 60 seconds between each attempt).

If the attempts fail the startup script will be aborted and the NFS server will not be started.

When the health check is enabled, after 10 minutes the proxy instance will be replaced.

## Custom GCP labels for proxy VM instances

Added a new `PROXY_LABELS` variable to set custom labels on the proxy VM instances. This can aid with filtering metrics and logs when running multiple proxy clusters in a single project.


# v0.5.1

* Collect NFS metrics by operation
* Collect NFS metrics for read/write bytes by mount/export

## Collect NFS metrics by operation

These metrics show the counts for each NFS operation (e.g. READ, WRITE, READDIR, LOOKUP, etc). The metrics include:

* Number of Requests
* Bytes Sent
* Bytes Received
* Major Timeouts
* Errors

On the proxy this will show the types of operation requested between the proxy and the source. This can be used for diagnostics if a proxy shows poor performance to see the type of traffic (e.g. read/write heavy, vs metadata heavy).

These metrics can also be collected from the clients to see the types of traffic between the client and the proxy.

## Collect NFS metrics for read/write bytes by mount/export

This allows for better visualization of traffic between proxy and source, and between proxy and clients.

Monitoring this at the network level cannot show whether inbound traffic comes from clients writing data, or the proxy reading data from the source.

The read/write metrics for mounts will show the number of bytes read/wrote  between the proxy and source. These metrics are split by individual mount so the dashboards can indicate which specific mount consists of a majority of the traffic.

The read/write metrics for exports will show the number of bytes read/wrote between the proxy and clients. These metrics are only provided in aggregate, with a single total for all exports.

# v0.5.0

* (GCP) Use LTS versions of Ubuntu
* (GCP) Use a smaller machine type when building the image
* (GCP) Stop reporting file system usage metrics for NFS mounts
* (GCP) Implement the knfsd-agent which provides a HTTP API for interacting with Knfsd nodes
* (GCP) Remove --manage-gids from RPCMOUNTDOPTS
* (GCP) Add ability to explicitly disable NFS Versions in `nfs-kernel-server` and default to disabling NFS versions `4.0`, `4.1`, and `4.2`
* (GCP) Remove the metadata server read sleep from `proxy-startup.sh`
* (GCP) Packer build script
* (GCP) Use fixed ports for NFS services
* (GCP) Configure mount point timeout when building the image
* (GCP) Auto-discovery for NetApp exports using NetApp REST API
* (GCP) Removed DISCO_MOUNT_EXPORT_MAP
* (GCP) Set nohide on all exports by default
* (GCP) Metrics collection agent to replace bash scripts
* (GCP) wildcard support for `EXCLUDED_EXPORTS`
* (GCP) `EXCLUDED_EXPORTS` changed to `list(string)`
* (GCP) Add include filter patterns `INCLUDED_EXPORTS` for auto-discovery
* (GCP) Remove restriction on protected paths
* (GCP) Replaced Stackdriver agent with Cloud Ops Agent
* (GCP) Changed KNFSD Metrics Agent to use OpenTelemetry
* (GCP) Changed custom metric types
* (GCP) Update kernel version to 5.17.0
* (GCP) Change mounts to use async instead of sync

## (GCP) Use LTS versions of Ubuntu

LTS (Long-Term Support) versions of Ubuntu are preferred for stability. These versions are supported for longer a longer period of time thus require less frequent updates to new major versions. LTS versions are normally released every 2 years and supported for 5 years. Non-LTS versions are only supported for 9 months and released every 6 months.

## (GCP) Use a smaller machine type when building the image

Sometimes c2-standard-32 instances are unavailable in a specific zone. The build machine was changed to use a more available machine type. A larger machine type can be used to improve the build times if they're too slow.

## (GCP) Stop reporting file system usage metrics for NFS mounts

The default stackdriver collectd configuration for the `df` plugin includes metrics for NFS shares. The df plugin only collects basic metrics such as disk free space, inode free space, etc.

Collecting these metrics about NFS shares from the proxy is largely pointless as the same metrics are also available from the source server.

If the proxy has hundreds or thousands of NFS exports mounted this can greatly increase the volume of metrics being collected, leading to excessive charges for metrics ingestion.

## (GCP) Implement the knfsd-agent which provides a HTTP API for interacting with Knfsd nodes

Implements a new Golang based knfsd-agent which runs on all Knfsd nodes. It provides a HTTP API that can be used for interacting with Knfsd nodes. [See here](../../image/knfsd-agent/README.md) for more information.

If upgrading from `v.0.4.0` and below you will need to build a [new version of the Knfsd Image](../../image).

## (GCP) Remove --manage-gids from RPCMOUNTDOPTS

The `--manage-gids` option causes NFS to ignore the user’s auxiliary groups and look them up based on the local system. The reason for this option is NFS v3 supports a maximum of 16 auxiliary groups.

This was causing the NFS proxy to ignore the auxiliary groups in the incoming NFS request and replace them with supplementary groups from the local system. Since the NFS proxy does not have any users configured and is not connected to LDAP this effectively removes all the supplementary groups from the user.

When the proxy then sends the request to the source server, the new request was missing the auxiliary groups. This would cause permission errors when accessing files that depended on those auxiliary groups.

The `--manage-gids` option only makes sense if the proxy is connected to LDAP.

## (GCP) Add ability to explicitly disable NFS Versions in `nfs-kernel-server` and default to disabling NFS versions `4.0`, `4.1`, and `4.2`

Adds the ability to explicitly disable NFS Versions in `nfs-kernel-server`. Explicitly disabling unwanted NFS versions prevents clients from accidentally auto-negotiating an undesired NFS version.

With this change, by default, NFS versions `4.0`, `4.1`, and `4.2` are now disabled on all proxies. To enable it, set a custom value for `DISABLED_NFS_VERSIONS`.

## (GCP) Remove the metadata server read sleep from `proxy-startup.sh`

Removes the legacy 1 second sleep that is performed before each call to the GCP Metadata server. This speeds up proxy startup by ~20 seconds.

## (GCP) Packer build script

Packer script to automate building the knfsd image on GCP.

## (GCP) Use fixed ports for NFS services

This allows defining a firewall rule that only permits clients to access the ports required by NFS.
Previously the firewall rule had to allow any port.

This also improves stability when the load balancer connects a client to a different instance.
This was especially problematic if the client used UDP to access the portmapper or mountd services, as the UDP load balancer could connect the client to a different instance compared with TCP.

## (GCP) Configure mount point timeout when building the image

Moved this configuration into the modprobe options when the image is built.
You will need to build a new image when updating to the latest Terraform to avoid issues with stale file handles.

## (GCP) Auto-discovery for NetApp exports using NetApp REST API

This can be used to discover exports on a NetApp system using the NetApp REST API when the `showmount` command is disabled on the source server.

This replaces the old `DISCO_MOUNT_EXPORT_MAP` that used the `tree` command to discover nested mounts (junction points).

## (GCP) Removed DISCO_MOUNT_EXPORT_MAP

This command would cause excessive I/O due to the use of the tree command to discover nested exports. Most nested exports were on NetApp file servers due to junction points.

Support for automatically discovering nested mounts is now handled using the NetApp REST API.

## (GCP) Set nohide on all exports by default

This allows nested mounts, even those explicitly defined using `EXPORT_MAP` to be exported automatically to clients without the client needing to mount them explicitly.

A NOHIDE option has been added to the Terraform to disable this option if required. To remove the nohide option set `NOHIDE = true` in Terraform.

## (GCP) Metrics collection agent to replace bash scripts

A custom metrics agent fixes several issues with the original bash scripts:

* Only read the slabinfo and mountstats once. Mountstats makes the biggest difference when you have a lot of exports.

* Single thread monitoring mountstats. Previously a separate bash process was spawned per export.

* Future support for a wider range of custom metrics.

## (GCP) wildcard support for `EXCLUDED_EXPORTS`

`EXCLUDED_EXPORTS` now supports wildcard patterns such as `/home/**`. See the [deployment README](../../deployment/README.md#filter-patterns) for full details.

## (GCP) `EXCLUDED_EXPORTS` changed to `list(string)`

`EXCLUDED_EXPORTS` now uses a list of strings instead of a comma delimited string.

```terraform
# Old Format
EXCLUDED_EXPORTS = "/home,/bin"

# New Format
EXCLUDED_EXPORTS = ["/home", "/bin"]
```

## (GCP) Add include filter patterns `INCLUDED_EXPORTS` for auto-discovery

When `INCLUDED_EXPORTS` is set, auto-discovery will only re-export exports that match an include pattern. This can be combined with `EXCLUDED_EXPORTS`, to only export paths that match an include pattern but do not match an exclude pattern. See the [deployment README](../../deployment/README.md#filter-patterns) for full details.

## (GCP) Remove restriction on protected paths

The restriction on protected paths has been removed. This requires building a new proxy image to apply the NFS `rootdir` setting.

**WARNING:** If you use the current Terraform with an old image, the old image will be vulnerable to the bug where auto-discovery can overwrite a system path such as `/home` with a mount.

## (GCP) Replaced Stackdriver agent with Cloud Ops Agent

The Stackdriver Agent is obsolete. The last supported Ubuntu version is 20.04 LTS.

## (GCP) Changed KNFSD Metrics Agent to use OpenTelemetry

The previous KNFSD Metrics Agent relied on reporting metrics via collectd (using the Stackdriver Agent).

Update the KNFSD metrics agent to use same OpenTelemetry Collector as the Cloud Ops Agent. This allows including additional metadata such as separating out the server and path for NFS mounts.

## (GCP) Changed custom metric types

A limitation of the old collectd based KNFSD Metrics Agent is that all gauges had to be floats.

The new KNFSD Metrics Agent can now report gauges such as `knfsd/nfs_connections` using the correct data type (integers).

You will need to apply the [knfsd-cache-utils/deployment/metrics/](../../deployment/metrics/) Terraform to update the custom metrics.

## (GCP) Update kernel version to 5.17.0

The 5.17.0 kernel introduces a new implementation of FS-Cache that has better handling for culling old objects when the cache is full.

## (GCP) Change mounts to use async instead of sync

Previously the source NFS exports were being mounted with the sync option. This has been changed to async because the new FS-Cache implementation has an issue where it fails to read cached objects when using the sync option.

The sync option is not required and was overly cautious. In testing when using async the proxy may buffer some writes before committing them back to the source server. However, when the writes are buffered the client is informed that the write is unstable. When the client issues a `COMMIT` command, or a write with the stable flag as `FILE_SYNC` or `DATA_SYNC` the proxy will write any buffered data to the source server with the stable flag set, and only respond to the client with success once the source server has indicated that the data has been committed to stable storage.

# v0.4.0

* (GCP) Fixed specifying project/region/zone
* (GCP) Changed `LOCAL_SSDS` to a simple count of the number of drives.
* (GCP) Prevent mounting over system directories
* (GCP) Added `EXCLUDED_EXPORTS` option to exclude exports from auto-discovery
* (GCP) Added optional UDP load balancer
* (GCP) Fixed remove duplicate and stale exports when restarting
* (GCP) Added configuration for NFS mount options
* (GCP) Added configuration for NFS export options
* (GCP) Added configuration option for read ahead
* (GCP) Explicitly set MIG `update_policy` to proactively replace instances, and make parameters configurable

## (GCP) Fixed specifying project/region/zone

Project, region and zone no longer need to be configured on the Google provider and provided parameters to the module.

If project, region and/or zone are set on the module, these values will be used instead of the values on the provider.

Project, region and zone will default to the provider's value if not set on the module.

## (GCP) LOCAL_SSDS changed to count

`LOCAL_SSDS` is now configured as a simple count of the number of drives. The local SSDS will be named sequentially with the prefix, `local-ssd-`.

**Old:**

```terraform
LOCAL_SSDS = ["local-ssd-1", "local-ssd-2", "local-ssd-3", "local-ssd-4"]
```

**New:**

```terraform
LOCAL_SSDS = 4
```

## (GCP) Prevent mounting over system directories

The `proxy-startup.sh` script now contains a list of protected directories such as `/bin` and `/usr`. Any exports that

When the proxy starts up, check the logs entries such as:

> startup-script: ERROR: Cannot mount 10.0.0.2:/home because /home is a system path

The `/home` directory is included in the list of protected directories to avoid unintended interactions, or issues with the GCP infrastructure such as SSH keys. These can be provisioned automatically on compute instances via OS Login or metadata. Commands such as `gcloud compute ssh` can also create SSH keys. These keys will be created in user home folders in the `/home` directory.

For a full list of the paths, see `PROTECTED_PATHS` in [proxy-startup.sh](../../deployment/terraform-module-knfsd/resources/proxy-startup.sh).

## (GCP) Added `EXCLUDED_EXPORTS` option to exclude exports from auto-discovery

This can be used to exclude specific exports when using auto-discovery such as
`EXPORT_HOST_AUTO_DETECT`. The main use is to exclude any exports that would
try to mount over a a protected directory such as `/home`.

## (GCP) Added optional UDP load balancer

Added an option `ENABLE_UDP` that will deploy a UDP load balancer for the NFS proxy (sharing the same IP as the TCP load balancer).

This is mainly aimed at support for the mount protocol for older clients that default to using UDP. NFS does not recommend using UDP.

### Upgrading existing deployments to support UDP

Existing deployments created the reserved address using the purpose `GCE_ENDPOINT`. To share an IP with multiple load balancers the reserved address' purpose needs to be changed to `SHARED_LOADBALANCER_VIP`.

**NOTE:** You only need to follow these instructions if you want to use the UDP load balancer for existing deployments using v0.3.0 or earlier.

To avoid breaking existing deployments the `google_compute_address` is set to ignore changes to `purpose` in Terraform as most existing deployments will not require UDP. New deployments will set the purpose to `SHARED_LOADBALANCER_VIP`.

**IMPORTANT:** The purpose cannot be changed while the reserved address is in use. To update the purpose you will first need to delete the current TCP load balancer (forwarding rule). **This will prevent the clients from accessing the NFS proxy during the update.**

If you try and set `UDP_ENABLED = true` on an existing deployment you will receive the following error (the IP will match your load balancer's IP):

```text
Error: Error creating ForwardingRule: operation received error: error code "IP_IN_USE_BY_ANOTHER_RESOURCE", message: IP '10.0.0.2' is already being used by another resource.
```

**IMPORTANT:** These instructions only applies when setting `UDP_ENABLED = true` on an existing deployment. If this error occurs when deploying the proxy for the first time check that `LOADBALANCER_IP` is not set to an IP address that is already in use.

Configure your environment (change these values to match your deployment):

```bash
export CLOUDSDK_CORE_PROJECT=my-project
export CLOUDSDK_COMPUTE_REGION=us-west1
export PROXY_BASENAME=rendercluster1
```

Remove the TCP forwarding rule. You can get a list of forwarding rules by running `gcloud compute forwarding-rules list`. The forwarding rule will have the same name as the `PROXY_BASENAME` variable.

```bash
gcloud compute forwarding-rules delete "$PROXY_BASENAME"
```

Get the reserved IP of the address. You can get a list of addresses by running `gcloud compute addresses list`. The address will be named `PROXY_BASENAME-static-ip`.

```bash
gcloud compute addresses describe "$PROXY_BASENAME-static-ip" --format='value(address)'
```

Update Terraform with the IP address for the proxy:

```terraform
module "proxy" {
  source = "github.com/GoogleCloudPlatform/knfsd-cache-utils//deployment/terraform-module-knfsd?ref=v0.3.0"

  LOADBALANCER_IP = "10.0.0.2" # Use the value from the command above
}
```

To update the address' purpose the address needs to be re-created. Delete the address using the gcloud command. Due to dependencies between resources, `terraform taint` cannot be used to automatically delete and re-create the address.

```bash
gcloud compute addresses delete "$PROXY_BASENAME-static-ip"
```

Use Terraform to re-create the reserved address and the forwarding rules:

```bash
terraform apply
```

## (GCP) Fixed remove duplicate and stale exports when restarting

The `/etc/exports` file was not cleared when running the start up script. When rebooting a proxy instance this would create duplicate entries (or leave stale entries) in the `/etc/exports` file.

The `/etc/exports` file is now cleared by the start up script before appending any exports.

## (GCP) Added configuration for NFS mount options

Added variables to Terraform for:

* `ACDIRMIN`
* `ACDIRMAX`
* `ACREGMIN`
* `ACREGMAX`
* `RSIZE`
* `WSIZE`

Also added `MOUNT_OPTIONS` to allow specifying any additional NFS mount options not covered by existing Terraform variables.

## (GCP) Added configuration for NFS export options

Added `EXPORT_OPTIONS` to allow specifying custom NFS export options.

## (GCP) Added configuration option for read ahead

This allows tuning the read ahead value of the proxy for performance based upon the workload.

## (GCP) Explicitly set MIG `update_policy` to proactively replace instances, and make parameters configurable

This ensures that a change in instance configuration is rolled out immediately to ensure that all caches are running an identical configuration.
