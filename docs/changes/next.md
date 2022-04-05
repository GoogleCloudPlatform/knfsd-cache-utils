# Next

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
