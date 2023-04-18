# FAQ

## Error creating HealthCheck: already exists

> Error creating HealthCheck: googleapi: Error 409: The resource 'projects/my-gcp-project/global/healthChecks/rendercluster1-autohealing-health-check' already exists, alreadyExists

Some of the resources created by Terraform (such as health checks) have to have a globally unique name. When deploying multiple knfsd proxy clusters (even if the clusters are in different regions) you *MUST* give each knfsd proxy cluster a unique `PROXY_BASENAME`.

See [firewall configuration](../deployment/firewall.md) for more details.

## Error creating Firewall: already exists

> Error creating Firewall: googleapi: Error 409: The resource 'projects/my-gcp-project/global/firewalls/allow-nfs-tcp-healthcheck' already exists, alreadyExists

The firewall rules can only be deployed once. When deploying multiple knfsd proxy clusters the best practice is to set `AUTO_CREATE_FIREWALL_RULES = false` for all the clusters and create the firewall rules yourself.

## Proxy instances are replaced every 10 minutes

The knfsd proxy instances in the Managed Instance Group (MIG) never become healthy. When the initial grace period expires the instances are replaced. The initial grace period defaults to 10 minutes (600 seconds) and can be changed by setting `HEALTHCHECK_INITIAL_DELAY_SECONDS`.

See [Check the knfsd proxy instance is starting correctly](./check-startup.md).

## Every file shows an I/O error in ls, or when trying to read/write

This could have one of two causes:

* The filehandle is too large to be re-exported.
* The directory is a nested mount (aka crossmnt).

If this happens for every file it is likely an issue with the filehandle. See "filehandle limits" in [known issues](known-issues.md).

If this only happens on some directories then it's likely an issue with nested mounts. See "Nested mounts (aka crossmnt)" in [known issues](known-issues.md).

## Some directories are empty

Check if the directory is a nested mount on the source server (some NFS servers might refer to this by another name such as junction).

See "Nested mounts (aka crossmnt)" in [known issues](known-issues.md).

## What if the source server's IP address changes

NFS resolves the IP address from the DNS or host name when the remote is mounted.

If the source server is restarted (e.g. due to failure) and a different IP address is allocated the proxies will need to be restarted.

If the clients connect via a load balancer they will be unaffected as the load balancer will reconnect the client to one of the new proxy instances.

If the clients connect to proxy instances directly (e.g. using DNS round robin) then the new proxy instances *MUST* have the same IP addresses.

In GCP this can be done using "Restart/Replace VMs" on the Managed Instance Group. This will not affect the clients as the clients are connected via a load balancer. The load balancer will retain the same static IP.

## What if the proxy's IP address changes

NFS resolves the IP address from the DNS or host name when the proxy is mounted.

If the proxy is restarted (e.g. due to failure) and a different IP address then the clients will need to be restarted.

Ideally the clients will be connected to the proxy instances using a load balancer. The load balancer provides a static IP that does not change.

However, if the clients are connected directly to the proxy instances (e.g. using DNS round robin) then the proxy instances *MUST* have static IP address.

**NOTE:** It is technically possible to recover the clients without restarting but the process is complex. You need to kill any processes that are waiting on NFS operations and then remount the shares. Trying to do this across hundreds of clients is likely infeasible.

## What if the client loses connection to the proxy

NFS resolves the IP address from the DNS or host name when the proxy is mounted.

If the proxy is restarted (e.g. due to failure) or the client cannot communicate with the proxy due to a network issue the client will wait until the proxy recovers when using the `hard` option (recommended).

## What if the proxy loses connection to the source server

Because the proxy is mounted using the `hard` recovery behaviour the proxy will
wait indefinitely for the source to recover. Once the source is available the
proxy will automatically resume function without any manual intervention.

This recovery may take up to 10 minutes (600 seconds), as per the NFS documentation on the `timeo` flag:

> For NFS over TCP the default timeo value is 600 (60 seconds). The NFS client performs linear backoff: After each retransmission the timeout is increased by timeo up to the maximum of 600 seconds.

## What if a proxy instance fails when using a load balancer

If one of the proxy is restarted or replaced (e.g. due to failure) the load balancer will redirect the client to another proxy server that is online.

This will not require an manual intervention. The client will see this as a momentary interruption in the network and will re-establish a connection. This new connection will be routed to one of the other proxy instances that are online.

## What if the client loses connection to the load balancer

