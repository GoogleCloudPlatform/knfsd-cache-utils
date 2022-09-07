# Known Issues

## Nested mounts (aka crossmnt)

When the source server has nested mounts, each nested mount must be explicitly re-exported by the proxy so that the mount is assigned a unique `fsid`.

If the nested mount is not explicitly re-exported you will see one of two issues on the client:

* An empty directory.
* An I/O error trying to access the nested mount.

If this occurs, consider using auto-discovery to automatically find and mount all the exports from the source server.

If you're already using `EXPORT_HOST_AUTO_DETECT`, check that `showmount -e SOURCE-SERVER` lists all the nested mounts. If the source server does not reply with all the nested mounts then you might have to list the exports explicitly using `EXPORT_MAP`.

## filehandle limits

When a filehandle is too large, the client will receive general I/O errors or permission errors when trying to list, read or write files via the proxy.

NFSv3 only supports up to 64 bytes for a filehandle, and the proxy server adds up to an additional 25 bytes (22 bytes, rounded up to the nearest multiple of 4).

The largest filehandle that can be re-exported by NFSv3 is 42 bytes, for a total of 64 bytes. Some NFS servers such as NetApp (especially when using qtrees) use filehandles greater than 42 bytes, these filehandles cannot be re-exported using NFSv3.

To fix the issue, re-export using NFSv4 (the proxy can still mount the source using NFSv3). NFSv3 should be disabled on the proxy to avoid clients attempting to mount using a protocol that will fail.

```terraform
# Only enable NFS 4.1 on re-export
DISABLED_NFS_VERSIONS = "3,4.0,4.2"
```

For further details see:
* [Reexporting NFS filesystems - Filehandle limits](https://www.kernel.org/doc/html/latest/filesystems/nfs/reexport.html#filehandle-limits)
* [NFS wiki - filehandle limits](https://linux-nfs.org/wiki/index.php/NFS_re-export#filehandle_limits)

## Knfsd proxy stops caching new data

Sometimes the cachefilesd will stop culling old data from the cache. When this happens the cache will fill up and be unable to cache any new data.

See [culling](./culling.md) for further details.

## knfsd-metrics-agent reports incorrect values for NFS transport metrics

Transport level metrics, such as `nfs.mount.ops_per_second` (aka `custom.google.com/knfsd/nfsiostat_ops_per_second`) are reported a value that is too low.

Generally this value has the correct shape, but is 16 times smaller than it should be.

This is because the information comes from the transport (`xprt`) lines from `/proc/self/mountstats`. Historically each mount only had a single transport, however, that is no longer true since the addition of the `nconnect` value.

The both `nfsiostat` and the current Go module used to parse the mount stats only reports a single transport (`xprt`) line. If the mount has multiple transport lines either the first or last line is chosen (depending on implementation).

Where possible the per-operation statistics should be summarised as these will give the correct value.

## NFS transport metrics add up to the wrong value

Transport level metrics come from the transport (`xprt`) lines from `/proc/self/mountstats`.

While these metrics are reported per mount, the same transport may be shared by multiple mounts. This occurs multiple mounts share the same source server, normally one TCP connection will be created per source server and shared by all the mounts. This can be changed by the `nconnect` value, for the knfsd proxy this defaults to 16 TCP connections per source server.

If you sum the transport level metrics such as `nfs.mount.ops_per_second` (aka `custom.google.com/knfsd/nfsiostat_ops_per_second`) the total value will be higher than expected due to counting the same TCP connection multiple times.

Where possible the per-operation statistics should be summarised as these will give the correct value.
