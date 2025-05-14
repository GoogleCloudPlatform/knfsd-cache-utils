# Next

* Update to Ubuntu 24.04 LTS (Noble Numbat) with kernel 6.11.0
* Disable unattended-upgrade.service
* Change default build machine type to e2-standard-4
* Update minimum Terraform version to 1.5
* Update knfsd metrics agent to support v6.6+ kernel versions

## Update to Ubuntu 24.04 LTS (Noble Numbat) with kernel 6.11.0

Update the image to use Ubuntu 24.04 LTS (Noble Numbat). The GCP image we're using comes with the 6.11.0 HWE (Hardware Enablement) Ubuntu 24.04 kernel installed. This fixes some issues with CacheFiles that were present in the original 6.8 kernel.

## Disable unattended-upgrade.service

There is a known issue in the 6.11.0 kernel that can cause a kernel panic when the NFS server is restarted due to a race condition. This issue is fixed in later kernels (tested with 6.14.6 mainline), but these kernels are not yet available for Ubuntu 24.04.

In normal operation the NFS server will not be restarted while the proxy is running. The `unattended-upgrade.service` can trigger a restart of the NFS server if it updates any of the NFS server packages, or libraries the NFS server relies on.

Disabling the `unattended-upgrade.service` to prevent restarting the NFS server. Security and OS updates can be managed by building new images using the latest GCP Ubuntu 24.04 image (update the `source_image` in `image/nfs-proxy.pkr.hcl`).

## Change default build machine type to e2-standard-4

The more powerful machine type is no longer required when building an image as we're not compiling a custom kernel.

## Update minimum Terraform version to 1.5

When changing the network variables to use self links the Terraform code was  changed to use the `strcontains` function to aid with backwards compatibility so that existing configurations could continue to use simple names. This function was not added until Terraform 1.5.

## Update knfsd metrics agent to support v6.6+ kernel versions

The 6.6 kernel introduced a new `wdeleg_getattr` metric to the `/proc/net/rpc/nfsd` file. This was not supported by the 0.10.1 Prometheus ProcFS parser. Updated the parser to 0.15.1 to support the new attribute.

# v1.0.0-beta8

* Fetch Ubuntu Kernel source from launchpad
* Display Kernel and OS version information at end of packer build
* Fix proxy startup script always prints "Error starting proxy"
* Fix auto-reexporting the root of an NFS v4 server
* Assign a public IP to the build machine by default
* Removed `SUBNETWORK_PROJECT` (using self links instead)

## Fetch Ubuntu Kernel source from launchpad

The kernel.ubuntu.com URLs have been removed in favour of using launchpad URLs directly.

## Display Kernel and OS version information at end of packer build

This helps verify that the correct Kernel version is in use and is
useful when looking at past build logs.

## Fix proxy startup script always prints "Error starting proxy"

The proxy startup script would always print the error message "Error starting proxy", even after the "Reached Proxy Startup Exit. Happy caching!" message. This was due to a typo when checking the startup_complete variable.

## Fix auto-reexporting the root of an NFS v4 server

When exporting the root "/" of an NFS v4 server with AUTO_REEXPORT enabled; attempting to navigate to nested volumes would result in an error such as "/mnt/files is not a directory".

This is because the logic that ensures the root export always has `fsid=0` was overriding the re-export logic. The updated logic correctly applies both conditions to the export.

## Assign a public IP to the build machine by default

When building using packer, assign a public IP (`omit_external_ip = false`) to the build machine by default. This matches the manual build process and makes building the image simpler when getting started.

When `omit_external_ip = true` the GCP network will require that Cloud NAT is configured so that the build instance can fetch packages and source code from the public internet.

## Removed `SUBNETWORK_PROJECT` (using self links instead)

Changed the format for shared VPC to use ID or self link instead. This is because some resources such as Cloud SQL always require self links, and other resources require self links when using shared VPC.

* Network ID format: `projects/{{project}}/global/networks/{{name}}`
* Subnetwork ID format: `projects/{{project}}/regions/{{region}}/subnetworks/{{name}}`

Simple names such as "default" are still supported. They will be converted automatically to a network/subnetwork ID. The project will be assumed to be the same project as the knfsd proxy cluster.

