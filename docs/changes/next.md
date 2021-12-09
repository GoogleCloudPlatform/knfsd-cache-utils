# Next

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
