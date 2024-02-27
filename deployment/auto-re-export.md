# Auto Re-export

With older kernel versions all re-exported filesystems had to be explicitly re-exported. This limitation existed because each re-export required an explicit FSID.

As of kernel 6.1 a new auto re-export feature was added; made possible by the addition of the new `fsidd` service that can automatically assign an FSID to an export when required.

Auto re-export is useful in the following conditions:

* The source server has nested mounts on separate filesystems (FSIDs).

  This will depend upon the implementation of the source NFS server. For example, NetApp Junctions are presented as separate filesystems and operate similar to Linux NFS' `nohide`/`crossmnt` option. Other NFS servers such as Isilon OneFS present nested volumes as if they were a single filesystem.

* The source server uses `crossmnt`/`nohide` (or similar) and you already use it with the clients. For example, when using NetApp Junctions.

* The source server has a large number of nested mounts (over 1,000) making it impractical or impossible to explicitly re-export every mount.

## Example

Assume the source server is exporting the following nested mounts:

```text
/dev/sda1 on /files
/dev/sda2 on /files/archive
```

With auto re-export enabled the proxy server only needs to explicitly export `/files`. When a client tries to access `/files/archive` the proxy will detect the change in file system, automatically re-export `/files/archive` using the FSID service to allocate a unique FSID to the export.

In NFS v4, this style of nested mount is directly supported by the NFS protocol. In NFS v3 this style of re-export relies on the `crossmnt` or `nohide` flags. There are some caveats that come with `crossmnt` or `nohide`, see `man exports.5` for more details.

## NFS protocol versions

### NFS version 3

NFS version 3 does not directly support nested volumes. Officially clients have to use the mount protocol to explicitly mount nested volumes.

Some NFS servers, including the Linux Kernel NFS server (knfsd), support a non-standard option that allows clients to see the nested volume without first mounting the volume. The Linux NFS client then detects the change in FSID and automatically mounts the nested volume.

On Linux this option is `nohide`/`crossmnt`. This can cause issues with NFS clients that do not support this, see `man exports.5` for details.

If the source server does not support an equivalent of `nohide` then auto re-export might not work. In this case the knfsd proxy will only see an empty directory instead of the nested volume.

For NFS v3 all root volumes still need to be explicitly re-exported. If you have a large number of root volumes consider mounting using NFS v4 instead.

### NFS version 4

The NFS v4 protocol was updated to directly support nested volumes. Under the NFS v4 protocol all exported volumes are considered to be nested volumes of a pseudo-root volume.

Because NFS v4 directly supports nested volumes the `nohide`/`crossmnt` issues of NFS v3 do not apply. However, the knfsd proxy still has to assign an FSID for each nested volume; either explicitly or by using auto re-export.

If you connect to the source server using NFS v4 you can re-export the pseudo-root path `/` with auto re-export enabled to automatically re-export every export.

**NOTE:** When using NFS v4, it is recommended that you use NFS v4.1 or greater. NFS v4.1 has many improvements to fix limitations of the NFS v4.0 protocol.

## Configuration

Use of an FSID service to automatically allocate FSIDs for exports is required when using auto re-export. An `FSID_MODE="external"` is recommended so that all the knfsd proxy instances in the cluster use the same FSID number for each export.

```terraform
module "nfs_proxy" {

    source = "github.com/GoogleCloudPlatform/knfsd-cache-utils//deployment/terraform-module-knfsd?ref=v1.0.0-beta8"

    # Include your standard KNFSD configuration, this example only shows the
    # configuration values specific to the auto re-export feature.

    # Explicitly re-export only the root volumes from the source server.
    # Not using auto-discovery such as EXPORT_HOST_AUTO_DETECT to avoid
    # explicitly re-exporting nested volumes.
    EXPORT_MAP                    = "10.0.5.5;/remoteexport;/remoteexport"

    # Enable the auto re-export feature
    AUTO_REEXPORT                 = true

    # Store FSID mappings in an external database so that all knfsd proxy
    # instances in the cluster allocate the same FSID to each export.
    FSID_MODE                     = "external"

    # If possible, use an NFS v4 protocol version for the proxy clients as
    # the NFS v4 protocol has direct support nested volumes.
    DISABLED_NFS_VERSIONS         = "3,4.0,4.2"
}
```