# v1.0.0-beta7

* Fix error applying Terraform when nodes greater than 1

## Fix error applying Terraform when nodes greater than 1

When the nodes are greater than 1 and using DNS round robin, Terraform cannot create a plan due to the IP addresses having an indeterminate count. Change this to use the input variable that specifies the number of knfsd nodes.

# v1.0.0-beta6

* Update kernel to Ubuntu mainline 6.4.0
* Fix scaling to zero when using round robin DNS
* Support deploying Cloud SQL instance with a private IP
* Make project/region/zone required
* Add new `RESERVE_KNFSD_CAPACITY` configuration option to reserve cluster capacity

## Update kernel to Ubuntu mainline 6.4.0

Update the kernel from 6.4-rc5 to the final 6.4.0 release. Still need to build our own kernel for now to include an additional patch to enable use of FS-Cache.

## Fix scaling to zero when using round robin DNS

The A record always requires at least one entry. Remove the A record from the DNS zone when scaling the proxy cluster to zero. Scaling can be useful if you only want to temporarily shutdown (or restart) the cluster without having to destroy all the other resources (such as Cloud SQL) that can take several minutes to re-create.

## Support deploying Cloud SQL instance with a private IP

Add a new configuration option `FSID_DATABASE_PRIVATE_IP` to support deploying the Cloud SQL instance with either a public or private IP.

## Make project/region/zone required

Relying on provider defaults for project/region/zone can make the deployments unreliable for some resources such as Cloud SQL. To ensure everything is deployed correctly require that these variables are explicitly set instead of relying on provider default values.

## Add new `RESERVE_KNFSD_CAPACITY` configuration option to reserve cluster capacity

Optionally create a Compute Engine Reservation for the cluster. The Knfsd nodes are often large instances with lots of Local SSDs. This means they can sometimes be difficult to schedule which can cause delays when replacing unhealthy instances, or performing rolling replacements.

A reservation ensures that the capacity for the Knfsd Cluster is always available in Google Cloud, regardless of the state of the instances. A reservation is not a commitment, and can be deleted at any time.

# v1.0.0-beta5

* FSID service to store FSID to export path mappings in an external database
* Automatically re-export nested volumes to support `crossmnt` and NFSv4
* Allow gVNIC without requiring the high bandwidth option
* Update kernel to 6.4-rc5
* Update nfs-utils to 2.6.3
* Fix configuring NFSD process
* Updated instructions on configuring manage-gids
* Bump prometheus/procfs version to resolve issues with high numbers of NFS mounts

## FSID service to store FSID to export path mappings in an external database

A new `FSID_MODE` option has been added to control how FSIDs are assigned to export paths. The default (and recommended) option is `external`, which uses a custom `knfsd-fsidd` service to store FSIDs in an external database.

See [Filesystem Identifiers](../../deployment/fsids.md) for more detail.

## Automatically re-export nested volumes to support `crossmnt` and NFSv4

Previously all nested volumes had to be explicitly re-exported. This could cause issues with slow start-up times on servers with a large number of nested volumes.

The main reason nested volumes had to be explicitly re-exported was to assign the nested volume an FSID. With the new FSID service nested volumes can be automatically assigned an FSID when the nested volume is uncovered.

See [Auto Re-export](../../deployment/auto-re-export.md) for more detail.

## Allow gVNIC without requiring the high bandwidth option

Add a new configuration option, `ENABLE_GVNIC`, to use the `gVNIC` network interface type even if  `ENABLE_HIGH_BANDWIDTH_CONFIGURATION` is not enabled.

This will allow the use of the more performant `gVNIC` instead of the default `virtio` driver on smaller instances, or where there is no need for the TIER_1 network performance.

## Update kernel to 6.4-rc5

The 6.4 kernel includes most of the FS-Cache performance patches. Also updated the image build process to use the Ubuntu build scripts.

## Update nfs-utils to 2.6.3

nfs-utils 2.6.3 is required to use the new re-export and fsidd features that were introduced in the 6.3 kernel.

## Fix configuring NFSD process

