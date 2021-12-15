# Next

* (GCP) Stop reporting file system usage metrics for NFS mounts
* (GCP) Add ability to explicitly disable NFS Versions in `nfs-kernel-server` and default to disabling NFS versions `4.0`, `4.1`, and `4.2`

## (GCP) Stop reporting file system usage metrics for NFS mounts

The default stackdriver collectd configuration for the `df` plugin includes metrics for NFS shares. The df plugin only collects basic metrics such as disk free space, inode free space, etc.

Collecting these metrics about NFS shares from the proxy is largely pointless as the same metrics are also available from the source server.

If the proxy has hundreds or thousands of NFS exports mounted this can greatly increase the volume of metrics being collected, leading to excessive charges for metrics ingestion.

## (GCP) Add ability to explicitly disable NFS Versions in `nfs-kernel-server` and default to disabling NFS versions `4.0`, `4.1`, and `4.2`

Adds the ability to explicitly disable NFS Versions in `nfs-kernel-server`. Explicitly disabling unwanted NFS versions prevents clients from accidentally auto-negotiating an undesired NFS version.

With this change, by default, NFS versions `4.0`, `4.1`, and `4.2` are now disabled on all proxies. To enable it, set a custom value for `DISABLED_NFS_VERSIONS`.
