# Next

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
