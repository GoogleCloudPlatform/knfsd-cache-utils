# FAQ

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
