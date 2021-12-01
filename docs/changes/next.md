# Next

* (GCP) Stop reporting file system usage metrics for NFS mounts
* (GCP) Implement the knfsd-agent which provides a HTTP API for interacting with Knfsd nodes

## (GCP) Stop reporting file system usage metrics for NFS mounts

The default stackdriver collectd configuration for the `df` plugin includes metrics for NFS shares. The df plugin only collects basic metrics such as disk free space, inode free space, etc.

Collecting these metrics about NFS shares from the proxy is largely pointless as the same metrics are also available from the source server.

If the proxy has hundreds or thousands of NFS exports mounted this can greatly increase the volume of metrics being collected, leading to excessive charges for metrics ingestion.


## (GCP) Implement the knfsd-agent which provides a HTTP API for interacting with Knfsd nodes

Implements a new Golang based knfsd-agent which runs on all Knfsd nodes. It provides a HTTP API that can be used for interacting with Knfsd nodes. [See here](../../image/knfsd-agent/README.md) for more information.

If upgrading from `v.0.4.0` and below you will need to build a [new version of the Knfsd Image](../../image).
