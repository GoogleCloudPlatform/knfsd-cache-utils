# Known Issues

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
