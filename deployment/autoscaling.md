# Autoscaling

**Important: To use Autoscaling you MUST have metrics enabled as they are used as a scaling metric**

These deployment scripts also support configuring autoscaling for the Knfsd Managed Instance Group. Scaling based on the standard metric of CPU Usage is not optimal for the caching use case. Instead the custom metric `custom.googleapis.com/knfsd/nfs_connections` is used for triggering an autoscaling event.

Autoscaling can be enabled by setting the `ENABLE_KNFSD_AUTOSCALING` environment variable to `true` (defaults to `false`). There are also some other configuration options detailed in the [Configuration Variables](README.md#configuration-variables) section (such as how many NFS Connections a Knfsd node should be handling before a scale-up).

To avoid interruptions to existing NFS client mounts, and by extension render operations the autoscaler behaviour is set to **SCALE UP ONLY**. When a Knfsd Client exceeds the number of connections defined in the `KNFSD_AUTOSCALING_NFS_CONNECTIONS_THRESHOLD` variable a new instance will be added to the Knfsd Managed Instance group. If the number of Knfsd Connections subsequently falls significantly the Knfsd cluster **will not** automatically scale down. This is not a GCP limitation but an intentional design consideration to avoid:

- Loss of FS-Cache data that would need to be re-pulled on scale up
- Interruption to existing NFS Client Connections

You can change this behaviour if you wish, but it is not recommended.

There is a slight delay for metric ingestion (1-2 mins) and then for a new node to spin up and initialise (~2 mins). When a scaling event occurs new traffic will continue to be sent to the existing healthy nodes in the cluster until there is a new node ready to handle the connections. It is therefore recommended that you set your `KNFSD_AUTOSCALING_NFS_CONNECTIONS_THRESHOLD` slightly lower than the maximum number of connections a single Knfsd node can handle. This will start the scaling event early and make sure a new node is ready before your existing nodes become overloaded.