Ubuntu 22.04 (Jammy Jellyfish) deprecated `/etc/default/nfs-kernel-server`. NFS configuration is now managed using `/etc/nfs.conf` and `/etc/nfs.conf.d/*.conf`.

This means that `NUM_NFS_THREADS` and `DISABLED_NFS_VERSIONS` was being ignored since `v1.0.0-beta1`.

The proxy startup script will now create an `/etc/nfs.conf.d/knfsd.conf` file to configure the NFSD process.

## Updated instructions on configuring manage-gids

Ubuntu 22.04 (Jammy Jellyfish) deprecated `/etc/default/nfs-kernel-server`. This means that the old method of updating `RPCMOUNTDOPTS` to contain `--manage-gids` will no longer work.

Instead the `/etc/nfs.conf` needs to be updated as either part of the image build process, or using a `CUSTOM_PRE_STARTUP_SCRIPT`.

## Bump prometheus/procfs version to resolve issues with high numbers of NFS mounts

See [https://github.com/GoogleCloudPlatform/knfsd-cache-utils/pull/29](https://github.com/GoogleCloudPlatform/knfsd-cache-utils/pull/29)

# v1.0.0-beta4

* Change the default build machine type to c2-standard-16
* Update to the latest FS-Cache performance patches (v11)
* Assign static IPs to proxy instances
* DNS Round Robin based load balancing

## Change the default build machine type to c2-standard-16

This improves the build times for the custom kernel. These are generally available and tend to offer the best build speed to price. Some additional speed may be gained by using a c2-standard-30.

## Update to the latest FS-Cache performance patches (v11)

Builds a custom version of the kernel based on `6.2.0-rc5`. This updates the custom patches that resolve the FS-Cache single page caching performance issue to the v11 revision of the patch set.

This includes the following patch sets:

* Initial conversion of NFS basic I/O to use folios (v2)
  <https://lore.kernel.org/linux-nfs/0FEB407A-5D01-4430-AEE4-13A45B4840D8@hammerspace.com/>
* Convert NFS with fscache to the netfs API (v11)
  <https://lore.kernel.org/linux-nfs/20230220134308.1193219-1-dwysocha@redhat.com/>
* mm, netfs, fscache: Stop read optimisation when folio removed from pagecache (v6)
  <https://lore.kernel.org/linux-nfs/20230216150701.3654894-1-dhowells@redhat.com/>
* vfs, security: Fix automount superblock LSM init problem, preventing NFS sb sharing (v5)
  <https://lore.kernel.org/linux-kernel/217595.1662033775@warthog.procyon.org.uk/>

## Assign static IPs to proxy instances

Added a new configuration option, `ASSIGN_STATIC_IPS`. This configures the MIG to use [stateful IP addresses](https://cloud.google.com/compute/docs/instance-groups/configuring-stateful-ip-addresses-in-migs).

When using stateful IP addresses, if an instance needs to be replaced due to an update, or auto-healing the new instance will have the same IP as the original instance.

This allows using the cluster without a load balancer, where the clients connected directly to a specific proxy instance via the instances internal IP address.

## DNS Round Robin based load balancing

Add a new configuration option, `TRAFFIC_DISTRIBUTION_MODE`, to choose between using DNS round robin or the internal TCP load balancer.

DNS round robin uses Cloud DNS to distribute the traffic between the different proxy instances in a KNFSD proxy instance group.

DNS round robin is the recommended method, however making this the default could cause unintended changes to existing deployments. To avoid this the `TRAFFIC_DISTRIBUTION_MODE` is required and has no default.

# v1.0.0-beta3

* Temporary fix for cachefilesd intermittently terminating

## Temporary fix for cachefilesd intermittently terminating

The cachefilesd service keeps terminating causing the cache to be withdrawn. This results in the proxy no longer caching any data and just acting as a pass-through server.

This patch provides a temporary fix while the linux-cachefs maintainers decide the best way to solve the issue.

# v1.0.0-beta2

* Stop pinning to specific package versions
* Update packer to support rsa-ssh2-256 and rsa-ssh2-512 key algorithms

## Stop pinning to specific package versions

Pinning is causing issues with maintaining the repository and building images because the apt repository only keeps the latest version for many of these packages. This causes the image for release tags to stop building when the package they depend upon is removed from the apt repository.

Most of the packages are already unpinned and install the latest version. The only components where the version really matters is the kernel and nfs-tools, which are both installed separately from apt.

## Update packer to support rsa-ssh2-256 and rsa-ssh2-512 key algorithms

The older `ssh-rsa` (RSA with SHA1) key algorithm is no longer secure and is not supported by the latest versions of OpenSSH or Ubuntu Jammy. The google computer packer plugin needs to be a minimum of 1.0.13 to support the newer `rsa-ssh2-256` and `rsa-ssh2-512` key algorithms.

# v1.0.0-beta1

* Update Monitoring Dashboard to support new Persistent Disk FS-Cache Volumes
* Upgrade to Ubuntu 22.04 LTS
* Build and use custom kernel with FS-Cache performance patches
* Remove custom culling agent and custom cachefilesd package

## Remove custom culling agent and custom cachefilesd package

Version `5.17`+ of the kernel does not contain the FS-Cache culling bug, therefore the custom culling agent and custom cachefilesd package is no longer required.

## Build and use custom kernel with FS-Cache performance patches

Builds a custom version of the kernel based on `6.1.0-rc5`. This custom version contains additional patches that resolve the FS-Cache single page caching performance issue. See [here](https://github.com/benjamin-maynard/kernel/commits/nfs-fscache-netfs) for more details.

## Upgrade to Ubuntu 22.04 LTS

Upgrades the Ubuntu image to 22.04.1 LTS (Jammy Jellyfish).

## Update Monitoring Dashboard to support new Persistent Disk FS-Cache Volumes

Updated the monitoring dashboard to corretly show FS-Cache usage, and read write throughput when using Persistent Disk for the FS-Cache volume.

# v0.10.0

* MIG scaler workflow and command line tool
* Added nested mounts to known issues
* Added filehandle limits to known issues
* Added support for higher bandwidth configurations

## MIG scaler workflow and command line tool

The MIG scaler workflow will scale up a MIG in increments over a period of time to avoid starting too many instances at the same time.

The command line tool is used to submit jobs to the workflow and manage the jobs.

## Added support for higher bandwidth configurations

Added new `ENABLE_HIGH_BANDWIDTH_CONFIGURATION` variable. When set to `true` instances will be configured with [gVNIC](https://cloud.google.com/compute/docs/networking/using-gvnic) network adapters and [Tier 1](https://cloud.google.com/compute/docs/networking/configure-vm-with-high-bandwidth-configuration) bandwidth.

To fully take advantage of this higher bandwidth, you need to ensure you are using `N2`, `N2D`, `C2` or `C2D` instances. You should also make sure your VM has [enough CPU's allocated](https://cloud.google.com/compute/docs/networking/configure-vm-with-high-bandwidth-configuration#bandwidth-tiers) to benefit from the new configuration.

The `network_performance_config` configuration block in the `google_compute_instance_template` Terraform resource is still in beta, so this version also requires the `google-beta` provider for this resource.

# v0.9.0

* Custom metrics configuration via Terraform
* Total client NFS operations metric
* Support Persistent Disk instead of Local SSD for the Cachefilesd volume
* Support fanout architecture deployment

## Custom metrics configuration via Terraform

Supply a custom YAML configuration for the metrics agent when deploying the module using Terraform. Previously this was only available by updating the configuration when building the image.

## Total client NFS operations metric

Added a metric for total NFS operations received by the proxy from clients.

This can be used to see what kind of load the proxy is under, and compared with
the total proxy to source operations.

## Support Persistent Disk instead of Local SSD for the Cachefilesd volume

Added the ability to use Persistent Disk instead of Local SSD for the Cachefilesd `/var/cache/fscache` volume.

This may be slightly less performant than Local SSD, but allows for cache volumes up to 64TB.

## Support fanout architecture deployment

Adds minor tweaks to the deployment scripts to support NFSv4 `Knfsfd-->Source` mounts which is required for the fanout architecutre.

Adds documentation and example configuration for deploying Knfsd in a fanout architecture.

# v0.8.0

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