Load balancer is using a static IP or the clients resolve the IP address from the DNS.

Because the client mounts the NFS volume using the `hard` option, if the client loses the connection in case of a network issue, the client will wait indefinitely for it to recover.

Once the issue is resolved the client will automatically resume function without any manual intervention.

## What if the entire proxy cluster fails

There a few reasons why the entire proxy might be unavailable, such as:

* Instances restart while the source server is unavailable (start up script will fail to complete)
* Configuration error, such as deleting, or scaling the managed instance group to zero.
* Missing firewall rules to permit NFS traffic, or rules denying NFS traffic

Because the client mounts the NFS volume using the `hard` option, if the client loses the connection in case of a network issue, the client will wait indefinitely for it to recover.

Once the issue is resolved the client will automatically resume function without any manual intervention.

## Behaviour of ls changes when using the proxy

When going directly to the source server `ls` will show changes immediately,
but when connecting through the proxy `ls` will continue to show cached changes. This isn't `ls` specific, as the behaviour is built into the kernel's `readdir` function. As any application that performs a directory listing should have the same behaviour.

This happens because `ls` initially bypasses the local cache on the client and always sends `GETATTR` to check the remote directory's `mtime`. If the `mtime` has changed then the client knows its cache is stale and performs a `READDIR`, otherwise the client can use its cache.

When using the proxy, the proxy cannot distinguish between a standard `GETATTR` that should be served from the metadata cache, and a `GETATTR` that is being used for cache invalidation.

`acdirmin` and `acdirmax` can be used to adjust directory metadata's expiry time on the proxy. Lowering this will force the proxy to re-validate the metadata with the source more often. This will result in the proxy detecting changes to the source more quickly but increase the number of metadata requests sent by the proxy to the source.

The proxy will still cache the `READDIR` results even after the directory metadata has expired. If the directory metadata (`mtime`) still matches the proxy will continue to use the cached `READDIR` results.

## Supplementary Groups

The NFS protocol only supports a maximum of 16 supplementary (auxiliary) groups when using UNIX (`sec=sys`) authentication.

If your system relies on users with more than 16 supplementary groups the NFS proxy will need to be connected to LDAP so that the proxy can resolve the full list of groups for a user.

Once connected to LDAP you need to set:

```ini
[mountd]
manage-gids=yes
```

You can set this by:

1. Changing `image/resources/etc/nfs.conf` and building a new image.
2. Running `nfsconf --set mountd manage-gids yes` in a `CUSTOM_PRE_STARTUP_SCRIPT`.

## Cache only writes, never reads, on 5.17 kernel

When mounting NFS shares using the `sync` option the proxy will continuously write to the cache but never read.

The `/proc/fs/fscache/stats` will show that there are 0 reads on the IO line:

```text
IO     : rd=0 wr=141739641
```

If kernel tracing for the `fscache_access` event is enabled and tracing observed, there will be constant `io_write` events, but no `io_read` events.

```bash
echo 1 >/sys/kernel/debug/tracing/events/fscache/fscache_access/enable
cat /sys/kernel/debug/tracing/trace_pipe
```

To view only `io_read`:

```bash
cat /sys/kernel/debug/tracing/trace_pipe | grep io_read
```

The trace will show:

```text
kworker/u32:135-4409    [005] ..... 13156.260267: fscache_access: c=000000df BEGIN io_writ r=2 a=9
          <...>-4350    [001] ..... 13156.260268: fscache_access: c=00000afd BEGIN io_writ r=2 a=5
 kworker/u32:72-4346    [004] ..... 13156.260268: fscache_access: c=0000093c BEGIN io_writ r=2 a=4
          <...>-4404    [010] ..... 13156.260272: fscache_access: c=000003b5 END   io      r=2 a=3
kworker/u32:207-4481    [015] ..... 13156.260272: fscache_access: c=000008e7 END   io      r=2 a=6
kworker/u32:162-4436    [002] ..... 13156.260274: fscache_access: c=000004f7 END   io      r=2 a=7
kworker/u32:207-4481    [015] ..... 13156.260274: fscache_access: c=000008e7 BEGIN io_writ r=2 a=7
kworker/u32:224-4498    [012] ..... 13156.260274: fscache_access: c=00000c55 END   io      r=2 a=1
kworker/u32:162-4436    [002] ..... 13156.260275: fscache_access: c=000004f7 BEGIN io_writ r=2 a=8
```
