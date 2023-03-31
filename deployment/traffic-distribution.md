# Client Traffic Distribution

This module supports two methods of distributing traffic from clients to KNFSD proxy instances. The third option of none allows defining your own method of distributing the traffic.

* DNS Round Robin (recommended)
* Internal TCP Load Balancer
* None

## DNS Round Robin

`TRAFFIC_DISTRIBUTION_MODE = "dns_round_robin"`

This uses a Cloud DNS zone and configures an A record using the [weighted round robin (WRR) routing policy](https://cloud.google.com/dns/docs/policies-overview#wrr-policy).

The clients need to be configured to use the DNS name of the KNFSD proxy cluster. The client will resolve the DNS name to the IP of a specific proxy instance. This address resolution is stable, so resolving the DNS address multiple times on the same client will result in the same IP each time.

### Limitations

DNS round robin imposes the following limitations on the KNFSD proxy configuration:

* The KNFSD proxy cluster *must* be resized using Terraform so that the DNS records are updated.

* Cannot resize the KNFSD proxy cluster while it has active clients. This is because the clients will remain connected to the same IP.

  Scaling down the cluster will result in some clients trying indefinitely to connect to the IPs of the instances that were removed.

  Scaling up the cluster *will not* redistribute existing clients to the new instances. When new clients are added, the will be distributed evenly across all instances in the cluster. This will result in the original instances being over-utilized, while the new instances are under-utilized.

* Auto scaling is not supported (`ENABLE_KNFSD_AUTOSCALING`). This is because the MIG must be resized using Terraform so that the DNS records are updated.

* Static (stateful) IPs are required (`ASSIGN_STATIC_IPS`). This ensures that if a proxy instance is replaced due to a failed health check the new instance has the same IP so that existing clients can reconnect.

* `MIG_REPLACEMENT_METHOD` must be `RECREATE`. `SUBSTITUTE` is not supported, as substitute will create a new instance before removing the old instance. This is not supported when using stateful IPs as the IP address of the new instance is already in use by the old instance. The old instance must be removed first, before the new instance is created.

### Failure Handling

If a proxy instance fails, the clients associated with that proxy instance will keep trying to connect to the same instance until the instance has recovered.

If the clients use the `hard` mount option, then the NFS client will try indefinitely to reconnect. However, the application might have its own timeout that aborts the read/write request.

If health checks are disabled the MIG will not automatically replace the failed instance. It is recommended when using DNS round robin that the auto healing health checks are enabled (`ENABLE_AUTOHEALING_HEALTHCHECKS`).

Because the MIG uses stateful IPs, the new instance will have the same IP as the original instance. The clients should reconnect to the new instance.

## Internal TCP Load Balancer

`TRAFFIC_DISTRIBUTION_MODE = "loadbalancer"`

This uses an [Internal TCP Cloud Load Balancer](https://cloud.google.com/load-balancing/docs/internal) to distribute client traffic between the proxy instances.

The TCP Load Balancer provides a single load balancer IP for the KNFSD proxy cluster. Clients connect to the load balancer IP, and the load balancer distributes the client connections between all the proxy instances.

Before using this option there are two things that need to be considered:

* Failure handling if an proxy instance is replaced.
* Cost of egress traffic

### Limitations

When scaling up a MIG with connected clients, due to the long lived persistent connections the traffic *will not* be redistributed to the new instances. This is identical to the failure handling, only instead of replacing a failed instance, new instances are being added.

### Failure Handling

The NFS client uses long lived persistent connections. This means that once a connection has been established the client will continue to use the same proxy instance until the connection is interrupted.

These long lived connections limit the load balancer's ability to re-distribute the client traffic. The load balancer is better suited for shorter lived connections.

If a proxy instance fails, the clients connected to that proxy instance will try to re-establish their connections. This causes the load balancer to distribute the clients of the failed instance between the remaining instances.

Once the failed instance has recovered, the client's will not be re-distributed. If the number of clients are increased, the new clients will be evenly distributed between all the nodes.

For example:

1. A cluster initially has 3 nodes, with 30 clients per node.
2. One of the nodes fails, the 30 clients on that node are distributed between the remaining two nodes. The two remaining nodes have 45 clients each.
3. When the node recovers the clients are not redistributed. The clients remain connected to only two of the nodes. Two of the nodes still have 45 clients each, and the node that was replaced has 0 clients.
4. An additional 90 clients are added. These clients are evenly distributed between all three nodes, resulting in two of the nodes having 75 clients each, and the third node only has 30 clients.

|                             | Node 1 | Node 2 | Node 3 |
|-----------------------------|-------:|-------:|-------:|
| Initial                     |     30 |     30 |     30 |
| Node 3 fails                |     45 |     45 |      - |
| Node 3 recovered            |     45 |     45 |      0 |
| Additional 90 clients added |     75 |     75 |     30 |

### Cost of egress traffic

There is a [charge for outbound traffic](https://cloud.google.com/vpc/network-pricing#lb) from the load balancer.

In general the load balancer is aimed at request/response style traffic similar to that of HTTP where the volume of data tends to be small in comparison to the number of requests.

However, with NFS traffic when performing jobs such as rendering a relatively small number of clients can easily fully saturate the bandwidth of several proxy instances for a sustained period of time. This means that the egress costs from the load balancer can be a significant portion of the cost to running a KNFSD proxy cluster.

## None

`TRAFFIC_DISTRIBUTION_MODE = "none"`.

Setting the traffic distribution mode to `none` disables the in-built traffic distribution options. This allows you to deploy your own solution to handle distributing client traffic between the proxy instances.

Some method of traffic distribution is required, so if using `none` you will need to provide your own solution.
